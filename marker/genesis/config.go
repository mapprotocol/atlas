package genesis

import (
	"github.com/mapprotocol/atlas/helper/decimal/bigintstr"
	"github.com/mapprotocol/atlas/helper/decimal/fixed"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/marker/internal/utils"
	"github.com/mapprotocol/atlas/params"
)

// durations in seconds
const (
	Second = 1
	Minute = 60 * Second
	Hour   = 60 * Minute
	Day    = 24 * Hour
	Week   = 7 * Day
	Year   = 365 * Day
)

// Config represent all atlas-blockchain configuration options for the genesis block
type Config struct {
	ChainID          *big.Int              `json:"chainId"` // chainId identifies the current chain and is used for replay protection
	Istanbul         params.IstanbulConfig `json:"istanbul"`
	Hardforks        HardforkConfig        `json:"hardforks"`
	GenesisTimestamp uint64                `json:"genesisTimestamp"`

	GasPriceMinimum      GasPriceMinimumParameters
	LockedGold           LockedGoldParameters
	GoldToken            GoldTokenParameters
	Validators           ValidatorsParameters
	Election             ElectionParameters
	EpochRewards         EpochRewardsParameters
	Blockchain           BlockchainParameters
	Random               RandomParameters
	DoubleSigningSlasher DoubleSigningSlasherParameters
	DowntimeSlasher      DowntimeSlasherParameters
}

// Save will write config into a json file
func (cfg *Config) Save(filepath string) error {
	return utils.WriteJson(cfg, filepath)
}

// ChainConfig returns the chain config objt for the blockchain
func (cfg *Config) ChainConfig() *params.ChainConfig {
	return &params.ChainConfig{
		ChainID:             cfg.ChainID,
		HomesteadBlock:      common.Big0,
		EIP150Block:         common.Big0,
		EIP150Hash:          common.Hash{},
		EIP155Block:         common.Big0,
		EIP158Block:         common.Big0,
		ByzantiumBlock:      common.Big0,
		ConstantinopleBlock: common.Big0,
		PetersburgBlock:     common.Big0,
		IstanbulBlock:       common.Big0,

		//ChurritoBlock: cfg.Hardforks.ChurritoBlock,
		DonutBlock: cfg.Hardforks.DonutBlock,

		Istanbul: &params.IstanbulConfig{
			Epoch:          cfg.Istanbul.Epoch,
			ProposerPolicy: cfg.Istanbul.ProposerPolicy,
			LookbackWindow: cfg.Istanbul.LookbackWindow,
			BlockPeriod:    cfg.Istanbul.BlockPeriod,
			RequestTimeout: cfg.Istanbul.RequestTimeout,
		},
	}
}

// HardforkConfig contains atlas hardforks activation blocks
type HardforkConfig struct {
	ChurritoBlock *big.Int `json:"churritoBlock"`
	DonutBlock    *big.Int `json:"donutBlock"`
}

// LockedGoldRequirements represents value/duration requirments on locked gold
type LockedGoldRequirements struct {
	Value    *big.Int `json:"value"`
	Duration uint64   `json:"duration"`
}

type LockedgoldRequirementsMarshaling struct {
	Value *bigintstr.BigIntStr `json:"value"`
}

// ElectionParameters are the initial configuration parameters for Elections
type ElectionParameters struct {
	MinElectableValidators uint64       `json:"minElectableValidators"`
	MaxElectableValidators uint64       `json:"maxElectableValidators"`
	MaxVotesPerAccount     *big.Int     `json:"maxVotesPerAccount"`
	ElectabilityThreshold  *fixed.Fixed `json:"electabilityThreshold"`
}

type ElectionParametersMarshaling struct {
	MaxVotesPerAccount *bigintstr.BigIntStr `json:"maxVotesPerAccount"`
}

// Version represents an artifact version number
type Version struct {
	Major int64 `json:"major"`
	Minor int64 `json:"minor"`
	Patch int64 `json:"patch"`
}

// BlockchainParameters are the initial configuration parameters for Blockchain
type BlockchainParameters struct {
	Version                 Version `json:"version"`
	GasForNonGoldCurrencies uint64  `json:"gasForNonGoldCurrencies"`
	BlockGasLimit           uint64  `json:"blockGasLimit"`
}

