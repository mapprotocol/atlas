package chainsdb

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	eth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/chains/ethereum"
	"github.com/mapprotocol/atlas/core/rawdb"
	"math/big"
	"reflect"
	"testing"
	"time"
)

// This test checks status reporting of InsertHeaderChain.
func TestHeaderInsertion01(t *testing.T) {

	chainDb0, _ := OpenDatabase("data222", 20, 20)

	db := HeaderChainStore{
		chainDb: chainDb0,
	}
	storeMgr = &db
	chainType := rawdb.ChainType(321)
	var (
		db001   = rawdb.NewMemoryDatabase()
		genesis = (&core.Genesis{Nonce: 111}).MustCommit(db001)
	)

	rawdb.WriteTdChains(chainDb0, genesis.Hash(), genesis.NumberU64(), genesis.Difficulty(), chainType)
	rawdb.WriteReceiptsChains(chainDb0, genesis.Hash(), genesis.NumberU64(), nil, chainType)
	rawdb.WriteCanonicalHashChains(chainDb0, genesis.Hash(), genesis.NumberU64(), chainType)
	rawdb.WriteHeadBlockHashChains(chainDb0, genesis.Hash(), chainType)
	rawdb.WriteHeadFastBlockHashChains(chainDb0, genesis.Hash(), chainType)
	rawdb.WriteHeadHeaderHashChains(chainDb0, genesis.Hash(), chainType)
	rawdb.WriteChainConfigChains(chainDb0, genesis.Hash(), (&core.Genesis{}).Config, chainType)

	hc, _ := GetStoreMgr(chainType)
	// chain A: G->A1->A2...A128
	chainA := makeHeaderChain(genesis.Header(), 128, ethash.NewFaker(), db001, 10)

	chainA001 := converChainList(chainA)

	// chain B: G->A1->B2...B128
	chainB := makeHeaderChain(chainA[0], 128, ethash.NewFaker(), db001, 10)
	chainB001 := converChainList(chainB)
	log.Root().SetHandler(log.StdoutHandler)

	// Inserting 64 headers on an empty chain, expecting
	// 1 callbacks, 1 canon-status, 0 sidestatus,
	testInsert01(t, hc, chainA001[:64], CanonStatTyState, nil)

	// Inserting 64 identical headers, expecting
	// 0 callbacks, 0 canon-status, 0 sidestatus,
	testInsert01(t, hc, chainA001[:64], NonStatTyState, nil)

	// Inserting the same some old, some new headers
	// 1 callbacks, 1 canon, 0 side
	testInsert01(t, hc, chainA001[32:96], CanonStatTyState, nil)

	// Inserting side blocks, but not overtaking the canon chain
	testInsert01(t, hc, chainB001[0:32], SideStatTyState, nil)

	// Inserting more side blocks, but we don't have the parent
	testInsert01(t, hc, chainB001[34:36], NonStatTyState, nil)

	// Inserting more sideblocks, overtaking the canon chain
	testInsert01(t, hc, chainB001[32:97], CanonStatTyState, nil)

	// Inserting more A-headers, taking back the canonicality
	testInsert01(t, hc, chainA001[90:100], CanonStatTyState, nil)

	// And B becomes canon again
	testInsert01(t, hc, chainB001[97:107], CanonStatTyState, nil)

	// And B becomes even longer
	testInsert01(t, hc, chainB001[107:128], CanonStatTyState, nil)
}

func testInsert01(t *testing.T, hc *HeaderChainStore, chain []*ethereum.Header, wantStatus WriteStatus, wantErr error) {
	t.Helper()
	status, _ := hc.InsertHeaderChain(chain, time.Now())
	if status != wantStatus {
		t.Errorf("wrong write status from InsertHeaderChain: got %v, want %v", status, wantStatus)
	}
}

func converChainList(headers []*eth_types.Header) (newChains1 []*ethereum.Header) {
	l := len(headers)
	newChains := make([]ethereum.Header, l)
	for i := 0; i < l; i++ {
		newChains1 = append(newChains1, convertChain(&newChains[i], headers[i]))
	}
	return
}

