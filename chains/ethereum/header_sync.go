package ethereum

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	lru "github.com/hashicorp/golang-lru"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/params"
	"github.com/mapprotocol/atlas/tools"
)

const (
	SyncCacheSize = 20
)

var (
	syncCache *HeaderSyncCache
)

func init() {
	syncCache = &HeaderSyncCache{
		size: SyncCacheSize,
	}
	syncCache.Cache, _ = lru.New(syncCache.size)
}

type HeaderSyncCache struct {
	Cache *lru.Cache
	size  int
}

type RelayerSyncInfo struct {
	Relayer common.Address
	Times   uint64
	Reward  *big.Int
}

type HeaderSync struct {
	height2receiveTimes map[uint64]uint64
	epoch2syncInfo      map[uint64][]*RelayerSyncInfo
}

func NewHeaderSync() *HeaderSync {
	return &HeaderSync{
		height2receiveTimes: make(map[uint64]uint64),
		epoch2syncInfo:      make(map[uint64][]*RelayerSyncInfo),
	}
}

func CloneHeaderSync(src *HeaderSync) (dst *HeaderSync, err error) {
	dst = NewHeaderSync()
	err = DeepCopy(src.height2receiveTimes, &dst.height2receiveTimes)
	if err != nil {
		return nil, err
	}
	err = DeepCopy(src.epoch2syncInfo, &dst.epoch2syncInfo)
	if err != nil {
		return nil, err
	}
	return dst, nil
}

