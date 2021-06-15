// Copyright 2019 The go-ethereum Authors
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

package atlasdb

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
)

// ReadSnapshotDisabled retrieves if the snapshot maintenance is disabled.
func ReadSnapshotDisabled(db ethdb.KeyValueReader, m mark) bool {
	disabled, _ := db.Has(markKey(snapshotDisabledKey, m))
	return disabled
}

// WriteSnapshotDisabled stores the snapshot pause flag.
func WriteSnapshotDisabled(db ethdb.KeyValueWriter, m mark) {
	if err := db.Put(markKey(snapshotDisabledKey, m), []byte("42")); err != nil {
		log.Crit("Failed to store snapshot disabled flag", "err", err)
	}
}

// DeleteSnapshotDisabled deletes the flag keeping the snapshot maintenance disabled.
func DeleteSnapshotDisabled(db ethdb.KeyValueWriter, m mark) {
	if err := db.Delete(markKey(snapshotDisabledKey, m)); err != nil {
		log.Crit("Failed to remove snapshot disabled flag", "err", err)
	}
}

// ReadSnapshotRoot retrieves the root of the block whose state is contained in
// the persisted snapshot.
func ReadSnapshotRoot(db ethdb.KeyValueReader, m mark) common.Hash {
	data, _ := db.Get(markKey(snapshotRootKey, m))
	if len(data) != common.HashLength {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteSnapshotRoot stores the root of the block whose state is contained in
// the persisted snapshot.
func WriteSnapshotRoot(db ethdb.KeyValueWriter, root common.Hash, m mark) {
	if err := db.Put(markKey(snapshotRootKey, m), root[:]); err != nil {
		log.Crit("Failed to store snapshot root", "err", err)
	}
}

// DeleteSnapshotRoot deletes the hash of the block whose state is contained in
// the persisted snapshot. Since snapshots are not immutable, this  method can
// be used during updates, so a crash or failure will mark the entire snapshot
// invalid.
func DeleteSnapshotRoot(db ethdb.KeyValueWriter, m mark) {
	if err := db.Delete(markKey(snapshotRootKey, m)); err != nil {
		log.Crit("Failed to remove snapshot root", "err", err)
	}
}

// ReadAccountSnapshot retrieves the snapshot entry of an account trie leaf.
func ReadAccountSnapshot(db ethdb.KeyValueReader, hash common.Hash) []byte {
	data, _ := db.Get(accountSnapshotKey(hash))
	return data
}

// WriteAccountSnapshot stores the snapshot entry of an account trie leaf.
func WriteAccountSnapshot(db ethdb.KeyValueWriter, hash common.Hash, entry []byte) {
	if err := db.Put(accountSnapshotKey(hash), entry); err != nil {
		log.Crit("Failed to store account snapshot", "err", err)
	}
}

// DeleteAccountSnapshot removes the snapshot entry of an account trie leaf.
func DeleteAccountSnapshot(db ethdb.KeyValueWriter, hash common.Hash) {
	if err := db.Delete(accountSnapshotKey(hash)); err != nil {
		log.Crit("Failed to delete account snapshot", "err", err)
	}
}

// ReadStorageSnapshot retrieves the snapshot entry of an storage trie leaf.
func ReadStorageSnapshot(db ethdb.KeyValueReader, accountHash, storageHash common.Hash) []byte {
	data, _ := db.Get(storageSnapshotKey(accountHash, storageHash))
	return data
}

// WriteStorageSnapshot stores the snapshot entry of an storage trie leaf.
func WriteStorageSnapshot(db ethdb.KeyValueWriter, accountHash, storageHash common.Hash, entry []byte) {
	if err := db.Put(storageSnapshotKey(accountHash, storageHash), entry); err != nil {
		log.Crit("Failed to store storage snapshot", "err", err)
	}
}

// DeleteStorageSnapshot removes the snapshot entry of an storage trie leaf.
func DeleteStorageSnapshot(db ethdb.KeyValueWriter, accountHash, storageHash common.Hash) {
	if err := db.Delete(storageSnapshotKey(accountHash, storageHash)); err != nil {
		log.Crit("Failed to delete storage snapshot", "err", err)
	}
}

// IterateStorageSnapshots returns an iterator for walking the entire storage
// space of a specific account.
func IterateStorageSnapshots(db ethdb.Iteratee, accountHash common.Hash) ethdb.Iterator {
	return db.NewIterator(storageSnapshotsKey(accountHash), nil)
}

// ReadSnapshotJournal retrieves the serialized in-memory diff layers saved at
// the last shutdown. The blob is expected to be max a few 10s of megabytes.
func ReadSnapshotJournal(db ethdb.KeyValueReader, m mark) []byte {
	data, _ := db.Get(markKey(snapshotJournalKey, m))
	return data
}

// WriteSnapshotJournal stores the serialized in-memory diff layers to save at
// shutdown. The blob is expected to be max a few 10s of megabytes.
func WriteSnapshotJournal(db ethdb.KeyValueWriter, journal []byte, m mark) {
	if err := db.Put(markKey(snapshotJournalKey, m), journal); err != nil {
		log.Crit("Failed to store snapshot journal", "err", err)
	}
}

// DeleteSnapshotJournal deletes the serialized in-memory diff layers saved at
// the last shutdown
func DeleteSnapshotJournal(db ethdb.KeyValueWriter, m mark) {
	if err := db.Delete(markKey(snapshotJournalKey, m)); err != nil {
		log.Crit("Failed to remove snapshot journal", "err", err)
	}
}

// ReadSnapshotGenerator retrieves the serialized snapshot generator saved at
// the last shutdown.
func ReadSnapshotGenerator(db ethdb.KeyValueReader, m mark) []byte {
	data, _ := db.Get(markKey(snapshotGeneratorKey, m))
	return data
}

// WriteSnapshotGenerator stores the serialized snapshot generator to save at
// shutdown.
func WriteSnapshotGenerator(db ethdb.KeyValueWriter, generator []byte, m mark) {
	if err := db.Put(markKey(snapshotGeneratorKey, m), generator); err != nil {
		log.Crit("Failed to store snapshot generator", "err", err)
	}
}

// DeleteSnapshotGenerator deletes the serialized snapshot generator saved at
// the last shutdown
func DeleteSnapshotGenerator(db ethdb.KeyValueWriter, m mark) {
	if err := db.Delete(markKey(snapshotGeneratorKey, m)); err != nil {
		log.Crit("Failed to remove snapshot generator", "err", err)
	}
}

// ReadSnapshotRecoveryNumber retrieves the block number of the last persisted
// snapshot layer.
func ReadSnapshotRecoveryNumber(db ethdb.KeyValueReader, m mark) *uint64 {
	data, _ := db.Get(markKey(snapshotRecoveryKey, m))
	if len(data) == 0 {
		return nil
	}
	if len(data) != 8 {
		return nil
	}
	number := binary.BigEndian.Uint64(data)
	return &number
}

// WriteSnapshotRecoveryNumber stores the block number of the last persisted
// snapshot layer.
func WriteSnapshotRecoveryNumber(db ethdb.KeyValueWriter, m mark, number uint64) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], number)
	if err := db.Put(markKey(snapshotRecoveryKey, m), buf[:]); err != nil {
		log.Crit("Failed to store snapshot recovery number", "err", err)
	}
}