func convertChain(header *ethereum.Header, e *eth_types.Header) *ethereum.Header {
	header.ParentHash = e.ParentHash
	header.UncleHash = e.UncleHash
	header.Coinbase = e.Coinbase
	header.Root = e.Root
	header.TxHash = e.TxHash
	header.ReceiptHash = e.ReceiptHash
	header.GasLimit = e.GasLimit
	header.GasUsed = e.GasUsed
	header.Time = e.Time
	header.MixDigest = e.MixDigest
	header.Nonce = eth_types.EncodeNonce(e.Nonce.Uint64())
	header.Bloom.SetBytes(e.Bloom.Bytes())
	if header.Difficulty = new(big.Int); e.Difficulty != nil {
		header.Difficulty.Set(e.Difficulty)
	}
	if header.Number = new(big.Int); e.Number != nil {
		header.Number.Set(e.Number)
	}
	if len(e.Extra) > 0 {
		header.Extra = make([]byte, len(e.Extra))
		copy(header.Extra, e.Extra)
	}
	// test rlp
	//fmt.Println(e.Hash(), "/n", header.Hash())
	return header
}

func TestHeaderChainStore_CurrentHeaderHash(t *testing.T) {
	chainDb0, _ := OpenDatabase("data222", 20, 20)

	db := HeaderChainStore{
		chainDb: chainDb0,
	}
	storeMgr = &db
	hs, _ := GetStoreMgr(rawdb.ChainType(321))

	tests := []struct {
		name   string
		fields *HeaderChainStore
		want   common.Hash
	}{
		{
			"1",
			hs,
			common.Hash{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := &HeaderChainStore{
				chainDb:          tt.fields.chainDb,
				currentChainType: tt.fields.currentChainType,
				Mu:               tt.fields.Mu,
				rand:             tt.fields.rand,
			}
			currentHash := hc.CurrentHeaderHash()
			td := hc.GetTdByHash(currentHash)
			GetHeaderByHash := hc.GetHeaderByHash(currentHash)
			fmt.Printf("CurrentHeaderHash() = %v GetTdByHash() %v GetHeaderByHash() %v ", currentHash, td, GetHeaderByHash)
		})
	}
}

func TestHeaderChainStore_GetHeaderByNumber(t *testing.T) {
	chainDb0, _ := OpenDatabase("data222", 20, 20)

	db := HeaderChainStore{
		chainDb: chainDb0,
	}
	storeMgr = &db
	hs, _ := GetStoreMgr(rawdb.ChainType(123))

	type args struct {
		number uint64
	}
	tests := []struct {
		name   string
		fields *HeaderChainStore
		args   args
		want   common.Hash
	}{
		{
			"1",
			hs,
			args{uint64(1)},
			common.Hash{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := &HeaderChainStore{
				chainDb:          tt.fields.chainDb,
				currentChainType: tt.fields.currentChainType,
				Mu:               tt.fields.Mu,
				rand:             tt.fields.rand,
			}
			if got := hc.GetHeaderByNumber(tt.args.number); !reflect.DeepEqual(got, tt.want) {
				fmt.Printf("GetHeaderByNumber() = %v", got)
			}
		})
	}
}

func TestHeaderChainStore_CurrentHeaderNumber(t *testing.T) {
	chainDb0, _ := OpenDatabase("data222", 20, 20)

	db := HeaderChainStore{
		chainDb: chainDb0,
	}
	storeMgr = &db
	hs, _ := GetStoreMgr(rawdb.ChainType(321))

	tests := []struct {
		name   string
		fields *HeaderChainStore
	}{
		{
			"1",
			hs,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := &HeaderChainStore{
				chainDb:          tt.fields.chainDb,
				currentChainType: tt.fields.currentChainType,
				Mu:               tt.fields.Mu,
				rand:             tt.fields.rand,
			}
			got := hc.CurrentHeaderNumber()
			fmt.Printf("CurrentHeaderNumber() = %v", got)
		})
	}
}

