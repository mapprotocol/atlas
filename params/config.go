package params

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
)

// Genesis hashes to enforce below configs on.
var (
	MainnetGenesisHash = common.HexToHash("0xf6285fd285d6c15aae581220e9b13f4d0ac75428ee90076e737f4e6125d31723") //"0xe9f6bb6cf0660cbbb1be839a9d8dda51116ab0577033d0bd77278f1044a674c0")
	TestnetGenesisHash = common.HexToHash("0x367cf53562c26325cfe1a827547f806d94bf891a8e70c222f3b66c89d31a1df0")
	DevnetGenesisHash  = common.HexToHash("0x1c00a47a70d32300cf336207d290ccc2838d3ea03b2ba73c07bafdd6070ff23a")
)

var (
	MainnetChainConfig = &params.ChainConfig{
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
		Ethash:              new(params.EthashConfig),
	}

	TestnetConfig = &params.ChainConfig{
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
		Ethash:              new(params.EthashConfig),
	}

	DevnetConfig = &params.ChainConfig{
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
		Ethash:              new(params.EthashConfig),
	}
	MainnetNetWorkID uint64 = 211
	TestnetWorkID    uint64 = 212
	DevnetWorkID     uint64 = 213

	//under params in cmd/node/defaults.go
	//DefaultHTTPPort    = 7445
	//DefaultWSPort      = 7446
	//ListenAddr         = 20201
	SingleChainID = big.NewInt(1234)
	SingleNetCfg  = &params.ChainConfig{
		SingleChainID,
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
		new(params.EthashConfig),
		nil}
)
