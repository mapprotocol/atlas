package ethereum

import (
	"io"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/rlp"
)

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
