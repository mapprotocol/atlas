package chains

import (
	"github.com/mapprotocol/atlas/core/rawdb"
)

const (
	ChainTypeMAP     rawdb.ChainType = 211
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
	ChainTypeETH,
	ChainTypeETHTest,
	ChainTypeETHDev,
}

var chainType2ChainGroup = map[rawdb.ChainType]ChainGroup{
	ChainTypeMAP:     ChainGroupMAP,
	ChainTypeETH:     ChainGroupETH,
	ChainTypeETHDev:  ChainGroupETH,
	ChainTypeETHTest: ChainGroupETH,
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
