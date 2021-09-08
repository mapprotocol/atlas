// Copyright 2015 The go-ethereum Authors
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

package atlas

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/mapprotocol/atlas/atlas/downloader"
	"github.com/mapprotocol/atlas/atlas/fetcher"
	"github.com/mapprotocol/atlas/atlas/protocols/eth"
	"github.com/mapprotocol/atlas/consensus"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/core"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/core/forkid"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/p2p"
	params2 "github.com/mapprotocol/atlas/params"
	"math"
	"math/big"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var (
	syncChallengeTimeout = 15 * time.Second // Time allowance for a node to reply to the sync progress challenge
)

const (
	softResponseLimit = 2 * 1024 * 1024 // Target maximum size of returned blocks, headers or node data.
	estHeaderRlpSize  = 500             // Approximate size of an RLP encoded block header

	// txChanSize is the size of channel listening to NewTxsEvent.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096
)

// txPool defines the methods needed from a transaction pool implementation to
// support all the operations needed by the Ethereum chain protocols.
type txPool interface {
	// Has returns an indicator whether txpool has a transaction
	// cached with the given hash.
	Has(hash common.Hash) bool

	// Get retrieves the transaction from local txpool with given
	// tx hash.
	Get(hash common.Hash) *types.Transaction

	// AddRemotes should add the given transactions to the pool.
	AddRemotes([]*types.Transaction) []error

	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	Pending() (map[common.Address]types.Transactions, error)

	// SubscribeNewTxsEvent should return an event subscription of
	// NewTxsEvent and send events to the given channel.
	SubscribeNewTxsEvent(chan<- core.NewTxsEvent) event.Subscription
}

// handlerConfig is the collection of initialization parameters to create a full
// node network handler.
type handlerConfig struct {
	Database   ethdb.Database            // Database for direct sync insertions
	Chain      *chain.BlockChain         // Blockchain to serve data from
	TxPool     txPool                    // Transaction pool to propagate from
	Network    uint64                    // Network identifier to adfvertise
	Sync       downloader.SyncMode       // Whether to fast or full sync
	BloomCache uint64                    // Megabytes to alloc for fast sync bloom
	EventMux   *event.TypeMux            // Legacy event mux, deprecate for `feed`
	Checkpoint *params.TrustedCheckpoint // Hard coded checkpoint for sync challenges
	Whitelist  map[uint64]common.Hash    // Hard coded whitelist for sync challenged
}

type ProtocolManager struct {
	networkID  uint64
	forkFilter forkid.Filter // Fork ID filter, constant across the lifetime of the node

	fastSync  uint32 // Flag whether fast sync is enabled (gets disabled if we already have blocks)
	acceptTxs uint32 // Flag whether we're considered synchronised (enables transaction processing)

	checkpointNumber uint64      // Block number for the sync progress validator to cross reference
	checkpointHash   common.Hash // Block hash for the sync progress validator to cross reference

	txpool     txPool
	blockchain *chain.BlockChain
	chaindb    ethdb.Database
	maxPeers   int

	downloader   *downloader.Downloader
	blockFetcher *fetcher.BlockFetcher
	txFetcher    *fetcher.TxFetcher
	peers        *peerSet

	eventMux      *event.TypeMux
	txsCh         chan core.NewTxsEvent
	txsSub        event.Subscription
	minedBlockSub *event.TypeMuxSubscription

	whitelist map[uint64]common.Hash

	// channels for fetcher, syncer, txsyncLoop
	txsyncCh chan *txsync
	quitSync chan struct{}

	chainSync *chainSyncer
	wg        sync.WaitGroup
	peerWG    sync.WaitGroup

	engine consensus.Engine

	server      *p2p.Server
	proxyServer *p2p.Server

	// Test fields or hooks
	broadcastTxAnnouncesOnly bool // Testing field, disable transaction propagation
}

