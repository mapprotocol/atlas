package vm

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/big"
	"sort"

	"github.com/mapprotocol/atlas/params"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	lru "github.com/hashicorp/golang-lru"
)

const (
	CacheSize = 20
)

var (
	//hsLock  sync.Mutex
	hsCache *HeaderStoreCache
)

func init() {
	hsCache = &HeaderStoreCache{
		size: CacheSize,
	}
	hsCache.Cache, _ = lru.New(hsCache.size)
}

type HeaderStoreCache struct {
	Cache *lru.Cache
	size  int
}

//type HeaderStore struct {
//	height2receiveTimes map[uint64]uint8
//	// the first layer key is the epoch id
//	// the second layer key is the relayer address
//	// the value is the number of times the repeater has been synchronized
//	relayerSyncTimes map[uint64]map[common.Address]uint64
//	// the first layer key is the relayer address
//	// the second layer key is the height of the block
//	// the value is abnormal msg
//	abnormalMsg map[common.Address]map[uint64]string
//}

type AbnormalInfo struct {
	Height *big.Int
	Msg    string
}

type RelayerSyncInfo struct {
	Relayer        common.Address
	Times          uint64
	InvalidHeaders []*AbnormalInfo
}

type HeaderStore struct {
	epoch2reward        map[uint64]*big.Int
	height2receiveTimes map[uint64]uint64
	epoch2syncInfo      map[uint64][]*RelayerSyncInfo
}

func NewHeaderStore() *HeaderStore {
	return &HeaderStore{
		epoch2reward:        make(map[uint64]*big.Int),
		height2receiveTimes: make(map[uint64]uint64),
		epoch2syncInfo:      make(map[uint64][]*RelayerSyncInfo),
	}
}

func (h *HeaderStore) SetEpoch2reward(epochID uint64) {
	if _, ok := h.epoch2reward[epochID]; !ok {
		h.epoch2reward[epochID] = big.NewInt(0)
	}
}

