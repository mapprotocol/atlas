package interfaces

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/ethereum"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/params"

)

type IChain interface {
	IValidate
	IHeaderStore
}

type Chain struct {
	Validate    IValidate
	HeaderStore IHeaderStore
}

func (c *Chain) ValidateHeaderChain(db types.StateDB, headers []byte, chainType chains.ChainType) (int, error) {
	return c.Validate.ValidateHeaderChain(db, headers, chainType)
}

func (c *Chain) ResetHeaderStore(db types.StateDB, header []byte, td *big.Int) error {
	return c.HeaderStore.ResetHeaderStore(db, header, td)
}

func (c *Chain) InsertHeaders(db types.StateDB, headers []byte) ([]*params.NumberHash, error) {
	return c.HeaderStore.InsertHeaders(db, headers)
}

func (c *Chain) GetCurrentNumberAndHash(db types.StateDB) (uint64, common.Hash, error) {
	return c.HeaderStore.GetCurrentNumberAndHash(db)
}

func (c *Chain) GetHashByNumber(db types.StateDB, number uint64) (common.Hash, error) {
	return c.HeaderStore.GetHashByNumber(db, number)
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
