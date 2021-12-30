package ethereum

import (
	"math/big"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

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
