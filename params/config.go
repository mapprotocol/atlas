package params

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	MainNetChainID uint64 = 214
	TestNetChainID uint64 = 212
	DevNetChainID  uint64 = 213
	Epoch          uint64 = 30000
)

// network id
const (
	MainnetNetWorkID = MainNetChainID
	TestnetWorkID    = TestNetChainID
	DevnetWorkID     = DevNetChainID
)

// Genesis hashes to enforce below configs on.
var (
	MainnetGenesisHash = common.HexToHash("0x9a2c09dc9f15e67f86dbf339e148ba0b4d0170fbfb72e420e30eaae1604b6669")
	TestnetGenesisHash = common.HexToHash("0xfc7c54a959dc7b00c5ba13d9bae0eccae4946f8169e0b8a89bd50a4ed8fae203")
	DevnetGenesisHash  = common.HexToHash("0xbdb60a85ba40c846fd02cb683b112d72a74751c96fe6b337ee631b6542cf6f9b")
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
			Epoch:          20000,
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
