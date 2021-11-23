package params

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	ethparams "github.com/ethereum/go-ethereum/params"

	"math/big"
)

var (
	baseUnit           = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	FbaseUnit          = new(big.Float).SetFloat64(float64(baseUnit.Int64()))
	Base               = new(big.Int).SetUint64(ethparams.InitialBaseFee)
	InvalidFee         = big.NewInt(65535)
	RelayerAddress     = common.BytesToAddress([]byte("RelayerAddress"))
	HeaderStoreAddress = common.BytesToAddress([]byte("headerstoreAddress"))
	TxVerifyAddress    = common.BytesToAddress([]byte("txVerifyAddress"))
)

var RelayerGas = map[string]uint64{
	"getBalance":      450000,
	"register":        2400000,
	"append":          2400000,
	"withdraw":        2520000,
	"getPeriodHeight": 450000,
	"getRelayers":     450000,
}

var (
	CountInEpoch                       = 100
	MaxRedeemHeight                    = Epoch
	NewEpochLength                     = Epoch
	ElectionPoint               uint64 = 20
	FirstNewEpochID             uint64 = 1
	PowForkPoint                uint64 = 0
	ElectionMinLimitForRegister        = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
	MinWorkEfficiency           uint64 = 1 //every relayer sync 1 block at least
)

var (
	ErrInvalidParam      = errors.New("Invalid Param")
	ErrOverEpochID       = errors.New("Over epoch id")
	ErrNotSequential     = errors.New("epoch id not sequential")
	ErrInvalidEpochInfo  = errors.New("Invalid epoch info")
	ErrNotFoundEpoch     = errors.New("cann't found the epoch info")
	ErrInvalidRegister   = errors.New("Invalid register account")
	ErrMatchEpochID      = errors.New("wrong match epoch id in a reward block")
	ErrNotRegister       = errors.New("Not match the register account")
	ErrNotDelegation     = errors.New("Not match the account")
	ErrNotMatchEpochInfo = errors.New("the epoch info is not match with accounts")
	ErrNotElectionTime   = errors.New("not time to election the next relayer")
	ErrAmountOver        = errors.New("the amount more than register amount")
	ErrDelegationSelf    = errors.New("wrong")
	ErrRedeemAmount      = errors.New("wrong redeem amount")
	ErrForbidAddress     = errors.New("Forbidding Address")
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
	FreezerRegistryId              = makeRegistryId("Freezer")
	GasPriceMinimumRegistryId      = makeRegistryId("GasPriceMinimum")
	GoldTokenRegistryId            = makeRegistryId("GoldToken")
	GovernanceRegistryId           = makeRegistryId("Governance")
	LockedGoldRegistryId           = makeRegistryId("LockedGold")
	RandomRegistryId               = makeRegistryId("Random")
	ReserveRegistryId              = makeRegistryId("Reserve")
	SortedOraclesRegistryId        = makeRegistryId("SortedOracles")
	StableTokenRegistryId          = makeRegistryId("StableToken")
	//TransferWhitelistRegistryId    = makeRegistryId("TransferWhitelist")
	ValidatorsRegistryId = makeRegistryId("Validators")

	// Function is "getOrComputeTobinTax()"
	// selector is first 4 bytes of keccak256 of "getOrComputeTobinTax()"
	// Source:
	// pip3 install pyethereum
	// python3 -c 'from ethereum.utils import sha3; print(sha3("getOrComputeTobinTax()")[0:4].hex())'
	TobinTaxFunctionSelector = hexutil.MustDecode("0x17f9a6f7")

	// Scale factor for the solidity fixidity library
	Fixidity1 = math.BigPow(10, 24)
)

const (
	MaximumExtraDataSize uint64 = 32 // Maximum size extra data may be after Genesis.
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
