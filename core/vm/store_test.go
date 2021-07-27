package vm

import (
	"fmt"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/params"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func getHeaderStore() *HeaderStore {
	var hs = &HeaderStore{
		epoch2reward: map[uint64]*big.Int{
			1: big.NewInt(48976200),
		},
		height2receiveTimes: map[uint64]uint64{
			510980: 3,
		},
		epoch2syncInfo: map[uint64][]*RelayerSyncInfo{
			1: {
				{
					Relayer: common.HexToAddress("0x3f98da321de0a"),
					Times:   862,
				},
			},
		},
	}
	return hs
}

func modifyHeaderStore(hs *HeaderStore) *HeaderStore {
	hs.epoch2reward[1] = big.NewInt(1111111111111)
	hs.epoch2reward[2] = big.NewInt(2222222222222)
	hs.height2receiveTimes[510980] = 1
	hs.height2receiveTimes[8888888] = 2
	hs.epoch2syncInfo[1] = append(hs.epoch2syncInfo[1], &RelayerSyncInfo{
		Relayer: common.HexToAddress("0x3f98da321de0a"),
		Times:   66666,
	})

	return hs
}

func TestCloneHeaderStore(t *testing.T) {
	type args struct {
		src *HeaderStore
	}
	tests := []struct {
		name    string
		args    args
		wantDst *HeaderStore
		wantErr bool
	}{
		{
			name: "t-1",
			args: args{
				src: getHeaderStore(),
			},
			wantDst: getHeaderStore(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDst, err := CloneHeaderStore(tt.args.src)
			modifyHeaderStore(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("CloneHeaderStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDst, tt.wantDst) {
				t.Errorf("CloneHeaderStore() gotDst = %v, want %v", gotDst, tt.wantDst)
			}
		})
	}
}

func TestHeaderStore_AddEpochReward(t *testing.T) {

	type args struct {
		epochID uint64
		reward  *big.Int
	}
	tests := []struct {
		name       string
		hs         *HeaderStore
		args       args
		fn         func(hs *HeaderStore)
		wantReward *big.Int
	}{
		{
			name: "not-set-epoch2reward",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 0,
				reward:  big.NewInt(100),
			},
			fn:         func(hs *HeaderStore) {},
			wantReward: big.NewInt(100),
		},
		{
			name: "not-set-epoch2reward-two-call-AddEpochReward",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 2,
				reward:  big.NewInt(100),
			},
			fn: func(hs *HeaderStore) {
				hs.AddEpochReward(2, big.NewInt(20))
			},
			wantReward: big.NewInt(120),
		},
		{
			name: "set-epoch2reward",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 5,
				reward:  big.NewInt(100),
			},
			fn: func(hs *HeaderStore) {
				hs.SetEpoch2reward(5)
			},
			wantReward: big.NewInt(100),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(tt.hs)
			tt.hs.AddEpochReward(tt.args.epochID, tt.args.reward)
			getReward := tt.hs.GetEpochReward(tt.args.epochID)
			if getReward.Cmp(tt.wantReward) != 0 {
				t.Errorf("AddEpochReward() getReward = %v, want %v", getReward, tt.wantReward)
			}
		})
	}
}