// NewProtocolManager returns a new Ethereum sub protocol manager. The Ethereum sub protocol manages peers capable
// with the Ethereum network.
func newHandler(config *params2.ChainConfig, checkpoint *params.TrustedCheckpoint, mode downloader.SyncMode, networkID uint64, mux *event.TypeMux,
	txpool txPool, engine consensus.Engine, blockchain *chain.BlockChain, chaindb ethdb.Database,
	cacheLimit int, whitelist map[uint64]common.Hash, server *p2p.Server, proxyServer *p2p.Server) (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		networkID:   networkID,
		forkFilter:  forkid.NewFilter(blockchain),
		eventMux:    mux,
		txpool:      txpool,
		blockchain:  blockchain,
		chaindb:     chaindb,
		peers:       newPeerSet(),
		whitelist:   whitelist,
		txsyncCh:    make(chan *txsync),
		quitSync:    make(chan struct{}),
		engine:      engine,
		server:      server,
		proxyServer: proxyServer,
	}

	if handler, ok := manager.engine.(consensus.Handler); ok {
		handler.SetBroadcaster(manager)
		handler.SetP2PServer(server)
	}

	if mode == downloader.FullSync {
		// The database seems empty as the current block is the genesis. Yet the fast
		// block is ahead, so fast sync was enabled for this node at a certain point.
		// The scenarios where this can happen is
		// * if the user manually (or via a bad block) rolled back a fast sync node
		//   below the sync point.
		// * the last fast sync is not finished while user specifies a full sync this
		//   time. But we don't have any recent state for full sync.
		// In these cases however it's safe to reenable fast sync.
		fullBlock, fastBlock := blockchain.CurrentBlock(), blockchain.CurrentFastBlock()
		if fullBlock.NumberU64() == 0 && fastBlock.NumberU64() > 0 {
			manager.fastSync = uint32(1)
			log.Warn("Switch sync mode from full sync to fast sync")
		}
	} else {
		if blockchain.CurrentBlock().NumberU64() > 0 {
			// Print warning log if database is not empty to run fast sync.
			log.Warn("Switch sync mode from fast sync to full sync")
		} else {
			// If fast sync was requested and our database is empty, grant it
			manager.fastSync = uint32(1)
		}
	}

	// If we have trusted checkpoints, enforce them on the chain
	if checkpoint != nil {
		manager.checkpointNumber = (checkpoint.SectionIndex+1)*params.CHTFrequency - 1
		manager.checkpointHash = checkpoint.SectionHead
	}

	// Construct the downloader (long sync) and its backing state bloom if fast
	// sync is requested. The downloader is responsible for deallocating the state
	// bloom when it's done.
	var stateBloom *trie.SyncBloom
	if atomic.LoadUint32(&manager.fastSync) == 1 {
		stateBloom = trie.NewSyncBloom(uint64(cacheLimit), chaindb)
	}
	manager.downloader = downloader.New(manager.checkpointNumber, chaindb, stateBloom, manager.eventMux, blockchain, nil, manager.removePeer)

	// Construct the fetcher (short sync)
	validator := func(header *types.Header) error {
		return engine.VerifyHeader(blockchain, header, true)
	}
	heighter := func() uint64 {
		return blockchain.CurrentBlock().NumberU64()
	}
	inserter := func(blocks types.Blocks) (int, error) {
		// If sync hasn't reached the checkpoint yet, deny importing weird blocks.
		//
		// Ideally we would also compare the head block's timestamp and similarly reject
		// the propagated block if the head is too old. Unfortunately there is a corner
		// case when starting new networks, where the genesis might be ancient (0 unix)
		// which would prevent full nodes from accepting it.
		if manager.blockchain.CurrentBlock().NumberU64() < manager.checkpointNumber {
			log.Warn("Unsynced yet, discarded propagated block", "number", blocks[0].Number(), "hash", blocks[0].Hash())
			return 0, nil
		}
		// If fast sync is running, deny importing weird blocks. This is a problematic
		// clause when starting up a new network, because fast-syncing miners might not
		// accept each others' blocks until a restart. Unfortunately we haven't figured
		// out a way yet where nodes can decide unilaterally whether the network is new
		// or not. This should be fixed if we figure out a solution.
		if atomic.LoadUint32(&manager.fastSync) == 1 {
			log.Warn("Fast syncing, discarded propagated block", "number", blocks[0].Number(), "hash", blocks[0].Hash())
			return 0, nil
		}
		n, err := manager.blockchain.InsertChain(blocks)
		if err == nil {
			atomic.StoreUint32(&manager.acceptTxs, 1) // Mark initial sync done on any fetcher import
		}
		return n, err
	}
	manager.blockFetcher = fetcher.NewBlockFetcher(false, nil, blockchain.GetBlockByHash, validator, manager.BroadcastBlock, heighter, nil, inserter, manager.removePeer)

	fetchTx := func(peer string, hashes []common.Hash) error {
		p := manager.peers.Peer(peer)
		if p == nil {
			return errors.New("unknown peer")
		}
		return p.RequestTxs(hashes)
	}
	manager.txFetcher = fetcher.NewTxFetcher(txpool.Has, txpool.AddRemotes, fetchTx)

	manager.chainSync = newChainSyncer(manager)

	return manager, nil
}

// newHandler returns a handler for all Ethereum chain management protocol.
//func newHandler(config *handlerConfig, engine consensus.Engine, server *p2p.Server) (*handler, error) {
//
//}

func (pm *ProtocolManager) makeProtocol(version uint) p2p.Protocol {
	length, ok := istanbul.ProtocolLengths[version]
	if !ok {
		panic("makeProtocol for unknown version")
	}

	return p2p.Protocol{
		Name:    istanbul.ProtocolName,
		Version: version,
		Length:  length,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			return pm.runPeer(pm.newPeer(int(version), p, rw, pm.txpool.Get))
		},
		NodeInfo: func() interface{} {
			return pm.NodeInfo()
		},
		PeerInfo: func(id enode.ID) interface{} {
			if p := pm.peers.Peer(fmt.Sprintf("%x", id[:8])); p != nil {
				return p.Info()
			}
			return nil
		},
	}
}

// unregisterPeer unregisters the peer from the protocol manager and its various services,
// but does not disconnect the peer at the p2p networking layer.
// It returns the peer, or nil if the peer had already been cleaned up previously.
func (pm *ProtocolManager) unregisterPeer(id string) *peer {
	// Short circuit if the peer has already been unregistered
	peer := pm.peers.Peer(id)
	if peer == nil {
		return nil
	}
	log.Debug("Unregistering peer from istanbul protocol manager", "peer", id)

	// Unregister the peer from the downloader, tx fetcher, consensus engine, and Ethereum peer set
	if err := pm.downloader.UnregisterPeer(id); err != nil {
		log.Error("Peer removal from downloader failed", "peer", id, "err", err)
	}
	if err := pm.txFetcher.Drop(id); err != nil {
		log.Error("Peer removal from tx fetcher  failed", "peer", id, "err", err)
	}
	if handler, ok := pm.engine.(consensus.Handler); ok {
		handler.UnregisterPeer(peer, peer.Peer.Server == pm.proxyServer)
	}
	if err := pm.peers.Unregister(id); err != nil {
		log.Error("Unregistering peer failed", "peer", id, "err", err)
	}
	return peer
}

