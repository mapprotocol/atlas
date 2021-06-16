package synchr

import "github.com/ethereum/go-ethereum/common"
import "github.com/mapprotocol/atlas/core/types"

type HeaderCacheOne struct {
	Header    *types.Header
	Forwarded bool
}

type HeaderCache struct {
	headers           map[string]*HeaderCacheOne
	hashesByHeight    map[uint64][]string
	maxHeight         uint64
	minHeight         uint64
	numHeightsToTrack uint64
}

func NewHeaderCache(numHeightsToTrack uint64) *HeaderCache {
	return &HeaderCache{
		headers:           make(map[string]*HeaderCacheOne, numHeightsToTrack),
		hashesByHeight:    make(map[uint64][]string, numHeightsToTrack),
		maxHeight:         0,
		minHeight:         ^uint64(0),
		numHeightsToTrack: numHeightsToTrack,
	}
}

func (hc *HeaderCache) Add(header *types.Header) bool {
	hash := header.Hash().Hex()
	_, exists := hc.headers[hash]
	if exists {
		return true
	}

	height := header.Number.Uint64()
	if hc.maxHeight >= hc.numHeightsToTrack && height <= hc.maxHeight-hc.numHeightsToTrack {
		return false
	}

	hc.headers[hash] = &HeaderCacheOne{
		Header:    header,
		Forwarded: false,
	}

	hashesAtHeight, heightExists := hc.hashesByHeight[height]
	if heightExists {
		hc.hashesByHeight[height] = append(hashesAtHeight, hash)
	} else {
		hc.hashesByHeight[height] = []string{hash}
		if height < hc.minHeight {
			hc.minHeight = height
		} else if height > hc.maxHeight {
			hc.maxHeight = height
		}
	}

	if hc.maxHeight >= hc.numHeightsToTrack {
		hc.removeTo(hc.maxHeight - hc.numHeightsToTrack + 1)
	}
	return true
}

func (hc *HeaderCache) removeTo(minHeightToKeep uint64) {
	for hc.minHeight < minHeightToKeep {
		hashesToRemove := hc.hashesByHeight[hc.minHeight]
		delete(hc.hashesByHeight, hc.minHeight)
		hc.minHeight++
		for _, hashToRemove := range hashesToRemove {
			delete(hc.headers, hashToRemove)
		}
	}
}

func (hc *HeaderCache) Get(hash common.Hash) (*HeaderCacheOne, bool) {
	hashHex := hash.Hex()
	item, exists := hc.headers[hashHex]
	if exists {
		return item, true
	}
	return nil, false
}
