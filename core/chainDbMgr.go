package core

import (
	"fmt"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/mapprotocol/atlas/atlasdb"
	"github.com/mapprotocol/atlas/core/rawdb"
	ethHeader "github.com/mapprotocol/atlas/core/vm/sync"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	mrand "math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var (
	StoreMgr *HeaderChainStore
)

func GetStoreMgr(chainType rawdb.ChainType) *HeaderChainStore {
	StoreMgr.currentChainType = chainType
	return StoreMgr
}

const (
	DefaultChainType = rawdb.ChainType(0)
)

type HeaderChainStore struct {
	chainDb           atlasdb.Database
	currentChainType  rawdb.ChainType
	currentHeaderHash common.Hash
	currentHeader     atomic.Value // Current head of the header chain (may be above the block chain!)
	Mu                sync.RWMutex // blockchaindb insertion lock
	rand              *mrand.Rand
}

func OpenDatabase(file string, cache, handles int) (atlasdb.Database, error) {
	return atlasdb.NewLDBDatabase(file, 10, 10)
}

func NewStoreDb(ctx *cli.Context, config *ethconfig.Config) {
	path := node.DefaultDataDir()
	if ctx.GlobalIsSet(utils.DataDirFlag.Name) {
		path = ctx.GlobalString(utils.DataDirFlag.Name)
	}
	chainDb, _ := OpenDatabase(path, config.DatabaseCache, config.DatabaseHandles)
	db := &HeaderChainStore{
		chainDb:          chainDb,
		currentChainType: DefaultChainType,
	}
	db.SetHead(0)
	db.currentHeader.Store(db.GetHeaderByNumber(0))
	db.currentHeaderHash = db.CurrentHeader().Hash()
	StoreMgr = db
}

func (lc *HeaderChainStore) SetHead(head uint64) {
	lc.Mu.Lock()
	defer lc.Mu.Unlock()
}

func (db *HeaderChainStore) SetChainType(m rawdb.ChainType) {
	db.currentChainType = m
}

func (db *HeaderChainStore) ReadHeader(Hash common.Hash, number uint64) *ethHeader.ETHHeader {
	return rawdb.ReadHeader(db.chainDb, Hash, number, db.currentChainType)
}

func (db *HeaderChainStore) WriteHeader(header *ethHeader.ETHHeader) {
	batch := db.chainDb.NewBatch()
	// Flush all accumulated deletions.
	if err := batch.Write(); err != nil {
		log.Crit("Failed to rewind block", "error", err)
	}
	rawdb.WriteHeader(db.chainDb, header, db.currentChainType)
}
func (db *HeaderChainStore) DeleteHeader(hash common.Hash, number uint64) {
	rawdb.DeleteHeader(db.chainDb, hash, number, db.currentChainType)
}

func (hc *HeaderChainStore) InsertHeaderChain(chains []*ethHeader.ETHHeader, start time.Time) (WriteStatus, error) {
	res, err := hc.writeHeaders(chains)

	// Report some public statistics so the user has a clue what's going on
	context := []interface{}{
		"count", res.imported,
		"elapsed", common.PrettyDuration(time.Since(start)),
	}
	if err != nil {
		context = append(context, "err", err)
	}
	if last := res.lastHeader; last != nil {
		context = append(context, "number", last.Number, "hash", res.lastHash)
		if timestamp := time.Unix(int64(last.Time), 0); time.Since(timestamp) > time.Minute {
			context = append(context, []interface{}{"age", common.PrettyAge(timestamp)}...)
		}
	}
	if res.ignored > 0 {
		context = append(context, []interface{}{"ignored", res.ignored}...)
	}
	log.Info("Imported new block headers", context...)
	return res.status, err
}

// GetBlockNumber retrieves the block number belonging to the given hash
// from the cache or database
func (hc *HeaderChainStore) GetBlockNumber(hash common.Hash) *uint64 {
	number := rawdb.ReadHeaderNumber(hc.chainDb, hash, hc.currentChainType)
	return number
}

