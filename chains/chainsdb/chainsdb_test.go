package chainsdb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	eth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mapprotocol/atlas/chains/headers/ethereum"
	"github.com/mapprotocol/atlas/core/rawdb"
)

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

func TestRead_chaintype_config(t *testing.T) {
	data, err := ioutil.ReadFile(fmt.Sprintf("config/chaintype_config.json"))
	if err != nil {
		log.Error("readFile Err", err)
	}
	var config []struct {
		Name string
		Id   rawdb.ChainType
	}
	_ = json.Unmarshal(data, &config)
	fmt.Println(config)
}
func TestRead_ethconfig(t *testing.T) {
	data, err := ioutil.ReadFile(fmt.Sprintf("config/%v_config.json", "eth"))
	if err != nil {
		log.Error("read eht store config err", err)
	}
	genesis := &ethereum.Header{}
	err = json.Unmarshal(data, genesis)
	if err != nil {
		log.Error("Unmarshal Err", err.Error())
	}
	fmt.Println(genesis.Hash())
}
