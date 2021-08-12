package chains

import (
	"github.com/mapprotocol/atlas/core/rawdb"
)

const (
	ChainTypeMAP      rawdb.ChainType = 1000
	ChainTypeETH      rawdb.ChainType = 1001
	ChainTypeETH_test rawdb.ChainType = 1002 // start 800
	ChainTypeETH_dev  rawdb.ChainType = 1003
)

var ChainTypeList = []rawdb.ChainType{
	ChainTypeMAP,
	ChainTypeETH,
	ChainTypeETH_test,
	ChainTypeETH_dev,
}

func IsSupportedChain(chain rawdb.ChainType) bool {
	for _, c := range ChainTypeList {
		if c == chain {
			return true
		}
	}
	return false
}
