package txverify

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/txverify/ethereum"
)

type IVerify interface {
	Verify(router common.Address, srcChain, dstChain *big.Int, txProveBytes []byte) error
}

func Factory(group chains.ChainGroup) (IVerify, error) {
	switch group {
	case chains.ChainGroupETH:
		return new(ethereum.Verify), nil
	}
	return nil, chains.ErrNotSupportChain
}
