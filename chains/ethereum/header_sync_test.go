package ethereum

import (
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func getHeaderStore() *HeaderSync {
	var hs = &HeaderSync{
		//epoch2reward: map[uint64]*big.Int{
		//	1: big.NewInt(48976200),
		//},
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

func modifyHeaderStore(hs *HeaderSync) *HeaderSync {
	//hs.epoch2reward[1] = big.NewInt(1111111111111)
	//hs.epoch2reward[2] = big.NewInt(2222222222222)
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
		src *HeaderSync
	}
	tests := []struct {
		name    string
		args    args
		wantDst *HeaderSync
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

//func TestHeaderStore_AddEpochReward(t *testing.T) {
//
//	type args struct {
//		epochID uint64
//		reward  *big.Int
//	}
//	tests := []struct {
//		name       string
//		hs         *HeaderSync
//		args       args
//		fn         func(hs *HeaderSync)
//		wantReward *big.Int
//	}{
//		{
//			name: "don`t-set-epoch2reward",
//			hs:   NewHeaderStore(),
//			args: args{
//				epochID: 0,
//				reward:  big.NewInt(100),
//			},
//			fn:         func(hs *HeaderSync) {},
//			wantReward: big.NewInt(100),
//		},
//		{
//			name: "add-epoch-reward",
//			hs:   NewHeaderStore(),
//			args: args{
//				epochID: 2,
//				reward:  big.NewInt(100),
//			},
//			fn: func(hs *HeaderSync) {
//				//hs.AddEpochReward(2, big.NewInt(20))
//			},
//			wantReward: big.NewInt(120),
//		},
//		{
//			name: "set-epoch2reward",
//			hs:   NewHeaderStore(),
//			args: args{
//				epochID: 5,
//				reward:  big.NewInt(100),
//			},
//			fn: func(hs *HeaderSync) {
//				hs.SetEpoch2reward(5)
//			},
//			wantReward: big.NewInt(100),
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			tt.fn(tt.hs)
//			//tt.hs.AddEpochReward(tt.args.epochID, tt.args.reward)
//			getReward := tt.hs.GetEpochReward(tt.args.epochID)
//			if getReward.Cmp(tt.wantReward) != 0 {
//				t.Errorf("AddEpochReward() getReward = %v, want %v", getReward, tt.wantReward)
//			}
//		})
//	}
//}

//func TestHeaderStore_GetEpochReward(t *testing.T) {
//	type args struct {
//		epochID uint64
//	}
//	tests := []struct {
//		name string
//		hs   *HeaderSync
//		args args
//		fn   func(hs *HeaderSync)
//		want *big.Int
//	}{
//		{
//			name: "t-1",
//			hs:   NewHeaderStore(),
//			args: args{
//				epochID: 10,
//			},
//			fn: func(hs *HeaderSync) {
//				hs.AddEpochReward(10, big.NewInt(998))
//			},
//			want: big.NewInt(998),
//		},
//		{
//			name: "t-2",
//			hs:   NewHeaderStore(),
//			args: args{
//				epochID: 10,
//			},
//			fn:   func(hs *HeaderSync) {},
//			want: big.NewInt(0),
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			tt.fn(tt.hs)
//			if got := tt.hs.GetEpochReward(tt.args.epochID); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetEpochReward() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

func TestHeaderStore_GetReceiveTimes(t *testing.T) {
	type args struct {
		height uint64
	}
	tests := []struct {
		name string
		hs   *HeaderSync
		args args
		fn   func(hs *HeaderSync)
		want uint64
	}{
		{
			name: "2-times",
			hs:   NewHeaderSync(),
			args: args{
				height: 10,
			},
			fn: func(hs *HeaderSync) {
				hs.IncrReceiveTimes(10)
				hs.IncrReceiveTimes(10)
			},
			want: 2,
		},
		{
			name: "0-times",
			hs:   NewHeaderSync(),
			args: args{
				height: 10,
			},
			fn:   func(hs *HeaderSync) {},
			want: 0,
		},
		{
			name: "3-times",
			hs:   NewHeaderSync(),
			args: args{
				height: 10,
			},
			fn: func(hs *HeaderSync) {
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
		hs        *HeaderSync
		args      args
		fn        func(hs *HeaderSync)
		wantTimes uint64
	}{
		{
			name: "0-times",
			hs:   NewHeaderSync(),
			args: args{
				height: 9898,
			},
			fn:        func(hs *HeaderSync) {},
			wantTimes: 0,
		},
		{
			name: "1-times",
			hs:   NewHeaderSync(),
			args: args{
				height: 5433,
			},
			fn: func(hs *HeaderSync) {
				hs.IncrReceiveTimes(5433)
			},
			wantTimes: 1,
		},
		{
			name: "2-times",
			hs:   NewHeaderSync(),
			args: args{
				height: 6160,
			},
			fn: func(hs *HeaderSync) {
				hs.IncrReceiveTimes(6160)
				hs.IncrReceiveTimes(6160)
			},
			wantTimes: 2,
		},
		{
			name: "3-times",
			hs:   NewHeaderSync(),
			args: args{
				height: 8211,
			},
			fn: func(hs *HeaderSync) {
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
		hs     *HeaderSync
		args   args
		before func(hs *HeaderSync)
		after  func(hs *HeaderSync)
		want   *big.Int
	}{
		{
			name: "epoch-not-exist",
			hs:   NewHeaderSync(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderSync) {},
			after:  func(hs *HeaderSync) {},
			want:   big.NewInt(0),
		},
		{
			name: "relayer-sync-info-is-nil",
			hs:   NewHeaderSync(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderSync) {
				hs.epoch2syncInfo[1] = nil
			},
			after: func(hs *HeaderSync) {},
			want:  big.NewInt(0),
		},
		{
			name: "relayer-sync-info-is-zero-value",
			hs:   NewHeaderSync(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderSync) {
				hs.epoch2syncInfo[1] = []*RelayerSyncInfo{}
			},
			after: func(hs *HeaderSync) {},
			want:  big.NewInt(0),
		},
		{
			name: "epoch-exist-reward-is-nil",
			hs:   NewHeaderSync(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderSync) {
				hs.epoch2syncInfo[1] = []*RelayerSyncInfo{
					{
						Relayer: common.HexToAddress("0xae90c87d2e80"),
						Times:   0,
					},
				}
			},
			after: func(hs *HeaderSync) {},
			want:  big.NewInt(123456789),
		},
		{
			name: "store-reward",
			hs:   NewHeaderSync(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(2222222233333333),
			},
			before: func(hs *HeaderSync) {
				hs.epoch2syncInfo[1] = []*RelayerSyncInfo{
					{
						Relayer: common.HexToAddress("0xae90c87d2e80"),
						Times:   0,
						Reward:  new(big.Int),
					},
				}
			},
			after: func(hs *HeaderSync) {},
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
		hs     *HeaderSync
		args   args
		before func(hs *HeaderSync)
		want   *big.Int
	}{
		{
			name: "epoch-not-exist",
			hs:   NewHeaderSync(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderSync) {},
			want:   big.NewInt(0),
		},
		{
			name: "epoch-exist-reward-is-nil",
			hs:   NewHeaderSync(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderSync) {
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
			hs:   NewHeaderSync(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderSync) {
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
			hs:   NewHeaderSync(),
			args: args{
				epochID: 1,
				relayer: common.HexToAddress("0xae90c87d2e80"),
				reward:  big.NewInt(123456789),
			},
			before: func(hs *HeaderSync) {
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
		hs        *HeaderSync
		args      args
		before    func(hs *HeaderSync)
		wantTimes uint64
	}{
		{
			name: "epoch-not-exist",
			hs:   NewHeaderSync(),
			args: args{
				epochID: 1,
				amount:  200,
				relayer: common.HexToAddress("0xae90c87d2e80"),
			},
			before:    func(hs *HeaderSync) {},
			wantTimes: 200,
		},
		{
			name: "relayer-not-exist",
			hs:   NewHeaderSync(),
			args: args{
				epochID: 18,
				amount:  60,
				relayer: common.HexToAddress("0xae90c87d2e80"),
			},
			before: func(hs *HeaderSync) {
				hs.AddSyncTimes(18, 100, common.HexToAddress("0xae90c87d2e80"))
			},
			wantTimes: 160,
		},
		{
			name: "epoch-relayer-exist",
			hs:   NewHeaderSync(),
			args: args{
				epochID: 1,
				amount:  200,
				relayer: common.HexToAddress("0xae90c87d2e80"),
			},
			before:    func(hs *HeaderSync) {},
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
		hs     *HeaderSync
		args   args
		before func(hs *HeaderSync)
		want   []common.Address
	}{
		{
			name: "t-1",
			hs:   NewHeaderSync(),
			args: args{
				epochID: 3,
			},
			before: func(hs *HeaderSync) {
				hs.AddSyncTimes(3, 100, common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"))
				hs.AddSyncTimes(3, 150, common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f5"))
				hs.AddSyncTimes(3, 128, common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f6"))
				hs.AddSyncTimes(3, 232, common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f7"))
			},
			want: []common.Address{
				common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f5"),
				common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f7"),
				common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f6"),
				common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"),
			},
		},
	}
	for i := 0; i < 1000000; i++ {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.before(tt.hs)
				if got := tt.hs.GetSortedRelayers(tt.args.epochID); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetSortedRelayers() = %v, want %v", got, tt.want)
				}
			})
		}
	}
}

func getStateDB() *state.StateDB {
	finalDb := rawdb.NewMemoryDatabase()
	finalState, _ := state.New(common.Hash{}, state.NewDatabase(finalDb), nil)
	return finalState
}

func TestHeaderStore_Store(t *testing.T) {
	type args struct {
		state   types.StateDB
		address common.Address
	}
	tests := []struct {
		name    string
		hs      *HeaderSync
		args    args
		wantErr bool
	}{
		{
			name: "t-1",
			hs: &HeaderSync{
				//epoch2reward: map[uint64]*big.Int{
				//	1: big.NewInt(111),
				//	2: big.NewInt(222),
				//	3: big.NewInt(333),
				//},
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
		state   types.StateDB
		address common.Address
	}
	tests := []struct {
		name    string
		hs      *HeaderSync
		args    args
		before  func(hs *HeaderSync, state types.StateDB)
		after   func(hs *HeaderSync)
		wantErr bool
	}{
		{
			name: "cache-exist",
			hs: &HeaderSync{
				//epoch2reward: map[uint64]*big.Int{
				//	1: big.NewInt(1111111111111111),
				//	2: big.NewInt(2222222222222222),
				//	3: big.NewInt(3333333333333333),
				//},
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
			before: func(hs *HeaderSync, state types.StateDB) {
				_ = hs.Store(state, params.HeaderStoreAddress)
			},
			after: func(hs *HeaderSync) {
				//for e, r := range hs.epoch2reward {
				//	fmt.Printf("epoch: %v, reward: %v\n", e, r)
				//}
			},
			wantErr: false,
		},
		{
			name: "cache-not-exist",
			hs: &HeaderSync{
				//epoch2reward: map[uint64]*big.Int{
				//	1: big.NewInt(1111111111111111),
				//	2: big.NewInt(2222222222222222),
				//	3: big.NewInt(3333333333333333),
				//},
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
			before: func(hs *HeaderSync, state types.StateDB) {
				_ = hs.Store(state, params.HeaderStoreAddress)
				// remove cache
				key := common.BytesToHash(params.HeaderStoreAddress[:])
				data := state.GetPOWState(params.HeaderStoreAddress, key)
				hash := vm.RlpHash(data)
				hsCache.Cache.Remove(hash)

			},
			after: func(hs *HeaderSync) {
				//for e, r := range hs.epoch2reward {
				//	fmt.Printf("epoch: %v, reward: %v\n", e, r)
				//}
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

func TestHistoryWorkEfficiency(t *testing.T) {
	type args struct {
		state   types.StateDB
		epochId uint64
		relayer common.Address
	}
	tests := []struct {
		name    string
		hs      *HeaderSync
		args    args
		before  func(hs *HeaderSync, args args)
		want    uint64
		wantErr bool
	}{
		{
			name: "didn't-store",
			hs:   NewHeaderSync(),
			args: args{
				state:   getStateDB(),
				epochId: 0,
				relayer: common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"),
			},
			before: func(hs *HeaderSync, args args) {
				hs.AddSyncTimes(args.epochId, 258, args.relayer)
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "success",
			hs:   NewHeaderSync(),
			args: args{
				state:   getStateDB(),
				epochId: 0,
				relayer: common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"),
			},
			before: func(hs *HeaderSync, args args) {
				hs.AddSyncTimes(args.epochId, 101, args.relayer)
				_ = hs.Store(args.state, params.HeaderStoreAddress)
			},
			want:    101,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(tt.hs, tt.args)

			got, err := HistoryWorkEfficiency(tt.args.state, tt.args.epochId, tt.args.relayer)
			if (err != nil) != tt.wantErr {
				t.Errorf("HistoryWorkEfficiency() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HistoryWorkEfficiency() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHeaderStore_CalcReward(t *testing.T) {
	hs := &HeaderSync{
		height2receiveTimes: map[uint64]uint64{},
		epoch2syncInfo: map[uint64][]*RelayerSyncInfo{
			1: {
				{
					Relayer: common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"),
					Times:   1,
					Reward:  &big.Int{},
				},
				{
					Relayer: common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f5"),
					Times:   1,
					Reward:  &big.Int{},
				},
				{
					Relayer: common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f6"),
					Times:   1,
					Reward:  &big.Int{},
				},
			},
			2: {
				{
					Relayer: common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"),
					Times:   1,
					Reward:  &big.Int{},
				},
				{
					Relayer: common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f5"),
					Times:   2,
					Reward:  &big.Int{},
				},
				{
					Relayer: common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f6"),
					Times:   1,
					Reward:  &big.Int{},
				},
			},
			3: {
				{
					Relayer: common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"),
					Times:   1,
					Reward:  &big.Int{},
				},
				{
					Relayer: common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f5"),
					Times:   2,
					Reward:  &big.Int{},
				},
			},
		},
	}
	type args struct {
		epochID   uint64
		allAmount *big.Int
	}
	tests := []struct {
		name string
		hs   *HeaderSync
		args args
		want map[common.Address]*big.Int
	}{
		{
			name: "t-1",
			hs:   hs,
			args: args{
				epochID:   1,
				allAmount: big.NewInt(5000000),
			},
			want: map[common.Address]*big.Int{
				common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"): big.NewInt(1666668),
				common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f5"): big.NewInt(1666666),
				common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f6"): big.NewInt(1666666),
			},
		},
		{
			name: "t-2",
			hs:   hs,
			args: args{
				epochID:   2,
				allAmount: big.NewInt(5000000),
			},
			want: map[common.Address]*big.Int{
				common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"): big.NewInt(1250000),
				common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f5"): big.NewInt(2500000),
				common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f6"): big.NewInt(1250000),
			},
		},
		{
			name: "t-3",
			hs:   hs,
			args: args{
				epochID:   3,
				allAmount: big.NewInt(5000000),
			},
			want: map[common.Address]*big.Int{
				common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"): big.NewInt(1666668),
				common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f5"): big.NewInt(3333332),
			},
		},
		{
			name: "t-4",
			hs: &HeaderSync{
				epoch2syncInfo: map[uint64][]*RelayerSyncInfo{
					4: {
						{
							Relayer: common.HexToAddress("0x32CD75ca677e9C37FD989272afA8504CB8F6eB52"),
							Times:   30,
							Reward:  &big.Int{},
						},
						{
							Relayer: common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"),
							Times:   30,
							Reward:  &big.Int{},
						},
					},
				},
			},
			args: args{
				epochID:   4,
				allAmount: new(big.Int).Mul(big.NewInt(1e18), big.NewInt(10)),
			},
			want: map[common.Address]*big.Int{
				common.HexToAddress("0x32CD75ca677e9C37FD989272afA8504CB8F6eB52"): big.NewInt(4999999999999999980),
				common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"): big.NewInt(5000000000000000020),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.hs.CalcReward(tt.args.epochID, tt.args.allAmount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalcReward() = %v, want %v", got, tt.want)
			}
		})
	}
}
