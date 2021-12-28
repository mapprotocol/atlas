package ethereum

import (
	"io"
	"sort"

	"github.com/ethereum/go-ethereum/rlp"
)

type extHeaderStore struct {
	Heights      []uint64
	ReceiveTimes []uint64
	SyncEpochs   []uint64
	SyncInfo     [][]*RelayerSyncInfo
}

func (h *HeaderSync) EncodeRLP(w io.Writer) error {
	var (
		heights      []uint64
		receiveTimes []uint64
		syncEpochs   []uint64
		syncInfo     [][]*RelayerSyncInfo
	)

	for height := range h.height2receiveTimes {
		heights = append(heights, height)
	}
	sort.Slice(heights, func(i, j int) bool {
		return heights[i] < heights[j]
	})
	for _, height := range heights {
		receiveTimes = append(receiveTimes, h.height2receiveTimes[height])
	}

	for e := range h.epoch2syncInfo {
		syncEpochs = append(syncEpochs, e)
	}
	sort.Slice(syncEpochs, func(i, j int) bool {
		return syncEpochs[i] < syncEpochs[j]
	})
	for _, e := range syncEpochs {
		syncInfo = append(syncInfo, h.epoch2syncInfo[e])
	}

	return rlp.Encode(w, extHeaderStore{
		Heights:      heights,
		ReceiveTimes: receiveTimes,
		SyncEpochs:   syncEpochs,
		SyncInfo:     syncInfo,
	})
}

func (h *HeaderSync) DecodeRLP(s *rlp.Stream) error {
	var eh extHeaderStore
	if err := s.Decode(&eh); err != nil {
		return err
	}
	height2receiveTimes := make(map[uint64]uint64)
	epoch2syncInfo := make(map[uint64][]*RelayerSyncInfo)

	for i, r := range eh.ReceiveTimes {
		height2receiveTimes[eh.Heights[i]] = r
	}

	for i, si := range eh.SyncInfo {
		epoch2syncInfo[eh.SyncEpochs[i]] = si
	}

	h.height2receiveTimes, h.epoch2syncInfo = height2receiveTimes, epoch2syncInfo
	return nil
}
