package params

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	MainNetChainID   uint64 = 22776
	TestNetChainID   uint64 = 212
	DevNetChainID    uint64 = 213
	SingleNetChainID uint64 = 214
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
	MainnetGenesisHash = common.HexToHash("0x6b2bd27bee0f7675550204c541a30cc6a14aa1738431cb60e21e666b2fec8014")
	TestnetGenesisHash = common.HexToHash("0xede9749324b65e9df7b74830696637ed25c24ae4ffd07f8c36b8e5b6eac7dfa0")
	DevnetGenesisHash  = common.HexToHash("0xa7712fd6f430d32fbc796665289bf9702b6991e96393fd670e7834c48e15755f")
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
		EnableRewardBlock:   big.NewInt(1125000),
		BN256ForkBlock:      big.NewInt(2350001),
		DeregisterBlock:     big.NewInt(2990000),
		CalcBaseBlock:       big.NewInt(3790000),
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
		EnableRewardBlock:   big.NewInt(5000),
		BN256ForkBlock:      big.NewInt(20000),
		DeregisterBlock:     big.NewInt(1242000),
		CalcBaseBlock:       big.NewInt(1242000),
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
		EnableRewardBlock:   big.NewInt(0),
		BN256ForkBlock:      big.NewInt(0),
		DeregisterBlock:     big.NewInt(0),
		CalcBaseBlock:       big.NewInt(0),
		ShanghaiBlock:       big.NewInt(1000),
		CancunBlock:         big.NewInt(1000),
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