// WriteStatus status of write
type WriteStatus byte

const (
	NonStatTy   WriteStatus = iota // the no
	CanonStatTy                    // the Canonical
	SideStatTy                     // the branch
)

type headerWriteResult struct {
	status     WriteStatus
	ignored    int
	imported   int
	lastHash   common.Hash
	lastHeader *ethHeader.ETHHeader
}

// numberHash is just a container for a number and a hash, to represent a block
type numberHash struct {
	number uint64
	hash   common.Hash
}

func (hc *HeaderChainStore) GetTd(hash common.Hash, number uint64) *big.Int {
	td := rawdb.ReadTd(hc.chainDb, hash, number, hc.currentChainType)
	if td == nil {
		return nil
	}
	return td
}
func (hc *HeaderChainStore) HasHeader(hash common.Hash, number uint64) bool {
	return rawdb.HasHeader(hc.chainDb, hash, number, hc.currentChainType)
}
func (hc *HeaderChainStore) CurrentHeader() *ethHeader.ETHHeader {
	return hc.currentHeader.Load().(*ethHeader.ETHHeader)
}

func (hc *HeaderChainStore) GetHeader(hash common.Hash, number uint64) *ethHeader.ETHHeader {

	header := rawdb.ReadHeader(hc.chainDb, hash, number, hc.currentChainType)
	if header == nil {
		return nil
	}

	return header
}

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyHeader(h *ethHeader.ETHHeader) *ethHeader.ETHHeader {
	cpy := *h
	if cpy.Difficulty = new(big.Int); h.Difficulty != nil {
		cpy.Difficulty.Set(h.Difficulty)
	}
	if cpy.Number = new(big.Int); h.Number != nil {
		cpy.Number.Set(h.Number)
	}
	if len(h.Extra) > 0 {
		cpy.Extra = make([]byte, len(h.Extra))
		copy(cpy.Extra, h.Extra)
	}
	return &cpy
}