// removePeer unregisters the peer and then disconnects it at the p2p (networking) layer.
// The caller of removePeer is responsible for logging any relevant error information.
func (pm *ProtocolManager) removePeer(id string) {
	peer := pm.unregisterPeer(id)
	if peer != nil {
		// Hard disconnect at the networking layer
		peer.Peer.Disconnect(p2p.DiscSubprotocolError)
	}
}

func (pm *ProtocolManager) Start(maxPeers int) {
	pm.maxPeers = maxPeers

	// broadcast transactions
	pm.wg.Add(1)
	pm.txsCh = make(chan core.NewTxsEvent, txChanSize)
	pm.txsSub = pm.txpool.SubscribeNewTxsEvent(pm.txsCh)
	go pm.txBroadcastLoop()

	// broadcast mined blocks
	pm.wg.Add(1)
	pm.minedBlockSub = pm.eventMux.Subscribe(core.NewMinedBlockEvent{})
	go pm.minedBroadcastLoop()

	// start sync handlers
	pm.wg.Add(2)
	go pm.chainSync.loop()
	go pm.txsyncLoop64() // TODO(karalabe): Legacy initial tx echange, drop with eth/64.
}

func (pm *ProtocolManager) Stop() {
	pm.txsSub.Unsubscribe()        // quits txBroadcastLoop
	pm.minedBlockSub.Unsubscribe() // quits blockBroadcastLoop

	// Quit chainSync and txsync64.
	// After this is done, no new peers will be accepted.
	close(pm.quitSync)
	pm.wg.Wait()

	// Disconnect existing sessions.
	// This also closes the gate for any new registrations on the peer set.
	// sessions which are already established but not added to pm.peers yet
	// will exit when they try to register.
	pm.peers.Close()
	pm.peerWG.Wait()

	log.Info("Ethereum protocol stopped")
}

func (pm *ProtocolManager) newPeer(pv int, p *p2p.Peer, rw p2p.MsgReadWriter, getPooledTx func(hash common.Hash) *types.Transaction) *peer {
	return newPeer(pv, p, rw, getPooledTx)
}

func (pm *ProtocolManager) runPeer(p *peer) error {
	if !pm.chainSync.handlePeerEvent(p) {
		return p2p.DiscQuitting
	}
	pm.peerWG.Add(1)
	defer pm.peerWG.Done()
	return pm.handle(p)
}

