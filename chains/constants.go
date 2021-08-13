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

var ChainTypeList = []rawdb.ChainType{
	ChainTypeMAP,
	ChainTypeETH,
	ChainTypeETHTest,
	ChainTypeETHDev,
}

func IsSupportedChain(chain rawdb.ChainType) bool {
	for _, c := range ChainTypeList {
		if c == chain {
			return true
		}
	}
	return false
}
