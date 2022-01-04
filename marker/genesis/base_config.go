package genesis

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/helper/decimal/fixed"
	"github.com/mapprotocol/atlas/params"
	"math/big"
)

// BaseConfig creates base parameters for atlas
// Callers must complete missing pieces
func BaseConfig() *Config {
	bigInt := big.NewInt
	bigIntStr := params.MustBigInt
	fixed := fixed.MustNew

	return &Config{

		GasPriceMinimum: GasPriceMinimumParameters{
			MinimumFloor:    bigInt(100000000),
			AdjustmentSpeed: fixed("0.5"),
			TargetDensity:   fixed("0.5"),
		},

		Validators: ValidatorsParameters{

			ValidatorLockedGoldRequirements: LockedGoldRequirements{
				Value: bigIntStr("10000000000000000000000"), // 10k Atlas
				// MUST BE KEPT IN SYNC WITH MEMBERSHIP HISTORY LENGTH
				//Duration: 60 * Day,// todo zhangwei
				Duration: 1 * Second, // todo zhangwei
			},
			ValidatorScoreExponent:        10,
			ValidatorScoreAdjustmentSpeed: fixed("0.1"),
			PledgeMultiplierInReward:      fixed("1"),
			CommissionUpdateDelay:         (3 * Day) / 5, // Approximately 3 days with 5s block times

			SlashingPenaltyResetPeriod: 30 * Day,

			DowntimeGracePeriod: 0,

			Commission: fixed("0.1"),
		},
		Election: ElectionParameters{
			MinElectableValidators: 1,
			MaxElectableValidators: 100,
			MaxVotesPerAccount:     bigInt(10),
			ElectabilityThreshold:  fixed("0"),
		},

		EpochRewards: EpochRewardsParameters{
			MaxValidatorEpochPayment: bigIntStr("10000000000000000000000"), //10K map
			CommunityRewardFraction:  fixed("0.1"),
			CommunityPartner:         common.Address{},
		},
		LockedGold: LockedGoldParameters{
			UnlockingPeriod: 1,
		},
		Random: RandomParameters{
			RandomnessBlockRetentionWindow: 720,
		},
		GoldToken: GoldTokenParameters{},
		Blockchain: BlockchainParameters{
			Version:                 Version{1, 0, 0},
			GasForNonGoldCurrencies: 50000,
			BlockGasLimit:           13000000,
		},
		DoubleSigningSlasher: DoubleSigningSlasherParameters{
			Reward:  bigIntStr("1000000000000000000000"), // 1000 cGLD
			Penalty: bigIntStr("9000000000000000000000"), // 9000 cGLD
		},
		DowntimeSlasher: DowntimeSlasherParameters{
			Reward:            bigIntStr("10000000000000000000"),  // 10 cGLD
			Penalty:           bigIntStr("100000000000000000000"), // 100 cGLD
			SlashableDowntime: 4,                                  // make it small so it works with small epoch sizes, e.g. 10
		},
	}
}
