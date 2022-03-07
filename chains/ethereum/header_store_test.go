package ethereum

import (
	"errors"
	"math/big"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/tools"
)

func init() {
	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.LvlInfo)
	log.Root().SetHandler(glogger)
}

func TestHeaderStore_delOldHeaders(t *testing.T) {
	type fields struct {
		headers map[string][]byte
		tds     map[string]*big.Int
	}
	tests := []struct {
		name       string
		fields     fields
		fn         func(hs *HeaderStore)
		wantLength int
	}{
		{
			name: "",
			fields: fields{
				headers: make(map[string][]byte),
				tds:     make(map[string]*big.Int),
			},
			fn: func(hs *HeaderStore) {
				for i := 1; i <= 10; i++ {
					key1 := headerKey(uint64(i), common.BigToHash(big.NewInt(int64(i*23))))
					key2 := headerKey(uint64(i), common.BigToHash(big.NewInt(int64(i*654))))
					header := encodeHeader(&Header{
						Number: big.NewInt(int64(i)),
					})
					hs.Headers[key1] = header
					hs.Headers[key2] = header
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := &HeaderStore{
				Headers: tt.fields.headers,
				TDs:     tt.fields.tds,
			}

			tt.fn(hs)
			hs.delOldHeaders()
			if len(hs.Headers) > MaxHeaderLimit {
				t.Errorf("delOldHeaders() failed, want length: %d, got length: %d", tt.wantLength, len(hs.Headers))
			}

			t.Log("header length: ", len(hs.Headers))
			for k := range hs.Headers {
				t.Log("key-2: ", k)
			}
		})
	}
}

func TestHeaderStore_Store(t *testing.T) {
	type fields struct {
		canonicalNumberToHash map[uint64]common.Hash
		headers               map[string][]byte
		tds                   map[string]*big.Int
		curNumber             uint64
		curHash               common.Hash
	}
	type args struct {
		state types.StateDB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "store-header",
			fields: fields{
				canonicalNumberToHash: map[uint64]common.Hash{
					1: common.HexToHash("0x0c769c23a93066eaf5e4c83976b44e4b51e954cae0afcf32c8cd0fb89950e76f"),
				},
				headers: map[string][]byte{
					headerKey(1, common.HexToHash("0x0c769c23a93066eaf5e4c83976b44e4b51e954cae0afcf32c8cd0fb89950e76f")): encodeHeader(&Header{
						ParentHash:  common.Hash{},
						UncleHash:   common.Hash{},
						Coinbase:    common.Address{},
						Root:        common.Hash{},
						TxHash:      common.Hash{},
						ReceiptHash: common.Hash{},
						Bloom:       ethtypes.Bloom{},
						Difficulty:  big.NewInt(2),
						Number:      big.NewInt(1),
						GasLimit:    556790,
						GasUsed:     55800,
						Time:        uint64(time.Now().Unix()),
						Extra:       nil,
						MixDigest:   common.Hash{},
						Nonce:       ethtypes.BlockNonce{},
						BaseFee:     nil,
					}),
				},
				tds: map[string]*big.Int{
					headerKey(1, common.HexToHash("0x0c769c23a93066eaf5e4c83976b44e4b51e954cae0afcf32c8cd0fb89950e76f")): big.NewInt(2),
				},
				curNumber: 1,
				curHash:   common.HexToHash("0x0c769c23a93066eaf5e4c83976b44e4b51e954cae0afcf32c8cd0fb89950e76f"),
			},
			args: args{
				state: getStateDB(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := &HeaderStore{
				CanonicalNumberToHash: tt.fields.canonicalNumberToHash,
				Headers:               tt.fields.headers,
				TDs:                   tt.fields.tds,
				CurNumber:             tt.fields.curNumber,
				CurHash:               tt.fields.curHash,
			}
			if err := hs.Store(tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
			}
			h := NewHeaderStore()
			if err := h.Load(tt.args.state); err != nil {
				t.Errorf("Load() error = %v", err)
			}
			if !reflect.DeepEqual(hs, h) {
				t.Error("not equal")
			}

			t.Log("hs: ", hs)
			t.Log("load hs: ", h)
		})
	}
}

func TestHeaderStore_Load(t *testing.T) {
	type fields struct {
		CanonicalNumberToHash map[uint64]common.Hash
		Headers               map[string][]byte
		TDs                   map[string]*big.Int
		CurNumber             uint64
		CurHash               common.Hash
	}
	type args struct {
		state types.StateDB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		before  func(hs *HeaderStore, state types.StateDB)
		wantErr bool
	}{
		{
			name: "load-header-from-cache",
			fields: fields{
				CanonicalNumberToHash: map[uint64]common.Hash{
					1: common.HexToHash("0x0c769c23a93066eaf5e4c83976b44e4b51e954cae0afcf32c8cd0fb89950e76f"),
				},
				Headers: map[string][]byte{
					headerKey(1, common.HexToHash("0x0c769c23a93066eaf5e4c83976b44e4b51e954cae0afcf32c8cd0fb89950e76f")): encodeHeader(&Header{
						ParentHash:  common.Hash{},
						UncleHash:   common.Hash{},
						Coinbase:    common.Address{},
						Root:        common.Hash{},
						TxHash:      common.Hash{},
						ReceiptHash: common.Hash{},
						Bloom:       ethtypes.Bloom{},
						Difficulty:  big.NewInt(2),
						Number:      big.NewInt(1),
						GasLimit:    556790,
						GasUsed:     55800,
						Time:        uint64(time.Now().Unix()),
						Extra:       nil,
						MixDigest:   common.Hash{},
						Nonce:       ethtypes.BlockNonce{},
						BaseFee:     nil,
					}),
				},
				TDs: map[string]*big.Int{
					headerKey(1, common.HexToHash("0x0c769c23a93066eaf5e4c83976b44e4b51e954cae0afcf32c8cd0fb89950e76f")): big.NewInt(2),
				},
				CurNumber: 1,
				CurHash:   common.HexToHash("0x0c769c23a93066eaf5e4c83976b44e4b51e954cae0afcf32c8cd0fb89950e76f"),
			},
			args: args{
				state: getStateDB(),
			},
			before: func(hs *HeaderStore, state types.StateDB) {
				_ = hs.Store(state)
			},
			wantErr: false,
		},
		{
			name: "load-header-from-decode",
			fields: fields{
				CanonicalNumberToHash: map[uint64]common.Hash{
					1: common.HexToHash("0x0c769c23a93066eaf5e4c83976b44e4b51e954cae0afcf32c8cd0fb89950e76f"),
				},
				Headers: map[string][]byte{
					headerKey(1, common.HexToHash("0x0c769c23a93066eaf5e4c83976b44e4b51e954cae0afcf32c8cd0fb89950e76f")): encodeHeader(&Header{
						ParentHash:  common.Hash{},
						UncleHash:   common.Hash{},
						Coinbase:    common.Address{},
						Root:        common.Hash{},
						TxHash:      common.Hash{},
						ReceiptHash: common.Hash{},
						Bloom:       ethtypes.Bloom{},
						Difficulty:  big.NewInt(2),
						Number:      big.NewInt(1),
						GasLimit:    556790,
						GasUsed:     55800,
						Time:        uint64(time.Now().Unix()),
						Extra:       nil,
						MixDigest:   common.Hash{},
						Nonce:       ethtypes.BlockNonce{},
						BaseFee:     nil,
					}),
				},
				TDs: map[string]*big.Int{
					headerKey(1, common.HexToHash("0x0c769c23a93066eaf5e4c83976b44e4b51e954cae0afcf32c8cd0fb89950e76f")): big.NewInt(2),
				},
				CurNumber: 1,
				CurHash:   common.HexToHash("0x0c769c23a93066eaf5e4c83976b44e4b51e954cae0afcf32c8cd0fb89950e76f"),
			},
			args: args{
				state: getStateDB(),
			},
			before: func(hs *HeaderStore, state types.StateDB) {
				_ = hs.Store(state)
				key := common.BytesToHash(chains.EthereumHeaderStoreAddress[:])
				data := state.GetPOWState(chains.EthereumHeaderStoreAddress, key)
				hash := tools.RlpHash(data)
				storeCache.Cache.Remove(hash)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := &HeaderStore{
				CanonicalNumberToHash: tt.fields.CanonicalNumberToHash,
				Headers:               tt.fields.Headers,
				TDs:                   tt.fields.TDs,
				CurNumber:             tt.fields.CurNumber,
				CurHash:               tt.fields.CurHash,
			}
			tt.before(hs, tt.args.state)

			h := NewHeaderStore()
			if err := h.Load(tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(hs, h) {
				t.Error("not equal")
			}
			t.Log("hs: ", hs)
			t.Log("load hs: ", h)
		})
	}
}

func testInsert(t *testing.T, db types.StateDB, hs *HeaderStore, chain []byte, wantStatus WriteStatus, wantErr error) {
	t.Helper()

	res, err := hs.WriteHeaders(db, chain)
	if !errors.Is(err, wantErr) {
		t.Fatalf("unexpected error from WriteHeaders: got %v, want %v", err, wantErr)
	}
	if res.status != wantStatus {
		t.Errorf("wrong write status from WriteHeaders: got %v, want %v", res.status, wantStatus)
	}
}

func rlpEncode(headers []*ethtypes.Header) []byte {
	bs, err := rlp.EncodeToBytes(headers)
	if err != nil {
		panic(err)
	}
	return bs
}

func convertHeader(header *ethtypes.Header) *Header {
	bs, err := rlp.EncodeToBytes(header)
	if err != nil {
		panic(err)
	}
	var headers *Header
	if err := rlp.DecodeBytes(bs, &headers); err != nil {
		panic(err)
	}
	return headers
}

// This test checks status reporting of InsertHeader
func TestHeaderInsertion(t *testing.T) {
	var (
		db      = rawdb.NewMemoryDatabase()
		statedb = getStateDB()
		genesis = (&core.Genesis{BaseFee: big.NewInt(ethparams.InitialBaseFee)}).MustCommit(db)
	)

	hs := NewHeaderStore()
	hs.WriteCanonicalHash(genesis.Hash(), genesis.Number().Uint64())
	hs.WriteHeader(convertHeader(genesis.Header()))
	hs.WriteTd(genesis.Hash(), genesis.Number().Uint64(), big.NewInt(0))
	hs.CurHash = genesis.Hash()
	hs.CurNumber = genesis.Number().Uint64()
	if err := hs.Store(statedb); err != nil {
		t.Fatalf(err.Error())
	}
	// chain A: G->A1->A2...A128
	chainA := makeHeaderChain(genesis.Header(), 128, ethash.NewFaker(), db, 10)
	// chain B: G->A1->B2...B128

	chainB := makeHeaderChain(chainA[0], 128, ethash.NewFaker(), db, 10)
	log.Root().SetHandler(log.StdoutHandler)

	// Inserting 64 headers on an empty chain, expecting
	// 1 callbacks, 1 canon-status, 0 sidestatus,
	testInsert(t, statedb, hs, rlpEncode(chainA[:64]), CanonStatTy, nil)

	// Inserting 64 identical headers, expecting
	// 0 callbacks, 0 canon-status, 0 sidestatus,
	testInsert(t, statedb, hs, rlpEncode(chainA[:64]), NonStatTy, nil)

	// Inserting the same some old, some new headers
	// 1 callbacks, 1 canon, 0 side
	testInsert(t, statedb, hs, rlpEncode(chainA[32:96]), CanonStatTy, nil)

	// Inserting side blocks, but not overtaking the canon chain
	testInsert(t, statedb, hs, rlpEncode(chainB[0:32]), SideStatTy, nil)

	// Inserting more side blocks, but we don't have the parent
	testInsert(t, statedb, hs, rlpEncode(chainB[34:36]), NonStatTy, errUnknownAncestor)

	// Inserting more sideblocks, overtaking the canon chain
	testInsert(t, statedb, hs, rlpEncode(chainB[32:97]), CanonStatTy, nil)

	// Inserting more A-headers, taking back the canonicality
	testInsert(t, statedb, hs, rlpEncode(chainA[90:100]), CanonStatTy, nil)

	// And B becomes canon again
	testInsert(t, statedb, hs, rlpEncode(chainB[97:107]), CanonStatTy, nil)

	// And B becomes even longer
	testInsert(t, statedb, hs, rlpEncode(chainB[107:128]), CanonStatTy, nil)

	//var ns []int
	//for n := range hs.CanonicalNumberToHash {
	//	ns = append(ns, int(n))
	//}
	//sort.Ints(ns)
	//fmt.Println("============================== ns: ", ns)
}
