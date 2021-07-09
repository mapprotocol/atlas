package vm

import (
	"io"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/rlp"
)

type extHeaderStore struct {
	RewardEpochs []uint64
	Rewards      []*big.Int
	Heights      []uint64
	ReceiveTimes []uint64
	SyncEpochs   []uint64
	SyncInfo     [][]*RelayerSyncInfo
}

func (h *HeaderStore) EncodeRLP(w io.Writer) error {
	var (
		rewards      []*big.Int
		rewardEpochs []uint64
		heights      []uint64
		receiveTimes []uint64
		syncEpochs   []uint64
		syncInfo     [][]*RelayerSyncInfo
	)

	for e := range h.epoch2reward {
		rewardEpochs = append(rewardEpochs, e)
	}
	for _, e := range rewardEpochs {
		rewards = append(rewards, h.epoch2reward[e])
	}

	for height := range h.height2receiveTimes {
		heights = append(heights, height)
	}
	for _, height := range heights {
		receiveTimes = append(receiveTimes, h.height2receiveTimes[height])
	}
	sort.Slice(syncEpochs, func(i, j int) bool {
		return syncEpochs[i] < syncEpochs[j]
	})

	for e := range h.epoch2syncInfo {
		syncEpochs = append(syncEpochs, e)
	}
	for _, e := range syncEpochs {
		syncInfo = append(syncInfo, h.epoch2syncInfo[e])
	}

	return rlp.Encode(w, extHeaderStore{
		Rewards:      rewards,
		RewardEpochs: rewardEpochs,
		Heights:      heights,
		ReceiveTimes: receiveTimes,
		SyncEpochs:   syncEpochs,
		SyncInfo:     syncInfo,
	})
}

func (h *HeaderStore) DecodeRLP(s *rlp.Stream) error {
	var eh extHeaderStore
	if err := s.Decode(&eh); err != nil {
		return err
	}
	epoch2reward := make(map[uint64]*big.Int)
	height2receiveTimes := make(map[uint64]uint64)
	epoch2syncInfo := make(map[uint64][]*RelayerSyncInfo)

	for i, r := range eh.Rewards {
		epoch2reward[eh.RewardEpochs[i]] = r
	}
	for i, r := range eh.ReceiveTimes {
		height2receiveTimes[eh.Heights[i]] = r
	}

	for i, si := range eh.SyncInfo {
		epoch2syncInfo[eh.SyncEpochs[i]] = si
	}

	h.epoch2reward, h.height2receiveTimes, h.epoch2syncInfo = epoch2reward, height2receiveTimes, epoch2syncInfo
	return nil
}
