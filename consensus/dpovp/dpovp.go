// implements for Dpovp consensus
package dpovp

import (
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/dpovp"
	"github.com/LemoFoundationLtd/lemochain-go/consensus"
	"github.com/LemoFoundationLtd/lemochain-go/core/state"
	"github.com/LemoFoundationLtd/lemochain-go/core/types"
	"github.com/LemoFoundationLtd/lemochain-go/crypto"
	"github.com/LemoFoundationLtd/lemochain-go/lemodb"
	"github.com/LemoFoundationLtd/lemochain-go/params"
	"github.com/LemoFoundationLtd/lemochain-go/rpc"
)

type Dpovp struct {
	config *params.DpovpConfig // Consensus engine configuration parameters
	db     lemodb.Database     // Database to store and retrieve snapshot checkpoints

	coinbase        common.Address      // Lemochain address of the signing key
	currentBlock    func() *types.Block // 获取当前block的回调
	isTurn          bool                // 是否可出块
	timeoutTime     int64       // 超时时间
	blockInternal   int64       // 出块间隔
	blockMinerTimer *time.Timer         // 出块timer
	isTurnMu        sync.Mutex          // isTurn mutex
}

// 修改定时器
func (d *Dpovp) ModifyTimer() {
	d.isTurnMu.Lock()
	defer d.isTurnMu.Unlock()

	time_dur := d.getTimespan() // 获取当前时间与最新块的时间差
	//log.Warn(`time_dur`, `time:`, string(time_dur))
	slot := d.getSlot()         // 获取新块离本节点索引的距离
	if slot == 1 {              // 说明下一个区块就该本节点产生了
		if time_dur >= d.blockInternal { // 如果上一个区块的时间与当前时间差大或等于3s（区块间的最小间隔为3s），则直接出块无需休眠
			d.isTurn = true
		} else {
			need_dur := d.blockInternal - time_dur // 如果上一个块时间与当前时间非常近（小于3s），则设置休眠
			d.resetMinerTimer(need_dur)
		}
	} else { // 说明还不该自己出块，但是需要修改超时时间了
		time_dur = int64(slot-1)*d.timeoutTime - time_dur
		d.resetMinerTimer(time_dur)
	}
}

// 重置出块定时器
func (d *Dpovp) resetMinerTimer(time_dur int64) {
	d.isTurnMu.Lock()
	defer d.isTurnMu.Unlock()

	// 停掉之前的定时器
	if d.blockMinerTimer != nil {
		d.blockMinerTimer.Stop()
	}
	// 重开新的定时器
	d.blockMinerTimer = time.AfterFunc(time.Duration(time_dur * int64(time.Millisecond)), func() {
		//d.isTurnMu.Lock()
		//defer d.isTurnMu.Unlock()
		d.isTurn = true
	})
	d.isTurn = false
}

// 获取最新块的出块者序号与本节点序号差
func (d *Dpovp) getSlot() int {
	lst_addr := d.currentBlock().Header().Coinbase
	lst_index := dpovp.GetCoreNodeIndex(lst_addr)
	me_index := dpovp.GetCoreNodeIndex(d.coinbase)
	tmp:=lst_addr.Hex()
	if tmp == `0x0000000000000000000000000000000000000000` {
		return me_index + 1
	}
	nodeCount := dpovp.GetCorNodesCount()
	if nodeCount == 1{
		return 1
	}
	return (me_index - lst_index + nodeCount) % nodeCount
}

// 获取最新区块的时间戳离当前时间的距离 单位：ms
func (d *Dpovp) getTimespan() int64 {
	lst_span := d.currentBlock().Header().Time.Int64()
	now:=time.Now().Unix()
	return (now - lst_span) * 1000
}

// 新增一个DPOVP共识机
func New(config *params.DpovpConfig, db lemodb.Database, coinbase common.Address, currentblock func() *types.Block) *Dpovp {
	//TODO
	conf := *config

	return &Dpovp{
		config:        &conf,
		db:            db,
		isTurn:        false,
		coinbase:      coinbase,
		timeoutTime:   config.Timeout,
		blockInternal: config.Sleeptime,
		currentBlock:  currentblock,
	}
}
// 设置coinbase
func (d *Dpovp) SetCoinbase(coinbase common.Address){
	d.coinbase = coinbase
}

// Author implements consensus.Engine, returning the Lemochain address recovered
// from the signature in the header's extra-data section.
// Author implements consensus.Engine, returning the header's coinbase as the
// proof-of-work verified author of the block.
// 暂未搞明白到底需要返回谁的地址
// 貌似是谁挖的矿返回谁的地址
func (d *Dpovp) Author(header *types.Header) (common.Address, error) {

	return header.Coinbase, nil
}

// VerifyHeader checks whether a header conforms to the consensus rules of a
// given engine. Verifying the seal may be done optionally here, or explicitly
// via the VerifySeal method.
func (d *Dpovp) VerifyHeader(chain consensus.ChainReader, header *types.Header, seal bool) error {

	return nil
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications (the order is that of
// the input slice).
func (d *Dpovp) VerifyHeaders(chain consensus.ChainReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {

	return nil, nil
}

// VerifyUncles verifies that the given block's uncles conform to the consensus
// rules of a given engine.
func (d *Dpovp) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	return nil
}

// VerifySeal checks whether the crypto seal on a header is valid according to
// the consensus rules of the given engine.
func (d *Dpovp) VerifySeal(chain consensus.ChainReader, header *types.Header) error {
	return nil
}

// Prepare initializes the consensus fields of a block header according to the
// rules of a particular engine. The changes are executed inline.
func (d *Dpovp) Prepare(chain consensus.ChainReader, header *types.Header) error {

	return nil
}

// Finalize runs any post-transaction state modifications (e.g. block rewards)
// and assembles the final block.
// Note: The block header and state database might be updated to reflect any
// consensus rules that happen at finalization (e.g. block rewards).
func (d *Dpovp) Finalize(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	// No block rewards in PoA, so the state remains as is and uncles are dropped
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	header.UncleHash = types.CalcUncleHash(nil)

	// Assemble and return the final block for sealing
	return types.NewBlock(header, txs, nil, receipts), nil
}

// Seal generates a new block for the given input block with the local miner's
// seal place on top.
func (d *Dpovp) Seal(chain consensus.ChainReader, block *types.Block, stop <-chan struct{}) (*types.Block, error) {
	if !d.isTurn {
		err := errors.New(`it's not turn to produce block`)
		return nil, err
	}
	// 出块
	header := block.Header()
	// Sealing the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		err := errors.New(`unknownblock`)
		return nil, err
	}
	// 对区块进行签名
	hash := header.Hash()
	privKey := dpovp.GetPrivKey()
	if signInfo, err := crypto.Sign(hash[:], &privKey); err != nil {
		return nil, err
	} else {
		header.SignInfo = make([]byte, len(signInfo))
		copy(header.SignInfo, signInfo)
	}

	// 出块之后需要重置定时器
	nodeCount := dpovp.GetCorNodesCount()
	var tim_dur int64
	if nodeCount > 1 {
		tim_dur = int64(nodeCount-1) * d.timeoutTime
	} else {
		tim_dur = d.blockInternal
	}
	d.resetMinerTimer(tim_dur)

	return block.WithSeal(header), nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have.
func (d *Dpovp) CalcDifficulty(chain consensus.ChainReader, time uint64, parent *types.Header) *big.Int {
	return new(big.Int)
}

// APIs returns the RPC APIs this consensus engine provides.
func (d *Dpovp) APIs(chain consensus.ChainReader) []rpc.API {
	return nil
}
