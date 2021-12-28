package ethereum

import (
	"bytes"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	lru "github.com/hashicorp/golang-lru"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/tools"
)

const (
	StoreCacheSize      = 20
	MaxHeaderLimit = 5
	SplicingSymbol = "-"
)

var (
	storeCache *Cache
)

func init() {
	storeCache = &Cache{
		size: StoreCacheSize,
	}
	storeCache.Cache, _ = lru.New(storeCache.size)
}

type Cache struct {
	Cache *lru.Cache
	size  int
}

type HeaderStore struct {
	canonicalNumberToHash map[uint64]common.Hash
	headers               map[string][]byte
	tds                   map[string]*big.Int
	//hash2number map[common.Hash]uint64
	curNumber uint64
	curHash   common.Hash
}

func headerKey(number uint64, hash common.Hash) string {
	return fmt.Sprintf("%d%s%s", number, SplicingSymbol, hash.Hex())
}

func (hs *HeaderStore) delOldHeaders() {
	length := len(hs.headers)
	if length <= MaxHeaderLimit {
		return
	}

	numbers := make([]uint64, 0, length)
	number2key := make(map[uint64][]string)
	for key := range hs.headers {
		numberStr := strings.Split(key, SplicingSymbol)[0]
		number, _ := strconv.ParseUint(numberStr, 10, 64)
		numbers = append(numbers, number)
		number2key[number] = append(number2key[number], key)
	}

	sort.Slice(numbers, func(i, j int) bool {
		return numbers[i] < numbers[j]
	})

	delTotal := length - MaxHeaderLimit
	for i := 0; i < delTotal; i++ {
		number := numbers[i]
		for _, key := range number2key[number] {
			delete(hs.headers, key)
			delete(hs.tds, key)
		}
	}
}

func encodeHeader(header *Header) []byte {
	data, err := rlp.EncodeToBytes(header)
	if err != nil {
		log.Crit("Failed to RLP encode header", "err", err)
	}
	return data
}

func decodeHeader(data []byte, hash common.Hash) *Header {
	header := new(Header)
	if err := rlp.Decode(bytes.NewReader(data), header); err != nil {
		log.Error("Invalid block header RLP", "hash", hash, "err", err)
		return nil
	}
	return header
}

func NewHeaderStore() *HeaderStore {
	return &HeaderStore{
		canonicalNumberToHash: make(map[uint64]common.Hash),
		headers:               make(map[string][]byte),
		tds:                   make(map[string]*big.Int),
	}
}

func cloneHeaderStore(src *HeaderStore) (dst *HeaderStore, err error) {
	dst = NewHeaderStore()
	if err := tools.DeepCopy(src, dst); err != nil {
		return nil, err
	}
	return dst, nil
}

func (hs *HeaderStore) Store(state types.StateDB) error {
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

	clone, err := cloneHeaderStore(hs)
	if err != nil {
		return err
	}
	hash := tools.RlpHash(data)
	storeCache.Cache.Add(hash, clone)
	return nil
}

func (hs *HeaderStore) Load(state types.StateDB) (err error) {
	var (
		h       HeaderStore
		address = chains.EthereumHeaderStoreAddress
		key     = common.BytesToHash(address[:])
	)

	data := state.GetPOWState(address, key)
	hash := tools.RlpHash(data)
	if cc, ok := storeCache.Cache.Get(hash); ok {
		cp, err := cloneHeaderStore(cc.(*HeaderStore))
		if err != nil {
			return err
		}
		h = *cp
		hs.canonicalNumberToHash, hs.headers, hs.tds, hs.curHash, hs.curNumber = h.canonicalNumberToHash, h.headers, h.tds, h.curHash, h.curNumber
		return nil
	}

	if err := rlp.DecodeBytes(data, &hs); err != nil {
		log.Error("HeaderStore RLP decode failed", "err", err, "HeaderStore", data)
		return fmt.Errorf("HeaderStore RLP decode failed, error: %s", err.Error())
	}

	clone, err := cloneHeaderStore(&h)
	if err != nil {
		return err
	}
	storeCache.Cache.Add(hash, clone)
	hs.canonicalNumberToHash, hs.headers, hs.tds, hs.curHash, hs.curNumber = h.canonicalNumberToHash, h.headers, h.tds, h.curHash, h.curNumber
	return nil
}

func (hs *HeaderStore) WriteHeader(header *Header) {
	var (
		hash   = header.Hash()
		number = header.Number.Uint64()
	)

	// Write the hash -> number mapping
	//hs.hash2number[hash] = number

	//if !hs.HasHeader(hash, number) {
	hs.headers[headerKey(number, hash)] = encodeHeader(header)
	//}
}

