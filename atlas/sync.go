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
	"math/big"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/atlas/downloader"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/types"
)

const (
	forceSyncCycle      = 10 * time.Second // Time interval to force syncs, even if few peers are available
	defaultMinSyncPeers = 5                // Amount of peers desired to start syncing

	// This is the target size for the packs of transactions sent by txsyncLoop64.
	// A pack can get larger than this if a single transactions exceeds this size.
	txsyncPackSize = 100 * 1024
)

type txsync struct {
	p   *peer
	txs []*types.Transaction
}

// chainSyncer coordinates blockchain sync components.
type chainSyncer struct {
	pm          *ProtocolManager
	force       *time.Timer
	forced      bool // true when force timer fired
	peerEventCh chan struct{}
	doneCh      chan error // non-nil when sync is running
}

// chainSyncOp is a scheduled sync operation.
type chainSyncOp struct {
	mode downloader.SyncMode
	peer *peer
	td   *big.Int
	head common.Hash
}

// handlePeerEvent notifies the syncer about a change in the peer set.
// This is called for new peers and every time a peer announces a new
// chain head.
func (cs *chainSyncer) handlePeerEvent(p *peer) bool {
	select {
	case cs.peerEventCh <- struct{}{}:
		return true
	case <-cs.pm.quitSync:
		return false
	}
}

// loop runs in its own goroutine and launches the sync when necessary.
func (cs *chainSyncer) loop() {
	defer cs.pm.wg.Done()

	cs.pm.blockFetcher.Start()
	cs.pm.txFetcher.Start()
	defer cs.pm.blockFetcher.Stop()
	defer cs.pm.txFetcher.Stop()

	// The force timer lowers the peer count threshold down to one when it fires.
	// This ensures we'll always start sync even if there aren't enough peers.
	cs.force = time.NewTimer(forceSyncCycle)
	defer cs.force.Stop()

	for {
		if op := cs.nextSyncOp(); op != nil {
			cs.startSync(op)
		}

		select {
		case <-cs.peerEventCh:
			// Peer information changed, recheck.
		case <-cs.doneCh:
			cs.doneCh = nil
			cs.force.Reset(forceSyncCycle)
			cs.forced = false
		case <-cs.force.C:
			cs.forced = true

		case <-cs.pm.quitSync:
			// Disable all insertion on the blockchain. This needs to happen before
			// terminating the downloader because the downloader waits for blockchain
			// inserts, and these can take a long time to finish.
			cs.pm.blockchain.StopInsert()
			cs.pm.downloader.Terminate()
			if cs.doneCh != nil {
				// Wait for the current sync to end.
				<-cs.doneCh
			}
			return
		}
	}
}

// nextSyncOp determines whether sync is required at this time.
func (cs *chainSyncer) nextSyncOp() *chainSyncOp {
	if cs.doneCh != nil {
		return nil // Sync already running.
	}

	// Ensure we're at minimum peer count.
	minPeers := defaultMinSyncPeers
	if cs.forced {
		minPeers = 1
	} else if minPeers > cs.pm.maxPeers {
		minPeers = cs.pm.maxPeers
	}
	if cs.pm.peers.Len() < minPeers {
		return nil
	}

	// We have enough peers, check TD.
	peer := cs.pm.peers.BestPeer()
	if peer == nil {
		return nil
	}
	mode, ourTD := cs.modeAndLocalHead()
	op := peerToSyncOp(mode, peer)
	if op.td.Cmp(ourTD) <= 0 {
		return nil // We're in sync.
	}
	return op
}

func peerToSyncOp(mode downloader.SyncMode, p *peer) *chainSyncOp {
	peerHead, peerTD := p.Head()
	return &chainSyncOp{mode: mode, peer: p, td: peerTD, head: peerHead}
}

func (cs *chainSyncer) modeAndLocalHead() (downloader.SyncMode, *big.Int) {
	// If we're in fast sync mode, return that directly
	if atomic.LoadUint32(&cs.pm.fastSync) == 1 {
		block := cs.pm.blockchain.CurrentFastBlock()
		td := cs.pm.blockchain.GetTdByHash(block.Hash())
		return downloader.FastSync, td
	}
	// We are probably in full sync, but we might have rewound to before the
	// fast sync pivot, check if we should reenable
	if pivot := rawdb.ReadLastPivotNumber(cs.pm.chaindb); pivot != nil {
		if head := cs.pm.blockchain.CurrentBlock(); head.NumberU64() < *pivot {
			block := cs.pm.blockchain.CurrentFastBlock()
			td := cs.pm.blockchain.GetTdByHash(block.Hash())
			return downloader.FastSync, td
		}
	}
	// Nope, we're really full syncing
	head := cs.pm.blockchain.CurrentHeader()
	td := cs.pm.blockchain.GetTd(head.Hash(), head.Number.Uint64())
	return downloader.FullSync, td
}

// startSync launches doSync in a new goroutine.
func (cs *chainSyncer) startSync(op *chainSyncOp) {
	cs.doneCh = make(chan error, 1)
	go func() { cs.doneCh <- cs.pm.doSync(op) }()
}

// newChainSyncer creates a chainSyncer.
func newChainSyncer(pm *ProtocolManager) *chainSyncer {
	return &chainSyncer{
		pm:          pm,
		peerEventCh: make(chan struct{}),
	}
}