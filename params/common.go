package params

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	ethparams "github.com/ethereum/go-ethereum/params"
)

var (
	Base       = big.NewInt(1e8)
	MaxBaseFee = big.NewInt(5000 * ethparams.GWei)
	MinBaseFee = big.NewInt(100 * ethparams.GWei)
)

var (
	NewRelayerAddress  = common.BytesToAddress([]byte("relayerAddress"))
	HeaderStoreAddress = common.BytesToAddress([]byte("headerstoreAddress"))
	TxVerifyAddress    = common.BytesToAddress([]byte("txVerifyAddress"))
)

const (
	// StateRegisterOnce can be election only once
	StateRegisterOnce uint8 = 1 << iota
	// StateResgisterAuto can be election in every epoch
	StateResgisterAuto
	// StateUnregister can be redeem real time (after MaxRedeemHeight block)
	StateUnregister
	// StateUnregistered flag the asset which is unregistered in the height is redeemed
	StateUnregistered
)
const (
	OpQueryRegister uint8 = 1 << iota
	OpQueryLocked
	OpQueryUnlocking
	OpQueryReward
	OpQueryFine
)

const (
	StateUnusedFlag    = 0xa0
	StateUsedFlag      = 0xa1
	StateSwitchingFlag = 0xa2
	StateRemovedFlag   = 0xa3
	StateAppendFlag    = 0xa4
	// health enter type
	TypeFixed  = 0xa1
	TypeWorked = 0xa2
	TypeBack   = 0xa3
)

var (
	ZeroAddress                  = BytesToAddress([]byte{})
	RegistrySmartContractAddress = common.HexToAddress("0x000000000000000000000000000000000000ce10")

	//AttestationsRegistryId         = makeRegistryId("Attestations")
	BlockchainParametersRegistryId = makeRegistryId("BlockchainParameters")
	ElectionRegistryId             = makeRegistryId("Election")
	EpochRewardsRegistryId         = makeRegistryId("EpochRewards")
	FeeCurrencyWhitelistRegistryId = makeRegistryId("FeeCurrencyWhitelist")
	GasPriceMinimumRegistryId      = makeRegistryId("GasPriceMinimum")
	GoldTokenRegistryId            = makeRegistryId("GoldToken")
	GovernanceRegistryId           = makeRegistryId("Governance")
	LockedGoldRegistryId           = makeRegistryId("LockedGold")
	RandomRegistryId               = makeRegistryId("Random")

	//TransferWhitelistRegistryId    = makeRegistryId("TransferWhitelist")
	ValidatorsRegistryId = makeRegistryId("Validators")
	AccountsId           = makeRegistryId("Accounts")

	// Function is "getOrComputeTobinTax()"
	// selector is first 4 bytes of keccak256 of "getOrComputeTobinTax()"
	// Source:
	// pip3 install pyethereum
	// python3 -c 'from ethereum.utils import sha3; print(sha3("getOrComputeTobinTax()")[0:4].hex())'
	TobinTaxFunctionSelector = hexutil.MustDecode("0x17f9a6f7")

	// Scale factor for the solidity fixidity library
	Fixidity1 = math.BigPow(10, 24)
)


var (
	RegistryProxyAddress      = common.HexToAddress("0xce10")
	ProxyOwnerStorageLocation = common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")
)

func makeRegistryId(contractName string) [32]byte {
	hash := crypto.Keccak256([]byte(contractName))
	var id [32]byte
	copy(id[:], hash)

	return id
}

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped from the left.
func BytesToAddress(b []byte) common.Address {
	var a common.Address
	a.SetBytes(b)
	return a
}
func MustBigInt(str string) *big.Int {
	i, ok := new(big.Int).SetString(str, 10)
	if !ok {
		panic(fmt.Errorf("Invalid string for big.Int: %s", str))
	}
	return i
}

////////////////////////////////////////////////////////////////////////

// DAOForkBlockExtra is the block header extra-data field to set for the DAO fork
// point and a number of consecutive blocks to allow fast/light syncers to correctly
// pick the side they want  ("dao-hard-fork").
var DAOForkBlockExtra = common.FromHex("0x64616f2d686172642d666f726b")

// DAOForkExtraRange is the number of consecutive blocks from the DAO fork point
// to override the extra-data in to prevent no-fork attacks.
var DAOForkExtraRange = big.NewInt(10)

// DAORefundContract is the address of the refund contract to send DAO balances to.
var DAORefundContract = common.HexToAddress("0xbf4ed7b27f1d666546e30d74d50d173d20bca754")

// DAODrainList is the list of accounts whose full balances will be moved into a
// refund contract at the beginning of the dao-fork block.
func DAODrainList() []common.Address {
	return []common.Address{}
}
