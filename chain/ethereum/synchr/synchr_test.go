package synchr_test

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/mapprotocol/atlas/chain/ethereum/synchr"
	"math/big"
	"math/rand"
	"testing"
)

func makeHeaderChain(num int, seed int64) []*types.Header {
	r := rand.New(rand.NewSource(seed))
	headers := make([]*types.Header, num)

	for i := 0; i < num; i++ {
		header := types.Header{
			Number: big.NewInt(int64(i)),
			Nonce:  types.EncodeNonce(r.Uint64()),
		}
		if i > 0 {
			header.ParentHash = headers[i-1].Hash()
		}
		headers[i] = &header
	}

	return headers
}

func Test_HeaderCache(t *testing.T) {
	cache := synchr.NewHeaderCache(3)
	headers := makeHeaderChain(5, 0)

	print(cache)
	print(headers)
}