func (hc *HeaderChainStore) writeHeaders(headers []*ethHeader.ETHHeader) (result *headerWriteResult, err error) {
	if len(headers) == 0 {
		return &headerWriteResult{}, nil
	}
	ptd := hc.GetTd(headers[0].ParentHash, headers[0].Number.Uint64()-1)
	if ptd == nil {
		return &headerWriteResult{}, consensus.ErrUnknownAncestor
	}
	var (
		lastNumber = headers[0].Number.Uint64() - 1 // Last successfully imported number
		lastHash   = headers[0].ParentHash          // Last imported header hash
		newTD      = new(big.Int).Set(ptd)          // Total difficulty of inserted chain

		lastHeader    *ethHeader.ETHHeader
		inserted      []numberHash // Ephemeral lookup of number/hash for the chain
		firstInserted = -1         // Index of the first non-ignored header
	)

	batch := hc.chainDb.NewBatch()
	for i, header := range headers {
		var hash common.Hash
		// The headers have already been validated at this point, so we already
		// know that it's a contiguous chain, where
		// headers[i].Hash() == headers[i+1].ParentHash
		if i < len(headers)-1 {
			hash = headers[i+1].ParentHash
		} else {
			hash = header.Hash()
		}
		number := header.Number.Uint64()
		newTD.Add(newTD, header.Difficulty)

		// If the header is already known, skip it, otherwise store
		if !hc.HasHeader(hash, number) {
			// Irrelevant of the canonical status, write the TD and header to the database.
			rawdb.WriteTd(batch, hash, number, newTD, hc.currentChainType)

			rawdb.WriteHeader(batch, header, hc.currentChainType)
			inserted = append(inserted, numberHash{number, hash})

			if firstInserted < 0 {
				firstInserted = i
			}
		}
		lastHeader, lastHash, lastNumber = header, hash, number
	}

	// Commit to disk!
	if err := batch.Write(); err != nil {
		log.Crit("Failed to write headers", "error", err)
	}
	batch.Reset()

	var (
		head    = hc.CurrentHeader().Number.Uint64()
		localTD = hc.GetTd(hc.currentHeaderHash, head)
		status  = SideStatTy
	)
	// If the total difficulty is higher than our known, add it to the canonical chain
	// Second clause in the if statement reduces the vulnerability to selfish mining.
	// Please refer to http://www.cs.cornell.edu/~ie53/publications/btcProcFC.pdf
	reorg := newTD.Cmp(localTD) > 0
	if !reorg && newTD.Cmp(localTD) == 0 {
		if lastNumber < head {
			reorg = true
		} else if lastNumber == head {
			reorg = mrand.Float64() < 0.5 //Random decision
		}
	}
	// If the parent of the (first) block is already the canon header,
	// we don't have to go backwards to delete canon blocks, but
	// simply pile them onto the existing chain
	chainAlreadyCanon := headers[0].ParentHash == hc.currentHeaderHash
	if reorg {
		// If the header can be added into canonical chain, adjust the
		// header chain markers(canonical indexes and head header flag).
		//
		// Note all markers should be written atomically.
		markerBatch := batch // we can reuse the batch to keep allocs down
		if !chainAlreadyCanon {
			// Delete any canonical number assignments above the new head
			for i := lastNumber + 1; ; i++ {
				hash := rawdb.ReadCanonicalHash(hc.chainDb, i, hc.currentChainType)
				if hash == (common.Hash{}) {
					break
				}
				rawdb.DeleteCanonicalHash(markerBatch, i, hc.currentChainType)
			}
			// Overwrite any stale canonical number assignments, going
			// backwards from the first header in this import
			var (
				headHash   = headers[0].ParentHash          // inserted[0].parent?
				headNumber = headers[0].Number.Uint64() - 1 // inserted[0].num-1 ?
				headHeader = hc.GetHeader(headHash, headNumber)
			)
			for rawdb.ReadCanonicalHash(hc.chainDb, headNumber, hc.currentChainType) != headHash {
				rawdb.WriteCanonicalHash(markerBatch, headHash, headNumber, hc.currentChainType)
				headHash = headHeader.ParentHash
				headNumber = headHeader.Number.Uint64() - 1
				headHeader = hc.GetHeader(headHash, headNumber)
			}
			// If some of the older headers were already known, but obtained canon-status
			// during this import batch, then we need to write that now
			// Further down, we continue writing the staus for the ones that
			// were not already known
			for i := 0; i < firstInserted; i++ {
				hash := headers[i].Hash()
				num := headers[i].Number.Uint64()
				rawdb.WriteCanonicalHash(markerBatch, hash, num, hc.currentChainType)
				rawdb.WriteHeadHeaderHash(markerBatch, hash, hc.currentChainType)
			}
		}
		// Extend the canonical chain with the new headers
		for _, hn := range inserted {
			rawdb.WriteCanonicalHash(markerBatch, hn.hash, hn.number, hc.currentChainType)
			rawdb.WriteHeadHeaderHash(markerBatch, hn.hash, hc.currentChainType)
		}
		if err := markerBatch.Write(); err != nil {
			log.Crit("Failed to write header markers into disk", "err", err)
		}
		markerBatch.Reset()
		// Last step update all in-memory head header markers
		hc.currentHeaderHash = lastHash
		hc.currentHeader.Store(CopyHeader(lastHeader))

		// Chain status is canonical since this insert was a reorg.
		// Note that all inserts which have higher TD than existing are 'reorg'.
		status = CanonStatTy
	}

	if len(inserted) == 0 {
		status = NonStatTy
	}
	return &headerWriteResult{
		status:     status,
		ignored:    len(headers) - len(inserted),
		imported:   len(inserted),
		lastHash:   lastHash,
		lastHeader: lastHeader,
	}, nil
}

