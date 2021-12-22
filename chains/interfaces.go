package chains

import (
	"github.com/mapprotocol/atlas/chains/ethereum"
	"github.com/mapprotocol/atlas/core/vm"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type IChain interface {
	IVerify
	IHeaderStore
}

type IVerify interface {
	Verify(router common.Address, srcChain, dstChain *big.Int, txProveBytes []byte) error
}

func VerifyFactory(group ChainGroup) (IVerify, error) {
	switch group {
	case ChainGroupETH:
		return new(ethereum.Verify), nil
	}
	return nil, ErrNotSupportChain
}

type StoreLoad interface {
	Store(state vm.StateDB) error
	Load(state vm.StateDB) error
}

type IHeaderStore interface {
	StoreLoad
	CurrentNumber() uint64
	Push(v interface{}) error
}

func HeaderStoreFactory(group ChainGroup) (IHeaderStore, error) {
	switch group {
	case ChainGroupETH:
		return new(ethereum.HeaderStore), nil
	}
	return nil, ErrNotSupportChain
}
