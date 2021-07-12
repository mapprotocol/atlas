package vm

import (
	"github.com/ethereum/go-ethereum/rlp"
	"math/big"
	"testing"
)

func TestHeaderStore_Decode(t *testing.T) {
	tests := []struct {
		name    string
		hs      *HeaderStore
		wantErr bool
	}{
		{
			name: "",
			hs: &HeaderStore{
				epoch2reward: map[uint64]*big.Int{
					1: big.NewInt(111),
					2: big.NewInt(222),
					3: big.NewInt(333),
				},
				height2receiveTimes: map[uint64]uint64{
					101: 1,
					202: 2,
					303: 3,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs, err := rlp.EncodeToBytes(&tt.hs)
			if err != nil {
				t.Error(err)
			}

			hs2 := HeaderStore{}
			if err := rlp.DecodeBytes(bs, &hs2); err != nil {
				t.Error(err)
			}
			t.Logf("hs2: %+v\n", hs2)
		})
	}
}
