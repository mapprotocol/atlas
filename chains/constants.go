package chains

import (
	"github.com/mapprotocol/atlas/core/rawdb"
)

const (
	ChainTypeMAP rawdb.ChainType = 1000
	ChainTypeETH rawdb.ChainType = 1001
)

var ChainTypeList = []rawdb.ChainType{
	ChainTypeMAP,
	ChainTypeETH,
}

func IsSupportedChain(chain rawdb.ChainType) bool {
	for _, c := range ChainTypeList {
		if c == chain {
			return true
		}
	}
	return false
}
