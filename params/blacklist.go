package params

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// blacklistedAccounts lists frozen accounts per chain ID (incident 2026-10-15).
// Consensus enforcement starts at ChainConfig.BlacklistBlock; the tx pool rejects them immediately.
var blacklistedAccounts = map[uint64]map[common.Address]struct{}{
	MainNetChainID: {
		common.HexToAddress("0x8d0B87a76C8D38a96dEdA851e737461eF9d03D92"): {}, // exploiter
		common.HexToAddress("0xF1b596F08aD6A323Ca43A31B92559c5dD8EE5184"): {}, // received 10M MAPO from exploiter
	},
	TestNetChainID: {
		common.HexToAddress("0xf855a761f9182c4b22A04753681A1F6324Ed3449"): {}, // drill only
	},
	DevNetChainID: {
		common.HexToAddress("0xf855a761f9182c4b22A04753681A1F6324Ed3449"): {}, // drill only
	},
}

// IsBlacklistedTx reports whether a tx with the given sender/recipient must be rejected.
// to is nil for contract creation.
func IsBlacklistedTx(chainID *big.Int, from common.Address, to *common.Address) bool {
	if chainID == nil || !chainID.IsUint64() {
		return false
	}
	list, ok := blacklistedAccounts[chainID.Uint64()]
	if !ok {
		return false
	}
	if _, hit := list[from]; hit {
		return true
	}
	if to != nil {
		if _, hit := list[*to]; hit {
			return true
		}
	}
	return false
}
