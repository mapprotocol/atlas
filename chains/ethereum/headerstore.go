package ethereum

import (
	"errors"
	"fmt"
	"github.com/mapprotocol/atlas/core/vm"
	"math/big"
	mrand "math/rand"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	lru "github.com/hashicorp/golang-lru"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/tools"
)

const (
	CacheSize = 20
)

var (
	hsCache *Cache
)

func init() {
	hsCache = &Cache{
		size: CacheSize,
	}
	hsCache.Cache, _ = lru.New(hsCache.size)
}

type Cache struct {
	Cache *lru.Cache
	size  int
}

type HeaderStore struct {
	canonicalNumberToHash map[string][]byte
	headers               map[string][]byte
	canonicalHeaders      map[uint64][]byte
	tds                   map[string][]byte
	hash2number           map[common.Hash]uint64
	lastNumber            uint64
	lastHash              common.Hash
	curNumber             uint64
	curHash               common.Hash
	length                uint
}

func headerKey(number uint64, hash common.Hash) string {
	return fmt.Sprintf("%d-%s", number, hash.Hex())
}

//func encodeNumber(number *big.Int) []byte {
//	data, err := rlp.EncodeToBytes(number)
//	if err != nil {
//		log.Crit("Failed to RLP encode block total difficulty", "err", err)
//	}
//	return data
//}
//
func encodeHeader(header *Header) []byte {
	data, err := rlp.EncodeToBytes(header)
	if err != nil {
		log.Crit("Failed to RLP encode header", "err", err)
	}
	return data
}

func NewHeaderStore() *HeaderStore {
	// todo
	return &HeaderStore{}
}

func CloneHeaderStore(src *HeaderStore) (dst *HeaderStore, err error) {
	dst = NewHeaderStore()
	// todo
	//err = DeepCopy()
	//if err != nil {
	//	return nil, err
	//}
	//err = DeepCopy()
	//if err != nil {
	//	return nil, err
	//}
	return dst, nil
}

func (hs *HeaderStore) Store(state vm.StateDB) error {
	var (
		address = chains.EthereumHeaderStoreAddress
		key     = common.BytesToHash(address[:])
	)

	data, err := rlp.EncodeToBytes(hs)
	if err != nil {
		log.Error("Failed to RLP encode HeaderStore", "err", err)
		return err
	}

	state.SetPOWState(address, key, data)

	clone, err := CloneHeaderStore(hs)
	if err != nil {
		return err
	}
	hash := tools.RlpHash(data)
	hsCache.Cache.Add(hash, clone)
	return nil
}

func (hs *HeaderStore) Load(state vm.StateDB) (err error) {
	var (
		h       HeaderStore
		address = chains.EthereumHeaderStoreAddress
		key     = common.BytesToHash(address[:])
	)

	data := state.GetPOWState(address, key)
	hash := tools.RlpHash(data)
	if cc, ok := hsCache.Cache.Get(hash); ok {
		cp, err := CloneHeaderStore(cc.(*HeaderStore))
		if err != nil {
			return err
		}
		h = *cp
		// todo
		//h.height2receiveTimes, h.epoch2syncInfo = hs.height2receiveTimes, hs.epoch2syncInfo
		return nil
	}

	if err := rlp.DecodeBytes(data, &hs); err != nil {
		log.Error("HeaderStore RLP decode failed", "err", err, "HeaderStore", data)
		return fmt.Errorf("HeaderStore RLP decode failed, error: %s", err.Error())
	}

	clone, err := CloneHeaderStore(&h)
	if err != nil {
		return err
	}
	hsCache.Cache.Add(hash, clone)
	// todo
	//h.height2receiveTimes, h.epoch2syncInfo = hs.height2receiveTimes, hs.epoch2syncInfo
	return nil
}

func (hs *HeaderStore) pop() {
	// todo
	numbers := make([]uint64, hs.length)
	for number := range hs.canonicalHeaders {
		numbers = append(numbers, number)
	}
	sort.Slice(numbers, func(i, j int) bool {
		return numbers[i] < numbers[j]
	})

	delete(hs.canonicalHeaders, numbers[0])
	hs.length--
}

//// WriteHeaderNumber stores the hash->number mapping.
//func WriteHeaderNumber(hash common.Hash, number uint64) {
//	key := headerNumberKey(hash)
//	enc := encodeBlockNumber(number)
//	if err := db.Put(key, enc); err != nil {
//		log.Crit("Failed to store hash to number mapping", "err", err)
//	}
//}

func (hs *HeaderStore) WriteHeader(header *Header) {
	hash := header.Hash()
	number := header.Number.Uint64()
	// todo
	// Write the hash -> number mapping
	//WriteHeaderNumber(db, hash, number)
	hs.hash2number[hash] = number

	if !hs.HasHeader(header.Hash(), number) {
		hs.canonicalHeaders[number] = encodeHeader(header)
		hs.length++
		hs.curNumber = number
	}
}

//func (hs *HeaderStore) Push(v interface{}) error {
//	header, ok := v.(Header)
//	if !ok {
//		return errors.New("invalid header")
//	}
//
//	if hs.length >= 2000 {
//		hs.pop()
//	}
//	hs.push(&header)
//	return nil
//}

// numberHash is just a container for a number and a hash, to represent a block
type numberHash struct {
	number uint64
	hash   common.Hash
}

