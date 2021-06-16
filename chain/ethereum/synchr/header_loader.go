package synchr

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/core/types"
)

type HeaderLoader interface {
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
}

type HeaderLoaderItem struct {
	client *ethclient.Client
}

func NewHeaderLoader(client *ethclient.Client) *HeaderLoaderItem {
	return &HeaderLoaderItem{client: client}
}

func (d *HeaderLoaderItem) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return d.client.HeaderByHash(ctx, hash)
}

func (d *HeaderLoaderItem) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return d.client.HeaderByNumber(ctx, number)
}

func (d *HeaderLoaderItem) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	return d.client.SubscribeNewHead(ctx, ch)
}
