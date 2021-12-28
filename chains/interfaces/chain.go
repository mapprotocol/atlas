package interfaces

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/ethereum"
	"github.com/mapprotocol/atlas/core/types"
)

type IChain interface {
	IValidate
	IHeaderStore
	//IHeaderSyncInfo
}

type Chain struct {
	Validate    IValidate
	HeaderStore IHeaderStore
	//HeaderSyncInfo IHeaderSyncInfo
}

func (c *Chain) ValidateHeaderChain(db types.StateDB, headers []byte) (int, error) {
	return c.Validate.ValidateHeaderChain(db, headers)
}

func (c *Chain) WriteHeaders(db types.StateDB, headers []byte) (int, error) {
	return c.HeaderStore.WriteHeaders(db, headers)
}

func (c *Chain) CurrentHash() common.Hash {
	return c.HeaderStore.CurrentHash()
}

func (c *Chain) CurrentNumber() uint64 {
	return c.HeaderStore.CurrentNumber()
}

func (c *Chain) GetHashByNumber(number uint64) common.Hash {
	return c.HeaderStore.GetHashByNumber(number)
}

func ChainFactory(group chains.ChainGroup) (IChain, error) {
	switch group {
	case chains.ChainGroupETH:
		return &Chain{
			Validate:    new(ethereum.Validate),
			HeaderStore: new(ethereum.HeaderStore),
		}, nil
	}

	return nil, errors.New("not support chain")
}
