// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package eth implements the Ethereum protocol.
package atlas

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/dnsdisc"
	"github.com/ethereum/go-ethereum/p2p/enode"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/mapprotocol/atlas/chains/ethereum"

	"github.com/mapprotocol/atlas/accounts"
	"github.com/mapprotocol/atlas/apis/atlasapi"
	"github.com/mapprotocol/atlas/atlas/downloader"
	"github.com/mapprotocol/atlas/atlas/ethconfig"
	"github.com/mapprotocol/atlas/atlas/filters"
	"github.com/mapprotocol/atlas/atlas/gasprice"
	"github.com/mapprotocol/atlas/atlas/protocols/eth"
	"github.com/mapprotocol/atlas/atlas/protocols/snap"
	"github.com/mapprotocol/atlas/cmd/node"
	"github.com/mapprotocol/atlas/consensus"
	"github.com/mapprotocol/atlas/consensus/consensustest"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	istanbulBackend "github.com/mapprotocol/atlas/consensus/istanbul/backend"
	"github.com/mapprotocol/atlas/core/bloombits"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/core/indexer"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/state/pruner"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/miner"
	"github.com/mapprotocol/atlas/p2p"
	"github.com/mapprotocol/atlas/params"
)

// Config contains the configuration options of the ETH protocol.
// Deprecated: use ethconfig.Config instead.
//type Config = ethconfig.Config

// Ethereum implements the Ethereum full node service.
type Ethereum struct {
	config *ethconfig.Config

	// Handlers
	txPool             *chain.TxPool
	blockchain         *chain.BlockChain
	handler            *handler
	ethDialCandidates  enode.Iterator
	snapDialCandidates enode.Iterator

	// DB interfaces
	chainDb ethdb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests     chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer      *indexer.ChainIndexer          // Bloom indexer operating during block imports
	closeBloomHandler chan struct{}

	APIBackend *EthAPIBackend

	miner     *miner.Miner
	gasPrice  *big.Int
	etherbase common.Address

	txFeeRecipient common.Address
	blsbase        common.Address

	networkID     uint64
	netRPCService *atlasapi.PublicNetAPI

	p2pServer *p2p.Server

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and etherbase)
}

