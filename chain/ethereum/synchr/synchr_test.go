package synchr_test

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/chain/ethereum/synchr"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/stretchr/testify/mock"
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

type TestSubscription struct{}

func (t *TestSubscription) Unsubscribe()      {}
func (t *TestSubscription) Err() <-chan error { return make(chan error) }

type TestHeaderLoader struct {
	mock.Mock
	NewHeaders chan<- *types.Header
}

func (t *TestHeaderLoader) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	args := t.Called(hash)
	return args.Get(0).(*types.Header), args.Error(1)
}
