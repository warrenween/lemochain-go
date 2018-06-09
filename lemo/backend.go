// Copyright 2014 The lemochain-go Authors
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

// Package lemo implements the Lemochain protocol.
package lemo

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/LemoFoundationLtd/lemochain-go/accounts"
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/hexutil"
	"github.com/LemoFoundationLtd/lemochain-go/consensus"
	"github.com/LemoFoundationLtd/lemochain-go/consensus/clique"
	"github.com/LemoFoundationLtd/lemochain-go/consensus/dpovp"
	"github.com/LemoFoundationLtd/lemochain-go/consensus/lemohash"
	"github.com/LemoFoundationLtd/lemochain-go/core"
	"github.com/LemoFoundationLtd/lemochain-go/core/bloombits"
	"github.com/LemoFoundationLtd/lemochain-go/core/types"
	"github.com/LemoFoundationLtd/lemochain-go/core/vm"
	"github.com/LemoFoundationLtd/lemochain-go/event"
	"github.com/LemoFoundationLtd/lemochain-go/internal/lemoapi"
	"github.com/LemoFoundationLtd/lemochain-go/lemo/downloader"
	"github.com/LemoFoundationLtd/lemochain-go/lemo/filters"
	"github.com/LemoFoundationLtd/lemochain-go/lemo/gasprice"
	"github.com/LemoFoundationLtd/lemochain-go/lemodb"
	"github.com/LemoFoundationLtd/lemochain-go/log"
	"github.com/LemoFoundationLtd/lemochain-go/miner"
	"github.com/LemoFoundationLtd/lemochain-go/node"
	"github.com/LemoFoundationLtd/lemochain-go/p2p"
	"github.com/LemoFoundationLtd/lemochain-go/params"
	"github.com/LemoFoundationLtd/lemochain-go/rlp"
	"github.com/LemoFoundationLtd/lemochain-go/rpc"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// Lemochain implements the Lemochain full node service.
type Lemochain struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan  chan bool    // Channel for shutting down the lemochain
	stopDbUpgrade func() error // stop chain db sequential key upgrade

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb lemodb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *LemoApiBackend

	miner    *miner.Miner
	gasPrice *big.Int
	lemobase common.Address

	networkId     uint64
	netRPCService *lemoapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and lemobase)
}

