// Copyright 2018 The go-ethereum Authors
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

// Package rawdb contains a collection of low level database accessors.
package rawdb

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/metrics"
)

// The fields below define the low level database schema prefixing.
var (
	// headHeaderKey tracks the latest know header's hash.
	headHeaderKey = []byte("LastHeader")

	// headBlockKey tracks the latest know full block's hash.
	headBlockKey = []byte("LastBlock")

	headRewardKey = []byte("LastReward")

	lastBlockKey = []byte("LastBlockIndex")

	// headFastBlockKey tracks the latest known incomplete block's hash duirng fast sync.
	headFastBlockKey = []byte("LastFast")

	// fastTrieProgressKey tracks the number of trie entries imported during fast sync.
	fastTrieProgressKey = []byte("TrieSync")

	// stateGcBodyReceiptKey tracks the number of body and receipt entries delete during state sync.
	stateGcBodyReceiptKey = []byte("LastState")

	// Data item prefixes (use single byte to avoid mixing data types, avoid `i`, used for indexes).
	headerPrefix       = []byte("h") // headerPrefix + ChainType + num (uint64 big endian) + hash -> header
	headerTDSuffix     = []byte("t") // headerPrefix + ChainType + num (uint64 big endian) + hash + headerTDSuffix -> td
	headerHashSuffix   = []byte("n") // headerPrefix + ChainType + num (uint64 big endian) + headerHashSuffix -> hash
	headerNumberPrefix = []byte("H") // headerNumberPrefix + hash -> num (uint64 big endian)

	blockBodyPrefix     = []byte("b") // blockBodyPrefix + ChainType + num (uint64 big endian) + hash -> block body
	blockReceiptsPrefix = []byte("r") // blockReceiptsPrefix + ChainType + num (uint64 big endian) + hash -> block receipts

	txLookupPrefix  = []byte("l") // txLookupPrefix + hash -> transaction/receipt lookup metadata
	bloomBitsPrefix = []byte("B") // bloomBitsPrefix + bit (uint16 big endian) + section (uint64 big endian) + hash -> bloom bits

	preimagePrefix = []byte("secure-key-")        // preimagePrefix + hash -> preimage
	configPrefix   = []byte("atlaschain-config-") // config prefix for the db

	// Chain index prefixes (use `i` + single byte to avoid mixing data types).
	BloomBitsIndexPrefix = []byte("iB") // BloomBitsIndexPrefix is the data table of a chain indexer to track its progress

	preimageCounter    = metrics.NewRegisteredCounter("db/preimage/total", nil)
	preimageHitCounter = metrics.NewRegisteredCounter("db/preimage/hits", nil)
)

//--------------- mark ---------
type ChainType uint64

func (t ChainType) toByte() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(t))
	return b
}
func (t ChainType) len() int {
	return 8
}

// change the visit Number by append chain mark   mark + number
func (t ChainType) setTypeNumber(n uint64) []byte {
	return append(t.toByte(), encodeBlockNumber(n)...)
}

//change the visit key by append chain mark      key  + mark
func (t ChainType) setTypeKey(b []byte) []byte {
	return append(b, t.toByte()...)
}

//  key  + mark + number
func (t ChainType) setTypeKeyNum(b []byte, n uint64) []byte {
	return append(append(b, t.toByte()...), encodeBlockNumber(n)...)
}

//--------------- mark ---------
// encodeBlockNumber encodes a block number as big endian uint64
func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

// headerKey = headerPrefix + num (uint64 big endian) + hash
func headerKey(t ChainType, number uint64, hash common.Hash) []byte {
	return append(append(headerPrefix, t.setTypeNumber(number)...), hash.Bytes()...)

}

// headerTDKey = headerPrefix + num (uint64 big endian) + hash + headerTDSuffix
func headerTDKey(t ChainType, number uint64, hash common.Hash) []byte {
	return append(headerKey(t, number, hash), headerTDSuffix...)
}

// headerHashKey = headerPrefix + num (uint64 big endian) + headerHashSuffix
func headerHashKey(t ChainType, number uint64) []byte {
	return append(append(headerPrefix, t.setTypeNumber(number)...), headerHashSuffix...)
}

// headerNumberKey = headerNumberPrefix + hash
func headerNumberKey(t ChainType, hash common.Hash) []byte {
	return append(t.setTypeKey(headerNumberPrefix), hash.Bytes()...)
}

// blockBodyKey = blockBodyPrefix + num (uint64 big endian) + hash
func blockBodyKey(t ChainType, number uint64, hash common.Hash) []byte {
	return append(append(blockBodyPrefix, t.setTypeNumber(number)...), hash.Bytes()...)
}

// blockReceiptsKey = blockReceiptsPrefix + num (uint64 big endian) + hash
func blockReceiptsKey(t ChainType, number uint64, hash common.Hash) []byte {
	return append(append(blockReceiptsPrefix, t.setTypeNumber(number)...), hash.Bytes()...)
}
