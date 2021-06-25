package vm

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var (
	errInsufficientBalanceForGas         = errors.New("insufficient balance to from for gas")
	errInsufficientBalanceForPayerForGas = errors.New("insufficient balance to payer for gas")
)

// Message represents a message sent to a contract.
type Message interface {
	Payment() common.Address
	From() common.Address
	//FromFrontier() (common.Address, error)
	To() *common.Address

	GasPrice() *big.Int
	Gas() uint64
	Value() *big.Int
	Fee() *big.Int

	Nonce() uint64
	CheckNonce() bool
	Data() []byte
}

