// Copyright 2016 The lemochain-go Authors
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

// Package les implements the Light Lemochain Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/LemoFoundationLtd/lemochain-go/accounts"
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/hexutil"
	"github.com/LemoFoundationLtd/lemochain-go/consensus"
	"github.com/LemoFoundationLtd/lemochain-go/core"
	"github.com/LemoFoundationLtd/lemochain-go/core/bloombits"
	"github.com/LemoFoundationLtd/lemochain-go/core/types"
	"github.com/LemoFoundationLtd/lemochain-go/lemo"
	"github.com/LemoFoundationLtd/lemochain-go/lemo/downloader"
	"github.com/LemoFoundationLtd/lemochain-go/lemo/filters"
	"github.com/LemoFoundationLtd/lemochain-go/lemo/gasprice"
	"github.com/LemoFoundationLtd/lemochain-go/lemodb"
	"github.com/LemoFoundationLtd/lemochain-go/event"
	"github.com/LemoFoundationLtd/lemochain-go/internal/lemoapi"
	"github.com/LemoFoundationLtd/lemochain-go/light"
	"github.com/LemoFoundationLtd/lemochain-go/log"
	"github.com/LemoFoundationLtd/lemochain-go/node"
	"github.com/LemoFoundationLtd/lemochain-go/p2p"
	"github.com/LemoFoundationLtd/lemochain-go/p2p/discv5"
	"github.com/LemoFoundationLtd/lemochain-go/params"
	rpc "github.com/LemoFoundationLtd/lemochain-go/rpc"
)

type LightLemochain struct {
	config *lemo.Config

	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool
	// Handlers
	peers           *peerSet
	txPool          *light.TxPool
	blockchain      *light.LightChain
	protocolManager *ProtocolManager
	serverPool      *serverPool
	reqDist         *requestDistributor
	retriever       *retrieveManager
	// DB interfaces
	chainDb lemodb.Database // Block chain database

	bloomRequests                              chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer, chtIndexer, bloomTrieIndexer *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *lemoapi.PublicNetAPI

	wg sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *lemo.Config) (*LightLemochain, error) {
	chainDb, err := lemo.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	llemo := &LightLemochain{
		config:           config,
		chainConfig:      chainConfig,
		chainDb:          chainDb,
		eventMux:         ctx.EventMux,
		peers:            peers,
		reqDist:          newRequestDistributor(peers, quitSync),
		accountManager:   ctx.AccountManager,
		engine:           lemo.CreateConsensusEngine(ctx, &config.Lemohash, chainConfig, chainDb),
		shutdownChan:     make(chan bool),
		networkId:        config.NetworkId,
		bloomRequests:    make(chan chan *bloombits.Retrieval),
		bloomIndexer:     lemo.NewBloomIndexer(chainDb, light.BloomTrieFrequency),
		chtIndexer:       light.NewChtIndexer(chainDb, true),
		bloomTrieIndexer: light.NewBloomTrieIndexer(chainDb, true),
	}

	llemo.relay = NewLesTxRelay(peers, llemo.reqDist)
	llemo.serverPool = newServerPool(chainDb, quitSync, &llemo.wg)
	llemo.retriever = newRetrieveManager(peers, llemo.reqDist, llemo.serverPool)
	llemo.odr = NewLesOdr(chainDb, llemo.chtIndexer, llemo.bloomTrieIndexer, llemo.bloomIndexer, llemo.retriever)
	if llemo.blockchain, err = light.NewLightChain(llemo.odr, llemo.chainConfig, llemo.engine); err != nil {
		return nil, err
	}
	llemo.bloomIndexer.Start(llemo.blockchain)
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		llemo.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	llemo.txPool = light.NewTxPool(llemo.chainConfig, llemo.blockchain, llemo.relay)
	if llemo.protocolManager, err = NewProtocolManager(llemo.chainConfig, true, ClientProtocolVersions, config.NetworkId, llemo.eventMux, llemo.engine, llemo.peers, llemo.blockchain, nil, chainDb, llemo.odr, llemo.relay, quitSync, &llemo.wg); err != nil {
		return nil, err
	}
	llemo.ApiBackend = &LesApiBackend{llemo, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	llemo.ApiBackend.gpo = gasprice.NewOracle(llemo.ApiBackend, gpoParams)
	return llemo, nil
}

func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
	var name string
	switch protocolVersion {
	case lpv1:
		name = "LES"
	case lpv2:
		name = "LES2"
	default:
		panic(nil)
	}
	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// Lemobase is the address that mining rewards will be send to
func (s *LightDummyAPI) Lemobase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for Lemobase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the lemochain package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightLemochain) APIs() []rpc.API {
	return append(lemoapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "lemo",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "lemo",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "lemo",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *LightLemochain) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightLemochain) BlockChain() *light.LightChain      { return s.blockchain }
func (s *LightLemochain) TxPool() *light.TxPool              { return s.txPool }
func (s *LightLemochain) Engine() consensus.Engine           { return s.engine }
func (s *LightLemochain) LesVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *LightLemochain) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *LightLemochain) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightLemochain) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// Lemochain protocol implementation.
func (s *LightLemochain) Start(srvr *p2p.Server) error {
	s.startBloomHandlers()
	log.Warn("Light client mode is an experimental feature")
	s.netRPCService = lemoapi.NewPublicNetAPI(srvr, s.networkId)
	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	s.protocolManager.Start(s.config.LightPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Lemochain protocol.
func (s *LightLemochain) Stop() error {
	s.odr.Stop()
	if s.bloomIndexer != nil {
		s.bloomIndexer.Close()
	}
	if s.chtIndexer != nil {
		s.chtIndexer.Close()
	}
	if s.bloomTrieIndexer != nil {
		s.bloomTrieIndexer.Close()
	}
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()

	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
