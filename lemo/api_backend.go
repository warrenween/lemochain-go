// Copyright 2015 The lemochain-go Authors
// This file is part of the lemochain-go library.
//
// The lemochain-go library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The lemochain-go library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the lemochain-go library. If not, see <http://www.gnu.org/licenses/>.

package lemo

import (
	"context"
	"math/big"

	"github.com/LemoFoundationLtd/lemochain-go/accounts"
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/math"
	"github.com/LemoFoundationLtd/lemochain-go/core"
	"github.com/LemoFoundationLtd/lemochain-go/core/bloombits"
	"github.com/LemoFoundationLtd/lemochain-go/core/state"
	"github.com/LemoFoundationLtd/lemochain-go/core/types"
	"github.com/LemoFoundationLtd/lemochain-go/core/vm"
	"github.com/LemoFoundationLtd/lemochain-go/lemo/downloader"
	"github.com/LemoFoundationLtd/lemochain-go/lemo/gasprice"
	"github.com/LemoFoundationLtd/lemochain-go/lemodb"
	"github.com/LemoFoundationLtd/lemochain-go/event"
	"github.com/LemoFoundationLtd/lemochain-go/params"
	"github.com/LemoFoundationLtd/lemochain-go/rpc"
)

// LemoApiBackend implements lemoapi.Backend for full nodes
type LemoApiBackend struct {
	lemo *Lemochain
	gpo *gasprice.Oracle
}

func (b *LemoApiBackend) ChainConfig() *params.ChainConfig {
	return b.lemo.chainConfig
}

func (b *LemoApiBackend) CurrentBlock() *types.Block {
	return b.lemo.blockchain.CurrentBlock()
}

func (b *LemoApiBackend) SetHead(number uint64) {
	b.lemo.protocolManager.downloader.Cancel()
	b.lemo.blockchain.SetHead(number)
}

func (b *LemoApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.lemo.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.lemo.blockchain.CurrentBlock().Header(), nil
	}
	return b.lemo.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *LemoApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.lemo.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.lemo.blockchain.CurrentBlock(), nil
	}
	return b.lemo.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *LemoApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.lemo.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.lemo.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *LemoApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.lemo.blockchain.GetBlockByHash(blockHash), nil
}

func (b *LemoApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return core.GetBlockReceipts(b.lemo.chainDb, blockHash, core.GetBlockNumber(b.lemo.chainDb, blockHash)), nil
}

func (b *LemoApiBackend) GetLogs(ctx context.Context, blockHash common.Hash) ([][]*types.Log, error) {
	receipts := core.GetBlockReceipts(b.lemo.chainDb, blockHash, core.GetBlockNumber(b.lemo.chainDb, blockHash))
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *LemoApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.lemo.blockchain.GetTdByHash(blockHash)
}

func (b *LemoApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.lemo.BlockChain(), nil)
	return vm.NewEVM(context, state, b.lemo.chainConfig, vmCfg), vmError, nil
}

func (b *LemoApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.lemo.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *LemoApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.lemo.BlockChain().SubscribeChainEvent(ch)
}

func (b *LemoApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.lemo.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *LemoApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.lemo.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *LemoApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.lemo.BlockChain().SubscribeLogsEvent(ch)
}

func (b *LemoApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.lemo.txPool.AddLocal(signedTx)
}

func (b *LemoApiBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.lemo.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *LemoApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.lemo.txPool.Get(hash)
}

func (b *LemoApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.lemo.txPool.State().GetNonce(addr), nil
}

func (b *LemoApiBackend) Stats() (pending int, queued int) {
	return b.lemo.txPool.Stats()
}

func (b *LemoApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.lemo.TxPool().Content()
}

func (b *LemoApiBackend) SubscribeTxPreEvent(ch chan<- core.TxPreEvent) event.Subscription {
	return b.lemo.TxPool().SubscribeTxPreEvent(ch)
}

func (b *LemoApiBackend) Downloader() *downloader.Downloader {
	return b.lemo.Downloader()
}

func (b *LemoApiBackend) ProtocolVersion() int {
	return b.lemo.LemoVersion()
}

func (b *LemoApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *LemoApiBackend) ChainDb() lemodb.Database {
	return b.lemo.ChainDb()
}

func (b *LemoApiBackend) EventMux() *event.TypeMux {
	return b.lemo.EventMux()
}

func (b *LemoApiBackend) AccountManager() *accounts.Manager {
	return b.lemo.AccountManager()
}

func (b *LemoApiBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.lemo.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *LemoApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.lemo.bloomRequests)
	}
}
