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

// EthApiBackend implements lemoapi.Backend for full nodes
type EthApiBackend struct {
	lemo *Lemochain
	gpo *gasprice.Oracle
}

func (b *EthApiBackend) ChainConfig() *params.ChainConfig {
	return b.lemo.chainConfig
}

func (b *EthApiBackend) CurrentBlock() *types.Block {
	return b.lemo.blockchain.CurrentBlock()
}

func (b *EthApiBackend) SetHead(number uint64) {
	b.lemo.protocolManager.downloader.Cancel()
	b.lemo.blockchain.SetHead(number)
}

func (b *EthApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
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

func (b *EthApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
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

func (b *EthApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
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

func (b *EthApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.lemo.blockchain.GetBlockByHash(blockHash), nil
}

func (b *EthApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return core.GetBlockReceipts(b.lemo.chainDb, blockHash, core.GetBlockNumber(b.lemo.chainDb, blockHash)), nil
}

func (b *EthApiBackend) GetLogs(ctx context.Context, blockHash common.Hash) ([][]*types.Log, error) {
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

func (b *EthApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.lemo.blockchain.GetTdByHash(blockHash)
}

func (b *EthApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.lemo.BlockChain(), nil)
	return vm.NewEVM(context, state, b.lemo.chainConfig, vmCfg), vmError, nil
}

func (b *EthApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.lemo.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *EthApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.lemo.BlockChain().SubscribeChainEvent(ch)
}

func (b *EthApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.lemo.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *EthApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.lemo.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *EthApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.lemo.BlockChain().SubscribeLogsEvent(ch)
}

func (b *EthApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.lemo.txPool.AddLocal(signedTx)
}

func (b *EthApiBackend) GetPoolTransactions() (types.Transactions, error) {
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

func (b *EthApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.lemo.txPool.Get(hash)
}

func (b *EthApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.lemo.txPool.State().GetNonce(addr), nil
}

func (b *EthApiBackend) Stats() (pending int, queued int) {
	return b.lemo.txPool.Stats()
}

func (b *EthApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.lemo.TxPool().Content()
}

func (b *EthApiBackend) SubscribeTxPreEvent(ch chan<- core.TxPreEvent) event.Subscription {
	return b.lemo.TxPool().SubscribeTxPreEvent(ch)
}

func (b *EthApiBackend) Downloader() *downloader.Downloader {
	return b.lemo.Downloader()
}

func (b *EthApiBackend) ProtocolVersion() int {
	return b.lemo.EthVersion()
}

func (b *EthApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *EthApiBackend) ChainDb() lemodb.Database {
	return b.lemo.ChainDb()
}

func (b *EthApiBackend) EventMux() *event.TypeMux {
	return b.lemo.EventMux()
}

func (b *EthApiBackend) AccountManager() *accounts.Manager {
	return b.lemo.AccountManager()
}

func (b *EthApiBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.lemo.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *EthApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.lemo.bloomRequests)
	}
}
