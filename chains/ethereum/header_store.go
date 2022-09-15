package ethereum

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	ethrlp "github.com/ethereum/go-ethereum/rlp"
	lru "github.com/hashicorp/golang-lru"
	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/params"
	"github.com/mapprotocol/atlas/tools"
	"github.com/mapprotocol/atlas/tools/rlp"
)

const (
	StoreCacheSize = 20
	MaxHeaderLimit = 200000
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

// WriteStatus status of write
type WriteStatus byte

const (
	NonStatTy WriteStatus = iota
	CanonStatTy
	SideStatTy
)

type Cache struct {
	Cache *lru.Cache
	size  int
}

type HeaderStore struct {
	CanonicalNumberToHash map[uint64]common.Hash
	Headers               map[string][]byte
	TDs                   map[string]*big.Int
	CurNumber             uint64
	CurHash               common.Hash
}

func headerKey(number uint64, hash common.Hash) string {
	return fmt.Sprintf("%d%s%s", number, SplicingSymbol, hash.Hex())
}

func (hs *HeaderStore) delOldHeaders() {
	length := len(hs.Headers)
	if length <= MaxHeaderLimit {
		return
	}

	numbers := make([]uint64, 0, length)
	number2key := make(map[uint64][]string)
	for key := range hs.Headers {
		numberStr := strings.Split(key, SplicingSymbol)[0]
		number, _ := strconv.ParseUint(numberStr, 10, 64)
		numbers = append(numbers, number)
		number2key[number] = append(number2key[number], key)
	}

	sort.Slice(numbers, func(i, j int) bool {
		return numbers[i] < numbers[j]
	})

	delTotal := length - MaxHeaderLimit
	for i := 0; i < delTotal; {
		number := numbers[i]
		for _, key := range number2key[number] {
			delete(hs.Headers, key)
			delete(hs.TDs, key)
			log.Info("deleted ethereum header", "number", number)
			i++
		}
	}
	log.Info("before cleaning up the old ethereum headers", "headers length", length)
	log.Info("after cleaning up the old ethereum headers", "headers length", len(hs.Headers))
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
		CanonicalNumberToHash: make(map[uint64]common.Hash),
		Headers:               make(map[string][]byte),
		TDs:                   make(map[string]*big.Int),
	}
}

func InitHeaderStore(state types.StateDB, header *Header, td *big.Int) error {
	hash := header.Hash()
	number := header.Number.Uint64()
	key := headerKey(number, hash)

	h := &HeaderStore{
		CanonicalNumberToHash: map[uint64]common.Hash{
			number: hash,
		},
		Headers: map[string][]byte{
			key: encodeHeader(header),
		},
		TDs: map[string]*big.Int{
			key: td,
		},
		CurHash:   hash,
		CurNumber: number,
	}
	return h.Store(state)
}

func (hs *HeaderStore) ResetHeaderStore(state types.StateDB, ethHeaders []byte, td *big.Int) error {
	var header Header
	if err := ethrlp.DecodeBytes(ethHeaders, &header); err != nil {
		log.Error("rlp decode ethereum header failed.", "err", err)
		return chains.ErrRLPDecode
	}
	// pointer given to Decode must not be nil
	hash := header.Hash()
	number := header.Number.Uint64()
	key := headerKey(number, hash)

	h := &HeaderStore{
		CanonicalNumberToHash: map[uint64]common.Hash{
			number: hash,
		},
		Headers: map[string][]byte{
			key: encodeHeader(&header),
		},
		TDs: map[string]*big.Int{
			key: td,
		},
		CurHash:   hash,
		CurNumber: number,
	}
	return h.Store(state)
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

	//newData := make([]byte, 0, len(data))
	//for _, d := range data {
	//	newData = append(newData, d)
	//}

	state.SetPOWState(address, key, data)

	clone, err := cloneHeaderStore(hs)
	if err != nil {
		return err
	}
	log.Info("data", "len", len(data))
	total := int64(0)
	for _, v := range hs.Headers {
		total += int64(len(v))
	}
	log.Info("headers --------- ", "total", total, "headers", len(hs.Headers))
	hash := tools.RlpHash(data)
	storeCache.Cache.Add(hash, clone)
	data = data[:0]
	return nil
}