func TestHeaderStore_GetEpochReward(t *testing.T) {
	type args struct {
		epochID uint64
	}
	tests := []struct {
		name string
		hs   *HeaderStore
		args args
		fn   func(hs *HeaderStore)
		want *big.Int
	}{
		{
			name: "t-1",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 10,
			},
			fn: func(hs *HeaderStore) {
				hs.AddEpochReward(10, big.NewInt(998))
			},
			want: big.NewInt(998),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(tt.hs)
			if got := tt.hs.GetEpochReward(tt.args.epochID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEpochReward() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHeaderStore_GetReceiveTimes(t *testing.T) {
	type args struct {
		height uint64
	}
	tests := []struct {
		name string
		hs   *HeaderStore
		args args
		fn   func(hs *HeaderStore)
		want uint64
	}{
		{
			name: "2-times",
			hs:   NewHeaderStore(),
			args: args{
				height: 10,
			},
			fn: func(hs *HeaderStore) {
				hs.IncrReceiveTimes(10)
				hs.IncrReceiveTimes(10)
			},
			want: 2,
		},
		{
			name: "0-times",
			hs:   NewHeaderStore(),
			args: args{
				height: 10,
			},
			fn:   func(hs *HeaderStore) {},
			want: 0,
		},
		{
			name: "3-times",
			hs:   NewHeaderStore(),
			args: args{
				height: 10,
			},
			fn: func(hs *HeaderStore) {
				hs.IncrReceiveTimes(10)
				hs.IncrReceiveTimes(10)
				hs.IncrReceiveTimes(10)
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(tt.hs)
			if got := tt.hs.GetReceiveTimes(tt.args.height); got != tt.want {
				t.Errorf("GetReceiveTimes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHeaderStore_IncrReceiveTimes(t *testing.T) {
	type args struct {
		height uint64
	}
	tests := []struct {
		name      string
		hs        *HeaderStore
		args      args
		fn        func(hs *HeaderStore)
		wantTimes uint64
	}{
		{
			name: "0-times",
			hs:   NewHeaderStore(),
			args: args{
				height: 9898,
			},
			fn:        func(hs *HeaderStore) {},
			wantTimes: 0,
		},
		{
			name: "1-times",
			hs:   NewHeaderStore(),
			args: args{
				height: 5433,
			},
			fn: func(hs *HeaderStore) {
				hs.IncrReceiveTimes(5433)
			},
			wantTimes: 1,
		},
		{
			name: "2-times",
			hs:   NewHeaderStore(),
			args: args{
				height: 6160,
			},
			fn: func(hs *HeaderStore) {
				hs.IncrReceiveTimes(6160)
				hs.IncrReceiveTimes(6160)
			},
			wantTimes: 2,
		},
		{
			name: "3-times",
			hs:   NewHeaderStore(),
			args: args{
				height: 8211,
			},
			fn: func(hs *HeaderStore) {
				hs.IncrReceiveTimes(8211)
				hs.IncrReceiveTimes(8211)
				hs.IncrReceiveTimes(8211)
			},
			wantTimes: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(tt.hs)
			if getTimes := tt.hs.GetReceiveTimes(tt.args.height); getTimes != tt.wantTimes {
				t.Errorf("GetReceiveTimes() = %v, wantTimes %v", getTimes, tt.wantTimes)
			}
		})
	}
}

func TestHeaderStore_StoreReward(t *testing.T) {
	type args struct {
		epochID uint64
		relayer common.Address
		reward  *big.Int
	}
	tests := []struct {
		name   string
		hs     *HeaderStore
		args   args
		before func(hs *HeaderStore)
		after  func(hs *HeaderStore)
		want   *big.Int
	}{
		{
			name: "epoch-not-exist",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderStore) {},
			after:  func(hs *HeaderStore) {},
			want:   big.NewInt(0),
		},
		{
			name: "relayer-sync-info-is-nil",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderStore) {
				hs.epoch2syncInfo[1] = nil
			},
			after: func(hs *HeaderStore) {},
			want:  big.NewInt(0),
		},
		{
			name: "relayer-sync-info-is-zero-value",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderStore) {
				hs.epoch2syncInfo[1] = []*RelayerSyncInfo{}
			},
			after: func(hs *HeaderStore) {},
			want:  big.NewInt(0),
		},
		{
			name: "epoch-exist-reward-is-nil",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderStore) {
				hs.epoch2syncInfo[1] = []*RelayerSyncInfo{
					{
						Relayer: common.HexToAddress("0xae90c87d2e80"),
						Times:   0,
					},
				}
			},
			after: func(hs *HeaderStore) {},
			want:  big.NewInt(123456789),
		},
		{
			name: "store-reward",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(2222222233333333),
			},
			before: func(hs *HeaderStore) {
				hs.epoch2syncInfo[1] = []*RelayerSyncInfo{
					{
						Relayer: common.HexToAddress("0xae90c87d2e80"),
						Times:   0,
						Reward:  new(big.Int),
					},
				}
			},
			after: func(hs *HeaderStore) {},
			want:  big.NewInt(2222222233333333),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(tt.hs)

			tt.hs.StoreReward(tt.args.epochID, tt.args.relayer, tt.args.reward)

			if got := tt.hs.LoadReward(tt.args.epochID, tt.args.relayer); got.Cmp(tt.want) != 0 {
				t.Errorf("LoadReward() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHeaderStore_LoadReward(t *testing.T) {
	type args struct {
		epochID uint64
		relayer common.Address
		reward  *big.Int
	}
	tests := []struct {
		name   string
		hs     *HeaderStore
		args   args
		before func(hs *HeaderStore)
		want   *big.Int
	}{
		{
			name: "epoch-not-exist",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderStore) {},
			want:   big.NewInt(0),
		},
		{
			name: "epoch-exist-reward-is-nil",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderStore) {
				hs.epoch2syncInfo[1] = []*RelayerSyncInfo{
					{
						Relayer: common.HexToAddress("0xae90c87d2e80"),
						Times:   0,
					},
				}
			},
			want: big.NewInt(123456789),
		},
		{
			name: "",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderStore) {
				hs.epoch2syncInfo[1] = []*RelayerSyncInfo{
					{
						Relayer: common.HexToAddress("0xae90c87d2e80"),
						Times:   0,
						Reward:  new(big.Int),
					},
				}
			},
			want: big.NewInt(123456789),
		},
		{
			name: "load-reward",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderStore) {
				hs.epoch2syncInfo[1] = []*RelayerSyncInfo{
					{
						Relayer: common.HexToAddress("0xae90c87d2e80"),
						Times:   0,
						Reward:  new(big.Int),
					},
				}
			},
			want: big.NewInt(123456789),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(tt.hs)

			tt.hs.StoreReward(tt.args.epochID, tt.args.relayer, tt.args.reward)

			if got := tt.hs.LoadReward(tt.args.epochID, tt.args.relayer); got.Cmp(tt.want) != 0 {
				t.Errorf("LoadReward() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHeaderStore_AddSyncTimes(t *testing.T) {
	type args struct {
		epochID uint64
		amount  uint64
		relayer common.Address
	}
	tests := []struct {
		name      string
		hs        *HeaderStore
		args      args
		before    func(hs *HeaderStore)
		wantTimes uint64
	}{
		{
			name: "epoch-not-exist",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 1,
				amount:  200,
				relayer: common.HexToAddress("0xae90c87d2e80"),
			},
			before:    func(hs *HeaderStore) {},
			wantTimes: 200,
		},
		{
			name: "relayer-not-exist",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 18,
				amount:  60,
				relayer: common.HexToAddress("0xae90c87d2e80"),
			},
			before: func(hs *HeaderStore) {
				hs.AddSyncTimes(18, 100, common.HexToAddress("0xae90c87d2e80"))
			},
			wantTimes: 160,
		},
		{
			name: "epoch-relayer-exist",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 1,
				amount:  200,
				relayer: common.HexToAddress("0xae90c87d2e80"),
			},
			before:    func(hs *HeaderStore) {},
			wantTimes: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(tt.hs)
			tt.hs.AddSyncTimes(tt.args.epochID, tt.args.amount, tt.args.relayer)
			if gotTimes := tt.hs.LoadSyncTimes(tt.args.epochID, tt.args.relayer); gotTimes != tt.wantTimes {
				t.Errorf("LoadSyncTimes() = %v, wantTimes %v", gotTimes, tt.wantTimes)
			}
		})
	}
}

func TestHeaderStore_GetSortedRelayers(t *testing.T) {
	type args struct {
		epochID uint64
	}
	tests := []struct {
		name   string
		hs     *HeaderStore
		args   args
		before func(hs *HeaderStore)
		want   []common.Address
	}{
		{
			name: "t-1",
			hs:   NewHeaderStore(),
			args: args{
				epochID: 3,
			},
			before: func(hs *HeaderStore) {
				hs.AddSyncTimes(3, 100, common.HexToAddress("0xae90c87d2e80"))
				hs.AddSyncTimes(3, 150, common.HexToAddress("0xb890c87d2e80"))
				hs.AddSyncTimes(3, 128, common.HexToAddress("0xa490c87d2e80"))
				hs.AddSyncTimes(3, 232, common.HexToAddress("0xe090c87d2e80"))
			},
			want: []common.Address{
				common.HexToAddress("0x0000000000000000000000000000B890C87d2E80"),
				common.HexToAddress("0x0000000000000000000000000000E090c87D2e80"),
				common.HexToAddress("0x0000000000000000000000000000a490C87d2E80"),
				common.HexToAddress("0x0000000000000000000000000000aE90c87D2e80"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(tt.hs)
			if got := tt.hs.GetSortedRelayers(tt.args.epochID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSortedRelayers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getStateDB() *state.StateDB {
	finalDb := rawdb.NewMemoryDatabase()
	finalState, _ := state.New(common.Hash{}, state.NewDatabase(finalDb), nil)
	return finalState
}

func TestHeaderStore_Store(t *testing.T) {
	type args struct {
		state   StateDB
		address common.Address
	}
	tests := []struct {
		name    string
		hs      *HeaderStore
		args    args
		wantErr bool
	}{
		{
			name: "t-1",
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
			args: args{
				state:   getStateDB(),
				address: params.HeaderStoreAddress,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.hs.Store(tt.args.state, tt.args.address); (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHeaderStore_Load(t *testing.T) {
	type args struct {
		state   StateDB
		address common.Address
	}
	tests := []struct {
		name    string
		hs      *HeaderStore
		args    args
		before  func(hs *HeaderStore, state StateDB)
		after   func(hs *HeaderStore)
		wantErr bool
	}{
		{
			name: "cache-exist",
			hs: &HeaderStore{
				epoch2reward: map[uint64]*big.Int{
					1: big.NewInt(1111111111111111),
					2: big.NewInt(2222222222222222),
					3: big.NewInt(3333333333333333),
				},
				height2receiveTimes: map[uint64]uint64{
					101: 1,
					202: 2,
					303: 3,
				},
			},
			args: args{
				state:   getStateDB(),
				address: params.HeaderStoreAddress,
			},
			before: func(hs *HeaderStore, state StateDB) {
				_ = hs.Store(state, params.HeaderStoreAddress)
			},
			after: func(hs *HeaderStore) {
				for e, r := range hs.epoch2reward {
					fmt.Printf("epoch: %v, reward: %v\n", e, r)
				}
			},
			wantErr: false,
		},
		{
			name: "cache-not-exist",
			hs: &HeaderStore{
				epoch2reward: map[uint64]*big.Int{
					1: big.NewInt(1111111111111111),
					2: big.NewInt(2222222222222222),
					3: big.NewInt(3333333333333333),
				},
				height2receiveTimes: map[uint64]uint64{
					101: 1,
					202: 2,
					303: 3,
				},
			},
			args: args{
				state:   getStateDB(),
				address: params.HeaderStoreAddress,
			},
			before: func(hs *HeaderStore, state StateDB) {
				_ = hs.Store(state, params.HeaderStoreAddress)
				// remove cache
				key := common.BytesToHash(params.HeaderStoreAddress[:])
				data := state.GetPOWState(params.HeaderStoreAddress, key)
				hash := RlpHash(data)
				hsCache.Cache.Remove(hash)

			},
			after: func(hs *HeaderStore) {
				for e, r := range hs.epoch2reward {
					fmt.Printf("epoch: %v, reward: %v\n", e, r)
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(tt.hs, tt.args.state)
			if err := tt.hs.Load(tt.args.state, tt.args.address); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.after(tt.hs)
		})
	}
}