func (hs *HeaderStore) GetTd(hash common.Hash, number uint64) *big.Int {
	// todo
	return big.NewInt(0)
}

func (hs *HeaderStore) HasHeader(hash common.Hash, number uint64) bool {
	return false
}

func (hs *HeaderStore) WriteTd(hash common.Hash, number uint64, td *big.Int) {
	// todo
}

func (hs *HeaderStore) ReadCanonicalHash(number uint64) common.Hash {
	// todo
	return common.Hash{}
}

func (hs *HeaderStore) WriteCanonicalHash(hash common.Hash, number uint64) {

}

func (hs *HeaderStore) DeleteCanonicalHash(number uint64) {
	// todo
}

func (hs *HeaderStore) WriteHeaders(ethHeaders []*interface{}) error {
	if len(ethHeaders) == 0 {
		return nil
	}

	headers := make([]*Header, 0, len(ethHeaders))
	for _, header := range ethHeaders {
		h, ok := (*header).(*Header)
		if !ok {
			return errors.New("invalid header")
		}
		headers = append(headers, h)
	}

	ptd := hs.GetTd(headers[0].ParentHash, headers[0].Number.Uint64()-1)
	if ptd == nil {
		return errUnknownAncestor
	}
	var (
		lastNumber = headers[0].Number.Uint64() - 1 // Last successfully imported number
		lastHash   = headers[0].ParentHash          // Last imported header hash
		newTD      = new(big.Int).Set(ptd)          // Total difficulty of inserted chain

		//lastHeader    *Header
		inserted      []numberHash // Ephemeral lookup of number/hash for the chain
		firstInserted = -1         // Index of the first non-ignored header
	)

	parentKnown := true // Set to true to force hc.HasHeader check the first iteration
	for i, header := range headers {
		var hash common.Hash

		if i < len(headers)-1 {
			hash = headers[i+1].ParentHash
		} else {
			hash = header.Hash()
		}
		number := header.Number.Uint64()
		newTD.Add(newTD, big.NewInt(1))

		alreadyKnown := parentKnown && hs.HasHeader(hash, number)
		if !alreadyKnown {
			hs.WriteTd(hash, number, newTD)
			//hc.tdCache.Add(hash, new(big.Int).Set(newTD))
			hs.WriteHeader(header)
			inserted = append(inserted, numberHash{number, hash})
			//hc.headerCache.Add(hash, header)
			//hc.numberCache.Add(hash, number)
			if firstInserted < 0 {
				firstInserted = i
			}
		}
		parentKnown = alreadyKnown
		lastHash, lastNumber = hash, number
	}

	// todo store

	var (
		head    = hs.CurrentNumber()
		localTD = hs.GetTd(hs.CurrentHash(), head)
	)

	reorg := newTD.Cmp(localTD) > 0
	if !reorg && newTD.Cmp(localTD) == 0 {
		if lastNumber < head {
			reorg = true
		} else if lastNumber == head {
			reorg = mrand.Float64() < 0.5
		}
	}

	chainAlreadyCanon := headers[0].ParentHash == hs.CurrentHash()
	if reorg {
		if !chainAlreadyCanon {
			// 删除 head 之后的
			for i := lastNumber + 1; ; i++ {
				hash := hs.ReadCanonicalHash(i)
				if hash == (common.Hash{}) {
					break
				}
				hs.DeleteCanonicalHash(i)
			}

			// 覆盖 head 之前的
			var (
				headHash   = headers[0].ParentHash          // inserted[0].parent?
				headNumber = headers[0].Number.Uint64() - 1 // inserted[0].num-1 ?
				// 'h' + number + hash
				//headHeader = GetHeader(headHash, headNumber)
				headHeader = hs.GetHeader(headHash, headNumber)
			)
			// 'h' + number + 'n'
			for hs.ReadCanonicalHash(headNumber) != headHash {
				// 'h' + number + 'n'
				hs.WriteCanonicalHash(headHash, headNumber)
				headHash = headHeader.ParentHash
				headNumber = headHeader.Number.Uint64() - 1
				// 'h' + number + hash
				headHeader = hs.GetHeader(headHash, headNumber)
			}
			// todo ???
			// If some of the older headers were already known, but obtained canon-status
			// during this import batch, then we need to write that now
			// Further down, we continue writing the status for the ones that
			// were not already known
			for i := 0; i < firstInserted; i++ {
				hash := headers[i].Hash()
				num := headers[i].Number.Uint64()
				hs.WriteCanonicalHash(hash, num)
				//rawdb.WriteHeadHeaderHash(markerBatch, hash)
			}
		}
		// Extend the canonical chain with the new headers
		for _, hn := range inserted {
			hs.WriteCanonicalHash(hn.hash, hn.number)
			//rawdb.WriteHeadHeaderHash(markerBatch, hn.hash)
		}

		// todo store

		hs.curHash = lastHash
	}

	return nil
}

//func (hs *HeaderStore) HasHeader(number uint64) bool {
//	_, exist := hs.headers[number]
//	return exist
//}

func (hs *HeaderStore) CurrentNumber() uint64 {
	return hs.curNumber
}

func (hs *HeaderStore) CurrentHash() common.Hash {
	return hs.curHash
}

// GetHeader todo 分叉处理
func (hs *HeaderStore) GetHeader(hash common.Hash, number uint64) *Header {
	// todo
	return &Header{}
}