func (hc *HeaderChainStore) ValidateHeaderChain(chain []*ethHeader.ETHHeader, checkFreq int) (int, error) {
	// Do a sanity check that the provided chain is actually ordered and linked
	for i := 1; i < len(chain); i++ {
		if chain[i].Number.Uint64() != chain[i-1].Number.Uint64()+1 {
			hash := chain[i].Hash()
			parentHash := chain[i-1].Hash()
			// Chain broke ancestry, log a message (programming error) and skip insertion
			log.Error("Non contiguous header insert", "number", chain[i].Number, "hash", hash,
				"parent", chain[i].ParentHash, "prevnumber", chain[i-1].Number, "prevhash", parentHash)

			return 0, fmt.Errorf("non contiguous insert: item %d is #%d [%x..], item %d is #%d [%x..] (parent [%x..])", i-1, chain[i-1].Number,
				parentHash.Bytes()[:4], i, chain[i].Number, hash.Bytes()[:4], chain[i].ParentHash[:4])
		}

	}

	// Generate the list of seal verification requests, and start the parallel verifier
	seals := make([]bool, len(chain))
	if checkFreq != 0 {
		// In case of checkFreq == 0 all seals are left false.
		for i := 0; i <= len(seals)/checkFreq; i++ {
			index := i*checkFreq + hc.rand.Intn(checkFreq)
			if index >= len(seals) {
				index = len(seals) - 1
			}
			seals[index] = true
		}
		// Last should always be verified to avoid junk.
		seals[len(seals)-1] = true
	}

	// todo Validate

	return 0, nil
}

// GetBlockHashesFromHash retrieves a number of block hashes starting at a given
// hash, fetching towards the genesis block.
func (hc *HeaderChainStore) GetBlockHashesFromHash(hash common.Hash, max uint64) []common.Hash {
	// Get the origin header from which to fetch
	header := hc.GetHeaderByHash(hash)
	if header == nil {
		return nil
	}
	// Iterate the headers until enough is collected or the genesis reached
	chain := make([]common.Hash, 0, max)
	for i := uint64(0); i < max; i++ {
		next := header.ParentHash
		if header = hc.GetHeader(next, header.Number.Uint64()-1); header == nil {
			break
		}
		chain = append(chain, next)
		if header.Number.Sign() == 0 {
			break
		}
	}
	return chain
}

// GetTdByHash retrieves a block's total difficulty in the canonical chain from the
// database by hash, caching it if found.
func (hc *HeaderChainStore) GetTdByHash(hash common.Hash) *big.Int {
	number := hc.GetBlockNumber(hash)
	if number == nil {
		return nil
	}
	return hc.GetTd(hash, *number)
}

// GetHeaderByHash retrieves a block header from the database by hash, caching it if
// found.
func (hc *HeaderChainStore) GetHeaderByHash(hash common.Hash) *ethHeader.ETHHeader {
	number := hc.GetBlockNumber(hash)
	if number == nil {
		return nil
	}
	return hc.GetHeader(hash, *number)
}

// GetHeaderByNumber retrieves a block header from the database by number,
// caching it (associated with its hash) if found.
func (hc *HeaderChainStore) GetHeaderByNumber(number uint64) *ethHeader.ETHHeader {
	hash := rawdb.ReadCanonicalHash(hc.chainDb, number, hc.currentChainType)
	if hash == (common.Hash{}) {
		return nil
	}
	return hc.GetHeader(hash, number)
}

func (hc *HeaderChainStore) GetCanonicalHash(number uint64) common.Hash {
	return rawdb.ReadCanonicalHash(hc.chainDb, number, hc.currentChainType)
}

// SetCurrentHeader sets the in-memory head header marker of the canonical chan
// as the given header.
func (hc *HeaderChainStore) SetCurrentHeader(head *ethHeader.ETHHeader) {
	hc.currentHeader.Store(head)
	hc.currentHeaderHash = head.Hash()
}
