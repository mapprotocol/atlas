package interfaces

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/ethereum"
	"github.com/mapprotocol/atlas/core/types"
)

type StoreLoad interface {
	Store(db types.StateDB) error
	Load(db types.StateDB) error
}

type IHeaderStore interface {
	WriteHeaders(db types.StateDB, headers []byte) (int, error)
	CurrentHash() common.Hash
	CurrentNumber() uint64
	GetHashByNumber(number uint64) common.Hash
}

func HeaderStoreFactory(group chains.ChainGroup) (IHeaderStore, error) {
	switch group {
	case chains.ChainGroupETH:
		return new(ethereum.HeaderStore), nil
	}
	return nil, chains.ErrNotSupportChain
}