// DoubleSigningSlasherParameters are the initial configuration parameters for DoubleSigningSlasher
type DoubleSigningSlasherParameters struct {
	Penalty *big.Int `json:"penalty"`
	Reward  *big.Int `json:"reward"`
}

type DoubleSigningSlasherParametersMarshaling struct {
	Penalty *bigintstr.BigIntStr `json:"penalty"`
	Reward  *bigintstr.BigIntStr `json:"reward"`
}

// DowntimeSlasherParameters are the initial configuration parameters for DowntimeSlasher
type DowntimeSlasherParameters struct {
	Penalty           *big.Int `json:"penalty"`
	Reward            *big.Int `json:"reward"`
	SlashableDowntime uint64   `json:"slashableDowntime"`
}

type DowntimeSlasherParametersMarshaling struct {
	Penalty *bigintstr.BigIntStr `json:"penalty"`
	Reward  *bigintstr.BigIntStr `json:"reward"`
}

type GovernanceParametersMarshaling struct {
	MinDeposit *bigintstr.BigIntStr `json:"minDeposit"`
}

// ValidatorsParameters are the initial configuration parameters for Validators
type ValidatorsParameters struct {
	ValidatorLockedGoldRequirements LockedGoldRequirements `json:"validatorLockedGoldRequirements"`
	ValidatorScoreExponent          uint64                 `json:"validatorScoreExponent"`
	ValidatorScoreAdjustmentSpeed   *fixed.Fixed           `json:"validatorScoreAdjustmentSpeed"`
	SlashingPenaltyResetPeriod      uint64                 `json:"slashingPenaltyResetPeriod"`
	CommissionUpdateDelay           uint64                 `json:"commissionUpdateDelay"`
	PledgeMultiplierInReward        *fixed.Fixed           `json:"pledgeMultiplierInReward"`
	DowntimeGracePeriod             uint64                 `json:"downtimeGracePeriod"`
	Commission                      *fixed.Fixed           `json:"commission"` // commission for genesis registered validator
}

// EpochRewardsParameters are the initial configuration parameters for EpochRewards
type EpochRewardsParameters struct {
	MaxValidatorEpochPayment *big.Int       `json:"maxValidatorEpochPayment"`
	CommunityRewardFraction  *fixed.Fixed   `json:"communityRewardFraction"`
	CommunityPartner         common.Address `json:"communityPartner"`
}

type EpochRewardsParametersMarshaling struct {
	MaxValidatorEpochPayment *bigintstr.BigIntStr `json:"maxValidatorEpochPayment"`
}

// GoldTokenParameters are the initial configuration parameters for GoldToken
type GoldTokenParameters struct {
	InitialBalances BalanceList `json:"initialBalances"`
}

// RandomParameters are the initial configuration parameters for Random
type RandomParameters struct {
	RandomnessBlockRetentionWindow uint64 `json:"randomnessBlockRetentionWindow"`
}

// GasPriceMinimumParameters are the initial configuration parameters for GasPriceMinimum
type GasPriceMinimumParameters struct {
	MinimumFloor    *big.Int     `json:"minimumFloor"`
	TargetDensity   *fixed.Fixed `json:"targetDensity"`
	AdjustmentSpeed *fixed.Fixed `json:"adjustmentSpeed"`
}

type GasPriceMinimumParametersMarshaling struct {
	MinimumFloor *bigintstr.BigIntStr `json:"minimumFloor"`
}

// LockedGoldParameters are the initial configuration parameters for LockedGold
type LockedGoldParameters struct {
	UnlockingPeriod uint64 `json:"unlockingPeriod"`
}

// Balance represents an account and it's initial balance in wei
type Balance struct {
	Account common.Address `json:"account"`
	Amount  *big.Int       `json:"amount"`
}

type BalanceMarshaling struct {
	Amount *bigintstr.BigIntStr `json:"amount"`
}

// BalanceList list of balances
type BalanceList []Balance

// Accounts returns all the addresses
func (bl BalanceList) Accounts() []common.Address {
	res := make([]common.Address, len(bl))
	for i, x := range bl {
		res[i] = x.Account
	}
	return res
}

// Amounts returns all the amounts
func (bl BalanceList) Amounts() []*big.Int {
	res := make([]*big.Int, len(bl))
	for i, x := range bl {
		res[i] = x.Amount
	}
	return res
}