// handle is the callback invoked to manage the life cycle of an eth peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p *peer) error {
	p.Log().Info("Ethereum peer connected", "name", p.Name())

	// Execute the Ethereum handshake
	var (
		genesis = pm.blockchain.Genesis()
		head    = pm.blockchain.CurrentHeader()
		hash    = head.Hash()
		number  = head.Number.Uint64()
		td      = pm.blockchain.GetTd(hash, number)
	)
	if err := p.Handshake(pm.networkID, td, hash, genesis.Hash(), forkid.NewID(pm.blockchain.Config(), pm.blockchain.Genesis().Hash(), pm.blockchain.CurrentHeader().Number.Uint64()), pm.forkFilter); err != nil {
		p.Log().Info("Ethereum handshake failed", "err", err)
		return err
	}
	forcePeer := false
	if handler, ok := pm.engine.(consensus.Handler); ok {
		isValidator, err := handler.Handshake(p)
		if err != nil {
			p.Log().Warn("Istanbul handshake failed", "err", err)
			return err
		}
		forcePeer = isValidator
		p.Log().Debug("Peer completed Istanbul handshake", "forcePeer", forcePeer)
	}
	// Ignore max peer and max inbound peer check if:
	//  - this is a trusted or statically dialed peer
	//  - the peer is from from the proxy server (e.g. peers connected to this node's internal network interface)
	//  - forcePeer is true
	if !forcePeer {
		// KJUE - Remove the server not nil check after restoring peer check in server.go
		if p.Peer.Server != nil {
			if err := p.Peer.Server.CheckPeerCounts(p.Peer); err != nil {
				return err
			}
		}
		// The p2p server CheckPeerCounts only checks if the total peer count
		// (eth and les) exceeds the total max peers. This checks if the number
		// of eth peers exceeds the eth max peers.
		isStaticOrTrusted := p.Peer.Info().Network.Trusted || p.Peer.Info().Network.Static
		if !isStaticOrTrusted && pm.peers.Len() >= pm.maxPeers && p.Peer.Server != pm.proxyServer {
			return p2p.DiscTooManyPeers
		}
	}

	// Register the peer locally
	if err := pm.peers.Register(p, pm.removePeer); err != nil {
		p.Log().Error("Ethereum peer registration failed", "err", err)
		return err
	}
	defer pm.unregisterPeer(p.id)

	// Register the peer in the downloader. If the downloader considers it banned, we disconnect
	if err := pm.downloader.RegisterPeer(p.id, uint(p.version), p); err != nil {
		return err
	}

	// Register the peer with the consensus engine.
	if handler, ok := pm.engine.(consensus.Handler); ok {
		if err := handler.RegisterPeer(p, p.Peer.Server == pm.proxyServer); err != nil {
			return err
		}
	}

	pm.chainSync.handlePeerEvent(p)

	// Propagate existing transactions. new transactions appearing
	// after this will be sent via broadcasts.
	pm.syncTransactions(p)

	// If we have a trusted CHT, reject all peers below that (avoid fast sync eclipse)
	if pm.checkpointHash != (common.Hash{}) {
		// Request the peer's checkpoint header for chain height/weight validation
		if err := p.RequestHeadersByNumber(pm.checkpointNumber, 1, 0, false); err != nil {
			return err
		}
		// Start a timer to disconnect if the peer doesn't reply in time
		p.syncDrop = time.AfterFunc(syncChallengeTimeout, func() {
			p.Log().Warn("Checkpoint challenge timed out, dropping", "addr", p.RemoteAddr(), "type", p.Name())
			pm.removePeer(p.id)
		})
		// Make sure it's cleaned up if the peer dies off
		defer func() {
			if p.syncDrop != nil {
				p.syncDrop.Stop()
				p.syncDrop = nil
			}
		}()
	}
	// If we have any explicit whitelist block hashes, request them
	for number := range pm.whitelist {
		if err := p.RequestHeadersByNumber(number, 1, 0, false); err != nil {
			return err
		}
	}
	// Handle incoming messages until the connection is torn down
	for {
		if err := pm.handleMsg(p); err != nil {
			p.Log().Debug("Ethereum message handling failed", "err", err)
			return err
		}
	}
}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (pm *ProtocolManager) handleMsg(p *peer) error {
	// Read the next message from the remote peer, and ensure it's fully consumed
	msg, err := p.ReadMsg()
	if err != nil {
		return err
	}
	defer msg.Discard()

	// Send messages to the consensus engine first. If they are consensus related,
	// e.g. for IBFT, let the consensus handler handle the message.
	if handler, ok := pm.engine.(consensus.Handler); ok {
		pubKey := p.Node().Pubkey()
		addr := crypto.PubkeyToAddress(*pubKey)
		handled, err := handler.HandleMsg(addr, msg, p)
		if handled {
			return err
		}
	}

	// Handle the message depending on its contents
	switch {
	case msg.Code == eth.StatusMsg:
		// Status messages should never arrive after the handshake
		return errResp(ErrExtraStatusMsg, "uncontrolled status message")

	// Block header query, collect the requested headers and reply
	case msg.Code == eth.GetBlockHeadersMsg:
		// Decode the complex header query
		var query getBlockHeadersData
		if err := msg.Decode(&query); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		hashMode := query.Origin.Hash != (common.Hash{})
		first := true
		maxNonCanonical := uint64(100)

		// Gather headers until the fetch or network limits is reached
		var (
			bytes   common.StorageSize
			headers []*types.Header
			unknown bool
		)
		for !unknown && len(headers) < int(query.Amount) && bytes < softResponseLimit && len(headers) < downloader.MaxHeaderFetch {
			// Retrieve the next header satisfying the query
			var origin *types.Header
			if hashMode {
				if first {
					first = false
					origin = pm.blockchain.GetHeaderByHash(query.Origin.Hash)
					if origin != nil {
						query.Origin.Number = origin.Number.Uint64()
					}
				} else {
					origin = pm.blockchain.GetHeader(query.Origin.Hash, query.Origin.Number)
				}
			} else {
				origin = pm.blockchain.GetHeaderByNumber(query.Origin.Number)
			}
			if origin == nil {
				break
			}
			headers = append(headers, origin)
			bytes += estHeaderRlpSize

			// Advance to the next header of the query
			switch {
			case hashMode && query.Reverse:
				// Hash based traversal towards the genesis block
				ancestor := query.Skip + 1
				if ancestor == 0 {
					unknown = true
				} else {
					query.Origin.Hash, query.Origin.Number = pm.blockchain.GetAncestor(query.Origin.Hash, query.Origin.Number, ancestor, &maxNonCanonical)
					unknown = (query.Origin.Hash == common.Hash{})
				}
			case hashMode && !query.Reverse:
				// Hash based traversal towards the leaf block
				var (
					current = origin.Number.Uint64()
					next    = current + query.Skip + 1
				)
				if next <= current {
					infos, _ := json.MarshalIndent(p.Peer.Info(), "", "  ")
					p.Log().Warn("GetBlockHeaders skip overflow attack", "current", current, "skip", query.Skip, "next", next, "attacker", infos)
					unknown = true
				} else {
					if header := pm.blockchain.GetHeaderByNumber(next); header != nil {
						nextHash := header.Hash()
						expOldHash, _ := pm.blockchain.GetAncestor(nextHash, next, query.Skip+1, &maxNonCanonical)
						if expOldHash == query.Origin.Hash {
							query.Origin.Hash, query.Origin.Number = nextHash, next
						} else {
							unknown = true
						}
					} else {
						unknown = true
					}
				}
			case query.Reverse:
				// Number based traversal towards the genesis block
				if query.Origin.Number >= query.Skip+1 {
					query.Origin.Number -= query.Skip + 1
				} else {
					unknown = true
				}

			case !query.Reverse:
				// Number based traversal towards the leaf block
				query.Origin.Number += query.Skip + 1
			}
		}
		return p.SendBlockHeaders(headers)

	case msg.Code == eth.BlockHeadersMsg:
		// A batch of headers arrived to one of our previous requests
		var headers []*types.Header
		if err := msg.Decode(&headers); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// If no headers were received, but we're expencting a checkpoint header, consider it that
		if len(headers) == 0 && p.syncDrop != nil {
			// Stop the timer either way, decide later to drop or not
			p.syncDrop.Stop()
			p.syncDrop = nil

			// If we're doing a fast sync, we must enforce the checkpoint block to avoid
			// eclipse attacks. Unsynced nodes are welcome to connect after we're done
			// joining the network
			if atomic.LoadUint32(&pm.fastSync) == 1 {
				p.Log().Warn("Dropping unsynced node during fast sync", "addr", p.RemoteAddr(), "type", p.Name())
				return errors.New("unsynced node cannot serve fast sync")
			}
		}
		// Filter out any explicitly requested headers, deliver the rest to the downloader
		filter := len(headers) == 1
		if filter {
			// If it's a potential sync progress check, validate the content and advertised chain weight
			if p.syncDrop != nil && headers[0].Number.Uint64() == pm.checkpointNumber {
				// Disable the sync drop timer
				p.syncDrop.Stop()
				p.syncDrop = nil

				// Validate the header and either drop the peer or continue
				if headers[0].Hash() != pm.checkpointHash {
					return errors.New("checkpoint hash mismatch")
				}
				return nil
			}
			// Otherwise if it's a whitelisted block, validate against the set
			if want, ok := pm.whitelist[headers[0].Number.Uint64()]; ok {
				if hash := headers[0].Hash(); want != hash {
					p.Log().Info("Whitelist mismatch, dropping peer", "number", headers[0].Number.Uint64(), "hash", hash, "want", want)
					return errors.New("whitelist block mismatch")
				}
				p.Log().Debug("Whitelist block verified", "number", headers[0].Number.Uint64(), "hash", want)
			}
			// Irrelevant of the fork checks, send the header to the fetcher just in case
			headers = pm.blockFetcher.FilterHeaders(p.id, headers, time.Now())
		}
		if len(headers) > 0 || !filter {
			err := pm.downloader.DeliverHeaders(p.id, headers)
			if err != nil {
				log.Debug("Failed to deliver headers", "err", err)
			}
		}

	case msg.Code == eth.GetBlockBodiesMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather blocks until the fetch or network limits is reached
		var (
			hash                 common.Hash
			bytes                int
			bodiesAndBlockHashes []rlp.RawValue
		)
		for bytes < softResponseLimit && len(bodiesAndBlockHashes) < downloader.MaxBlockFetch {
			// Retrieve the hash of the next block
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested block body, stopping if enough was found
			if body := pm.blockchain.GetBody(hash); body != nil {
				bh := &blockBodyWithBlockHash{BlockHash: hash, BlockBody: body}
				bhRLPbytes, err := rlp.EncodeToBytes(bh)
				if err != nil {
					return err
				}
				bhRLP := rlp.RawValue(bhRLPbytes)
				bodiesAndBlockHashes = append(bodiesAndBlockHashes, bhRLP)
				bytes += len(bhRLP)
			}
		}
		return p.SendBlockBodiesRLP(bodiesAndBlockHashes)

	case msg.Code == eth.BlockBodiesMsg:
		// A batch of block bodies arrived to one of our previous requests
		var request blockBodiesData
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver them all to the downloader for queuing
		blockHashes := make([]common.Hash, len(request))
		transactions := make([][]*types.Transaction, len(request))
		randomness := make([]*types.Randomness, len(request))
		epochSnarkData := make([]*types.EpochSnarkData, len(request))

		for i, blockBodyWithBlockHash := range request {
			blockHashes[i] = blockBodyWithBlockHash.BlockHash
			transactions[i] = blockBodyWithBlockHash.BlockBody.Transactions
			randomness[i] = blockBodyWithBlockHash.BlockBody.Randomness
			epochSnarkData[i] = blockBodyWithBlockHash.BlockBody.EpochSnarkData
		}
		// Filter out any explicitly requested bodies, deliver the rest to the downloader
		filter := len(blockHashes) > 0 || len(transactions) > 0 || len(randomness) > 0 || len(epochSnarkData) > 0
		if filter {
			blockHashes, transactions, randomness, epochSnarkData = pm.blockFetcher.FilterBodies(p.id, blockHashes, transactions, randomness, epochSnarkData, time.Now())
		}
		if len(blockHashes) > 0 || len(transactions) > 0 || len(randomness) > 0 || len(epochSnarkData) > 0 || !filter {
			err := pm.downloader.DeliverBodies(p.id, transactions, randomness, epochSnarkData)
			if err != nil {
				log.Debug("Failed to deliver bodies", "err", err)
			}
		}

	case msg.Code == eth.GetNodeDataMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather state data until the fetch or network limits is reached
		var (
			hash  common.Hash
			bytes int
			data  [][]byte
		)
		for bytes < softResponseLimit && len(data) < downloader.MaxStateFetch {
			// Retrieve the hash of the next state entry
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested state entry, stopping if enough was found
			if entry, err := pm.blockchain.TrieNode(hash); err == nil {
				data = append(data, entry)
				bytes += len(entry)
			}
		}
		return p.SendNodeData(data)

	case msg.Code == eth.NodeDataMsg:
		// A batch of node state data arrived to one of our previous requests
		var data [][]byte
		if err := msg.Decode(&data); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver all to the downloader
		if err := pm.downloader.DeliverNodeData(p.id, data); err != nil {
			log.Debug("Failed to deliver node state data", "err", err)
		}

	case msg.Code == eth.GetReceiptsMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather state data until the fetch or network limits is reached
		var (
			hash     common.Hash
			bytes    int
			receipts []rlp.RawValue
		)
		for bytes < softResponseLimit && len(receipts) < downloader.MaxReceiptFetch {
			// Retrieve the hash of the next block
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested block's receipts, skipping if unknown to us
			results := pm.blockchain.GetReceiptsByHash(hash)
			if results == nil {
				if header := pm.blockchain.GetHeaderByHash(hash); header == nil || header.ReceiptHash != types.EmptyRootHash {
					continue
				}
			}
			// If known, encode and queue for response packet
			if encoded, err := rlp.EncodeToBytes(results); err != nil {
				log.Error("Failed to encode receipt", "err", err)
			} else {
				receipts = append(receipts, encoded)
				bytes += len(encoded)
			}
		}
		return p.SendReceiptsRLP(receipts)

	case msg.Code == eth.ReceiptsMsg:
		// A batch of receipts arrived to one of our previous requests
		var receipts [][]*types.Receipt
		if err := msg.Decode(&receipts); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver all to the downloader
		if err := pm.downloader.DeliverReceipts(p.id, receipts); err != nil {
			log.Debug("Failed to deliver receipts", "err", err)
		}

	case msg.Code == eth.NewBlockHashesMsg:
		var announces newBlockHashesData
		if err := msg.Decode(&announces); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		// Mark the hashes as present at the remote node
		for _, block := range announces {
			p.MarkBlock(block.Hash)
		}
		// Schedule all the unknown hashes for retrieval
		unknown := make(newBlockHashesData, 0, len(announces))
		for _, block := range announces {
			if !pm.blockchain.HasBlock(block.Hash, block.Number) {
				unknown = append(unknown, block)
			}
		}
		for _, block := range unknown {
			pm.blockFetcher.Notify(p.id, block.Hash, block.Number, time.Now(), p.RequestOneHeader, p.RequestBodies)
		}

	case msg.Code == eth.NewBlockMsg:
		// Retrieve and decode the propagated block
		var request newBlockData
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if hash := types.DeriveSha(request.Block.Transactions(),trie.NewStackTrie(nil)); hash != request.Block.TxHash() {
			log.Warn("Propagated block has invalid body", "have", hash, "exp", request.Block.TxHash())
			break // TODO(karalabe): return error eventually, but wait a few releases
		}
		if err := request.sanityCheck(); err != nil {
			return err
		}
		request.Block.ReceivedAt = msg.ReceivedAt
		request.Block.ReceivedFrom = p

		// Mark the peer as owning the block and schedule it for import
		p.MarkBlock(request.Block.Hash())
		pm.blockFetcher.Enqueue(p.id, request.Block)

		// Assuming the block is importable by the peer, but possibly not yet done so,
		// calculate the head hash and TD that the peer truly must have.
		var (
			trueHead = request.Block.ParentHash()
			trueTD   = new(big.Int).Sub(request.TD, big.NewInt(1))
		)
		// Update the peer's total difficulty if better than the previous
		if _, td := p.Head(); trueTD.Cmp(td) > 0 {
			p.SetHead(trueHead, trueTD)
			pm.chainSync.handlePeerEvent(p)
		}

	case msg.Code == eth.NewPooledTransactionHashesMsg && p.version >= istanbul.Celo66:
		// New transaction announcement arrived, make sure we have
		// a valid and fresh chain to handle them
		if atomic.LoadUint32(&pm.acceptTxs) == 0 {
			break
		}
		var hashes []common.Hash
		if err := msg.Decode(&hashes); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Schedule all the unknown hashes for retrieval
		for _, hash := range hashes {
			p.MarkTransaction(hash)
		}
		pm.txFetcher.Notify(p.id, hashes)

	case msg.Code == eth.GetPooledTransactionsMsg && p.version >= istanbul.Celo66:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather transactions until the fetch or network limits is reached
		var (
			hash   common.Hash
			bytes  int
			hashes []common.Hash
			txs    []rlp.RawValue
		)
		for bytes < softResponseLimit {
			// Retrieve the hash of the next block
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested transaction, skipping if unknown to us
			tx := pm.txpool.Get(hash)
			if tx == nil {
				continue
			}
			// If known, encode and queue for response packet
			if encoded, err := rlp.EncodeToBytes(tx); err != nil {
				log.Error("Failed to encode transaction", "err", err)
			} else {
				hashes = append(hashes, hash)
				txs = append(txs, encoded)
				bytes += len(encoded)
			}
		}
		return p.SendPooledTransactionsRLP(hashes, txs)

	case msg.Code == eth.TransactionsMsg || (msg.Code == eth.PooledTransactionsMsg && p.version >= istanbul.Celo66):
		// Transactions arrived, make sure we have a valid and fresh chain to handle them
		if atomic.LoadUint32(&pm.acceptTxs) == 0 {
			break
		}
		// Transactions can be processed, parse all of them and deliver to the pool
		var txs []*types.Transaction
		if err := msg.Decode(&txs); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		for i, tx := range txs {
			// Validate and mark the remote transaction
			if tx == nil {
				return errResp(ErrDecode, "transaction %d is nil", i)
			}
			p.MarkTransaction(tx.Hash())
		}
		pm.txFetcher.Enqueue(p.id, txs, msg.Code == eth.PooledTransactionsMsg)

	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}

func (pm *ProtocolManager) Enqueue(id string, block *types.Block) {
	pm.blockFetcher.Enqueue(id, block)
}

// BroadcastBlock will either propagate a block to a subset of its peers, or
// will only announce its availability (depending what's requested).
func (pm *ProtocolManager) BroadcastBlock(block *types.Block, propagate bool) {
	hash := block.Hash()
	peers := pm.peers.PeersWithoutBlock(hash)

	// If propagation is requested, send to a subset of the peer
	if propagate {
		// Calculate the TD of the block (it's not imported yet, so block.Td is not valid)
		var td *big.Int
		if parent := pm.blockchain.GetBlock(block.ParentHash(), block.NumberU64()-1); parent != nil {
			td = new(big.Int).Add(big.NewInt(1), pm.blockchain.GetTd(block.ParentHash(), block.NumberU64()-1))
		} else {
			log.Error("Propagating dangling block", "number", block.Number(), "hash", hash)
			return
		}
		// Send the block to a subset of our peers
		transfer := peers[:int(math.Sqrt(float64(len(peers))))]
		for _, peer := range transfer {
			peer.AsyncSendNewBlock(block, td)
		}
		log.Trace("Propagated block", "hash", hash, "recipients", len(transfer), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
		return
	}
	// Otherwise if the block is indeed in out own chain, announce it
	if pm.blockchain.HasBlock(hash, block.NumberU64()) {
		for _, peer := range peers {
			peer.AsyncSendNewBlockHash(block)
		}
		log.Trace("Announced block", "hash", hash, "recipients", len(peers), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
	}
}

// BroadcastTransactions will propagate a batch of transactions to all peers which are not known to
// already have the given transaction.
func (pm *ProtocolManager) BroadcastTransactions(txs types.Transactions, propagate bool) {
	var (
		txset = make(map[*peer][]common.Hash)
		annos = make(map[*peer][]common.Hash)
	)
	// Broadcast transactions to a batch of peers not knowing about it
	if propagate {
		for _, tx := range txs {
			peers := pm.peers.PeersWithoutTx(tx.Hash())

			// Send the block to a subset of our peers
			transfer := peers[:int(math.Sqrt(float64(len(peers))))]
			for _, peer := range transfer {
				txset[peer] = append(txset[peer], tx.Hash())
			}
			log.Trace("Broadcast transaction", "hash", tx.Hash(), "recipients", len(peers))
		}
		for peer, hashes := range txset {
			peer.AsyncSendTransactions(hashes)
		}
		return
	}
	// Otherwise only broadcast the announcement to peers
	for _, tx := range txs {
		peers := pm.peers.PeersWithoutTx(tx.Hash())
		for _, peer := range peers {
			annos[peer] = append(annos[peer], tx.Hash())
		}
	}
	for peer, hashes := range annos {
		if peer.version >= istanbul.Celo66 {
			peer.AsyncSendPooledTransactionHashes(hashes)
		} else {
			peer.AsyncSendTransactions(hashes)
		}
	}
}

// minedBroadcastLoop sends mined blocks to connected peers.
func (pm *ProtocolManager) minedBroadcastLoop() {
	defer pm.wg.Done()

	for obj := range pm.minedBlockSub.Chan() {
		if ev, ok := obj.Data.(core.NewMinedBlockEvent); ok {
			pm.BroadcastBlock(ev.Block, true)  // First propagate block to peers
			pm.BroadcastBlock(ev.Block, false) // Only then announce to the rest
		}
	}
}

// txBroadcastLoop announces new transactions to connected peers.
func (pm *ProtocolManager) txBroadcastLoop() {
	defer pm.wg.Done()

	for {
		select {
		case event := <-pm.txsCh:
			// For testing purpose only, disable propagation
			if pm.broadcastTxAnnouncesOnly {
				pm.BroadcastTransactions(event.Txs, false)
				continue
			}
			pm.BroadcastTransactions(event.Txs, true)  // First propagate transactions to peers
			pm.BroadcastTransactions(event.Txs, false) // Only then announce to the rest

		case <-pm.txsSub.Err():
			return
		}
	}
}

// NodeInfo represents a short summary of the Ethereum sub-protocol metadata
// known about the host peer.
type NodeInfo struct {
	Network    uint64              `json:"network"`    // Ethereum network ID (1=Frontier, 2=Morden, Ropsten=3, Rinkeby=4)
	Difficulty *big.Int            `json:"difficulty"` // Total difficulty of the host's blockchain
	Genesis    common.Hash         `json:"genesis"`    // SHA3 hash of the host's genesis block
	Config     *params2.ChainConfig `json:"config"`     // Chain configuration for the fork rules
	Head       common.Hash         `json:"head"`       // SHA3 hash of the host's best owned block
}

// NodeInfo retrieves some protocol metadata about the running host node.
func (pm *ProtocolManager) NodeInfo() *NodeInfo {
	currentBlock := pm.blockchain.CurrentBlock()
	return &NodeInfo{
		Network:    pm.networkID,
		Difficulty: pm.blockchain.GetTd(currentBlock.Hash(), currentBlock.NumberU64()),
		Genesis:    pm.blockchain.Genesis().Hash(),
		Config:     pm.blockchain.Config(),
		Head:       currentBlock.Hash(),
	}
}

func (pm *ProtocolManager) FindPeers(targets map[enode.ID]bool, purpose p2p.PurposeFlag) map[enode.ID]consensus.Peer {
	m := make(map[enode.ID]consensus.Peer)
	for _, p := range pm.peers.Peers() {
		id := p.Node().ID()
		if targets[id] || (targets == nil) {
			if p.PurposeIsSet(purpose) {
				m[id] = p
			}
		}
	}
	return m
}

// syncTransactions starts sending all currently pending transactions to the given peer.
func (pm *ProtocolManager) syncTransactions(p *peer) {
	// Assemble the set of transaction to broadcast or announce to the remote
	// peer. Fun fact, this is quite an expensive operation as it needs to sort
	// the transactions if the sorting is not cached yet. However, with a random
	// order, insertions could overflow the non-executable queues and get dropped.
	//
	// TODO(karalabe): Figure out if we could get away with random order somehow
	var txs types.Transactions
	pending, _ := pm.txpool.Pending()
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	if len(txs) == 0 {
		return
	}
	// The eth/65 (celo/66) protocol introduces proper transaction announcements, so instead
	// of dripping transactions across multiple peers, just send the entire list as
	// an announcement and let the remote side decide what they need (likely nothing).
	if p.version >= istanbul.Celo66 {
		hashes := make([]common.Hash, len(txs))
		for i, tx := range txs {
			hashes[i] = tx.Hash()
		}
		p.AsyncSendPooledTransactionHashes(hashes)
		return
	}
	// Out of luck, peer is running legacy protocols, drop the txs over
	select {
	case pm.txsyncCh <- &txsync{p: p, txs: txs}:
	case <-pm.quitSync:
	}
}

// Needed for eth/64 (celo/65).  With eth/65+ (celo/66+), this runs but
// won't get any messages through pm.txsyncCh and so won't do anything.
// txsyncLoop64 takes care of the initial transaction sync for each new
// connection. When a new peer appears, we relay all currently pending
// transactions. In order to minimise egress bandwidth usage, we send
// the transactions in small packs to one peer at a time.
func (pm *ProtocolManager) txsyncLoop64() {
	defer pm.wg.Done()

	var (
		pending = make(map[enode.ID]*txsync)
		sending = false               // whether a send is active
		pack    = new(txsync)         // the pack that is being sent
		done    = make(chan error, 1) // result of the send
	)

	// send starts a sending a pack of transactions from the sync.
	send := func(s *txsync) {
		if s.p.version >= istanbul.Celo66 {
			panic("initial transaction syncer running on eth/65+ (celo/66+)")
		}
		// Fill pack with transactions up to the target size.
		size := common.StorageSize(0)
		pack.p = s.p
		pack.txs = pack.txs[:0]
		for i := 0; i < len(s.txs) && size < txsyncPackSize; i++ {
			pack.txs = append(pack.txs, s.txs[i])
			size += s.txs[i].Size()
		}
		// Remove the transactions that will be sent.
		s.txs = s.txs[:copy(s.txs, s.txs[len(pack.txs):])]
		if len(s.txs) == 0 {
			delete(pending, s.p.ID())
		}
		// Send the pack in the background.
		s.p.Log().Trace("Sending batch of transactions", "count", len(pack.txs), "bytes", size)
		sending = true
		go func() { done <- pack.p.SendTransactions64(pack.txs) }()
	}

	// pick chooses the next pending sync.
	pick := func() *txsync {
		if len(pending) == 0 {
			return nil
		}
		n := rand.Intn(len(pending)) + 1
		for _, s := range pending {
			if n--; n == 0 {
				return s
			}
		}
		return nil
	}

	for {
		select {
		case s := <-pm.txsyncCh:
			pending[s.p.ID()] = s
			if !sending {
				send(s)
			}
		case err := <-done:
			sending = false
			// Stop tracking peers that cause send failures.
			if err != nil {
				pack.p.Log().Debug("Transaction send failed", "err", err)
				delete(pending, pack.p.ID())
			}
			// Schedule the next send.
			if s := pick(); s != nil {
				send(s)
			}
		case <-pm.quitSync:
			return
		}
	}
}

// doSync synchronizes the local blockchain with a remote peer.
func (pm *ProtocolManager) doSync(op *chainSyncOp) error {
	if op.mode == downloader.FastSync {
		// Before launch the fast sync, we have to ensure user uses the same
		// txlookup limit.
		// The main concern here is: during the fast sync Geth won't index the
		// block(generate tx indices) before the HEAD-limit. But if user changes
		// the limit in the next fast sync(e.g. user kill Geth manually and
		// restart) then it will be hard for Geth to figure out the oldest block
		// has been indexed. So here for the user-experience wise, it's non-optimal
		// that user can't change limit during the fast sync. If changed, Geth
		// will just blindly use the original one.
		limit := pm.blockchain.TxLookupLimit()
		if stored := rawdb.ReadFastTxLookupLimit(pm.chaindb); stored == nil {
			rawdb.WriteFastTxLookupLimit(pm.chaindb, limit)
		} else if *stored != limit {
			pm.blockchain.SetTxLookupLimit(*stored)
			log.Warn("Update txLookup limit", "provided", limit, "updated", *stored)
		}
	}
	// Run the sync cycle, and disable fast sync if we're past the pivot block
	err := pm.downloader.Synchronise(op.peer.id, op.head, op.td, op.mode)
	if err != nil {
		return err
	}
	if atomic.LoadUint32(&pm.fastSync) == 1 {
		log.Info("Fast sync complete, auto disabling")
		atomic.StoreUint32(&pm.fastSync, 0)
	}

	// If we've successfully finished a sync cycle and passed any required checkpoint,
	// enable accepting transactions from the network.
	head := pm.blockchain.CurrentBlock()
	if head.NumberU64() >= pm.checkpointNumber {
		// Checkpoint passed, sanity check the timestamp to have a fallback mechanism
		// for non-checkpointed (number = 0) private networks.
		if head.Time() >= uint64(time.Now().AddDate(0, -1, 0).Unix()) {
			atomic.StoreUint32(&pm.acceptTxs, 1)
		}
	}

	if head.NumberU64() > 0 {
		// We've completed a sync cycle, notify all peers of new state. This path is
		// essential in star-topology networks where a gateway node needs to notify
		// all its out-of-date peers of the availability of a new block. This failure
		// scenario will most often crop up in private and hackathon networks with
		// degenerate connectivity, but it should be healthy for the mainnet too to
		// more reliably update peers or the local TD state.
		pm.BroadcastBlock(head, false)
	}

	return nil
}


// newBlockData is the network packet for the block propagation message.
type newBlockData struct {
	Block *types.Block
	TD    *big.Int
}

// sanityCheck verifies that the values are reasonable, as a DoS protection
func (request *newBlockData) sanityCheck() error {
	if err := request.Block.SanityCheck(); err != nil {
		return err
	}
	return nil
}

type blockBodyWithBlockHash struct {
	BlockHash common.Hash
	BlockBody *types.Body
}

// blockBodiesData is the network packet for block content distribution.
type blockBodiesData []*blockBodyWithBlockHash