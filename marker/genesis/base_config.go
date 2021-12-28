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
				Duration: 60 * Day,
			},
			ValidatorScoreExponent:        10,
			ValidatorScoreAdjustmentSpeed: fixed("0.1"),

			CommissionUpdateDelay: (3 * Day) / 5, // Approximately 3 days with 5s block times

			SlashingPenaltyResetPeriod: 30 * Day,

			DowntimeGracePeriod: 0,

			Commission: fixed("0.1"),
		},
		Election: ElectionParameters{
			MinElectableValidators: 1,
			MaxElectableValidators: 100,
			MaxVotesPerAccount:     bigInt(10),
			ElectabilityThreshold:  fixed("0.001"),
		},

		EpochRewards: EpochRewardsParameters{
			TargetVotingYieldInitial:                     fixed("0"),      // Change to (x + 1) ^ 365 = 1.06 once Mainnet activated.
			TargetVotingYieldAdjustmentFactor:            fixed("0"),      // Change to 1 / 3650 once Mainnet activated.,
			TargetVotingYieldMax:                         fixed("0.0005"), // (x + 1) ^ 365 = 1.20
			RewardsMultiplierMax:                         fixed("2"),
			RewardsMultiplierAdjustmentFactorsUnderspend: fixed("0.5"),
			RewardsMultiplierAdjustmentFactorsOverspend:  fixed("5"),

			// Intentionally set lower than the expected value at steady state to account for the fact that
			// users may take some time to start voting with their cGLD.
			TargetVotingGoldFraction: fixed("0.5"),
			MaxValidatorEpochPayment: bigIntStr("205479452054794520547"), // (75,000 / 365) * 10 ^ 18
			CommunityRewardFraction:  fixed("0.25"),
			//CarbonOffsettingPartner:  common.Address{},
			//CarbonOffsettingFraction: fixed("0.001"),

			Frozen: false,
		},
		LockedGold: LockedGoldParameters{
			UnlockingPeriod: 259200,
		},
		Random: RandomParameters{
			RandomnessBlockRetentionWindow: 720,
		},
		TransferWhitelist: TransferWhitelistParameters{},
		GoldToken: GoldTokenParameters{
			Frozen: false,
		},
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
		Governance: GovernanceParameters{
			UseMultiSig:             true,
			ConcurrentProposals:     3,
			MinDeposit:              bigIntStr("100000000000000000000"), // 100 cGLD
			QueueExpiry:             4 * Week,
			DequeueFrequency:        30 * Minute,
			ApprovalStageDuration:   30 * Minute,
			ReferendumStageDuration: Hour,
			ExecutionStageDuration:  Day,
			ParticipationBaseline:   fixed("0.005"),
			ParticipationFloor:      fixed("0.01"),
			BaselineUpdateFactor:    fixed("0.2"),
			BaselineQuorumFactor:    fixed("1"),
		},
	}
}
