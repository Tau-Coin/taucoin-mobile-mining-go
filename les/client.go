// Copyright 2016 The go-tau Authors
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

// Package les implements the Light Tau Subprotocol.
package les

import (
	"fmt"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/accounts"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/accounts/abi/bind"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common/hexutil"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common/mclock"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/consensus"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/bloombits"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/rawdb"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/types"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/tau"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/tau/downloader"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/tau/filters"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/tau/gasprice"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/event"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/internal/tauapi"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/light"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/log"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/node"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/p2p"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/p2p/enode"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/params"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/rpc"
)

type LightTau struct {
	lesCommons

	reqDist    *requestDistributor
	retriever  *retrieveManager
	odr        *LesOdr
	relay      *lesTxRelay
	handler    *clientHandler
	txPool     *light.TxPool
	blockchain *light.LightChain
	serverPool *serverPool

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend     *LesApiBackend
	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager
	netRPCService  *tauapi.PublicNetAPI
}

func New(ctx *node.ServiceContext, config *tau.Config) (*LightTau, error) {
	chainDb, err := ctx.OpenDatabase("lightchaindata", config.DatabaseCache, config.DatabaseHandles, "tau/db/chaindata/")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlockWithOverride(chainDb, config.Genesis, config.OverrideIstanbul)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	ltau := &LightTau{
		lesCommons: lesCommons{
			genesis:     genesisHash,
			config:      config,
			chainConfig: chainConfig,
			iConfig:     light.DefaultClientIndexerConfig,
			chainDb:     chainDb,
			peers:       peers,
			closeCh:     make(chan struct{}),
		},
		eventMux:       ctx.EventMux,
		reqDist:        newRequestDistributor(peers, &mclock.System{}),
		accountManager: ctx.AccountManager,
		engine:         tau.CreateConsensusEngine(ctx, chainConfig, &config.Tauash, nil, false, chainDb),
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   tau.NewBloomIndexer(chainDb, params.BloomBitsBlocksClient, params.HelperTrieConfirmations),
		serverPool:     newServerPool(chainDb, config.UltraLightServers),
	}
	ltau.retriever = newRetrieveManager(peers, ltau.reqDist, ltau.serverPool)
	ltau.relay = newLesTxRelay(peers, ltau.retriever)

	ltau.odr = NewLesOdr(chainDb, light.DefaultClientIndexerConfig, ltau.retriever)
	ltau.chtIndexer = light.NewChtIndexer(chainDb, ltau.odr, params.CHTFrequency, params.HelperTrieConfirmations)
	ltau.bloomTrieIndexer = light.NewBloomTrieIndexer(chainDb, ltau.odr, params.BloomBitsBlocksClient, params.BloomTrieFrequency)
	ltau.odr.SetIndexers(ltau.chtIndexer, ltau.bloomTrieIndexer, ltau.bloomIndexer)

	checkpoint := config.Checkpoint
	if checkpoint == nil {
		checkpoint = params.TrustedCheckpoints[genesisHash]
	}
	// Note: NewLightChain adds the trusted checkpoint so it needs an ODR with
	// indexers already set but not started yet
	if ltau.blockchain, err = light.NewLightChain(ltau.odr, ltau.chainConfig, ltau.engine, checkpoint); err != nil {
		return nil, err
	}
	ltau.chainReader = ltau.blockchain
	ltau.txPool = light.NewTxPool(ltau.chainConfig, ltau.blockchain, ltau.relay)

	// Set up checkpoint oracle.
	oracle := config.CheckpointOracle
	if oracle == nil {
		oracle = params.CheckpointOracles[genesisHash]
	}
	ltau.oracle = newCheckpointOracle(oracle, ltau.localCheckpoint)

	// Note: AddChildIndexer starts the update process for the child
	ltau.bloomIndexer.AddChildIndexer(ltau.bloomTrieIndexer)
	ltau.chtIndexer.Start(ltau.blockchain)
	ltau.bloomIndexer.Start(ltau.blockchain)

	ltau.handler = newClientHandler(config.UltraLightServers, config.UltraLightFraction, checkpoint, ltau)
	if ltau.handler.ulc != nil {
		log.Warn("Ultra light client is enabled", "trustedNodes", len(ltau.handler.ulc.keys), "minTrustedFraction", ltau.handler.ulc.fraction)
		ltau.blockchain.DisableCheckFreq()
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		ltau.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	ltau.ApiBackend = &LesApiBackend{ctx.ExtRPCEnabled(), ltau, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.Miner.GasPrice
	}
	ltau.ApiBackend.gpo = gasprice.NewOracle(ltau.ApiBackend, gpoParams)

	return ltau, nil
}

type LightDummyAPI struct{}

// Tauerbase is the address that mining rewards will be send to
func (s *LightDummyAPI) Tauerbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("mining is not supported in light mode")
}

// Coinbase is the address that mining rewards will be send to (alias for Tauerbase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("mining is not supported in light mode")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the tau package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightTau) APIs() []rpc.API {
	return append(tauapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "tau",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "tau",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.handler.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "tau",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		}, {
			Namespace: "les",
			Version:   "1.0",
			Service:   NewPrivateLightAPI(&s.lesCommons),
			Public:    false,
		},
	}...)
}

func (s *LightTau) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightTau) BlockChain() *light.LightChain      { return s.blockchain }
func (s *LightTau) TxPool() *light.TxPool              { return s.txPool }
func (s *LightTau) Engine() consensus.Engine           { return s.engine }
func (s *LightTau) LesVersion() int                    { return int(ClientProtocolVersions[0]) }
func (s *LightTau) Downloader() *downloader.Downloader { return s.handler.downloader }
func (s *LightTau) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightTau) Protocols() []p2p.Protocol {
	return s.makeProtocols(ClientProtocolVersions, s.handler.runPeer, func(id enode.ID) interface{} {
		if p := s.peers.Peer(peerIdToString(id)); p != nil {
			return p.Info()
		}
		return nil
	})
}

// Start implements node.Service, starting all internal goroutines needed by the
// light tau protocol implementation.
func (s *LightTau) Start(srvr *p2p.Server) error {
	log.Warn("Light client mode is an experimental feature")

	// Start bloom request workers.
	s.wg.Add(bloomServiceThreads)
	s.startBloomHandlers(params.BloomBitsBlocksClient)

	s.netRPCService = tauapi.NewPublicNetAPI(srvr, s.config.NetworkId)

	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Tau protocol.
func (s *LightTau) Stop() error {
	close(s.closeCh)
	s.peers.Close()
	s.reqDist.close()
	s.odr.Stop()
	s.relay.Stop()
	s.bloomIndexer.Close()
	s.chtIndexer.Close()
	s.blockchain.Stop()
	s.handler.stop()
	s.txPool.Stop()
	s.engine.Close()
	s.eventMux.Stop()
	s.serverPool.stop()
	s.chainDb.Close()
	s.wg.Wait()
	log.Info("Light tau stopped")
	return nil
}

// SetClient sets the rpc client and binds the registrar contract.
func (s *LightTau) SetContractBackend(backend bind.ContractBackend) {
	if s.oracle == nil {
		return
	}
	s.oracle.start(backend)
}