func (hs *HeaderStore) Load(state types.StateDB) (err error) {
	var (
		h       HeaderStore
		address = chains.EthereumHeaderStoreAddress
		key     = common.BytesToHash(address[:])
	)

	data := state.GetPOWState(address, key)
	if len(data) == 0 {
		return errors.New("please initialize header store")
	}

	hash := tools.RlpHash(data)
	if cc, ok := storeCache.Cache.Get(hash); ok {
		cp, err := cloneHeaderStore(cc.(*HeaderStore))
		if err != nil {
			return err
		}
		h = *cp
		hs.CurHash, hs.CurNumber = h.CurHash, h.CurNumber
		hs.CanonicalNumberToHash, hs.Headers, hs.TDs = h.CanonicalNumberToHash, h.Headers, h.TDs
		return nil
	}

	if err := ethrlp.DecodeBytes(data, &h); err != nil {
		log.Error("HeaderStore RLP decode failed", "err", err)
		return fmt.Errorf("HeaderStore RLP decode failed, error: %s", err.Error())
	}

	clone, err := cloneHeaderStore(&h)
	if err != nil {
		return err
	}
	storeCache.Cache.Add(hash, clone)
	hs.CurHash, hs.CurNumber = h.CurHash, h.CurNumber
	hs.CanonicalNumberToHash, hs.Headers, hs.TDs = h.CanonicalNumberToHash, h.Headers, h.TDs
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
	hs.Headers[headerKey(number, hash)] = encodeHeader(header)
	//}
}

func (hs *HeaderStore) GetTd(hash common.Hash, number uint64) *big.Int {
	return hs.TDs[headerKey(number, hash)]
}

func (hs *HeaderStore) HasHeader(hash common.Hash, number uint64) bool {
	_, isExist := hs.Headers[headerKey(number, hash)]
	return isExist
}

func (hs *HeaderStore) WriteTd(hash common.Hash, number uint64, td *big.Int) {
	hs.TDs[headerKey(number, hash)] = new(big.Int).SetUint64(td.Uint64())
}

func (hs *HeaderStore) ReadCanonicalHash(number uint64) common.Hash {
	return hs.CanonicalNumberToHash[number]
}

func (hs *HeaderStore) WriteCanonicalHash(hash common.Hash, number uint64) {
	// number -> hash mapping
	hs.CanonicalNumberToHash[number] = hash
}

func (hs *HeaderStore) DeleteCanonicalHash(number uint64) {
	delete(hs.CanonicalNumberToHash, number)
}

type headerWriteResult struct {
	status     WriteStatus
	ignored    int
	imported   []*params.NumberHash
	lastHash   common.Hash
	lastNumber uint64
}

func (hs *HeaderStore) InsertHeaders(db types.StateDB, ethHeaders []byte) ([]*params.NumberHash, error) {
	start := time.Now()
	res, err := hs.WriteHeaders(db, ethHeaders)

	// Report some public statistics so the user has a clue what's going on
	context := []interface{}{
		"count", len(res.imported),
		"elapsed", common.PrettyDuration(time.Since(start)),
	}
	if err != nil {
		context = append(context, "err", err)
	}

	if res.lastNumber != 0 {
		context = append(context, "number", res.lastNumber, "hash", res.lastHash)
	}
	if res.ignored > 0 {
		context = append(context, []interface{}{"ignored", res.ignored}...)
	}
	log.Info("stored new ethereum block headers", context...)
	return res.imported, err
}

