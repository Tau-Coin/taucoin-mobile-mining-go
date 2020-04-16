// Copyright 2014 The go-tau Authors
// This file is part of the go-tau library.
//
// The go-tau library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-tau library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-tau library. If not, see <http://www.gnu.org/licenses/>.

// Package tau implements the Tau protocol.
package tau

import (
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/accounts"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/consensus"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/consensus/tauhash"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/rawdb"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/types"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/event"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/internal/tauapi"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/log"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/miner"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/node"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/p2p"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/p2p/enr"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/params"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/rpc"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/tau/downloader"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/taudb"
)

// Tau implements the Tau full node service.
type Tau struct {
	config *Config

	// Channel for shutting down the service
	shutdownChan chan bool

	server *p2p.Server

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager

	// DB interfaces
	chainDb taudb.Database  // Block chain database
	ipfsDb  taudb.IpfsStore // Block chain IPFS database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	APIBackend *TauAPIBackend

	miner     *miner.Miner
	feeFloor  *big.Int
	tauerbase common.Address

	networkID     uint64
	netRPCService *tauapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and tauerbase)
}

// New creates a new Tau object (including the
// initialisation of the common Tau object)
func New(ctx *node.ServiceContext, config *Config) (*Tau, error) {
	// Ensure configuration values are compatible and sane
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}

	if config.NoPruning && config.TrieDirtyCache > 0 {
		config.TrieCleanCache += config.TrieDirtyCache
		config.TrieDirtyCache = 0
	}
	log.Info("Allocated trie memory caches", "clean", common.StorageSize(config.TrieCleanCache)*1024*1024, "dirty", common.StorageSize(config.TrieDirtyCache)*1024*1024)

	// Assemble the Tau object
	chainDb, err := ctx.OpenDatabaseWithFreezer("chaindata", config.DatabaseCache, config.DatabaseHandles, config.DatabaseFreezer, "tau/db/chaindata/")
	if err != nil {
		return nil, err
	}

	log.Info("Open Ipfs database")
	ipfsDb, err2 := ctx.OpenIpfsDatabase()
	if err2 != nil {
		return nil, err2
	}

	log.Info("Start set up genesis block", "genesis config", config.Genesis)
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlockWithOverride(chainDb, ipfsDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	tau := &Tau{
		config:         config,
		chainDb:        chainDb,
		ipfsDb:         ipfsDb,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, chainConfig, &config.Tauash),
		shutdownChan:   make(chan bool),
		networkID:      config.NetworkId,
		feeFloor:       config.Miner.FeeFloor,
		tauerbase:      config.Miner.Tauerbase,
	}

	bcVersion := rawdb.ReadDatabaseVersion(chainDb)
	var dbVer = "<nil>"
	if bcVersion != nil {
		dbVer = fmt.Sprintf("%d", *bcVersion)
	}
	log.Info("Initialising Tau protocol", "versions", ProtocolVersions, "network", config.NetworkId, "dbversion", dbVer)

	var (
		cacheConfig = &core.CacheConfig{
			TrieCleanLimit:      config.TrieCleanCache,
			TrieCleanNoPrefetch: config.NoPrefetch,
			TrieDirtyLimit:      config.TrieDirtyCache,
			TrieDirtyDisabled:   config.NoPruning,
			TrieTimeLimit:       config.TrieTimeout,
		}
	)

	tau.blockchain, err = core.NewBlockChain(chainDb, ipfsDb, cacheConfig, chainConfig, tau.engine, tau.shouldPreserve)
	if err != nil {
		return nil, err
	}

	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		tau.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	log.Info("New Tx Pool")
	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	tau.txPool = core.NewTxPool(config.TxPool, chainConfig, tau.blockchain)

	// Permit the downloader to use the trie cache allowance during fast sync
	cacheLimit := cacheConfig.TrieCleanLimit + cacheConfig.TrieDirtyLimit
	checkpoint := config.Checkpoint
	if checkpoint == nil {
		checkpoint = params.TrustedCheckpoints[genesisHash]
	}
	if tau.protocolManager, err = NewProtocolManager(chainConfig, checkpoint, config.SyncMode, config.NetworkId, tau.eventMux, tau.txPool, tau.engine, tau.blockchain, ipfsDb, cacheLimit, config.Whitelist); err != nil {
		return nil, err
	}
	tau.miner = miner.New(tau, &config.Miner, chainConfig, tau.EventMux(), tau.engine, tau.isLocalBlock)

	tau.APIBackend = &TauAPIBackend{ctx.ExtRPCEnabled(), tau}

	return tau, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Tau service
func CreateConsensusEngine(ctx *node.ServiceContext, chainConfig *params.ChainConfig, config *tauhash.Config) consensus.Engine {
	engine := tauhash.New(tauhash.Config{
		CacheDir:       ctx.ResolvePath(config.CacheDir),
		CachesInMem:    config.CachesInMem,
		CachesOnDisk:   config.CachesOnDisk,
		DatasetDir:     config.DatasetDir,
		DatasetsInMem:  config.DatasetsInMem,
		DatasetsOnDisk: config.DatasetsOnDisk,
	})
	engine.SetThreads(-1) // Disable CPU mining
	return engine
}

