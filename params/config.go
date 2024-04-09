package params

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	MainNetChainID   uint64 = 4120
	TestNetChainID   uint64 = 4121
	DevNetChainID    uint64 = 4122
	SingleNetChainID uint64 = 4123
	Epoch            uint64 = 50000
)

// network id
const (
	MainnetNetWorkID = MainNetChainID
	TestnetWorkID    = TestNetChainID
	DevnetWorkID     = DevNetChainID
	SingleNetworkID  = SingleNetChainID
)

// Genesis hashes to enforce below configs on.
var (
	MainnetGenesisHash = common.HexToHash("0x49f819f1db33812af359aa5d08aae4a699265f4e62d61308a2b637c515bd2ccd")
	TestnetGenesisHash = common.HexToHash("0x5420312471d7cf6a04beda2e05d824e1d4e268d8f55020a77bc2c3875bcd73c2")
	DevnetGenesisHash  = common.HexToHash("0xd8cdaa5ab22c001e3e03e6d1c1f4dbb08f62c23c11ead15435d244d66512e853")
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
		EnableRewardBlock:   big.NewInt(0),
		BN256ForkBlock:      big.NewInt(0),
		DeregisterBlock:     big.NewInt(0),
		CalcBaseBlock:       big.NewInt(0),
		Istanbul: &IstanbulConfig{
			Epoch:          Epoch,
			ProposerPolicy: 2,
			BlockPeriod:    5,
			RequestTimeout: 3000,
			LookbackWindow: 12,
		},
		FullHeaderChainAvailable: true,
	}

	TestnetConfig = &ChainConfig{
		ChainID:             big.NewInt(int64(TestnetWorkID)),
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
		EnableRewardBlock:   big.NewInt(0),
		BN256ForkBlock:      big.NewInt(0),
		DeregisterBlock:     big.NewInt(0),
		CalcBaseBlock:       big.NewInt(0),
		Istanbul: &IstanbulConfig{
			Epoch:          4000,
			ProposerPolicy: 2,
			BlockPeriod:    5,
			RequestTimeout: 3000,
			LookbackWindow: 12,
		},
		FullHeaderChainAvailable: true,
	}

	DevnetConfig = &ChainConfig{
		ChainID:             big.NewInt(int64(DevNetChainID)),
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
		EnableRewardBlock:   big.NewInt(0),
		BN256ForkBlock:      big.NewInt(0),
		DeregisterBlock:     big.NewInt(0),
		CalcBaseBlock:       big.NewInt(0),
		Istanbul: &IstanbulConfig{
			Epoch:          1000,
			ProposerPolicy: 2,
			BlockPeriod:    5,
			RequestTimeout: 3000,
			LookbackWindow: 12,
		},
	}

	SingleNetConfig = &ChainConfig{
		ChainID:             new(big.Int).SetUint64(SingleNetChainID),
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
		EnableRewardBlock:   big.NewInt(0),
		BN256ForkBlock:      big.NewInt(0),
		DeregisterBlock:     big.NewInt(0),
		CalcBaseBlock:       big.NewInt(0),
		Istanbul: &IstanbulConfig{
			Epoch:          1000,
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
		EnableRewardBlock:   big.NewInt(0),
		BN256ForkBlock:      big.NewInt(2000),
		DeregisterBlock:     big.NewInt(0),
		CalcBaseBlock:       big.NewInt(0),
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
		EnableRewardBlock:   big.NewInt(5000),
		BN256ForkBlock:      big.NewInt(20000),
		DeregisterBlock:     big.NewInt(0),
		CalcBaseBlock:       big.NewInt(0),
		DonutBlock:          nil,
		EWASMBlock:          nil,
		CatalystBlock:       nil,
		Istanbul: &IstanbulConfig{
			Epoch:          4000,
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
		EnableRewardBlock:   big.NewInt(0),
		BN256ForkBlock:      big.NewInt(2000),
		DeregisterBlock:     big.NewInt(0),
		CalcBaseBlock:       big.NewInt(0),
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
