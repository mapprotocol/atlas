package vm

import (
	"bytes"
	"encoding/gob"
	"fmt"

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
	// the key is the height of the block
	headerStoreInfo map[uint64]storeInfo
	// the first layer key is the epoch id
	// the second layer key is the relayer address
	// the value is the number of times the repeater has been synchronized
	epochStore map[uint64]map[common.Address]uint64
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
		h.epochStore, h.headerStoreInfo = hs.epochStore, hs.headerStoreInfo
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
	h.epochStore, h.headerStoreInfo = hs.epochStore, hs.headerStoreInfo
	return nil
}

func (h *HeaderStore) GetReceiveTimes(height uint64) uint8 {
	return h.headerStoreInfo[height].receiveTimes
}

func (h *HeaderStore) IncrReceiveTimes(height uint64) {
	s := h.headerStoreInfo[height]
	s.receiveTimes++
}

func (h *HeaderStore) StoreAbnormalMsg(height uint64, msg string) {
	s := h.headerStoreInfo[height]
	s.abnormalMsg = msg
}

func (h *HeaderStore) LoadAbnormalMsg(height uint64) string {
	return h.headerStoreInfo[height].abnormalMsg
}

func (h *HeaderStore) AddSyncTimes(epochID, amount uint64, relayer common.Address) {
	h.epochStore[epochID][relayer] += amount
}

func (h *HeaderStore) LoadSyncTimes(epochID uint64, relayer common.Address) uint64 {
	return h.epochStore[epochID][relayer]
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
