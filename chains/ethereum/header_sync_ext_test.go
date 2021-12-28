package ethereum

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"math/big"
	"testing"
)

func TestHeaderStore_Decode(t *testing.T) {
	var tests = []struct {
		name    string
		hs      *HeaderSync
		wantErr bool
	}{
		{
			name: "",
			hs: &HeaderSync{
				epoch2reward: map[uint64]*big.Int{
					1: big.NewInt(100),
					2: big.NewInt(200),
					3: big.NewInt(200),
				},
				height2receiveTimes: map[uint64]uint64{
					101: 1,
					202: 2,
					303: 3,
					404: 4,
					505: 5,
				},
				epoch2syncInfo: map[uint64][]*RelayerSyncInfo{
					1: {
						{
							Relayer: common.Address{},
							Times:   0,
							Reward:  nil,
						},
					},
					2: {
						{
							Relayer: common.Address{},
							Times:   0,
							Reward:  nil,
						},
					},
					3: {
						{
							Relayer: common.Address{},
							Times:   0,
							Reward:  nil,
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for i := 0; i < 100; i++ {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				bs, err := rlp.EncodeToBytes(&tt.hs)
				if err != nil {
					t.Error(err)
				}

				hs2 := HeaderSync{}
				if err := rlp.DecodeBytes(bs, &hs2); err != nil {
					t.Error(err)
				}
				//fmt.Printf("============================== hs2: %+v\n", hs2)
			})
		}
	}
}
