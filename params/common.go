package params

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var (
	baseUnit       = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	FbaseUnit      = new(big.Float).SetFloat64(float64(baseUnit.Int64()))
	Base           = new(big.Int).SetUint64(10000)
	InvalidFee     = big.NewInt(65535)
	RelayerAddress = common.BytesToAddress([]byte("truestaking"))
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
	CountInEpoch                       = 10
	MaxRedeemHeight             uint64 = 20000
	NewEpochLength              uint64 = 10000
	ElectionPoint               uint64 = 100
	FirstNewEpochID             uint64 = 1
	PowForkPoint                uint64 = 0
	ElectionMinLimitForRegister        = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
	MinWorkEfficiency           uint64 = 50 //every relayer generate 100 block at least
)

var (
	ErrInvalidParam      = errors.New("Invalid Param")
	ErrOverEpochID       = errors.New("Over epoch id")
	ErrNotSequential     = errors.New("epoch id not sequential")
	ErrInvalidEpochInfo  = errors.New("Invalid epoch info")
	ErrNotFoundEpoch     = errors.New("cann't found the epoch info")
	ErrInvalidStaking    = errors.New("Invalid staking account")
	ErrMatchEpochID      = errors.New("wrong match epoch id in a reward block")
	ErrNotStaking        = errors.New("Not match the staking account")
	ErrNotDelegation     = errors.New("Not match the delegation account")
	ErrNotMatchEpochInfo = errors.New("the epoch info is not match with accounts")
	ErrNotElectionTime   = errors.New("not time to election the next committee")
	ErrAmountOver        = errors.New("the amount more than staking amount")
	ErrDelegationSelf    = errors.New("Cann't delegation myself")
	ErrRedeemAmount      = errors.New("wrong redeem amount")
	ErrForbidAddress     = errors.New("Forbidding Address")
	ErrRepeatPk          = errors.New("repeat PK on staking tx")
)

const (
	// StateRegisterOnce can be election only once
	StateRegisterOnce uint8 = 1 << iota
	// StateResgisterAuto can be election in every epoch
	StateResgisterAuto
	StateStakingCancel
	// StateRedeem can be redeem real time (after MaxRedeemHeight block)
	StateRedeem
	// StateRedeemed flag the asset which is staking in the height is redeemed
	StateRedeemed
)
const (
	OpQueryStaking uint8 = 1 << iota
	OpQueryLocked
	OpQueryCancelable
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
	BaseBig       = big.NewInt(1e18)
	NewRewardCoin = new(big.Int).Mul(big.NewInt(570), BaseBig)
)

func GetReward() *big.Int {
	return new(big.Int).Set(NewRewardCoin)
}
