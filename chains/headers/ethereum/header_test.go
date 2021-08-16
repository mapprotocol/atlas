package ethereum

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/mapprotocol/atlas/params"
	"math/big"
	"reflect"
	"testing"
)

func TestHeader_Genesis(t *testing.T) {
	type fields struct {
		ParentHash  common.Hash
		UncleHash   common.Hash
		Coinbase    common.Address
		Root        common.Hash
		TxHash      common.Hash
		ReceiptHash common.Hash
		Bloom       types.Bloom
		Difficulty  *big.Int
		Number      *big.Int
		GasLimit    uint64
		GasUsed     uint64
		Time        uint64
		Extra       []byte
		MixDigest   common.Hash
		Nonce       types.BlockNonce
	}
	type args struct {
		chainID uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Header
	}{
		{
			name:   "",
			fields: fields{},
			args:   args{chainID: params.MainNetChainID},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := &Header{
				ParentHash:  tt.fields.ParentHash,
				UncleHash:   tt.fields.UncleHash,
				Coinbase:    tt.fields.Coinbase,
				Root:        tt.fields.Root,
				TxHash:      tt.fields.TxHash,
				ReceiptHash: tt.fields.ReceiptHash,
				Bloom:       tt.fields.Bloom,
				Difficulty:  tt.fields.Difficulty,
				Number:      tt.fields.Number,
				GasLimit:    tt.fields.GasLimit,
				GasUsed:     tt.fields.GasUsed,
				Time:        tt.fields.Time,
				Extra:       tt.fields.Extra,
				MixDigest:   tt.fields.MixDigest,
				Nonce:       tt.fields.Nonce,
			}
			if got := eh.Genesis(tt.args.chainID); !reflect.DeepEqual(got, tt.want) {

			}
		})
	}
}

func Test_configGenesis(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{"eth"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configGenesis(tt.args.name)
		})
	}
}
