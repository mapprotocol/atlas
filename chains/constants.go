package chains

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/mapprotocol/atlas/params"
)

const (
	ChainTypeMAP     = ChainType(params.MainNetChainID)
	ChainTypeMAPTest = ChainType(params.TestNetChainID)
	ChainTypeMAPDev  = ChainType(params.DevNetChainID)
)

const (
	ChainTypeETH     ChainType = 1
	ChainTypeETHTest ChainType = 34434
)

const (
	ChainGroupMAP = 1000
	ChainGroupETH = 1001
)

var ChainTypeList = []ChainType{
	ChainTypeMAP,
	ChainTypeMAPTest,
	ChainTypeMAPDev,
	ChainTypeETH,
	ChainTypeETHTest,
}

var chainType2ChainGroup = map[ChainType]ChainGroup{
	ChainTypeETH:     ChainGroupETH,
	ChainTypeETHTest: ChainGroupETH,
}

var chainType2ChainID = map[ChainType]uint64{
	ChainTypeETH:     params.MainNetChainID,
	ChainTypeETHTest: params.DevNetChainID,
}

var chainType2LondonBlock = map[ChainType]*big.Int{
	ChainTypeETH:     big.NewInt(12_965_000),
	ChainTypeETHTest: big.NewInt(10_499_401),
}

var (
	EthereumHeaderStoreAddress = common.BytesToAddress([]byte("EthereumHeaderStoreAddress"))
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
