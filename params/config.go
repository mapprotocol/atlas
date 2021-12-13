package params

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	MainNetChainID uint64 = 214
	TestNetChainID uint64 = 212
	DevNetChainID  uint64 = 213
	Epoch          uint64 = 20000
)

// network id
const (
	MainnetNetWorkID = MainNetChainID
	TestnetWorkID    = TestNetChainID
	DevnetWorkID     = DevNetChainID
)

// Genesis hashes to enforce below configs on.
var (
	MainnetGenesisHash = common.HexToHash("0xa6c12a7509006b29863f201aab1cc0f21f5abf2c81c63d716f2e6957cdb765da")
	TestnetGenesisHash = common.HexToHash("0x28ce321590b3f59bef95e5b59f423f33238401fbd1962844cd275ce38f2c4ed0")
	DevnetGenesisHash  = common.HexToHash("0x0161e9569a5ca784969f1c28218a22ff8ac7672a5d32fc8f493d3f5208a14f73")
)

var (
	MainnetChainConfig = &ChainConfig{
		ChainID:             big.NewInt(int64(MainNetChainID)),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        big.NewInt(0),
		DAOForkSupport:      true,
		EIP150Block:         big.NewInt(0),
		EIP150Hash:          common.Hash{},
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		Istanbul: &IstanbulConfig{
			Epoch:          Epoch,
			ProposerPolicy: 2,
			BlockPeriod:    5,
			RequestTimeout: 3000,
			LookbackWindow: 12,
		},
	}

	TestnetConfig = &ChainConfig{
		ChainID:             big.NewInt(212),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      true,
		EIP150Block:         big.NewInt(0),
		EIP150Hash:          common.Hash{},
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		Istanbul: &IstanbulConfig{
			Epoch:          80,
			ProposerPolicy: 2,
			BlockPeriod:    5,
			RequestTimeout: 3000,
			LookbackWindow: 12,
		},
	}

	DevnetConfig = &ChainConfig{
		ChainID:             big.NewInt(213),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      true,
		EIP150Block:         big.NewInt(0),
		EIP150Hash:          common.Hash{},
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		Istanbul: &IstanbulConfig{
			Epoch:          17280,
			ProposerPolicy: 2,
			BlockPeriod:    5,
			RequestTimeout: 3000,
			LookbackWindow: 12,
		},
	}

	AllEthashProtocolChanges = &ChainConfig{
		ChainID:             big.NewInt(213),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      false,
		EIP150Block:         big.NewInt(0),
		EIP150Hash:          common.Hash{},
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		DonutBlock:          nil,
		EWASMBlock:          nil,
		CatalystBlock:       nil,
		Istanbul: &IstanbulConfig{
			Epoch:          17280,
			ProposerPolicy: 2,
			BlockPeriod:    5,
			RequestTimeout: 3000,
			LookbackWindow: 12,
		},
		FullHeaderChainAvailable: true,
		Faker:                    true,
	}

	TestChainConfig = &ChainConfig{
		ChainID:             big.NewInt(1),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      false,
		EIP150Block:         big.NewInt(0),
		EIP150Hash:          common.Hash{},
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		DonutBlock:          nil,
		EWASMBlock:          nil,
		CatalystBlock:       nil,
		Istanbul: &IstanbulConfig{
			Epoch:          17280,
			ProposerPolicy: 2,
			BlockPeriod:    5,
			RequestTimeout: 3000,
			LookbackWindow: 12,
		},
		FullHeaderChainAvailable: true,
		Faker:                    true,
	}

	IstanbulTestChainConfig = &ChainConfig{
		ChainID:             big.NewInt(1337),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      false,
		EIP150Block:         big.NewInt(0),
		EIP150Hash:          common.Hash{},
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         nil,
		LondonBlock:         nil,
		DonutBlock:          nil,
		EWASMBlock:          big.NewInt(0),
		CatalystBlock:       nil,
		Istanbul: &IstanbulConfig{
			Epoch:          300,
			ProposerPolicy: 0,
			RequestTimeout: 1000,
			BlockPeriod:    1,
		},
		FullHeaderChainAvailable: true,
		Faker:                    false,
	}
)
