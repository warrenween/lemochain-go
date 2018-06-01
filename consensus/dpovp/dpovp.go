// implements for Dpovp consensus
package dpovp

import (
	"errors"
	"math/big"
	"sync"
	"time"

	"bytes"
	"fmt"

	"github.com/LemoFoundationLtd/lemochain-go/common"
	commonDpovp "github.com/LemoFoundationLtd/lemochain-go/common/dpovp"
	"github.com/LemoFoundationLtd/lemochain-go/consensus"
	"github.com/LemoFoundationLtd/lemochain-go/core/state"
	"github.com/LemoFoundationLtd/lemochain-go/core/types"
	"github.com/LemoFoundationLtd/lemochain-go/crypto"
	"github.com/LemoFoundationLtd/lemochain-go/lemodb"
	"github.com/LemoFoundationLtd/lemochain-go/params"
	"github.com/LemoFoundationLtd/lemochain-go/rpc"
)

// Ethash proof-of-work protocol constants.
var (
	FrontierBlockReward    *big.Int = big.NewInt(5e+18) // Block reward in wei for successfully mining a block
)

type Dpovp struct {
	config *params.DpovpConfig // Consensus engine configuration parameters
	db     lemodb.Database     // Database to store and retrieve snapshot checkpoints

	coinbase        common.Address      // Lemochain address of the signing key
	currentBlock    func() *types.Block // 获取当前block的回调
	isTurn          bool                // 是否可出块
	timeoutTime     int64               // 超时时间
	blockInternal   int64               // 出块间隔
	blockMinerTimer *time.Timer         // 出块timer
	isTurnMu        sync.Mutex          // isTurn mutex
}

// 修改定时器
func (d *Dpovp) ModifyTimer() {
	d.isTurnMu.Lock()
	defer d.isTurnMu.Unlock()

	timeDur := d.getTimespan() // 获取当前时间与最新块的时间差
	//log.Warn(`timeDur`, `time:`, string(timeDur))
	slot := d.getSlot() // 获取新块离本节点索引的距离
	if slot == 1 {      // 说明下一个区块就该本节点产生了
		if timeDur >= d.blockInternal { // 如果上一个区块的时间与当前时间差大或等于3s（区块间的最小间隔为3s），则直接出块无需休眠
			d.isTurn = true
		} else {
			needDur := d.blockInternal - timeDur // 如果上一个块时间与当前时间非常近（小于3s），则设置休眠
			go d.resetMinerTimer(needDur)
		}
	} else { // 说明还不该自己出块，但是需要修改超时时间了
		timeDur = timeDur % (int64(commonDpovp.GetCorNodesCount()) * d.timeoutTime)
		timeDur = int64(slot-1)*d.timeoutTime - timeDur
		go d.resetMinerTimer(timeDur)
	}
}

// 重置出块定时器
func (d *Dpovp) resetMinerTimer(timeDur int64) {
	d.isTurnMu.Lock()
	defer d.isTurnMu.Unlock()

	// 停掉之前的定时器
	if d.blockMinerTimer != nil {
		d.blockMinerTimer.Stop()
	}
	// 重开新的定时器
	d.blockMinerTimer = time.AfterFunc(time.Duration(timeDur*int64(time.Millisecond)), func() {
		//d.isTurnMu.Lock()
		//defer d.isTurnMu.Unlock()
		d.isTurn = true
	})
	d.isTurn = false
}

// 获取最新块的出块者序号与本节点序号差
func (d *Dpovp) getSlot() int {
	lstAddr := d.currentBlock().Header().Coinbase
	lstIndex := commonDpovp.GetCoreNodeIndex(&lstAddr)
	meIndex := commonDpovp.GetCoreNodeIndex(&(d.coinbase))
	var tmp [20]byte // 空地址

	if bytes.Compare(lstAddr[:], tmp[:]) == 0 {
		return meIndex + 1
	}
	nodeCount := commonDpovp.GetCorNodesCount()
	if nodeCount == 1 {
		return 1
	}
	return (meIndex - lstIndex + nodeCount) % nodeCount
}

