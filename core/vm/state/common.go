package state

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)
var StakingAddress  = common.BytesToAddress([]byte("truestaking"))
type BalanceInfo struct {
	Address common.Address `json:"address"`
	Valid   *big.Int       `json:"valid"`
	Lock    *big.Int       `json:"lock"`
}
