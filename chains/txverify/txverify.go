package txverify

import (
	"math/big"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/txverify/ethereum"
)

type IVerify interface {
	Verify(srcChain, dstChain *big.Int, txProveBytes []byte) error
}

func Factory(group chains.ChainGroup) (IVerify, error) {
	switch group {
	case chains.ChainGroupETH:
		return new(ethereum.Verify), nil
	}
	return nil, chains.ErrNotSupportChain
}
