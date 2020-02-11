// Copyright 2015 The go-tau Authors
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

package tau

import (
	"context"
	"errors"
	"math/big"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/accounts"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common/math"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/bloombits"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/rawdb"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/state"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/types"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/vm"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/tau/downloader"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/tau/gasprice"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/taudb"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/event"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/params"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/rpc"
)

// TauAPIBackend implements tauapi.Backend for full nodes
type TauAPIBackend struct {
	extRPCEnabled bool
	tau           *Tau
	gpo           *gasprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *TauAPIBackend) ChainConfig() *params.ChainConfig {
	return b.tau.blockchain.Config()
}

func (b *TauAPIBackend) CurrentBlock() *types.Block {
	return b.tau.blockchain.CurrentBlock()
}

func (b *TauAPIBackend) SetHead(number uint64) {
	b.tau.protocolManager.downloader.Cancel()
	b.tau.blockchain.SetHead(number)
}

func (b *TauAPIBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		block := b.tau.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if number == rpc.LatestBlockNumber {
		return b.tau.blockchain.CurrentBlock().Header(), nil
	}
	return b.tau.blockchain.GetHeaderByNumber(uint64(number)), nil
}

func (b *TauAPIBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.tau.blockchain.GetHeaderByHash(hash), nil
}

func (b *TauAPIBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		block := b.tau.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if number == rpc.LatestBlockNumber {
		return b.tau.blockchain.CurrentBlock(), nil
	}
	return b.tau.blockchain.GetBlockByNumber(uint64(number)), nil
}

func (b *TauAPIBackend) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.tau.blockchain.GetBlockByHash(hash), nil
}

func (b *TauAPIBackend) StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if number == rpc.PendingBlockNumber {
		block, state := b.tau.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, number)
	if err != nil {
		return nil, nil, err
	}
	if header == nil {
		return nil, nil, errors.New("header not found")
	}
	stateDb, err := b.tau.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *TauAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	return b.tau.blockchain.GetReceiptsByHash(hash), nil
}

func (b *TauAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	receipts := b.tau.blockchain.GetReceiptsByHash(hash)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *TauAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.tau.blockchain.GetTdByHash(blockHash)
}

func (b *TauAPIBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.tau.BlockChain(), nil)
	return vm.NewEVM(context, state, b.tau.blockchain.Config(), *b.tau.blockchain.GetVMConfig()), vmError, nil
}

func (b *TauAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.tau.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *TauAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.tau.BlockChain().SubscribeChainEvent(ch)
}

func (b *TauAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.tau.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *TauAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.tau.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *TauAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.tau.BlockChain().SubscribeLogsEvent(ch)
}

func (b *TauAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.tau.txPool.AddLocal(signedTx)
}

func (b *TauAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.tau.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *TauAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.tau.txPool.Get(hash)
}

func (b *TauAPIBackend) GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error) {
	tx, blockHash, blockNumber, index := rawdb.ReadTransaction(b.tau.ChainDb(), txHash)
	return tx, blockHash, blockNumber, index, nil
}

func (b *TauAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.tau.txPool.Nonce(addr), nil
}

func (b *TauAPIBackend) Stats() (pending int, queued int) {
	return b.tau.txPool.Stats()
}

func (b *TauAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.tau.TxPool().Content()
}

func (b *TauAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.tau.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *TauAPIBackend) Downloader() *downloader.Downloader {
	return b.tau.Downloader()
}

func (b *TauAPIBackend) ProtocolVersion() int {
	return b.tau.TauVersion()
}

func (b *TauAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *TauAPIBackend) ChainDb() taudb.Database {
	return b.tau.ChainDb()
}

func (b *TauAPIBackend) EventMux() *event.TypeMux {
	return b.tau.EventMux()
}

func (b *TauAPIBackend) AccountManager() *accounts.Manager {
	return b.tau.AccountManager()
}

func (b *TauAPIBackend) ExtRPCEnabled() bool {
	return b.extRPCEnabled
}

func (b *TauAPIBackend) RPCGasCap() *big.Int {
	return b.tau.config.RPCGasCap
}

func (b *TauAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.tau.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *TauAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.tau.bloomRequests)
	}
}