func (s *Lemochain) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new Lemochain object (including the
// initialisation of the common Lemochain object)
func New(ctx *node.ServiceContext, config *Config) (*Lemochain, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run lemo.Lemochain in light sync mode, use les.LightLemochain")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	stopDbUpgrade := upgradeDeduplicateData(chainDb)
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	lemo := &Lemochain{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		shutdownChan:   make(chan bool),
		stopDbUpgrade:  stopDbUpgrade,
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		lemobase:       config.Lemobase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}
	// sman modify
	lemo.engine = CreateConsensusEngine(ctx, &config.Lemohash, chainConfig, chainDb, config.Lemobase, func() *types.Block { return lemo.blockchain.CurrentBlock() })

	log.Info("Initialising Lemochain protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := core.GetBlockChainVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run glemo upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		core.WriteBlockChainVersion(chainDb, core.BlockChainVersion)
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	)
	lemo.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, lemo.chainConfig, lemo.engine, vmConfig)
	if err != nil {
		return nil, err
	}

	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		lemo.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	lemo.bloomIndexer.Start(lemo.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	lemo.txPool = core.NewTxPool(config.TxPool, lemo.chainConfig, lemo.blockchain)

	if lemo.protocolManager, err = NewProtocolManager(lemo.chainConfig, config.SyncMode, config.NetworkId, lemo.eventMux, lemo.txPool, lemo.engine, lemo.blockchain, chainDb); err != nil {
		return nil, err
	}
	// sman modify for node mode
	if config.NodeMode == NodeModeStar {
		lemo.miner = miner.New(lemo, lemo.chainConfig, lemo.EventMux(), lemo.engine)
		lemo.miner.SetExtra(makeExtraData(config.ExtraData))
	}

	lemo.ApiBackend = &LemoApiBackend{lemo, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	lemo.ApiBackend.gpo = gasprice.NewOracle(lemo.ApiBackend, gpoParams)

	return lemo, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"glemo",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (lemodb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*lemodb.LDBDatabase); ok {
		db.Meter("lemo/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Lemochain service
func CreateConsensusEngine(ctx *node.ServiceContext, config *lemohash.Config, chainConfig *params.ChainConfig, db lemodb.Database, coinbase common.Address, currentblock func() *types.Block) consensus.Engine {
	// sman 此处路由到我们的新的共识方法：DPOVP
	if chainConfig.Dpovp != nil {
		return dpovp.New(chainConfig.Dpovp, db, coinbase, currentblock)
	} else {
		log.Error(`sman not dpovp`)
		return nil
	}

	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	// Otherwise assume proof-of-work
	switch {
	case config.PowMode == lemohash.ModeFake:
		log.Warn("Lemohash used in fake mode")
		return lemohash.NewFaker()
	case config.PowMode == lemohash.ModeTest:
		log.Warn("Lemohash used in test mode")
		return lemohash.NewTester()
	case config.PowMode == lemohash.ModeShared:
		log.Warn("Lemohash used in shared mode")
		return lemohash.NewShared()
	default:
		engine := lemohash.New(lemohash.Config{
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
}

// APIs returns the collection of RPC services the lemochain package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Lemochain) APIs() []rpc.API {
	apis := lemoapi.GetAPIs(s.ApiBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)
	apis = append(apis, []rpc.API{
		{
			Namespace: "lemo",
			Version:   "1.0",
			Service:   NewPublicLemochainAPI(s),
			Public:    true,
		}, {
			Namespace: "lemo",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "lemo",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
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
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
	// sman modify for node mode
	if s.config.NodeMode == NodeModeStar {
		apis = append(apis, []rpc.API{
			{
				Namespace: "lemo",
				Version:   "1.0",
				Service:   NewPublicMinerAPI(s),
				Public:    true,
			}, {
				Namespace: "miner",
				Version:   "1.0",
				Service:   NewPrivateMinerAPI(s),
				Public:    false,
			},
		}...)
	}
	// Append all the local APIs and return
	return apis
}

func (s *Lemochain) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Lemochain) Lemobase() (eb common.Address, err error) {
	s.lock.RLock()
	lemobase := s.lemobase
	s.lock.RUnlock()

	if lemobase != (common.Address{}) {
		return lemobase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			lemobase := accounts[0].Address

			s.lock.Lock()
			s.lemobase = lemobase
			s.lock.Unlock()

			log.Info("Lemobase automatically configured", "address", lemobase)
			return lemobase, nil
		}
	}
	return common.Address{}, fmt.Errorf("lemobase must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (self *Lemochain) SetLemobase(lemobase common.Address) {
	self.lock.Lock()
	self.lemobase = lemobase
	self.lock.Unlock()

	self.miner.SetLemobase(lemobase)
}

func (s *Lemochain) StartMining(local bool) error {
	eb, err := s.Lemobase()
	if err != nil {
		log.Error("Cannot start mining without lemobase", "err", err)
		return fmt.Errorf("lemobase missing: %v", err)
	}
	if clique, ok := s.engine.(*clique.Clique); ok {
		wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
		if wallet == nil || err != nil {
			log.Error("Lemobase account unavailable locally", "err", err)
			return fmt.Errorf("signer missing: %v", err)
		}
		clique.Authorize(eb, wallet.SignHash)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

func (s *Lemochain) StopMining()         { s.miner.Stop() }
func (s *Lemochain) IsMining() bool      { return s.miner.Mining() }
func (s *Lemochain) Miner() *miner.Miner { return s.miner }

func (s *Lemochain) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Lemochain) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Lemochain) TxPool() *core.TxPool               { return s.txPool }
func (s *Lemochain) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Lemochain) Engine() consensus.Engine           { return s.engine }
func (s *Lemochain) ChainDb() lemodb.Database           { return s.chainDb }
func (s *Lemochain) IsListening() bool                  { return true } // Always listening
func (s *Lemochain) LemoVersion() int                   { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Lemochain) NetVersion() uint64                 { return s.networkId }
func (s *Lemochain) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Lemochain) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Lemochain protocol implementation.
func (s *Lemochain) Start(srvr *p2p.Server) error {
	// sman set coinbase to blockchain
	coinbase, _ := s.Lemobase()
	s.blockchain.SetCoinbase(coinbase)

	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = lemoapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		if s.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", s.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= s.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Lemochain protocol.
func (s *Lemochain) Stop() error {
	if s.stopDbUpgrade != nil {
		s.stopDbUpgrade()
	}
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
