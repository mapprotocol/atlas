package vm

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"

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

type storeInfo struct {
	abnormalMsg  string
	receiveTimes uint8
}

type HeaderStore struct {
	height2receiveTimes map[uint64]uint8
	// the first layer key is the epoch id
	// the second layer key is the relayer address
	// the value is the number of times the repeater has been synchronized
	relayerSyncTimes map[uint64]map[common.Address]uint64
	// the first layer key is the relayer address
	// the second layer key is the height of the block
	// the value is abnormal msg
	abnormalMsg map[common.Address]map[uint64]string
}

func NewHeaderStore() *HeaderStore {
	return &HeaderStore{}
}

func CloneHeaderStore(src *HeaderStore) (dst *HeaderStore, err error) {
	cp, err := DeepCopy(src)
	if err != nil {
		return nil, err
	}
	return cp.(*HeaderStore), nil
}

func DeepCopy(src interface{}) (dst interface{}, err error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return nil, err
	}
	if err := gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst); err != nil {
		return nil, err
	}
	return dst, nil
}

func (h *HeaderStore) Store(state StateDB, address common.Address) error {
	key := common.BytesToHash(address[:])
	data, err := rlp.EncodeToBytes(h)
	if err != nil {
		log.Error("Failed to RLP encode HeaderStore", "err", err, "HeaderStore", h)
		return err
	}

	hash := RlpHash(data)
	state.SetState(address, key, hash)

	clone, err := CloneHeaderStore(h)
	if err != nil {
		return err
	}
	hsCache.Cache.Add(hash, clone)
	return nil
}

func (h *HeaderStore) Load(state StateDB, address common.Address) (err error) {
	key := common.BytesToHash(address[:])
	hash := state.GetState(address, key)
	data, err := rlp.EncodeToBytes(hash)
	if err != nil {
		log.Error("Failed to RLP encode HeaderStore", "err", err, "HeaderStore", h)
		return err
	}

	var hs HeaderStore
	if cc, ok := hsCache.Cache.Get(hash); ok {
		cp, err := CloneHeaderStore(cc.(*HeaderStore))
		if err != nil {
			return err
		}
		hs = *cp
		h.height2receiveTimes, h.relayerSyncTimes, h.abnormalMsg = hs.height2receiveTimes, hs.relayerSyncTimes, hs.abnormalMsg
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
	h.height2receiveTimes, h.relayerSyncTimes, h.abnormalMsg = hs.height2receiveTimes, hs.relayerSyncTimes, hs.abnormalMsg
	return nil
}

func (h *HeaderStore) GetReceiveTimes(height uint64) uint8 {
	return h.height2receiveTimes[height]
}

func (h *HeaderStore) IncrReceiveTimes(height uint64) {
	h.height2receiveTimes[height]++
}

func (h *HeaderStore) StoreAbnormalMsg(relayer common.Address, height uint64, msg string) {
	h.abnormalMsg[relayer][height] = msg
}

func (h *HeaderStore) LoadAbnormalMsg(relayer common.Address, height uint64) string {
	return h.abnormalMsg[relayer][height]
}

func (h *HeaderStore) AddSyncTimes(epochID, amount uint64, relayer common.Address) {
	h.relayerSyncTimes[epochID][relayer] += amount
}

func (h *HeaderStore) LoadSyncTimes(epochID uint64, relayer common.Address) uint64 {
	return h.relayerSyncTimes[epochID][relayer]
}

func (h *HeaderStore) GetSortedRelayers(epochID uint64) []common.Address {
	m := h.relayerSyncTimes[epochID]
	rss := make([]string, 0, len(m))
	rs := make([]common.Address, 0, len(m))
	for addr := range m {
		rss = append(rss, addr.String())
	}

	sort.Strings(rss)
	for _, r := range rss {
		rs = append(rs, common.HexToAddress(r))
	}
	return rs
}

func HistoryWorkEfficiency(state StateDB, epochId uint64, relayer common.Address) (uint64, error) {
	headerStore := NewHeaderStore()
	err := headerStore.Load(state, SyncAddress)
	if err != nil {
		log.Error("header store load error", "error", err)
		return 0, err
	}

	return headerStore.LoadSyncTimes(epochId, relayer), nil
}