// 获取最新区块的时间戳离当前时间的距离 单位：ms
func (d *Dpovp) getTimespan() int64 {
	lstSpan := d.currentBlock().Header().Time.Int64()
	now := time.Now().Unix()
	return (now - lstSpan) * 1000
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
func (d *Dpovp) SetCoinbase(coinbase common.Address) {
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
	return d.verifyHeader(chain, header, nil)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications (the order is that of
// the input slice).
func (d *Dpovp) VerifyHeaders(chain consensus.ChainReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))

	go func() {
		for i, header := range headers {
			err := d.verifyHeader(chain, header, headers[:i])

			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

// verifyHeader checks whether a header conforms to the consensus rules.The
// caller may optionally pass in a batch of parents (ascending order) to avoid
// looking those up from the database. This is useful for concurrently verifying
// a batch of new headers.
func (d *Dpovp) verifyHeader(chain consensus.ChainReader, header *types.Header, parents []*types.Header) error {
	if header.Number == nil {
		return consensus.ErrInvalidNumber
	}
	number := header.Number.Uint64()
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	// Don't waste time checking blocks from the future
	if header.Time.Cmp(big.NewInt(time.Now().Unix())) > 0 {
		return consensus.ErrFutureBlock
	}

	// 验证签名与coinbase对应的pubkey是否一致
	pubkey, err := crypto.Ecrecover(header.Hash().Bytes(), header.SignInfo)
	if err != nil {
		return fmt.Errorf(`Wrong signinfo`)
	}
	blkCbPubkey := commonDpovp.GetPubkeyByAddress(&(header.Coinbase))	// 获取出块者的node公钥
	if blkCbPubkey == nil{
		return fmt.Errorf("Verify header failed. Cann't get pubkey of %s", common.ToHex(header.Coinbase[:]))
	}
	if bytes.Compare(blkCbPubkey, pubkey) != 0 {
		return fmt.Errorf("Cann't verify block's signer")
	}

	// 是否该该节点出块
	timespan := int64(header.Time.Uint64()-parent.Time.Uint64()) * 1000 // 单位：ms
	nodeCount := commonDpovp.GetCorNodesCount()
	// 只有一个出块节点
	if nodeCount == 1 {
		if timespan < d.blockInternal { // 块间隔至少blockInternal
			return fmt.Errorf(`Not sleep enough time`)
		}
		return nil
	}
	// 所有节点全部超时时一轮的超时间隔
	oneTurnTimespan := int64(nodeCount) * d.timeoutTime
	// 去掉整轮后的间隔
	timespan = timespan % oneTurnTimespan
	// 当前块与父块的最近逻辑间距
	dist := commonDpovp.GetCoreNodeIndex(&(header.Coinbase)) - commonDpovp.GetCoreNodeIndex(&(parent.Coinbase))

	if dist == 0 {
		return fmt.Errorf(`one node can't produce block twice'`)
	}
	if dist == 1 {
		if timespan < d.timeoutTime {
			return fmt.Errorf(`not sleep enough time`)
		}
		return nil
	}
	if timespan < int64(dist)*d.timeoutTime || timespan >= int64(dist+1)*d.timeoutTime {
		return fmt.Errorf(`it's not turn`)
	}
	return nil
}

// VerifyUncles verifies that the given block's uncles conform to the consensus
// rules of a given engine.
func (d *Dpovp) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return errors.New("uncles not allowed")
	}
	return nil
}

// VerifySeal checks whether the crypto seal on a header is valid according to
// the consensus rules of the given engine.
func (d *Dpovp) VerifySeal(chain consensus.ChainReader, header *types.Header) error {
	// 验证签名与coinbase是否一致
	pubkey, err := crypto.Ecrecover(header.HashNoDpovp().Bytes(), header.SignInfo)
	if err != nil {
		return fmt.Errorf("Failed to verify Seal. hash:%s", header.Hash())
	}
	var signer common.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])
	if bytes.Compare(header.Coinbase[:], signer[:]) != 0 {
		return fmt.Errorf(`signer != coinbase`)
	}

	return nil
}

// Prepare initializes the consensus fields of a block header according to the
// rules of a particular engine. The changes are executed inline.
func (d *Dpovp) Prepare(chain consensus.ChainReader, header *types.Header) error {
	// Nonce is reserved for now, set to empty
	header.Nonce = types.BlockNonce{}
	// Set the difficulty to 1
	header.Difficulty = new(big.Int).SetInt64(1)
	// Mix digest is reserved for now, set to empty
	header.MixDigest = common.Hash{}
	header.Time = new(big.Int).SetUint64(uint64(time.Now().Unix()))
	return nil
}

// Finalize runs any post-transaction state modifications (e.g. block rewards)
// and assembles the final block.
// Note: The block header and state database might be updated to reflect any
// consensus rules that happen at finalization (e.g. block rewards).
func (d *Dpovp) Finalize(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	// No block rewards in PoA, so the state remains as is and uncles are dropped
	accumulateRewards(state, header)
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
	} else {
		d.isTurn = false
	}
	// 出块
	header := block.Header()
	// Sealing the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return nil, fmt.Errorf("unknownblock, number:%d", number)
	}
	// 对区块进行签名
	hash := header.HashNoDpovp()
	privKey := commonDpovp.GetPrivKey()
	if signInfo, err := crypto.Sign(hash[:], &privKey); err != nil {
		return nil, err
	} else {
		header.SignInfo = make([]byte, len(signInfo))
		copy(header.SignInfo, signInfo)
	}
	// 出块之后需要重置定时器
	nodeCount := commonDpovp.GetCorNodesCount()
	var timeDur int64
	if nodeCount > 1 {
		timeDur = int64(nodeCount-1) * d.timeoutTime
	} else {
		timeDur = d.blockInternal
	}
	d.resetMinerTimer(timeDur)

	return block.WithSeal(header), nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have.
func (d *Dpovp) CalcDifficulty(chain consensus.ChainReader, time uint64, parent *types.Header) *big.Int {
	return new(big.Int).SetInt64(1)
}

// APIs returns the RPC APIs this consensus engine provides.
func (d *Dpovp) APIs(chain consensus.ChainReader) []rpc.API {
	return nil
}

// AccumulateRewards credits the coinbase of the given block with the mining
// reward
func accumulateRewards(state *state.StateDB, header *types.Header) {
	blockReward := FrontierBlockReward
	reward := new(big.Int).Set(blockReward)
	state.AddBalance(header.Coinbase, reward)
}