func DeepCopy(src, dst interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func (h *HeaderSync) Store(state types.StateDB) error {
	var (
		address = chains.EthereumHeaderSyncAddress
		key     = common.BytesToHash(address[:])
	)

	data, err := rlp.EncodeToBytes(h)
	if err != nil {
		log.Error("Failed to RLP encode HeaderSync", "err", err, "HeaderSync", h)
		return err
	}

	state.SetPOWState(address, key, data)

	clone, err := CloneHeaderSync(h)
	if err != nil {
		return err
	}
	hash := tools.RlpHash(data)
	syncCache.Cache.Add(hash, clone)
	return nil
}

func (h *HeaderSync) Load(state types.StateDB) (err error) {
	var (
		hs      HeaderSync
		address = chains.EthereumHeaderSyncAddress
		key     = common.BytesToHash(address[:])
	)

	data := state.GetPOWState(address, key)
	hash := tools.RlpHash(data)
	if cc, ok := syncCache.Cache.Get(hash); ok {
		cp, err := CloneHeaderSync(cc.(*HeaderSync))
		if err != nil {
			return err
		}
		hs = *cp
		h.height2receiveTimes, h.epoch2syncInfo = hs.height2receiveTimes, hs.epoch2syncInfo
		return nil
	}

	if err := rlp.DecodeBytes(data, &hs); err != nil {
		log.Error("HeaderSync RLP decode failed", "err", err, "HeaderSync", data)
		return fmt.Errorf("HeaderSync RLP decode failed, error: %s", err.Error())
	}

	clone, err := CloneHeaderSync(&hs)
	if err != nil {
		return err
	}
	syncCache.Cache.Add(hash, clone)
	h.height2receiveTimes, h.epoch2syncInfo = hs.height2receiveTimes, hs.epoch2syncInfo
	return nil
}

func (h *HeaderSync) GetReceiveTimes(height uint64) uint64 {
	return h.height2receiveTimes[height]
}

func (h *HeaderSync) IncrReceiveTimes(height uint64) {
	h.height2receiveTimes[height]++
}

func (h *HeaderSync) StoreReward(epochID uint64, relayer common.Address, reward *big.Int) {
	for _, rsi := range h.epoch2syncInfo[epochID] {
		if bytes.Equal(rsi.Relayer.Bytes(), relayer.Bytes()) {
			if rsi.Reward == nil {
				rsi.Reward = new(big.Int)
			}
			rsi.Reward = reward
		}
	}
}

func (h *HeaderSync) LoadReward(epochID uint64, relayer common.Address) *big.Int {
	for _, rsi := range h.epoch2syncInfo[epochID] {
		if bytes.Equal(rsi.Relayer.Bytes(), relayer.Bytes()) {
			return rsi.Reward
		}
	}
	return big.NewInt(0)
}

func (h *HeaderSync) AddSyncTimes(epochID, amount uint64, relayer common.Address) {
	// epoch does not exist
	if _, ok := h.epoch2syncInfo[epochID]; !ok {
		h.epoch2syncInfo[epochID] = append(h.epoch2syncInfo[epochID], &RelayerSyncInfo{
			Relayer: relayer,
			Times:   amount,
			Reward:  new(big.Int),
		})
		return
	}

	relayerExist := false
	for i, rsi := range h.epoch2syncInfo[epochID] {
		if bytes.Equal(rsi.Relayer.Bytes(), relayer.Bytes()) {
			h.epoch2syncInfo[epochID][i].Times += amount
			relayerExist = true
		}
	}

	// relayer does not exist
	if !relayerExist {
		h.epoch2syncInfo[epochID] = append(h.epoch2syncInfo[epochID], &RelayerSyncInfo{
			Relayer: relayer,
			Times:   amount,
			Reward:  new(big.Int),
		})
	}
}

func (h *HeaderSync) LoadRelayerSyncTimes(db types.StateDB, epochID uint64, relayer common.Address) (uint64, error) {
	if err := h.Load(db); err != nil {
		return 0, err
	}
	for i, rsi := range h.epoch2syncInfo[epochID] {
		if bytes.Equal(rsi.Relayer.Bytes(), relayer.Bytes()) {
			return h.epoch2syncInfo[epochID][i].Times, nil
		}
	}
	return 0, nil
}

func (h *HeaderSync) GetSortedRelayers(epochID uint64) []common.Address {
	rsis := h.epoch2syncInfo[epochID]
	rss := make([]string, 0, len(rsis))
	rs := make([]common.Address, 0, len(rsis))
	for _, rsi := range rsis {
		rss = append(rss, rsi.Relayer.String())
	}

	sort.Strings(rss)
	for _, r := range rss {
		rs = append(rs, common.HexToAddress(r))
	}
	return rs
}

//func (h *HeaderSync) CalcReward(epochID uint64, allAmount *big.Int) map[common.Address]*big.Int {
//	residualReward := allAmount
//	relayers := h.GetSortedRelayers(epochID)
//	rewards := make(map[common.Address]*big.Int, len(relayers))
//
//	totalSyncTimes := uint64(0)
//	for _, s := range h.epoch2syncInfo[epochID] {
//		totalSyncTimes += s.Times
//	}
//	if totalSyncTimes == 0 {
//		return rewards
//	}
//	singleBlockReward := new(big.Int).Quo(allAmount, new(big.Int).SetUint64(totalSyncTimes))
//
//	for i, r := range relayers {
//		if i == len(relayers)-1 {
//			rewards[r] = residualReward
//			break
//		}
//
//		times := h.LoadRelayerSyncTimes(epochID, r)
//		relayerReward := new(big.Int).Mul(singleBlockReward, new(big.Int).SetUint64(times))
//		residualReward = new(big.Int).Sub(residualReward, relayerReward)
//		rewards[r] = relayerReward
//	}
//	return rewards
//}

func (h *HeaderSync) StoreSyncTimes(db types.StateDB, epochID uint64, relayer common.Address, headers []*params.NumberHash) error {
	if err := h.Load(db); err != nil {
		return err
	}

	// todo
	//var total uint64
	//for _, header := range headers {
	//	if h.GetReceiveTimes(header.Number) >= 3 {
	//		return fmt.Errorf("the number of synchronizations has reached the limit(%d)", 3)
	//	}
	//	total++
	//	h.IncrReceiveTimes(header.Number)
	//}

	h.AddSyncTimes(epochID, uint64(len(headers)), relayer)

	return h.Store(db)
}

func HistoryWorkEfficiency(state types.StateDB, epochId uint64, relayer common.Address) (uint64, error) {
	headerStore := NewHeaderSync()
	err := headerStore.Load(state)
	if err != nil {
		log.Error("header sync load error", "error", err)
		return 0, err
	}

	return headerStore.LoadRelayerSyncTimes(state, epochId, relayer)
}
