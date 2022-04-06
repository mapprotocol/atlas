package env

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/marker/internal/utils"
	"github.com/mapprotocol/atlas/params"
	"math/big"
	"testing"
)

func TestEnvironment_SaveGenesis(t *testing.T) {

	var ChainConfig = params.ChainConfig{
		ChainID: big.NewInt(211),

		HomesteadBlock: big.NewInt(211),
		DAOForkBlock:   big.NewInt(211),
		DAOForkSupport: true,

		EIP150Block: big.NewInt(211),
		EIP150Hash:  common.HexToHash("s"),

		EIP155Block: big.NewInt(211),
		EIP158Block: big.NewInt(211),

		//ByzantiumBlock:      big.NewInt(211),
		//ConstantinopleBlock: big.NewInt(211),
		//PetersburgBlock:     big.NewInt(211),
		//IstanbulBlock:       big.NewInt(211),
		//MuirGlacierBlock:    big.NewInt(211),
		//BerlinBlock:         big.NewInt(211),
		//LondonBlock:         big.NewInt(211),

		DonutBlock: big.NewInt(211),

		//YoloV3Block   *big.Int `json:"yoloV3Block,omitempty"`   // YOLO v3: Gas repricings TODO @holiman add EIP references
		EWASMBlock:    big.NewInt(211),
		CatalystBlock: big.NewInt(211),

		// Various consensus engines
		Istanbul: &params.IstanbulConfig{
			Epoch:          4,
			ProposerPolicy: 5,
			LookbackWindow: 6,
			BlockPeriod:    7,
			RequestTimeout: 8,
		},

		// This does not belong here but passing it to every function is not possible since that breaks
		// some implemented interfaces and introduces churn across the geth codebase.
		FullHeaderChainAvailable: true,

		// Requests mock engine if true
		Faker: true,
	}
	var genesis chain.Genesis
	genesis = chain.Genesis{
		Config:    &ChainConfig,
		Nonce:     1,
		Timestamp: 2,
		ExtraData: []byte{'a'},
		GasLimit:  3,
		Mixhash:   common.HexToHash("123"),
		Coinbase:  common.Address{},
		Alloc:     nil,

		// These fields are used for consensus tests. Please don't use them
		// in actual genesis blocks.
		Number:     4,
		GasUsed:    5,
		ParentHash: common.HexToHash("456"),
		//BaseFee:    big.NewInt(44),
	}

	err := utils.WriteJson(&genesis, "D:/root/test/test.json")
	if err != nil {
		t.Fatalf("WriteJson", "err", err)
	}
}
