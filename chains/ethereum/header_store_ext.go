package ethereum

import (
	"io"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

type extHeaderStore struct {
	Numbers []uint64
	Hashes  []common.Hash
	//HeadersKey   []*big.Int
	//HeadersValue [][]byte
	//TDsKey       []*big.Int
	//TDsValue     []*big.Int
	CurNumber uint64
	CurHash   common.Hash
}

func (hs *HeaderStore) EncodeRLP(w io.Writer) error {
	var (
		cl = len(hs.CanonicalNumberToHash)
		//hl = len(hs.HeaderNumber)
		//tl = len(hs.HeaderNumber)
	)

	var (
		Numbers = make([]uint64, 0, cl)
		Hashes  = make([]common.Hash, 0, cl)
		//HeadersKey = make([]*big.Int, 0, hl)
		//TDsKey     = make([]*big.Int, 0, tl)
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

	//for _, k := range hs.HeaderNumber {
	//	HeadersKey = append(HeadersKey, k)
	//}
	//
	//for _, k := range hs.HeaderNumber {
	//	TDsKey = append(TDsKey, k)
	//}

	return rlp.Encode(w, extHeaderStore{
		Numbers: Numbers,
		Hashes:  Hashes,
		//HeadersKey: HeadersKey,
		//TDsKey:     TDsKey,
		CurNumber: hs.CurNumber,
		CurHash:   hs.CurHash,
	})
}

func (hs *HeaderStore) DecodeRLP(s *rlp.Stream) error {
	var eh extHeaderStore
	if err := s.Decode(&eh); err != nil {
		return err
	}

	CanonicalNumberToHash := make(map[uint64]common.Hash)
	//headerNumber := make([]*big.Int, 0, len(eh.HeadersKey))

	for i, number := range eh.Numbers {
		CanonicalNumberToHash[number] = eh.Hashes[i]
	}
	//for _, v := range eh.HeadersKey {
	//	headerNumber = append(headerNumber, v)
	//}

	hs.CurNumber, hs.CurHash = eh.CurNumber, eh.CurHash
	//hs.CanonicalNumberToHash, hs.HeaderNumber = CanonicalNumberToHash, headerNumber
	hs.CanonicalNumberToHash = CanonicalNumberToHash
	return nil
}

type extLightHeader struct {
	HeadersKey   []string
	HeadersValue [][]byte
	TDsKey       []string
	TDsValue     []*big.Int
}

func (lh *LightHeader) EncodeRLP(w io.Writer) error {
	var (
		hl = len(lh.Headers)
		tl = len(lh.TDs)
	)

	var (
		HeadersKey   = make([]string, 0, hl)
		HeadersValue = make([][]byte, 0, hl)
		TDsKey       = make([]string, 0, tl)
		TDsValue     = make([]*big.Int, 0, tl)
	)

	for k := range lh.Headers {
		HeadersKey = append(HeadersKey, k)
	}
	sort.Strings(HeadersKey)
	for _, v := range HeadersKey {
		HeadersValue = append(HeadersValue, lh.Headers[v])
	}

	for k := range lh.TDs {
		TDsKey = append(TDsKey, k)
	}
	sort.Strings(TDsKey)
	for _, v := range TDsKey {
		TDsValue = append(TDsValue, lh.TDs[v])
	}

	return rlp.Encode(w, extLightHeader{
		HeadersKey:   HeadersKey,
		TDsKey:       TDsKey,
		HeadersValue: HeadersValue,
		TDsValue:     TDsValue,
	})
}

func (lh *LightHeader) DecodeRLP(s *rlp.Stream) error {
	var eh extLightHeader
	if err := s.Decode(&eh); err != nil {
		return err
	}

	Headers := make(map[string][]byte)
	TDs := make(map[string]*big.Int)

	for i, k := range eh.HeadersKey {
		Headers[k] = eh.HeadersValue[i]
	}
	for i, k := range eh.TDsKey {
		TDs[k] = eh.TDsValue[i]
	}

	lh.Headers, lh.TDs = Headers, TDs
	return nil
}
