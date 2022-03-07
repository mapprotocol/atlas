package interfaces

import (
	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/ethereum"
	"github.com/mapprotocol/atlas/core/types"
)

type IValidate interface {
	ValidateHeaderChain(db types.StateDB, headers []byte, chainType chains.ChainType) (int, error)
}

func ValidateFactory(group chains.ChainGroup) (IValidate, error) {
	switch group {
	case chains.ChainGroupETH:
		return new(ethereum.Validate), nil
	}
	return nil, chains.ErrNotSupportChain
}


type Validate interface {
    
}