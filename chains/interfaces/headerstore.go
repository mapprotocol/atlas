package interfaces

import (
	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/ethereum"
)

type StoreLoad interface {
	//Store(state vm.StateDB, address common.Address) error
	//Load(state vm.StateDB, address common.Address) error
}

type IHeaderStore interface {
	StoreLoad
	CurrentNumber() uint64
	Push(v interface{}) error
	GetHeaderByNumber(number uint64) interface{}
}

func HeaderStoreFactory(group chains.ChainGroup) (IHeaderStore, error) {
	switch group {
	case chains.ChainGroupETH:
		return new(ethereum.HeaderStore), nil
	}
	return nil, chains.ErrNotSupportChain
}