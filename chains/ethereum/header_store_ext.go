package ethereum

import (
	"io"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

type extHeaderStore struct {
	Numbers      []uint64
	Hashes       []common.Hash
	HeadersKey   []string
	HeadersValue [][]byte
	TDsKey       []string
	TDsValue     []*big.Int
	CurNumber    uint64
	CurHash      common.Hash
}

func (hs *HeaderStore) EncodeRLP(w io.Writer) error {
	var (
		cl = len(hs.CanonicalNumberToHash)
		hl = len(hs.Headers)
		tl = len(hs.TDs)
	)

	var (
		Numbers      = make([]uint64, 0, cl)
		Hashes       = make([]common.Hash, 0, cl)
		HeadersKey   = make([]string, 0, hl)
		HeadersValue = make([][]byte, 0, hl)
		TDsKey       = make([]string, 0, tl)
		TDsValue     = make([]*big.Int, 0, tl)
	)

	for number := range hs.CanonicalNumberToHash {
		Numbers = append(Numbers, number)
	}
	sort.Slice(Numbers, func(i, j int) bool {
		return Numbers[i] < Numbers[j]
	})
	for _, number := range Numbers {
		Hashes = append(Hashes, hs.CanonicalNumberToHash[number])
	}

	for k := range hs.Headers {
		HeadersKey = append(HeadersKey, k)
	}
	sort.Strings(HeadersKey)
	for _, v := range HeadersKey {
		HeadersValue = append(HeadersValue, hs.Headers[v])
	}

	for k := range hs.TDs {
		TDsKey = append(TDsKey, k)
	}
	sort.Strings(TDsKey)
	for _, v := range TDsKey {
		TDsValue = append(TDsValue, hs.TDs[v])
	}

	return rlp.Encode(w, extHeaderStore{
		Numbers:      Numbers,
		Hashes:       Hashes,
		HeadersKey:   HeadersKey,
		HeadersValue: HeadersValue,
		TDsKey:       TDsKey,
		TDsValue:     TDsValue,
		CurNumber:    hs.CurNumber,
		CurHash:      hs.CurHash,
	})
}

func (hs *HeaderStore) DecodeRLP(s *rlp.Stream) error {
	var eh extHeaderStore
	if err := s.Decode(&eh); err != nil {
		return err
	}

	CanonicalNumberToHash := make(map[uint64]common.Hash)
	Headers := make(map[string][]byte)
	TDs := make(map[string]*big.Int)

	for i, number := range eh.Numbers {
		CanonicalNumberToHash[number] = eh.Hashes[i]
	}
	for i, k := range eh.HeadersKey {
		Headers[k] = eh.HeadersValue[i]
	}
	for i, k := range eh.TDsKey {
		TDs[k] = eh.TDsValue[i]
	}

	hs.CurNumber, hs.CurHash = eh.CurNumber, eh.CurHash
	hs.CanonicalNumberToHash, hs.Headers, hs.TDs = CanonicalNumberToHash, Headers, TDs
	return nil
}
