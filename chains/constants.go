package chains

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"

	"github.com/mapprotocol/atlas/params"
)

const (
	//ChainTypeMAP     ChainType = ChainType(params.MainNetChainID)
	//ChainTypeMAPTest ChainType = ChainType(params.TestNetChainID)
	//ChainTypeMAPDev  ChainType = ChainType(params.DevNetChainID)
	ChainTypeETH     ChainType = 1
	ChainTypeETHTest ChainType = 3
	ChainTypeETHDev  ChainType = 10
)

const (
	ChainGroupMAP = 1000
	ChainGroupETH = 1001
)

var ChainTypeList = []ChainType{
	//ChainTypeMAP,
	//ChainTypeMAPTest,
	//ChainTypeMAPDev,
	ChainTypeETH,
	ChainTypeETHTest,
	ChainTypeETHDev,
}

var chainType2ChainGroup = map[ChainType]ChainGroup{
	ChainTypeETH:     ChainGroupETH,
	ChainTypeETHDev:  ChainGroupETH,
	ChainTypeETHTest: ChainGroupETH,
}

var chainType2ChainID = map[ChainType]uint64{
	ChainTypeETH:     params.MainNetChainID,
	ChainTypeETHDev:  params.DevNetChainID,
	ChainTypeETHTest: params.TestNetChainID,
}

var chainType2LondonBlock = map[ChainType]*big.Int{
	ChainTypeETH:     big.NewInt(12_965_000),
	ChainTypeETHTest: big.NewInt(10_499_401),
}

var (
	EthereumHeaderStoreAddress = common.BytesToAddress([]byte("EthereumHeaderStoreAddress"))
	EthereumHeaderSyncAddress  = common.BytesToAddress([]byte("EthereumHeaderSyncInfoAddress"))
)

type ChainType uint64
type ChainGroup uint64

func IsSupportedChain(chain ChainType) bool {
	for _, c := range ChainTypeList {
		if c == chain {
			return true
		}
	}
	return false
}

func ChainType2ChainGroup(chain ChainType) (ChainGroup, error) {
	group, ok := chainType2ChainGroup[chain]
	if !ok {
		return 0, ErrNotSupportChain
	}
	return group, nil
}

func ChainType2ChainID(chain ChainType) (uint64, error) {
	chainID, ok := chainType2ChainID[chain]
	if !ok {
		return 0, ErrNotSupportChain
	}
	return chainID, nil
}

func ChainType2LondonBlock(chain ChainType) (*big.Int, error) {
	lb, ok := chainType2LondonBlock[chain]
	if !ok {
		return nil, ErrNotSupportChain
	}
	return lb, nil
}