// DeleteSnapshotRecoveryNumber deletes the block number of the last persisted
// snapshot layer.
func DeleteSnapshotRecoveryNumber(db ethdb.KeyValueWriter, m mark) {
	if err := db.Delete(markKey(snapshotRecoveryKey, m)); err != nil {
		log.Crit("Failed to remove snapshot recovery number", "err", err)
	}
}

// ReadSnapshotSyncStatus retrieves the serialized sync status saved at shutdown.
func ReadSnapshotSyncStatus(db ethdb.KeyValueReader, m mark) []byte {
	data, _ := db.Get(markKey(snapshotSyncStatusKey, m))
	return data
}

// WriteSnapshotSyncStatus stores the serialized sync status to save at shutdown.
func WriteSnapshotSyncStatus(db ethdb.KeyValueWriter, status []byte, m mark) {
	if err := db.Put(markKey(snapshotSyncStatusKey, m), status); err != nil {
		log.Crit("Failed to store snapshot sync status", "err", err)
	}
}

// DeleteSnapshotSyncStatus deletes the serialized sync status saved at the last
// shutdown
func DeleteSnapshotSyncStatus(db ethdb.KeyValueWriter, m mark) {
	if err := db.Delete(markKey(snapshotSyncStatusKey, m)); err != nil {
		log.Crit("Failed to remove snapshot sync status", "err", err)
	}
}
