package relayer

import "math/big"

var (
	CountInEpoch                      = 20
	MaxRedeemHeight            uint64 = 250000   // about 15 days
	NewEpochLength             uint64 = 25000  // about 1.5 days
	ElectionPoint              uint64 = 200
	FirstNewEpochID            uint64 = 1
	DposForkPoint              uint64 = 0
	ElectionMinLimitForStaking        = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
)
