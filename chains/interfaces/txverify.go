package interfaces

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/ethereum"
	"github.com/mapprotocol/atlas/core/types"
)

type IVerify interface {
	Verify(db types.StateDB, router common.Address, srcChain, dstChain *big.Int, txProveBytes []byte) (logs []byte, err error)
}

func VerifyFactory(group chains.ChainGroup) (IVerify, error) {
	switch group {
	case chains.ChainGroupETH:
		return new(ethereum.Verify), nil
	}
	return nil, chains.ErrNotSupportChain
}
