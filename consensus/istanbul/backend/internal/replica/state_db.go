// Copyright 2021 MAP Protocol Authors.
// This file is part of MAP Protocol.

// MAP Protocol is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// MAP Protocol is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with MAP Protocol.  If not, see <http://www.gnu.org/licenses/>.

package replica

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/consensus/istanbul/backend/internal/db"
)

// Keys in the node database.
const (
	replicaStateDBVersion = 1
	replicaStateKey       = "replicaState" // Info about start/stop state

)

// ReplicaStateDB represents a Map that can be accessed either
// by address or enode
type ReplicaStateDB struct {
	gdb    *db.GenericDB
	lock   sync.RWMutex
	logger log.Logger
}

// OpenReplicaStateDB opens a validator enode database for storing and retrieving infos about validator
// enodes. If no path is given an in-memory, temporary database is constructed.
func OpenReplicaStateDB(path string) (*ReplicaStateDB, error) {
	logger := log.New("db", "ReplicaStateDB")

	gdb, err := db.New(int64(replicaStateDBVersion), path, logger, &opt.WriteOptions{NoWriteMerge: true})
	if err != nil {
		logger.Error("Error creating db", "err", err)
		return nil, err
	}

	return &ReplicaStateDB{
		gdb:    gdb,
		logger: logger,
	}, nil
}

// Close flushes and closes the database files.
func (rsdb *ReplicaStateDB) Close() error {
	rsdb.lock.Lock()
	defer rsdb.lock.Unlock()
	return rsdb.gdb.Close()
}

func (rsdb *ReplicaStateDB) GetReplicaState() (*replicaStateImpl, error) {
	rsdb.lock.Lock()
	defer rsdb.lock.Unlock()

	rawEntry, err := rsdb.gdb.Get([]byte(replicaStateKey))
	if err != nil {
		return nil, err
	}

	var entry replicaStateImpl
	if err = rlp.DecodeBytes(rawEntry, &entry); err != nil {
		return nil, err
	}
	return &entry, err
}

// StoreReplicaState will store the latest replica state
func (rsdb *ReplicaStateDB) StoreReplicaState(rs State) error {
	rsdb.lock.Lock()
	defer rsdb.lock.Unlock()
	logger := rsdb.logger.New("func", "StoreReplicaState")

	entryBytes, err := rlp.EncodeToBytes(rs)
	if err != nil {
		logger.Error("Failed to save roundState", "reason", "rlp encoding", "err", err)
		return err
	}

	batch := new(leveldb.Batch)
	batch.Put([]byte(replicaStateKey), entryBytes)
	err = rsdb.gdb.Write(batch)
	if err != nil {
		logger.Error("Failed to save roundState", "reason", "levelDB write", "err", err)
	}

	return err
}
