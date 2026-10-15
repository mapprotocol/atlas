package params

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestIsBlacklistedTx(t *testing.T) {
	var (
		mainnet   = new(big.Int).SetUint64(MainNetChainID)
		devnet    = new(big.Int).SetUint64(DevNetChainID)
		testnet   = new(big.Int).SetUint64(TestNetChainID)
		exploiter = common.HexToAddress("0x8d0b87a76c8d38a96deda851e737461ef9d03d92")
		receiver  = common.HexToAddress("0xF1b596F08aD6A323Ca43A31B92559c5dD8EE5184")
		drill     = common.HexToAddress("0xf855a761f9182c4b22A04753681A1F6324Ed3449")
		other     = common.HexToAddress("0x1111111111111111111111111111111111111111")
	)
	for i, tt := range []struct {
		chainID *big.Int
		from    common.Address
		to      *common.Address
		want    bool
	}{
		{mainnet, exploiter, &other, true},
		{mainnet, other, &exploiter, true},
		{mainnet, receiver, &other, true},
		{mainnet, other, &receiver, true},
		{mainnet, exploiter, nil, true},
		{mainnet, other, &other, false},
		{mainnet, other, nil, false},
		{mainnet, drill, &other, false}, // drill address is devnet/testnet only
		{devnet, drill, &other, true},
		{devnet, other, &drill, true},
		{devnet, exploiter, &other, false}, // mainnet list does not apply to devnet
		{testnet, drill, &other, true},
		{testnet, other, &drill, true},
		{testnet, exploiter, &other, false}, // mainnet list does not apply to testnet
		{nil, exploiter, &other, false},
	} {
		if have := IsBlacklistedTx(tt.chainID, tt.from, tt.to); have != tt.want {
			t.Errorf("test %d: have %v, want %v", i, have, tt.want)
		}
	}
}

func TestMainnetBlacklistBlock(t *testing.T) {
	if MainnetChainConfig.IsBlacklist(big.NewInt(26654674)) {
		t.Fatal("blacklist should not be active at the halt block 26654674")
	}
	if !MainnetChainConfig.IsBlacklist(big.NewInt(26654675)) {
		t.Fatal("blacklist should be active from block 26654675")
	}
}

func TestTestnetBlacklistBlock(t *testing.T) {
	if TestnetConfig.IsBlacklist(big.NewInt(24726499)) {
		t.Fatal("blacklist should not be active before block 24726500")
	}
	if !TestnetConfig.IsBlacklist(big.NewInt(24726500)) {
		t.Fatal("blacklist should be active from block 24726500")
	}
}
