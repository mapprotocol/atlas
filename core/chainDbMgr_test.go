package core

import (
	"fmt"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	eth_rawdb "github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/core/chain/eth"
	"github.com/mapprotocol/atlas/core/rawdb"
	"math/big"
	"testing"
	"time"
)

// This test checks status reporting of InsertHeaderChain.
func TestHeaderInsertion(t *testing.T) {

	chainDb0, _ := OpenDatabase("data2", 20, 20)

	db := HeaderChainStore{
		chainDb: chainDb0,
	}
	StoreMgr = &db

	var (
		db001   = eth_rawdb.NewMemoryDatabase()
		genesis = (&core.Genesis{}).MustCommit(db001)
	)

	rawdb.WriteTd(chainDb0, genesis.Hash(), genesis.NumberU64(), genesis.Difficulty(), rawdb.ChainType(123))
	rawdb.WriteReceipts(chainDb0, genesis.Hash(), genesis.NumberU64(), nil, rawdb.ChainType(123))
	rawdb.WriteCanonicalHash(chainDb0, genesis.Hash(), genesis.NumberU64(), rawdb.ChainType(123))
	rawdb.WriteHeadBlockHash(chainDb0, genesis.Hash(), rawdb.ChainType(123))
	rawdb.WriteHeadFastBlockHash(chainDb0, genesis.Hash(), rawdb.ChainType(123))
	rawdb.WriteHeadHeaderHash(chainDb0, genesis.Hash(), rawdb.ChainType(123))
	rawdb.WriteChainConfig(chainDb0, genesis.Hash(), (&core.Genesis{}).Config, rawdb.ChainType(123))

	hc, _ := GetStoreMgr(rawdb.ChainType(123))
	// chain A: G->A1->A2...A128
	chainA := makeHeaderChain(genesis.Header(), 128, ethash.NewFaker(), db001, 10)
	headers := make([]eth.Header, len(chainA))
	for i, block := range chainA {
		headers[i].ParentHash = block.ParentHash
		headers[i].UncleHash = block.UncleHash
		headers[i].Coinbase = block.Coinbase
		headers[i].Root = block.Root
		headers[i].TxHash = block.TxHash
		headers[i].Bloom = block.Bloom
		headers[i].Difficulty = block.Difficulty
		headers[i].Number = block.Number
		headers[i].GasLimit = block.GasLimit
		headers[i].GasUsed = block.GasUsed
		headers[i].Time = block.Time
		headers[i].Extra = block.Extra
		headers[i].MixDigest = block.MixDigest
		headers[i].Nonce = block.Nonce
	}
	headers01 := make([]*eth.Header, len(chainA))
	l := len(chainA)
	for i := 0; i < l; i++ {
		headers01[i] = &headers[i]
	}
	//// chain B: G->A1->B2...B128
	//makeHeaderChain(chainA[0], 128, ethash.NewFaker(), db, 10)
	log.Root().SetHandler(log.StdoutHandler)
	// Inserting 64 headers on an empty chain, expecting
	// 1 callbacks, 1 canon-status, 0 sidestatus,
	testInsert(t, hc, headers01[:4], CanonStatTy, nil)
	// Inserting 64 headers on an empty chain, expecting
	// 1 callbacks, 1 canon-status, 0 sidestatus,
	testInsert(t, hc, headers01[:64], NonStatTy, nil)

	// Inserting the same some old, some new headers
	// 1 callbacks, 1 canon, 0 side
	//testInsert(t, hc, headers01[2:5], CanonStatTy, nil)
}

func testInsert(t *testing.T, hc *HeaderChainStore, chain []*eth.Header, wantStatus WriteStatus, wantErr error) {
	t.Helper()
	status, _ := hc.InsertHeaderChain(chain, time.Now())
	if status != wantStatus {
		t.Errorf("wrong write status from InsertHeaderChain: got %v, want %v", status, wantStatus)
	}

}
func Test3(t *testing.T) {
	chainDb0, _ := OpenDatabase("data9", 20, 20)

	db := HeaderChainStore{
		chainDb: chainDb0,
	}
	StoreMgr = &db

	var (
		db001   = eth_rawdb.NewMemoryDatabase()
		genesis = (&core.Genesis{}).MustCommit(db001)
	)

	hc, _ := GetStoreMgr(rawdb.ChainType(123))
	batch := hc.chainDb.NewBatch()
	chainA := makeHeaderChain(genesis.Header(), 128, ethash.NewFaker(), db001, 10)
	headers := make([]eth.Header, len(chainA))
	for i, block := range chainA {
		headers[i].ParentHash = block.ParentHash
		headers[i].UncleHash = block.UncleHash
		headers[i].Coinbase = block.Coinbase
		headers[i].Root = block.Root
		headers[i].TxHash = block.TxHash
		headers[i].Bloom = block.Bloom
		headers[i].Difficulty = block.Difficulty
		headers[i].Number = block.Number
		headers[i].GasLimit = block.GasLimit
		headers[i].GasUsed = block.GasUsed
		headers[i].Time = block.Time
		headers[i].Extra = block.Extra
		headers[i].MixDigest = block.MixDigest
		headers[i].Nonce = block.Nonce
	}
	headers01 := make([]*eth.Header, len(chainA))
	l := len(chainA)
	for i := 0; i < l; i++ {
		headers01[i] = &headers[i]
	}
	number := new(big.Int).SetUint64(uint64(100))
	headers01[0].Number = number
	rawdb.WriteHeader(batch, headers01[3], hc.currentChainType)
	fmt.Println(headers01[3].Number.Uint64())
	batch.Write()
	hc.WriteCurrentHeaderHash(headers01[3].Hash())
	sss := hc.CurrentHeaderNumber()
	fmt.Println(sss)
	//sss := hc.HasHeader(headers01[0].Hash(), number.Uint64())
	//fmt.Println(sss)
}
