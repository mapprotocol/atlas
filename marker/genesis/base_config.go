package genesis

import (
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

	////a epoch award 1,500,000map = 300,000,000(one year award)/6000,000(number a year) *30000(one epoch number)
	perNumberReward := new(big.Int).Div(big.NewInt(300000000), big.NewInt(6000000))
	perEpochReward := new(big.Int).Mul(perNumberReward, big.NewInt(int64(params.Epoch))) //MAP
	perEpochReward = new(big.Int).Mul(perEpochReward, big.NewInt(1e18))                  //wei
	return &Config{
		Validators: ValidatorsParameters{

			ValidatorLockedGoldRequirements: LockedGoldRequirements{
				Value: bigIntStr("1000000000000000000000000"), //1000,000e18
				// MUST BE KEPT IN SYNC WITH MEMBERSHIP HISTORY LENGTH
				Duration: 60 * Day,
			},
			ValidatorScoreExponent:        10,
			ValidatorScoreAdjustmentSpeed: fixed("0.2"),
			PledgeMultiplierInReward:      fixed("1"),
			CommissionUpdateDelay:         (3 * Day) / 5, // Approximately 3 days with 5s block times

			SlashingPenaltyResetPeriod: 30 * Day,

			DowntimeGracePeriod: 0,

			//Commission: fixed("0.1"),
			Commission: bigInt(100000), // 0.1 be relative to 1000000
		},
		Election: ElectionParameters{
			MinElectableValidators: 1,
			MaxElectableValidators: 100,
			MaxVotesPerAccount:     bigInt(10),
			ElectabilityThreshold:  fixed("0.001"),
		},
		EpochRewards: EpochRewardsParameters{
			//a epoch award 1,500,000map = 300,000,000(one year award)/6000,000(number a year) *30000(one epoch number)
			//MaxValidatorEpochPayment = 1,500,000map *(2/3)
			//MaxRelayerEpochPayment   = 1,500,000map *(1/3)
			//MaxEpochPayment: bigIntStr("1500000000000000000000000") , //Validator Relayer
			MaxEpochPayment: perEpochReward, //Validator Relayer

			CommunityRewardFraction:     fixed("0"),
			CommunityPartner:            params.ZeroAddress,
			EpochRelayerPaymentFraction: fixed("0.33333333333333333333"),
		},
		LockedGold: LockedGoldParameters{
			UnlockingPeriod: 1296000, //15 day
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