// New creates a new Ethereum object (including the
// initialisation of the common Ethereum object)
func New(stack *node.Node, config *ethconfig.Config) (*Ethereum, error) {
	// Ensure configuration values are compatible and sane
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run atlas in light sync mode")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	if config.Miner.GasPrice == nil || config.Miner.GasPrice.Cmp(common.Big0) <= 0 {
		log.Warn("Sanitizing invalid miner gas price", "provided", config.Miner.GasPrice, "updated", ethconfig.Defaults.Miner.GasPrice)
		config.Miner.GasPrice = new(big.Int).Set(ethconfig.Defaults.Miner.GasPrice)
	}
	if config.NoPruning && config.TrieDirtyCache > 0 {
		if config.SnapshotCache > 0 {
			config.TrieCleanCache += config.TrieDirtyCache * 3 / 5
			config.SnapshotCache += config.TrieDirtyCache * 2 / 5
		} else {
			config.TrieCleanCache += config.TrieDirtyCache
		}
		config.TrieDirtyCache = 0
	}
	log.Info("Allocated trie memory caches", "clean", common.StorageSize(config.TrieCleanCache)*1024*1024, "dirty", common.StorageSize(config.TrieDirtyCache)*1024*1024)

	// Assemble the Ethereum object
	chainDb, err := stack.OpenDatabaseWithFreezer("chaindata", config.DatabaseCache, config.DatabaseHandles, config.DatabaseFreezer, "atlas/db/chaindata/", false)
	if err != nil {
		return nil, err
	}

	ethereum.MakeGlobalEthash(stack.DataDir())
	chainConfig, genesisHash, genesisErr := chain.SetupGenesisBlockWithOverride(chainDb, config.Genesis, config.OverrideChurrito)
	if _, ok := genesisErr.(*ethparams.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	if err := pruner.RecoverPruning(stack.ResolvePath(""), chainDb, stack.ResolvePath(config.TrieCleanCacheJournal)); err != nil {
		log.Error("Failed to recover state", "error", err)
	}
	eth := &Ethereum{
		config:            config,
		chainDb:           chainDb,
		eventMux:          stack.EventMux(),
		accountManager:    stack.AccountManager(),
		engine:            CreateConsensusEngine(stack, chainConfig, config, chainDb),
		closeBloomHandler: make(chan struct{}),
		networkID:         config.NetworkId,
		//gasPrice:          config.Miner.GasPrice,
		etherbase:      config.Miner.Etherbase,
		txFeeRecipient: config.TxFeeRecipient,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   indexer.NewBloomIndexer(chainDb, ethparams.BloomBitsBlocks, ethparams.BloomConfirms),
		p2pServer:      stack.Server(),
	}

	bcVersion := rawdb.ReadDatabaseVersion(chainDb)
	var dbVer = "<nil>"
	if bcVersion != nil {
		dbVer = fmt.Sprintf("%d", *bcVersion)
	}
	log.Info("Initialising Ethereum protocol", "network", config.NetworkId, "dbversion", dbVer)

	if !config.SkipBcVersionCheck {
		if bcVersion != nil && *bcVersion > chain.BlockChainVersion {
			return nil, fmt.Errorf("database version is v%d, Geth %s only supports v%d", *bcVersion, params.VersionWithMeta, chain.BlockChainVersion)
		} else if bcVersion == nil || *bcVersion < chain.BlockChainVersion {
			if bcVersion != nil { // only print warning on upgrade, not on init
				log.Warn("Upgrade blockchain database version", "from", dbVer, "to", chain.BlockChainVersion)
			}
			rawdb.WriteDatabaseVersion(chainDb, chain.BlockChainVersion)
		}
	}
	var (
		vmConfig = vm.Config{
			EnablePreimageRecording: config.EnablePreimageRecording,
		}
		cacheConfig = &chain.CacheConfig{
			TrieCleanLimit:      config.TrieCleanCache,
			TrieCleanJournal:    stack.ResolvePath(config.TrieCleanCacheJournal),
			TrieCleanRejournal:  config.TrieCleanCacheRejournal,
			TrieCleanNoPrefetch: config.NoPrefetch,
			TrieDirtyLimit:      config.TrieDirtyCache,
			TrieDirtyDisabled:   config.NoPruning,
			TrieTimeLimit:       config.TrieTimeout,
			SnapshotLimit:       config.SnapshotCache,
			Preimages:           config.Preimages,
		}
	)
	eth.blockchain, err = chain.NewBlockChain(chainDb, cacheConfig, chainConfig, eth.engine, vmConfig, eth.shouldPreserve, &config.TxLookupLimit)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*ethparams.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		eth.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	eth.bloomIndexer.Start(eth.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = stack.ResolvePath(config.TxPool.Journal)
	}
	eth.txPool = chain.NewTxPool(config.TxPool, chainConfig, eth.blockchain)

	// Permit the downloader to use the trie cache allowance during fast sync
	cacheLimit := cacheConfig.TrieCleanLimit + cacheConfig.TrieDirtyLimit + cacheConfig.SnapshotLimit
	checkpoint := config.Checkpoint
	if checkpoint == nil {
		checkpoint = ethparams.TrustedCheckpoints[genesisHash]
	}

	if eth.handler, err = newHandler(&handlerConfig{
		Database:   chainDb,
		Chain:      eth.blockchain,
		TxPool:     eth.txPool,
		Network:    config.NetworkId,
		Sync:       config.SyncMode,
		BloomCache: uint64(cacheLimit),
		EventMux:   eth.eventMux,
		Checkpoint: checkpoint,
		Whitelist:  config.Whitelist,
		//Engine:      eth.engine,
		Server:      stack.Server(),
		ProxyServer: nil,
	}); err != nil {
		return nil, err
	}

	// If the engine is istanbul, then inject the blockchain
	if istanbul, isIstanbul := eth.engine.(*istanbulBackend.Backend); isIstanbul {
		istanbul.SetChain(
			eth.blockchain, eth.blockchain.CurrentBlock,
			func(hash common.Hash) (*state.StateDB, error) {
				stateRoot := eth.blockchain.GetHeaderByHash(hash).Root
				return eth.blockchain.StateAt(stateRoot)
			})
	}

	eth.miner = miner.New(eth, &config.Miner, chainConfig, eth.EventMux(), eth.engine, eth.isLocalBlock, chainDb)

	eth.miner.SetExtra(makeExtraData(config.Miner.ExtraData))

	eth.APIBackend = &EthAPIBackend{stack.Config().ExtRPCEnabled(), stack.Config().AllowUnprotectedTxs, eth, nil}
	if eth.APIBackend.allowUnprotectedTxs {
		log.Info("Unprotected transactions allowed")
	}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.Miner.GasPrice
	}
	eth.APIBackend.gpo = gasprice.NewOracle(eth.APIBackend, gpoParams)

	// Setup DNS discovery iterators.
	dnsclient := dnsdisc.NewClient(dnsdisc.Config{})
	eth.ethDialCandidates, err = dnsclient.NewIterator(eth.config.EthDiscoveryURLs...)
	if err != nil {
		return nil, err
	}
	eth.snapDialCandidates, err = dnsclient.NewIterator(eth.config.SnapDiscoveryURLs...)
	if err != nil {
		return nil, err
	}

	// Start the RPC service
	eth.netRPCService = atlasapi.NewPublicNetAPI(eth.p2pServer, config.NetworkId)
	if config.VerifyCheckPoint {
		log.Info("[atlas start on verify the check point]")
		err := chain.VerifyCheckPoint(true, eth.blockchain)
		if err != nil {
			return nil, err
		}
		log.Info("[atlas start on verify the check point. pass....]")
	}
	// Register the backend on the node
	stack.RegisterAPIs(eth.APIs())
	stack.RegisterProtocols(eth.Protocols())
	stack.RegisterLifecycle(eth)
	// Check for unclean shutdown
	if uncleanShutdowns, discards, err := rawdb.PushUncleanShutdownMarker(chainDb); err != nil {
		log.Error("Could not update unclean-shutdown-marker list", "error", err)
	} else {
		if discards > 0 {
			log.Warn("Old unclean shutdowns found", "count", discards)
		}
		for _, tstamp := range uncleanShutdowns {
			t := time.Unix(int64(tstamp), 0)
			log.Warn("Unclean shutdown detected", "booted", t,
				"age", common.PrettyAge(t))
		}
	}
	return eth, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"geth",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > ethparams.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", ethparams.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// APIs return the collection of RPC services the ethereum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Ethereum) APIs() []rpc.API {
	apis := atlasapi.GetAPIs(s.APIBackend, s.engine)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.handler.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.APIBackend, false, 5*time.Minute),
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
			Service:   NewPrivateDebugAPI(s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Ethereum) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Ethereum) Etherbase() (eb common.Address, err error) {
	s.lock.RLock()
	etherbase := s.etherbase
	s.lock.RUnlock()

	if etherbase != (common.Address{}) {
		return etherbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			etherbase := accounts[0].Address

			s.lock.Lock()
			s.etherbase = etherbase
			s.lock.Unlock()

			log.Info("Etherbase automatically configured", "address", etherbase)
			return etherbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("etherbase must be explicitly specified")
}

func (s *Ethereum) TxFeeRecipient() (common.Address, error) {
	s.lock.RLock()
	txFeeRecipient := s.txFeeRecipient
	s.lock.RUnlock()

	if txFeeRecipient != (common.Address{}) {
		return txFeeRecipient, nil
	}
	return common.Address{}, fmt.Errorf("txFeeRecipient must be explicitly specified")
}

func (s *Ethereum) BLSbase() (eb common.Address, err error) {
	s.lock.RLock()
	blsbase := s.blsbase
	s.lock.RUnlock()

	if blsbase != (common.Address{}) {
		return blsbase, nil
	}

	return s.Etherbase()
}

// isLocalBlock checks whether the specified block is mined
// by local miner accounts.
//
// We regard two types of accounts as local miner account: etherbase
// and accounts specified via `txpool.locals` flag.
func (s *Ethereum) isLocalBlock(block *types.Block) bool {
	author, err := s.engine.Author(block.Header())
	if err != nil {
		log.Warn("Failed to retrieve block author", "number", block.NumberU64(), "hash", block.Hash(), "err", err)
		return false
	}
	// Check whether the given address is etherbase.
	s.lock.RLock()
	etherbase := s.etherbase
	s.lock.RUnlock()
	if author == etherbase {
		return true
	}
	// Check whether the given address is specified by `txpool.local`
	// CLI flag.
	for _, account := range s.config.TxPool.Locals {
		if account == author {
			return true
		}
	}
	return false
}

// shouldPreserve checks whether we should preserve the given block
// during the chain reorg depending on whether the author of block
// is a local account.
func (s *Ethereum) shouldPreserve(block *types.Block) bool {
	// The reason we need to disable the self-reorg preserving for clique
	// is it can be probable to introduce a deadlock.
	//
	// e.g. If there are 7 available signers
	//
	// r1   A
	// r2     B
	// r3       C
	// r4         D
	// r5   A      [X] F G
	// r6    [X]
	//
	// In the round5, the inturn signer E is offline, so the worst case
	// is A, F and G sign the block of round5 and reject the block of opponents
	// and in the round6, the last available signer B is offline, the whole
	// network is stuck.
	//if _, ok := s.engine.(*clique.Clique); ok {
	//	return false
	//}
	return s.isLocalBlock(block)
}

// SetValidator sets the mining reward address.
func (s *Ethereum) SetValidator(etherbase common.Address) {
	s.lock.Lock()
	s.etherbase = etherbase
	s.lock.Unlock()

	s.miner.SetValidator(etherbase)
}

// SetTxFeeRecipient sets the mining reward address.
func (s *Ethereum) SetTxFeeRecipient(txFeeRecipient common.Address) {
	s.lock.Lock()
	s.txFeeRecipient = txFeeRecipient
	s.lock.Unlock()

	s.miner.SetTxFeeRecipient(txFeeRecipient)
}

// StartMining starts the miner with the given number of CPU threads. If mining
// is already running, this method adjust the number of threads allowed to use
// and updates the minimum price required by the transaction pool.
// todo ibft
func (s *Ethereum) StartMining(threads int) error {
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

	if !s.IsMining() {

		// Configure the local mining address
		validator, err := s.Etherbase()
		if err != nil {
			log.Error("Cannot start mining without validator", "err", err)
			return fmt.Errorf("validator missing: %v", err)
		}

		txFeeRecipient, err := s.TxFeeRecipient()
		if err != nil {
			log.Error("Cannot start mining without txFeeRecipient", "err", err)
			return fmt.Errorf("txFeeRecipient missing: %v", err)
		}

		blsbase, err := s.BLSbase()
		if err != nil {
			log.Error("Cannot start mining without blsbase", "err", err)
			return fmt.Errorf("blsbase missing: %v", err)
		}

		if istanbul, isIstanbul := s.engine.(*istanbulBackend.Backend); isIstanbul {
			valAccount := accounts.Account{Address: validator}
			wallet, err := s.accountManager.Find(valAccount)
			if wallet == nil || err != nil {
				log.Error("Validator account unavailable locally", "err", err)
				return fmt.Errorf("signer missing: %v", err)
			}
			publicKey, err := wallet.GetPublicKey(valAccount)
			if err != nil {
				return fmt.Errorf("ECDSA public key missing: %v", err)
			}
			blswallet, err := s.accountManager.Find(accounts.Account{Address: blsbase})
			if blswallet == nil || err != nil {
				log.Error("BLSbase account unavailable locally", "err", err)
				return fmt.Errorf("BLS signer missing: %v", err)
			}

			istanbul.Authorize(validator, blsbase, publicKey, wallet.Decrypt, wallet.SignData, blswallet.SignBLS, wallet.SignHash)

			if istanbul.IsProxiedValidator() {
				if err := istanbul.StartProxiedValidatorEngine(); err != nil {
					log.Error("Error in starting proxied validator engine", "err", err)
					return err
				}
			}
		}

		// If mining is started, we can disable the transaction rejection mechanism
		// introduced to speed sync times.
		atomic.StoreUint32(&s.handler.acceptTxs, 1)

		go s.miner.Start(validator, txFeeRecipient)
	}
	return nil
}

// StopMining terminates the miner, both at the consensus engine level as well as
// at the block creation level.
// todo ibft
func (s *Ethereum) StopMining() {
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

func (s *Ethereum) startAnnounce() error {
	if istanbul, ok := s.engine.(consensus.Istanbul); ok {
		return istanbul.StartAnnouncing()
	}

	return nil
}

func (s *Ethereum) stopAnnounce() error {
	if istanbul, ok := s.engine.(consensus.Istanbul); ok {
		return istanbul.StopAnnouncing()
	}

	return nil
}

func (s *Ethereum) IsMining() bool      { return s.miner.Mining() }
func (s *Ethereum) Miner() *miner.Miner { return s.miner }

func (s *Ethereum) AccountManager() *accounts.Manager   { return s.accountManager }
func (s *Ethereum) BlockChain() *chain.BlockChain       { return s.blockchain }
func (s *Ethereum) Config() *ethconfig.Config           { return s.config }
func (s *Ethereum) TxPool() *chain.TxPool               { return s.txPool }
func (s *Ethereum) EventMux() *event.TypeMux            { return s.eventMux }
func (s *Ethereum) Engine() consensus.Engine            { return s.engine }
func (s *Ethereum) ChainDb() ethdb.Database             { return s.chainDb }
func (s *Ethereum) IsListening() bool                   { return true } // Always listening
func (s *Ethereum) EthVersion() int                     { return int(eth.ProtocolVersions[0]) }
func (s *Ethereum) NetVersion() uint64                  { return s.networkID }
func (s *Ethereum) Downloader() *downloader.Downloader  { return s.handler.downloader }
func (s *Ethereum) Synced() bool                        { return atomic.LoadUint32(&s.handler.acceptTxs) == 1 }
func (s *Ethereum) ArchiveMode() bool                   { return s.config.NoPruning }
func (s *Ethereum) BloomIndexer() *indexer.ChainIndexer { return s.bloomIndexer }

// Protocols returns all the currently configured
// network protocols to start.
func (s *Ethereum) Protocols() []p2p.Protocol {
	protos := eth.MakeProtocols((*ethHandler)(s.handler), s.networkID, s.ethDialCandidates)
	if s.config.SnapshotCache > 0 {
		protos = append(protos, snap.MakeProtocols((*snapHandler)(s.handler), s.snapDialCandidates)...)
	}
	return protos
}

// Start implements node.Lifecycle, starting all internal goroutines needed by the
// Ethereum protocol implementation.
func (s *Ethereum) Start() error {
	eth.StartENRUpdater(s.blockchain, s.p2pServer.LocalNode())

	// Start the bloom bits servicing goroutines
	s.startBloomHandlers(ethparams.BloomBitsBlocks)

	// Figure out a max peers count based on the server limits
	maxPeers := s.p2pServer.MaxPeers
	if s.config.LightServ > 0 {
		if s.config.LightPeers >= s.p2pServer.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", s.config.LightPeers, s.p2pServer.MaxPeers)
		}
		maxPeers -= s.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	s.handler.Start(maxPeers)

	if err := s.startAnnounce(); err != nil {
		return err
	}
	return nil
}

// Stop implements node.Lifecycle, terminating all internal goroutines used by the
// Ethereum protocol.
func (s *Ethereum) Stop() error {
	// Stop all the peer-related stuff first.
	s.ethDialCandidates.Close()
	s.snapDialCandidates.Close()
	s.stopAnnounce()
	s.handler.Stop()

	// Then stop everything else.
	s.bloomIndexer.Close()
	close(s.closeBloomHandler)
	s.txPool.Stop()
	s.miner.Close()
	s.blockchain.Stop()
	s.engine.Close()
	rawdb.PopUncleanShutdownMarker(s.chainDb)
	s.chainDb.Close()
	s.eventMux.Stop()

	return nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Ethereum service
func CreateConsensusEngine(stack *node.Node, chainConfig *params.ChainConfig, config *ethconfig.Config, db ethdb.Database) consensus.Engine {
	if chainConfig.Faker {
		return consensustest.NewFaker()
	}
	// If Istanbul is requested, set it up
	if chainConfig.Istanbul != nil {
		log.Debug("Setting up Istanbul consensus engine")
		if err := istanbul.ApplyParamsChainConfigToConfig(chainConfig, &config.Istanbul); err != nil {
			log.Crit("Invalid Configuration for Istanbul Engine", "err", err)
		}

		return istanbulBackend.New(&config.Istanbul, db)
	}
	log.Error(fmt.Sprintf("Only Istanbul Consensus is supported: %v", chainConfig))
	return nil
}
