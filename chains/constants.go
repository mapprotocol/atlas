package chains

import (
	"github.com/mapprotocol/atlas/core/rawdb"
)

const (
	ChainTypeMAP         rawdb.ChainType = 1000
	ChainTypeETH         rawdb.ChainType = 1001
	ChainTypeETH_test    rawdb.ChainType = 1002
	ChainTypeETH_dev     rawdb.ChainType = 1003
	ChainTypeETH_ropsten rawdb.ChainType = 1004 // start 800
)

var ChainTypeList = []rawdb.ChainType{
	ChainTypeMAP,
	ChainTypeETH,
	ChainTypeETH_test,
	ChainTypeETH_dev,
	ChainTypeETH_ropsten,
}

func IsSupportedChain(chain rawdb.ChainType) bool {
	for _, c := range ChainTypeList {
		if c == chain {
			return true
		}
	}
	return false
}
