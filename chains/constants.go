package chains

import (
	"math/big"

	"github.com/mapprotocol/atlas/core/rawdb"
	params2 "github.com/mapprotocol/atlas/params"
)

const (
	ChainTypeMAP     rawdb.ChainType = rawdb.ChainType(params2.MainNetChainID)
	ChainTypeMAPTest rawdb.ChainType = rawdb.ChainType(params2.TestNetChainID)
	ChainTypeMAPDev  rawdb.ChainType = rawdb.ChainType(params2.DevNetChainID)
	ChainTypeETH     rawdb.ChainType = 1
	ChainTypeETHTest rawdb.ChainType = 3 // start 800
	ChainTypeETHDev  rawdb.ChainType = 10
)

const (
	ChainGroupMAP = 1000
	ChainGroupETH = 1001
)

var ChainTypeList = []rawdb.ChainType{
	ChainTypeMAP,
	ChainTypeMAPTest,
	ChainTypeMAPDev,
	ChainTypeETH,
	ChainTypeETHTest,
	ChainTypeETHDev,
}

var chainType2ChainGroup = map[rawdb.ChainType]ChainGroup{
	ChainTypeMAP:     ChainGroupMAP,
	ChainTypeMAPTest: ChainGroupMAP,
	ChainTypeMAPDev:  ChainGroupMAP,
	ChainTypeETH:     ChainGroupETH,
	ChainTypeETHDev:  ChainGroupETH,
	ChainTypeETHTest: ChainGroupETH,
}

var chainType2LondonBlock = map[rawdb.ChainType]*big.Int{
	ChainTypeETH:     big.NewInt(12_965_000),
	ChainTypeETHTest: big.NewInt(10_499_401),
}

type ChainGroup uint64

func IsSupportedChain(chain rawdb.ChainType) bool {
	for _, c := range ChainTypeList {
		if c == chain {
			return true
		}
	}
	return false
}

func ChainType2ChainGroup(chain rawdb.ChainType) (ChainGroup, error) {
	group, ok := chainType2ChainGroup[chain]
	if !ok {
		return 0, ErrNotSupportChain
	}
	return group, nil
}

func ChainType2LondonBlock(chain rawdb.ChainType) (*big.Int, error) {
	lb, ok := chainType2LondonBlock[chain]
	if !ok {
		return nil, ErrNotSupportChain
	}
	return lb, nil
}
