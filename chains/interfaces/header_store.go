package interfaces

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/ethereum"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/params"
)

type StoreLoad interface {
	Store(db types.StateDB) error
	Load(db types.StateDB) error
}

type IHeaderStore interface {
	ResetHeaderStore(db types.StateDB, header []byte, td *big.Int) error
	InsertHeaders(db types.StateDB, headers []byte) ([]*params.NumberHash, error)
	GetCurrentNumberAndHash(db types.StateDB) (uint64, common.Hash, error)
	GetHashByNumber(db types.StateDB, number uint64) (common.Hash, error)
}

func HeaderStoreFactory(group chains.ChainGroup) (IHeaderStore, error) {
	switch group {
	case chains.ChainGroupETH:
		return new(ethereum.HeaderStore), nil
	}
	return nil, chains.ErrNotSupportChain
}