func (hs *HeaderStore) WriteHeaders(db types.StateDB, ethHeaders []byte) (*headerWriteResult, error) {
	var headers []*Header
	if err := ethrlp.DecodeBytes(ethHeaders, &headers); err != nil {
		log.Error("rlp decode ethereum headers failed.", "err", err)
		return &headerWriteResult{}, chains.ErrRLPDecode
	}
	if len(headers) == 0 {
		return &headerWriteResult{}, nil
	}

	if err := hs.Load(db); err != nil {
		return &headerWriteResult{}, err
	}
	ptd := hs.GetTd(headers[0].ParentHash, headers[0].Number.Uint64()-1)
	if ptd == nil {
		return &headerWriteResult{}, errUnknownAncestor
	}
	var (
		lastNumber = headers[0].Number.Uint64() - 1 // Last successfully imported number
		lastHash   = headers[0].ParentHash          // Last imported header hash
		newTD      = new(big.Int).Set(ptd)          // Total difficulty of inserted chain

		inserted      []*params.NumberHash // Ephemeral lookup of number/hash for the chain
		firstInserted = -1                 // Index of the first non-ignored header
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
		newTD.Add(newTD, header.Difficulty)

		alreadyKnown := parentKnown && hs.HasHeader(hash, number)
		if !alreadyKnown {
			hs.WriteTd(hash, number, newTD)
			hs.WriteHeader(header)

			inserted = append(inserted, &params.NumberHash{Number: number, Hash: hash})
			if firstInserted < 0 {
				firstInserted = i
			}
		}
		parentKnown = alreadyKnown
		lastHash, lastNumber = hash, number
	}

	var (
		head    = hs.CurNumber
		localTD = hs.GetTd(hs.CurHash, head)
		status  = SideStatTy
	)

	reorg := newTD.Cmp(localTD) > 0
	if !reorg && newTD.Cmp(localTD) == 0 {
		if lastNumber < head {
			reorg = true
		} else if lastNumber == head {
			//reorg = rand.Float64() < 0.5
			reorg = true
		}
	}

	// If the parent of the (first) block is already the canon header,
	// we don't have to go backwards to delete canon blocks, but
	// simply pile them onto the existing chain
	chainAlreadyCanon := headers[0].ParentHash == hs.CurHash
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
				return &headerWriteResult{}, fmt.Errorf("not found header, number: %d, hash: %s", headNumber, headHash)
			}
			for hs.ReadCanonicalHash(headNumber) != headHash {
				hs.WriteCanonicalHash(headHash, headNumber)
				headHash = headHeader.ParentHash
				headNumber = headHeader.Number.Uint64() - 1
				headHeader = hs.GetHeader(headHash, headNumber)
				if headHeader == nil {
					return &headerWriteResult{}, fmt.Errorf("not found header, number: %d, hash: %s", headNumber, headHash)
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
			hs.WriteCanonicalHash(hn.Hash, hn.Number)
		}

		hs.delOldHeaders()
		hs.CurHash = lastHash
		hs.CurNumber = lastNumber

		// Chain status is canonical since this insert was a reorg.
		// Note that all inserts which have higher TD than existing are 'reorg'.
		status = CanonStatTy
	}
	if err := hs.Store(db); err != nil {
		return &headerWriteResult{}, err
	}
	if len(inserted) == 0 {
		status = NonStatTy
	}
	return &headerWriteResult{
		status:     status,
		ignored:    len(headers) - len(inserted),
		imported:   inserted,
		lastHash:   lastHash,
		lastNumber: lastNumber,
	}, nil
}

func (hs *HeaderStore) CurrentNumber() uint64 {
	return hs.CurNumber
}

func (hs *HeaderStore) CurrentHash() common.Hash {
	return hs.CurHash
}

func (hs *HeaderStore) GetHeader(hash common.Hash, number uint64) *Header {
	data := hs.Headers[headerKey(number, hash)]
	if len(data) != 0 {
		return decodeHeader(data, hash)
	}
	return nil
}

func (hs *HeaderStore) GetHeaderByNumber(number uint64) *Header {
	hash := hs.ReadCanonicalHash(number)
	return hs.GetHeader(hash, number)
}

func (hs *HeaderStore) GetCurrentNumberAndHash(db types.StateDB) (uint64, common.Hash, error) {
	if err := hs.Load(db); err != nil {
		return 0, common.Hash{}, err
	}
	return hs.CurNumber, hs.CurHash, nil
}

func (hs *HeaderStore) GetHashByNumber(db types.StateDB, number uint64) (common.Hash, error) {
	if err := hs.Load(db); err != nil {
		return common.Hash{}, err
	}
	return hs.ReadCanonicalHash(number), nil
}
