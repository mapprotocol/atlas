package params

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

const (
	MainNetChainID    uint64 = 211
	TestNetChainID    uint64 = 212
	DevNetChainID     uint64 = 213
	SingleNodeChainID uint64 = 214
)

// Genesis hashes to enforce below configs on.
var (
	MainnetGenesisHash = common.HexToHash("0xf6285fd285d6c15aae581220e9b13f4d0ac75428ee90076e737f4e6125d31723")
	TestnetGenesisHash = common.HexToHash("0x63f425f4a8362103c2be5089223d8823cad0baf5827eeecd20ee4adbe7dec063")
	DevnetGenesisHash  = common.HexToHash("0x1c00a47a70d32300cf336207d290ccc2838d3ea03b2ba73c07bafdd6070ff23a")
)

var (
	MainnetChainConfig = &ChainConfig{
		ChainID:             big.NewInt(211),
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
		Ethash:              new(EthashConfig),
		Istanbul: &IstanbulConfig{
			Epoch:          17280,
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
		Ethash:              new(EthashConfig),
		Istanbul: &IstanbulConfig{
			Epoch:          17280,
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
		Ethash:              new(EthashConfig),
		Istanbul: &IstanbulConfig{
			Epoch:          17280,
			ProposerPolicy: 2,
			BlockPeriod:    5,
			RequestTimeout: 3000,
			LookbackWindow: 12,
		},
	}
	MainnetNetWorkID uint64 = 211
	TestnetWorkID    uint64 = 212
	DevnetWorkID     uint64 = 213
	SingleWorkID     uint64 = 214
	//under params in cmd/node/defaults.go
	//DefaultHTTPPort    = 7445
	//DefaultWSPort      = 7446
	//ListenAddr         = 20201
	SingleChainID   = big.NewInt(214)
	SingleNetConfig = &ChainConfig{
		ChainID:             SingleChainID,
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
		YoloV3Block:         nil,
		EWASMBlock:          nil,
		CatalystBlock:       nil,
		Ethash:              new(EthashConfig),
		Clique:              nil,
		Istanbul: &IstanbulConfig{
			Epoch:          17280,
			ProposerPolicy: 2,
			BlockPeriod:    5,
			RequestTimeout: 3000,
			LookbackWindow: 12,
		},
	}

	AllEthashProtocolChanges = &ChainConfig{
		big.NewInt(213),
		big.NewInt(0),
		nil,
		false,
		big.NewInt(0),
		common.Hash{},
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		nil,
		nil,
		nil,
		nil,
		new(EthashConfig),
		nil,

		&IstanbulConfig{
			Epoch:          17280,
			ProposerPolicy: 2,
			BlockPeriod:    5,
			RequestTimeout: 3000,
			LookbackWindow: 12,
		},

		nil,
	}

	TestChainConfig = &ChainConfig{
		big.NewInt(1),
		big.NewInt(0),
		nil,
		false,
		big.NewInt(0),
		common.Hash{},
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		nil,
		nil,
		nil,
		nil,
		new(EthashConfig),
		nil,
		&IstanbulConfig{
			Epoch:          17280,
			ProposerPolicy: 2,
			BlockPeriod:    5,
			RequestTimeout: 3000,
			LookbackWindow: 12,
		},
		nil,
	}

	IstanbulTestChainConfig = &ChainConfig{
		big.NewInt(1337),
		big.NewInt(0),
		nil,
		false,
		big.NewInt(0),
		common.Hash{},
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		nil,
		nil,
		nil,
		big.NewInt(0), nil,
		nil,
		nil,
		&IstanbulConfig{
			Epoch:          300,
			ProposerPolicy: 0,
			RequestTimeout: 1000,
			BlockPeriod:    1,
		},
		false,
	}
)
