// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// CallContext provides a basic interface for the EVM calling conventions. The EVM
// depends on this context being implemented for doing subcalls and initialising new EVM contracts.
type CallContext interface {
	// Call another contract
	Call(env *EVM, me ContractRef, addr common.Address, data []byte, gas, value *big.Int) ([]byte, error)
	// Take another's contract code and execute within our own context
	CallCode(env *EVM, me ContractRef, addr common.Address, data []byte, gas, value *big.Int) ([]byte, error)
	// Same as CallCode except sender and value is propagated from parent to child scope
	DelegateCall(env *EVM, me ContractRef, addr common.Address, data []byte, gas *big.Int) ([]byte, error)
	// Create a new contract
	Create(env *EVM, me ContractRef, data []byte, gas, value *big.Int) ([]byte, common.Address, error)
}

// EVMRunner provides a simplified API to run EVM calls
// EVM's sender, gasPrice, txFeeRecipient and state are set by the runner on each call
// This object can be re-used many times in contrast to the EVM's single use behaviour.
type EVMRunner interface {
	// Execute performs a potentially write operation over the runner's state
	// It can be seen as a message (input,value) from sender to recipient that returns `ret`
	Execute(recipient common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, err error)

	// ExecuteFrom is like Execute, but lets you specify the sender to use for the EVM call.
	// It exists only for use in the Tobin tax calculation done as part of TobinTransfer, because that
	// originally used the transaction's sender instead of the zero address.
	ExecuteFrom(sender, recipient common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, err error)

	// Query performs a read operation over the runner's state
	// It can be seen as a message (input,value) from sender to recipient that returns `ret`
	Query(recipient common.Address, input []byte, gas uint64) (ret []byte, err error)

	// StopGasMetering backward compatibility method to stop gas metering
	// Deprecated. DO NOT USE
	StopGasMetering()

	// StartGasMetering backward compatibility method to start gas metering
	// Deprecated. DO NOT USE
	StartGasMetering()
}