func TestHeaderChainStore_ReadCanonicalHash(t *testing.T) {
	chainDb0, _ := OpenDatabase("data222", 20, 20)

	db := HeaderChainStore{
		chainDb: chainDb0,
	}
	storeMgr = &db
	hs, _ := GetStoreMgr(rawdb.ChainType(123))

	tests := []struct {
		name   string
		fields *HeaderChainStore
		number uint64
	}{
		{
			"1",
			hs,
			uint64(128),
		},
		{
			"1",
			hs,
			uint64(127),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := &HeaderChainStore{
				chainDb:          tt.fields.chainDb,
				currentChainType: tt.fields.currentChainType,
				Mu:               tt.fields.Mu,
				rand:             tt.fields.rand,
			}
			got := hc.ReadCanonicalHash(tt.number)
			fmt.Printf("ReadCanonicalHash() = %v", got)
		})
	}
}
func Test_thread_InsertHeaderChain(t *testing.T) {
	chainDb0, _ := OpenDatabase("data333", 20, 20)

	db := HeaderChainStore{
		chainDb: chainDb0,
	}
	storeMgr = &db
	chainType := rawdb.ChainType(333)
	var (
		db001   = rawdb.NewMemoryDatabase()
		genesis = (&core.Genesis{Nonce: 111}).MustCommit(db001)
	)

	rawdb.WriteTdChains(chainDb0, genesis.Hash(), genesis.NumberU64(), genesis.Difficulty(), chainType)
	rawdb.WriteReceiptsChains(chainDb0, genesis.Hash(), genesis.NumberU64(), nil, chainType)
	rawdb.WriteCanonicalHashChains(chainDb0, genesis.Hash(), genesis.NumberU64(), chainType)
	rawdb.WriteHeadBlockHashChains(chainDb0, genesis.Hash(), chainType)
	rawdb.WriteHeadFastBlockHashChains(chainDb0, genesis.Hash(), chainType)
	rawdb.WriteHeadHeaderHashChains(chainDb0, genesis.Hash(), chainType)
	rawdb.WriteChainConfigChains(chainDb0, genesis.Hash(), (&core.Genesis{}).Config, chainType)

	hc, _ := GetStoreMgr(chainType)
	// chain A: G->A1->A2...A128
	chainA := makeHeaderChain(genesis.Header(), 128, ethash.NewFaker(), db001, 10)

	chainA001 := converChainList(chainA)

	for i := 0; i < 1000; i++ {

		go func() {
			hc.InsertHeaderChain(chainA001, time.Now())
			fmt.Println("111")
		}()
	}
	time.Sleep(5 * time.Second)

}
func Test_thread_WriteHeader(t *testing.T) {
	chainDb0, _ := OpenDatabase("data333", 20, 20)

	db := HeaderChainStore{
		chainDb: chainDb0,
	}
	storeMgr = &db
	chainType := rawdb.ChainType(333)
	var (
		db001   = rawdb.NewMemoryDatabase()
		genesis = (&core.Genesis{Nonce: 111}).MustCommit(db001)
	)

	rawdb.WriteTdChains(chainDb0, genesis.Hash(), genesis.NumberU64(), genesis.Difficulty(), chainType)
	rawdb.WriteReceiptsChains(chainDb0, genesis.Hash(), genesis.NumberU64(), nil, chainType)
	rawdb.WriteCanonicalHashChains(chainDb0, genesis.Hash(), genesis.NumberU64(), chainType)
	rawdb.WriteHeadBlockHashChains(chainDb0, genesis.Hash(), chainType)
	rawdb.WriteHeadFastBlockHashChains(chainDb0, genesis.Hash(), chainType)
	rawdb.WriteHeadHeaderHashChains(chainDb0, genesis.Hash(), chainType)
	rawdb.WriteChainConfigChains(chainDb0, genesis.Hash(), (&core.Genesis{}).Config, chainType)

	hc, _ := GetStoreMgr(chainType)
	// chain A: G->A1->A2...A128
	chainA := makeHeaderChain(genesis.Header(), 128, ethash.NewFaker(), db001, 10)

	chainA001 := converChainList(chainA)

	for i := 0; i < 1000; i++ {

		go func() {
			hc.WriteHeader(chainA001[0])
			fmt.Println(hc.ReadHeader(chainA001[0].Hash(), 1))
		}()
	}
	time.Sleep(5 * time.Second)

}
func Test_GetStoreMgr(t *testing.T) {
	chainDb0, _ := OpenDatabase("data333", 20, 20)

	db := HeaderChainStore{
		chainDb: chainDb0,
	}
	storeMgr = &db
	for i := 0; i < 1000; i++ {

		go func() {
			GetStoreMgr(rawdb.ChainType(i))
		}()
	}
	time.Sleep(5 * time.Second)

}
