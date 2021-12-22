package interfaces

import (
	"github.com/mapprotocol/atlas/chains/ethereum"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/mapprotocol/atlas/chains"
)

type IVerify interface {
	Verify(router common.Address, srcChain, dstChain *big.Int, txProveBytes []byte) error
}

func VerifyFactory(group chains.ChainGroup) (IVerify, error) {
	switch group {
	case chains.ChainGroupETH:
		return new(ethereum.Verify), nil
	}
	return nil, chains.ErrNotSupportChain
}