func CloneHeaderStore(src *HeaderStore) (dst *HeaderStore, err error) {
	dst = NewHeaderStore()
	err = DeepCopy(src.epoch2reward, &dst.epoch2reward)
	if err != nil {
		return nil, err
	}
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

func (h *HeaderStore) Store(state StateDB, address common.Address) error {
	key := common.BytesToHash(address[:])
	data, err := rlp.EncodeToBytes(h)
	if err != nil {
		log.Error("Failed to RLP encode HeaderStore", "err", err, "HeaderStore", h)
		return err
	}

	state.SetPOWState(address, key, data)

	clone, err := CloneHeaderStore(h)
	if err != nil {
		return err
	}
	hash := RlpHash(data)
	hsCache.Cache.Add(hash, clone)
	return nil
}

func (h *HeaderStore) Load(state StateDB, address common.Address) (err error) {
	key := common.BytesToHash(address[:])
	data := state.GetPOWState(address, key)
	var hs HeaderStore
	hash := RlpHash(data)
	if cc, ok := hsCache.Cache.Get(hash); ok {
		cp, err := CloneHeaderStore(cc.(*HeaderStore))
		if err != nil {
			return err
		}
		hs = *cp
		h.epoch2reward, h.height2receiveTimes, h.epoch2syncInfo = hs.epoch2reward, hs.height2receiveTimes, hs.epoch2syncInfo
		return nil
	}

	if err := rlp.DecodeBytes(data, &hs); err != nil {
		log.Error("HeaderStore RLP decode failed", "err", err, "HeaderStore", data)
		return fmt.Errorf("HeaderStore RLP decode failed, error: %s", err.Error())
	}

	clone, err := CloneHeaderStore(&hs)
	if err != nil {
		return err
	}
	hsCache.Cache.Add(hash, clone)
	h.epoch2reward, h.height2receiveTimes, h.epoch2syncInfo = hs.epoch2reward, hs.height2receiveTimes, hs.epoch2syncInfo
	return nil
}

func (h *HeaderStore) AddEpochReward(epochID uint64, reward *big.Int) {
	r := h.epoch2reward[epochID]
	if r == nil {
		h.epoch2reward[epochID] = new(big.Int).Add(big.NewInt(0), reward)
		return
	}
	h.epoch2reward[epochID] = new(big.Int).Add(h.epoch2reward[epochID], reward)
}

func (h *HeaderStore) GetEpochReward(epochID uint64) *big.Int {
	r := h.epoch2reward[epochID]
	if r == nil {
		return big.NewInt(0)
	}
	return h.epoch2reward[epochID]
}

func (h *HeaderStore) GetReceiveTimes(height uint64) uint64 {
	return h.height2receiveTimes[height]
}

func (h *HeaderStore) IncrReceiveTimes(height uint64) {
	h.height2receiveTimes[height]++
}

func (h *HeaderStore) StoreAbnormalMsg(relayer common.Address, height *big.Int, msg string) {
	epoch := GetEpochFromHeight(height.Uint64())
	// epoch does not exist
	if _, ok := h.epoch2syncInfo[epoch.EpochID]; !ok {
		h.epoch2syncInfo[epoch.EpochID] = append(h.epoch2syncInfo[epoch.EpochID], &RelayerSyncInfo{
			Relayer: relayer,
			Times:   0,
			InvalidHeaders: []*AbnormalInfo{
				{
					Height: height,
					Msg:    msg,
				},
			},
		})
		return
	}

	relayerExist := false
	for i, rsi := range h.epoch2syncInfo[epoch.EpochID] {
		if bytes.Equal(rsi.Relayer.Bytes(), relayer.Bytes()) {
			h.epoch2syncInfo[epoch.EpochID][i].InvalidHeaders = append(
				h.epoch2syncInfo[epoch.EpochID][i].InvalidHeaders,
				&AbnormalInfo{
					Height: height,
					Msg:    msg,
				})
			relayerExist = true
		}
	}
	// relayer does not exist
	if !relayerExist {
		h.epoch2syncInfo[epoch.EpochID] = append(h.epoch2syncInfo[epoch.EpochID], &RelayerSyncInfo{
			Relayer: relayer,
			Times:   0,
			InvalidHeaders: []*AbnormalInfo{
				{
					Height: height,
					Msg:    msg,
				},
			},
		})
	}
}

func (h *HeaderStore) LoadAbnormalMsg(relayer common.Address, height *big.Int) string {
	epoch := GetEpochFromHeight(height.Uint64())
	for _, rsi := range h.epoch2syncInfo[epoch.EpochID] {
		if bytes.Equal(rsi.Relayer.Bytes(), relayer.Bytes()) {
			for _, h := range rsi.InvalidHeaders {
				if h.Height.Cmp(height) == 0 {
					return h.Msg
				}
			}
		}
	}
	return ""
}

func (h *HeaderStore) AddSyncTimes(epochID, amount uint64, relayer common.Address) {
	// epoch does not exist
	if _, ok := h.epoch2syncInfo[epochID]; !ok {
		h.epoch2syncInfo[epochID] = append(h.epoch2syncInfo[epochID], &RelayerSyncInfo{
			Relayer:        relayer,
			Times:          amount,
			InvalidHeaders: []*AbnormalInfo{},
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
			Relayer:        relayer,
			Times:          amount,
			InvalidHeaders: []*AbnormalInfo{},
		})
	}
}

func (h *HeaderStore) LoadSyncTimes(epochID uint64, relayer common.Address) uint64 {
	for i, rsi := range h.epoch2syncInfo[epochID] {
		if bytes.Equal(rsi.Relayer.Bytes(), relayer.Bytes()) {
			return h.epoch2syncInfo[epochID][i].Times
		}
	}
	return 0
}

func (h *HeaderStore) GetSortedRelayers(epochID uint64) []common.Address {
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

func (h *HeaderStore) CalcReward(epochID uint64, allAmount *big.Int) map[common.Address]*big.Int {
	residualReward := allAmount
	relayers := h.GetSortedRelayers(epochID)
	rewards := make(map[common.Address]*big.Int, len(relayers))

	for i, r := range relayers {
		if i == len(relayers)-1 {
			rewards[r] = residualReward
			break
		}

		totalSyncTimes := uint64(0)
		for _, s := range h.epoch2syncInfo[epochID] {
			totalSyncTimes += s.Times
		}

		times := h.LoadSyncTimes(epochID, r)
		singleBlockReward := new(big.Int).Quo(allAmount, new(big.Int).SetUint64(totalSyncTimes))
		relayerReward := new(big.Int).Mul(singleBlockReward, new(big.Int).SetUint64(times))
		residualReward = new(big.Int).Sub(residualReward, relayerReward)
		rewards[r] = relayerReward
	}
	return rewards
}

func HistoryWorkEfficiency(state StateDB, epochId uint64, relayer common.Address) (uint64, error) {
	headerStore := NewHeaderStore()
	err := headerStore.Load(state, params.HeaderStoreAddress)
	if err != nil {
		log.Error("header store load error", "error", err)
		return 0, err
	}

	return headerStore.LoadSyncTimes(epochId, relayer), nil
}
