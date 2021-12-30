package interfaces

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/ethereum"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/params"

)

type IChain interface {
	IValidate
	IHeaderStore
	IHeaderSync
}

type Chain struct {
	Validate    IValidate
	HeaderStore IHeaderStore
	HeaderSync  IHeaderSync
}

func (c *Chain) ValidateHeaderChain(db types.StateDB, headers []byte) (int, error) {
	return c.Validate.ValidateHeaderChain(db, headers)
}

func (c *Chain) WriteHeaders(db types.StateDB, headers []byte) ([]*params.NumberHash, error) {
	return c.HeaderStore.WriteHeaders(db, headers)
}

func (c *Chain) GetCurrentNumberAndHash(db types.StateDB) (uint64, common.Hash, error) {
	return c.HeaderStore.GetCurrentNumberAndHash(db)
}

func (c *Chain) GetHashByNumber(db types.StateDB, number uint64) (common.Hash, error) {
	return c.HeaderStore.GetHashByNumber(db, number)
}

func (c *Chain) StoreSyncTimes(db types.StateDB, epochID uint64, relayer common.Address, headers []*params.NumberHash) error {
	return c.HeaderSync.StoreSyncTimes(db, epochID, relayer, headers)
}

func (c *Chain) LoadRelayerSyncTimes(db types.StateDB, epochID uint64, relayer common.Address) (uint64, error) {
	return c.HeaderSync.LoadRelayerSyncTimes(db, epochID, relayer)
}

func ChainFactory(group chains.ChainGroup) (IChain, error) {
	switch group {
	case chains.ChainGroupETH:
		return &Chain{
			Validate:    new(ethereum.Validate),
			HeaderStore: new(ethereum.HeaderStore),
			HeaderSync:  new(ethereum.HeaderSync),
		}, nil
	}

	return nil, errors.New("not support chain")
}
