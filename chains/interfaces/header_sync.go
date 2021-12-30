package interfaces

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/ethereum"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/params"
)

type IHeaderSync interface {
	StoreSyncTimes(db types.StateDB, epochID uint64, relayer common.Address, headers []*params.NumberHash) error
	LoadRelayerSyncTimes(db types.StateDB, epochID uint64, relayer common.Address) (uint64, error)
}

func HeaderSyncFactory(group chains.ChainGroup) (IHeaderSync, error) {
	switch group {
	case chains.ChainGroupETH:
		return new(ethereum.HeaderSync), nil
	}
	return nil, chains.ErrNotSupportChain
}