// numberHash is just a container for a number and a hash, to represent a block
type numberHash struct {
	number uint64
	hash   common.Hash
}

func (hs *HeaderStore) GetTd(hash common.Hash, number uint64) *big.Int {
	return hs.tds[headerKey(number, hash)]
}

func (hs *HeaderStore) HasHeader(hash common.Hash, number uint64) bool {
	_, isExist := hs.headers[headerKey(number, hash)]
	return isExist
}

func (hs *HeaderStore) WriteTd(hash common.Hash, number uint64, td *big.Int) {
	hs.tds[headerKey(number, hash)] = td
}

func (hs *HeaderStore) ReadCanonicalHash(number uint64) common.Hash {
	return hs.canonicalNumberToHash[number]
}

func (hs *HeaderStore) WriteCanonicalHash(hash common.Hash, number uint64) {
	// number -> hash mapping
	hs.canonicalNumberToHash[number] = hash
}

func (hs *HeaderStore) DeleteCanonicalHash(number uint64) {
	delete(hs.canonicalNumberToHash, number)
}

func (hs *HeaderStore) WriteHeaders(db types.StateDB, ethHeaders []byte) (int, error) {
	if len(ethHeaders) == 0 {
		return 0, nil
	}

	var headers []*Header
	if err := rlp.DecodeBytes(ethHeaders, &hs); err != nil {
		log.Error("rlp decode failed.", "err", err)
		return 0, chains.ErrRLPDecode
	}

	ptd := hs.GetTd(headers[0].ParentHash, headers[0].Number.Uint64()-1)
	if ptd == nil {
		return 0, errUnknownAncestor
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
			hs.WriteHeader(header)

			inserted = append(inserted, numberHash{number, hash})
			if firstInserted < 0 {
				firstInserted = i
			}
		}
		parentKnown = alreadyKnown
		lastHash, lastNumber = hash, number
	}

	var (
		head    = hs.curNumber
		localTD = hs.GetTd(hs.curHash, head)
	)

	reorg := newTD.Cmp(localTD) > 0
	if !reorg && newTD.Cmp(localTD) == 0 {
		if lastNumber < head {
			reorg = true
		} else if lastNumber == head {
			reorg = rand.Float64() < 0.5
		}
	}

	// If the parent of the (first) block is already the canon header,
	// we don't have to go backwards to delete canon blocks, but
	// simply pile them onto the existing chain
	chainAlreadyCanon := headers[0].ParentHash == hs.curHash
	if reorg {
		if !chainAlreadyCanon {
			for i := lastNumber + 1; ; i++ {
				hash := hs.ReadCanonicalHash(i)
				if hash == (common.Hash{}) {
					break
				}
				hs.DeleteCanonicalHash(i)
			}

			var (
				headHash   = headers[0].ParentHash          // inserted[0].parent?
				headNumber = headers[0].Number.Uint64() - 1 // inserted[0].num-1 ?
			)
			headHeader := hs.GetHeader(headHash, headNumber)
			if headHeader == nil {
				return 0, fmt.Errorf("not found header, number: %d, hash: %s", headNumber, headHash)
			}
			for hs.ReadCanonicalHash(headNumber) != headHash {
				hs.WriteCanonicalHash(headHash, headNumber)
				headHash = headHeader.ParentHash
				headNumber = headHeader.Number.Uint64() - 1
				headHeader = hs.GetHeader(headHash, headNumber)
				if headHeader == nil {
					return 0, fmt.Errorf("not found header, number: %d, hash: %s", headNumber, headHash)
				}
			}

			// If some of the older headers were already known, but obtained canon-status
			// during this import batch, then we need to write that now
			// Further down, we continue writing the status for the ones that
			// were not already known
			for i := 0; i < firstInserted; i++ {
				hash := headers[i].Hash()
				num := headers[i].Number.Uint64()
				hs.WriteCanonicalHash(hash, num)
			}
		}
		// Extend the canonical chain with the new headers
		for _, hn := range inserted {
			hs.WriteCanonicalHash(hn.hash, hn.number)
		}

		hs.delOldHeaders()
		if err := hs.Store(db); err != nil {
			return 0, err
		}
		hs.curHash = lastHash
	}

	return len(inserted), nil
}

func (hs *HeaderStore) CurrentNumber() uint64 {
	return hs.curNumber
}

func (hs *HeaderStore) CurrentHash() common.Hash {
	return hs.curHash
}

func (hs *HeaderStore) GetHeader(hash common.Hash, number uint64) *Header {
	data := hs.headers[headerKey(number, hash)]
	if len(data) != 0 {
		return decodeHeader(data, hash)
	}
	return nil
}

func (hs *HeaderStore) GetHashByNumber(number uint64) common.Hash {
	return hs.ReadCanonicalHash(number)
}
