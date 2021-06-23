package core

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/atlasdb"
	"github.com/mapprotocol/atlas/core/rawdb"
)

var err = errors.New("no db")

type StoreDb struct {
	Db atlasdb.Database

	currentMark rawdb.Mark
}

func OpenDatabase(file string, cache, handles int) (atlasdb.Database, error) {
	return atlasdb.NewLDBDatabase(file, cache, handles)
}
func NewStoreDb(chainDb atlasdb.Database, m rawdb.Mark, file string, cache, handles int) (*StoreDb, error) {
	if chainDb == nil {
		chainDb1, err := OpenDatabase(file, cache, handles)
		if err != nil {
			return nil, err
		}
		chainDb = chainDb1
	}
	db := &StoreDb{
		Db:          chainDb,
		currentMark: m,
	}
	return db, nil
}

func (db *StoreDb) SetMark(m rawdb.Mark) {
	db.currentMark = m
}

func (db *StoreDb) ReadHeader(Hash common.Hash, number uint64) *types.Header {
	return rawdb.ReadHeader(db.Db, Hash, number, db.currentMark)
}

func (db *StoreDb) WriteHeader(header *types.Header) {
	batch := db.Db.NewBatch()
	// Flush all accumulated deletions.
	if err := batch.Write(); err != nil {
		log.Crit("Failed to rewind block", "error", err)
	}
	rawdb.WriteHeader(db.Db, header, db.currentMark)
}
func (db *StoreDb) DeleteHeader(hash common.Hash, number uint64) {
	rawdb.DeleteHeader(db.Db, hash, number, db.currentMark)
}
