package vm

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/vm"
	"sort"

	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

var activators = map[int]func(*JumpTable){
	2929: enable2929,
	2200: enable2200,
	1884: enable1884,
	1344: enable1344,
}

// EnableEIP enables the given EIP on the config.
// This operation writes in-place, and callers need to ensure that the globally
// defined jump tables are not polluted.
func EnableEIP(eipNum int, jt *JumpTable) error {
	enablerFn, ok := activators[eipNum]
	if !ok {
		return fmt.Errorf("undefined eip %d", eipNum)
	}
	enablerFn(jt)
	return nil
}

func ValidEip(eipNum int) bool {
	_, ok := activators[eipNum]
	return ok
}
func ActivateableEips() []string {
	var nums []string
	for k := range activators {
		nums = append(nums, fmt.Sprintf("%d", k))
	}
	sort.Strings(nums)
	return nums
}

// enable1884 applies EIP-1884 to the given jump table:
// - Increase cost of BALANCE to 700
// - Increase cost of EXTCODEHASH to 700
// - Increase cost of SLOAD to 800
// - Define SELFBALANCE, with cost GasFastStep (5)
func enable1884(jt *JumpTable) {
	// Gas cost changes
	jt[vm.SLOAD].constantGas = params.SloadGasEIP1884
	jt[vm.BALANCE].constantGas = params.BalanceGasEIP1884
	jt[vm.EXTCODEHASH].constantGas = params.ExtcodeHashGasEIP1884

	// New opcode
	jt[vm.SELFBALANCE] = &operation{
		execute:     opSelfBalance,
		constantGas: vm.GasFastStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
	}
}

func opSelfBalance(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	balance, _ := uint256.FromBig(interpreter.evm.StateDB.GetBalance(scope.Contract.Address()))
	scope.Stack.push(balance)
	return nil, nil
}

// enable1344 applies EIP-1344 (ChainID Opcode)
// - Adds an opcode that returns the current chain’s EIP-155 unique identifier
func enable1344(jt *JumpTable) {
	// New opcode
	jt[vm.CHAINID] = &operation{
		execute:     opChainID,
		constantGas: vm.GasQuickStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
	}
}

// opChainID implements CHAINID opcode
func opChainID(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	chainId, _ := uint256.FromBig(interpreter.evm.chainConfig.ChainID)
	scope.Stack.push(chainId)
	return nil, nil
}

// enable2200 applies EIP-2200 (Rebalance net-metered SSTORE)
func enable2200(jt *JumpTable) {
	jt[vm.SLOAD].constantGas = params.SloadGasEIP2200
	jt[vm.SSTORE].dynamicGas = gasSStoreEIP2200
}

// enable2929 enables "EIP-2929: Gas cost increases for state access opcodes"
// https://eips.ethereum.org/EIPS/eip-2929
func enable2929(jt *JumpTable) {
	jt[vm.SSTORE].dynamicGas = gasSStoreEIP2929

	jt[vm.SLOAD].constantGas = 0
	jt[vm.SLOAD].dynamicGas = gasSLoadEIP2929

	jt[vm.EXTCODECOPY].constantGas = vm.WarmStorageReadCostEIP2929
	jt[vm.EXTCODECOPY].dynamicGas = gasExtCodeCopyEIP2929

	jt[vm.EXTCODESIZE].constantGas = vm.WarmStorageReadCostEIP2929
	jt[vm.EXTCODESIZE].dynamicGas = gasEip2929AccountCheck

	jt[vm.EXTCODEHASH].constantGas = vm.WarmStorageReadCostEIP2929
	jt[vm.EXTCODEHASH].dynamicGas = gasEip2929AccountCheck

	jt[vm.BALANCE].constantGas = vm.WarmStorageReadCostEIP2929
	jt[vm.BALANCE].dynamicGas = gasEip2929AccountCheck

	jt[vm.CALL].constantGas = vm.WarmStorageReadCostEIP2929
	jt[vm.CALL].dynamicGas = gasCallEIP2929

	jt[vm.CALLCODE].constantGas = vm.WarmStorageReadCostEIP2929
	jt[vm.CALLCODE].dynamicGas = gasCallCodeEIP2929

	jt[vm.STATICCALL].constantGas = vm.WarmStorageReadCostEIP2929
	jt[vm.STATICCALL].dynamicGas = gasStaticCallEIP2929

	jt[vm.DELEGATECALL].constantGas = vm.WarmStorageReadCostEIP2929
	jt[vm.DELEGATECALL].dynamicGas = gasDelegateCallEIP2929

	// This was previously part of the dynamic cost, but we're using it as a constantGas
	// factor here
	jt[vm.SELFDESTRUCT].constantGas = params.SelfdestructGasEIP150
	jt[vm.SELFDESTRUCT].dynamicGas = gasSelfdestructEIP2929
}