// APIs return the collection of RPC services the tau package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Tau) APIs() []rpc.API {
	apis := tauapi.GetAPIs(s.APIBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "tau",
			Version:   "1.0",
			Service:   NewPublicTauAPI(s),
			Public:    true,
		}, {
			Namespace: "tau",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "tau",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Tau) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Tau) Tauerbase() (eb common.Address, err error) {
	s.lock.RLock()
	tauerbase := s.tauerbase
	s.lock.RUnlock()

	if tauerbase != (common.Address{}) {
		return tauerbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			tauerbase := accounts[0].Address

			s.lock.Lock()
			s.tauerbase = tauerbase
			s.lock.Unlock()

			log.Info("Tauerbase automatically configured", "address", tauerbase)
			return tauerbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("tauerbase must be explicitly specified")
}

// isLocalBlock checks whtauer the specified block is mined
// by local miner accounts.
//
// We regard two types of accounts as local miner account: tauerbase
// and accounts specified via `txpool.locals` flag.
func (s *Tau) isLocalBlock(block *types.Block) bool {
	author, err := s.engine.Author(block.Header())
	if err != nil {
		log.Warn("Failed to retrieve block author", "number", block.NumberU64(), "hash", block.Hash(), "err", err)
		return false
	}
	// Check whtauer the given address is tauerbase.
	s.lock.RLock()
	tauerbase := s.tauerbase
	s.lock.RUnlock()
	if author == tauerbase {
		return true
	}
	// Check whtauer the given address is specified by `txpool.local`
	// CLI flag.
	for _, account := range s.config.TxPool.Locals {
		if account == author {
			return true
		}
	}
	return false
}

// shouldPreserve checks whtauer we should preserve the given block
// during the chain reorg depending on whtauer the author of block
// is a local account.
func (s *Tau) shouldPreserve(block *types.Block) bool {
	return s.isLocalBlock(block)
}

// SetTauerbase sets the mining reward address.
func (s *Tau) SetTauerbase(tauerbase common.Address) {
	s.lock.Lock()
	s.tauerbase = tauerbase
	s.lock.Unlock()

	s.miner.SetTauerbase(tauerbase)
}

// StartMining starts the miner with the given number of CPU threads. If mining
// is already running, this method adjust the number of threads allowed to use
// and updates the minimum price required by the transaction pool.
func (s *Tau) StartMining(threads int) error {
	// Update the thread count within the consensus engine
	type threaded interface {
		SetThreads(threads int)
	}
	if th, ok := s.engine.(threaded); ok {
		log.Info("Updated mining threads", "threads", threads)
		if threads == 0 {
			threads = -1 // Disable the miner from within
		}
		th.SetThreads(threads)
	}
	// If the miner was not running, initialize it
	if !s.IsMining() {
		// Propagate the initial price point to the transaction pool
		s.lock.RLock()
		price := s.feeFloor
		s.lock.RUnlock()
		s.txPool.SetFeeFloor(price)

		// Configure the local mining address
		eb, err := s.Tauerbase()
		if err != nil {
			log.Error("Cannot start mining without tauerbase", "err", err)
			return fmt.Errorf("tauerbase missing: %v", err)
		}
		// If mining is started, we can disable the transaction rejection mechanism
		// introduced to speed sync times.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)

		go s.miner.Start(eb)
	}
	return nil
}

// StopMining terminates the miner, both at the consensus engine level as well as
// at the block creation level.
func (s *Tau) StopMining() {
	// Update the thread count within the consensus engine
	type threaded interface {
		SetThreads(threads int)
	}
	if th, ok := s.engine.(threaded); ok {
		th.SetThreads(-1)
	}
	// Stop the block creating itself
	s.miner.Stop()
}

func (s *Tau) IsMining() bool      { return s.miner.Mining() }
func (s *Tau) Miner() *miner.Miner { return s.miner }

func (s *Tau) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Tau) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Tau) TxPool() *core.TxPool               { return s.txPool }
func (s *Tau) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Tau) Engine() consensus.Engine           { return s.engine }
func (s *Tau) ChainDb() taudb.Database            { return s.chainDb }
func (s *Tau) Ipfs() taudb.IpfsStore              { return s.ipfsDb }
func (s *Tau) IsListening() bool                  { return true } // Always listening
func (s *Tau) TauVersion() int                    { return int(ProtocolVersions[0]) }
func (s *Tau) NetVersion() uint64                 { return s.networkID }
func (s *Tau) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *Tau) Synced() bool                       { return atomic.LoadUint32(&s.protocolManager.acceptTxs) == 1 }
func (s *Tau) ArchiveMode() bool                  { return s.config.NoPruning }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Tau) Protocols() []p2p.Protocol {
	protos := make([]p2p.Protocol, len(ProtocolVersions))
	for i, vsn := range ProtocolVersions {
		protos[i] = s.protocolManager.makeProtocol(vsn)
		protos[i].Attributes = []enr.Entry{s.currentTauEntry()}
	}
	return protos
}

// Start implements node.Service, starting all internal goroutines needed by the
// Tau protocol implementation.
func (s *Tau) Start(srvr *p2p.Server) error {
	s.startTauEntryUpdate(srvr.LocalNode())

	// Start the RPC service
	s.netRPCService = tauapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers

	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Tau protocol.
func (s *Tau) Stop() error {
	s.blockchain.Stop()
	s.engine.Close()
	s.protocolManager.Stop()
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)
	return nil
}
