package vm

import (
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

type (
	executionFunc func(pc *uint64, interpreter *EVMInterpreter, callContext *ScopeContext) ([]byte, error)
	gasFunc       func(*EVM, *Contract, *Stack, *Memory, uint64) (uint64, error) // last parameter is the requested memory size as a uint64
	// memorySizeFunc returns the required size, and whether the operation overflowed a uint64
	memorySizeFunc func(*Stack) (size uint64, overflow bool)
)

type operation struct {
	// execute is the operation function
	execute     executionFunc
	constantGas uint64
	dynamicGas  gasFunc
	// minStack tells how many stack items are required
	minStack int
	// maxStack specifies the max length the stack can have for this operation
	// to not overflow the stack.
	maxStack int

	// memorySize returns the memory size required for the operation
	memorySize memorySizeFunc

	halts   bool // indicates whether the operation should halt further execution
	jumps   bool // indicates whether the program counter should not increment
	writes  bool // determines whether this a state modifying operation
	reverts bool // determines whether the operation reverts state (implicitly halts)
	returns bool // determines whether the operations sets the return data content
}

var (
	frontierInstructionSet         = newFrontierInstructionSet()
	homesteadInstructionSet        = newHomesteadInstructionSet()
	tangerineWhistleInstructionSet = newTangerineWhistleInstructionSet()
	spuriousDragonInstructionSet   = newSpuriousDragonInstructionSet()
	byzantiumInstructionSet        = newByzantiumInstructionSet()
	constantinopleInstructionSet   = newConstantinopleInstructionSet()
	istanbulInstructionSet         = newIstanbulInstructionSet()
	berlinInstructionSet           = newBerlinInstructionSet()
	//yoloV1InstructionSet           = newYoloV1InstructionSet()
)
type JumpTable [256]*operation

//func newYoloV1InstructionSet() JumpTable {
//	instructionSet := newIstanbulInstructionSet()
//
//	enable2315(&instructionSet) // Subroutines - https://eips.ethereum.org/EIPS/eip-2315
//
//	return instructionSet
//}
func newIstanbulInstructionSet() JumpTable {
	instructionSet := newConstantinopleInstructionSet()

	enable1344(&instructionSet) // ChainID opcode - https://eips.ethereum.org/EIPS/eip-1344
	enable1884(&instructionSet) // Reprice reader opcodes - https://eips.ethereum.org/EIPS/eip-1884
	enable2200(&instructionSet) // Net metered SSTORE - https://eips.ethereum.org/EIPS/eip-2200

	return instructionSet
}

func newBerlinInstructionSet() JumpTable {
	instructionSet := newIstanbulInstructionSet()
	enable2929(&instructionSet) // Access lists for trie accesses https://eips.ethereum.org/EIPS/eip-2929
	return instructionSet
}


func newHomesteadInstructionSet() JumpTable {
	instructionSet := newFrontierInstructionSet()
	instructionSet[vm.DELEGATECALL] = &operation{
		execute:     opDelegateCall,
		dynamicGas:  gasDelegateCall,
		constantGas: params.CallGasFrontier,
		minStack:    minStack(6, 1),
		maxStack:    maxStack(6, 1),
		memorySize:  memoryDelegateCall,
		returns:     true,
	}
	return instructionSet
}

func newTangerineWhistleInstructionSet() JumpTable {
	instructionSet := newHomesteadInstructionSet()
	instructionSet[vm.BALANCE].constantGas = params.BalanceGasEIP150
	instructionSet[vm.EXTCODESIZE].constantGas = params.ExtcodeSizeGasEIP150
	instructionSet[vm.SLOAD].constantGas = params.SloadGasEIP150
	instructionSet[vm.EXTCODECOPY].constantGas = params.ExtcodeCopyBaseEIP150
	instructionSet[vm.CALL].constantGas = params.CallGasEIP150
	instructionSet[vm.CALLCODE].constantGas = params.CallGasEIP150
	instructionSet[vm.DELEGATECALL].constantGas = params.CallGasEIP150
	return instructionSet
}
func newSpuriousDragonInstructionSet() JumpTable {
	instructionSet := newTangerineWhistleInstructionSet()
	instructionSet[vm.EXP].dynamicGas = gasExpEIP158
	return instructionSet
}

func newByzantiumInstructionSet() JumpTable {
	instructionSet := newSpuriousDragonInstructionSet()
	instructionSet[vm.STATICCALL] = &operation{
		execute:     opStaticCall,
		constantGas: params.CallGasEIP150,
		dynamicGas:  gasStaticCall,
		minStack:    minStack(6, 1),
		maxStack:    maxStack(6, 1),
		memorySize:  memoryStaticCall,
		returns:     true,
	}
	instructionSet[vm.RETURNDATASIZE] = &operation{
		execute:     opReturnDataSize,
		constantGas: GasQuickStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
	}
	instructionSet[vm.RETURNDATACOPY] = &operation{
		execute:     opReturnDataCopy,
		constantGas: GasFastestStep,
		dynamicGas:  gasReturnDataCopy,
		minStack:    minStack(3, 0),
		maxStack:    maxStack(3, 0),
		memorySize:  memoryReturnDataCopy,
	}
	instructionSet[vm.REVERT] = &operation{
		execute:    opRevert,
		dynamicGas: gasRevert,
		minStack:   minStack(2, 0),
		maxStack:   maxStack(2, 0),
		memorySize: memoryRevert,
		reverts:    true,
		returns:    true,
	}
	return instructionSet
}

func newConstantinopleInstructionSet() JumpTable {
	instructionSet := newByzantiumInstructionSet()
	instructionSet[vm.SHL] = &operation{
		execute:     opSHL,
		constantGas: GasFastestStep,
		minStack:    minStack(2, 1),
		maxStack:    maxStack(2, 1),
	}
	instructionSet[vm.SHR] = &operation{
		execute:     opSHR,
		constantGas: GasFastestStep,
		minStack:    minStack(2, 1),
		maxStack:    maxStack(2, 1),
	}
	instructionSet[vm.SAR] = &operation{
		execute:     opSAR,
		constantGas: GasFastestStep,
		minStack:    minStack(2, 1),
		maxStack:    maxStack(2, 1),
	}
	instructionSet[vm.EXTCODEHASH] = &operation{
		execute:     opExtCodeHash,
		constantGas: params.ExtcodeHashGasConstantinople,
		minStack:    minStack(1, 1),
		maxStack:    maxStack(1, 1),
	}
	instructionSet[vm.CREATE2] = &operation{
		execute:     opCreate2,
		constantGas: params.Create2Gas,
		dynamicGas:  gasCreate2,
		minStack:    minStack(4, 1),
		maxStack:    maxStack(4, 1),
		memorySize:  memoryCreate2,
		writes:      true,
		returns:     true,
	}
	return instructionSet
}

func newFrontierInstructionSet() JumpTable {
	return JumpTable{
		vm.STOP: {
			execute:     opStop,
			constantGas: 0,
			minStack:    minStack(0, 0),
			maxStack:    maxStack(0, 0),
			halts:       true,
		},
		vm.ADD: {
			execute:     opAdd,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.MUL: {
			execute:     opMul,
			constantGas: vm.GasFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.SUB: {
			execute:     opSub,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.DIV: {
			execute:     opDiv,
			constantGas: vm.GasFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.SDIV: {
			execute:     opSdiv,
			constantGas: vm.GasFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.MOD: {
			execute:     opMod,
			constantGas: vm.GasFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.SMOD: {
			execute:     opSmod,
			constantGas: vm.GasFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.ADDMOD: {
			execute:     opAddmod,
			constantGas: vm.GasMidStep,
			minStack:    minStack(3, 1),
			maxStack:    maxStack(3, 1),
		},
		vm.MULMOD: {
			execute:     opMulmod,
			constantGas: vm.GasMidStep,
			minStack:    minStack(3, 1),
			maxStack:    maxStack(3, 1),
		},
		vm.EXP: {
			execute:    opExp,
			dynamicGas: gasExpFrontier,
			minStack:   minStack(2, 1),
			maxStack:   maxStack(2, 1),
		},
		vm.SIGNEXTEND: {
			execute:     opSignExtend,
			constantGas: vm.GasFastStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.LT: {
			execute:     opLt,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.GT: {
			execute:     opGt,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.SLT: {
			execute:     opSlt,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.SGT: {
			execute:     opSgt,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.EQ: {
			execute:     opEq,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.ISZERO: {
			execute:     opIszero,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
		},
		vm.AND: {
			execute:     opAnd,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.XOR: {
			execute:     opXor,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.OR: {
			execute:     opOr,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.NOT: {
			execute:     opNot,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
		},
		vm.BYTE: {
			execute:     opByte,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
		},
		vm.SHA3: {
			execute:     opSha3,
			constantGas: params.Sha3Gas,
			dynamicGas:  gasSha3,
			minStack:    minStack(2, 1),
			maxStack:    maxStack(2, 1),
			memorySize:  memorySha3,
		},
		vm.ADDRESS: {
			execute:     opAddress,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.BALANCE: {
			execute:     opBalance,
			constantGas: params.BalanceGasFrontier,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
		},
		vm.ORIGIN: {
			execute:     opOrigin,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.CALLER: {
			execute:     opCaller,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.CALLVALUE: {
			execute:     opCallValue,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.CALLDATALOAD: {
			execute:     opCallDataLoad,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
		},
		vm.CALLDATASIZE: {
			execute:     opCallDataSize,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.CALLDATACOPY: {
			execute:     opCallDataCopy,
			constantGas: vm.GasFastestStep,
			dynamicGas:  gasCallDataCopy,
			minStack:    minStack(3, 0),
			maxStack:    maxStack(3, 0),
			memorySize:  memoryCallDataCopy,
		},
		vm.CODESIZE: {
			execute:     opCodeSize,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.CODECOPY: {
			execute:     opCodeCopy,
			constantGas: vm.GasFastestStep,
			dynamicGas:  gasCodeCopy,
			minStack:    minStack(3, 0),
			maxStack:    maxStack(3, 0),
			memorySize:  memoryCodeCopy,
		},
		vm.GASPRICE: {
			execute:     opGasprice,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.EXTCODESIZE: {
			execute:     opExtCodeSize,
			constantGas: params.ExtcodeSizeGasFrontier,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
		},
		vm.EXTCODECOPY: {
			execute:     opExtCodeCopy,
			constantGas: params.ExtcodeCopyBaseFrontier,
			dynamicGas:  gasExtCodeCopy,
			minStack:    minStack(4, 0),
			maxStack:    maxStack(4, 0),
			memorySize:  memoryExtCodeCopy,
		},
		vm.BLOCKHASH: {
			execute:     opBlockhash,
			constantGas: vm.GasExtStep,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
		},
		vm.COINBASE: {
			execute:     opCoinbase,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.TIMESTAMP: {
			execute:     opTimestamp,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	NUMBER: {
			execute:     opNumber,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	DIFFICULTY: {
			execute:     opDifficulty,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	GASLIMIT: {
			execute:     opGasLimit,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	POP: {
			execute:     opPop,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(1, 0),
			maxStack:    maxStack(1, 0),
		},
		vm.	MLOAD: {
			execute:     opMload,
			constantGas: vm.GasFastestStep,
			dynamicGas:  gasMLoad,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
			memorySize:  memoryMLoad,
		},
		vm.MSTORE: {
			execute:     opMstore,
			constantGas: vm.GasFastestStep,
			dynamicGas:  gasMStore,
			minStack:    minStack(2, 0),
			maxStack:    maxStack(2, 0),
			memorySize:  memoryMStore,
		},
		vm.	MSTORE8: {
			execute:     opMstore8,
			constantGas: vm.GasFastestStep,
			dynamicGas:  gasMStore8,
			memorySize:  memoryMStore8,
			minStack:    minStack(2, 0),
			maxStack:    maxStack(2, 0),
		},
		vm.	SLOAD: {
			execute:     opSload,
			constantGas: params.SloadGasFrontier,
			minStack:    minStack(1, 1),
			maxStack:    maxStack(1, 1),
		},
		vm.SSTORE: {
			execute:    opSstore,
			dynamicGas: gasSStore,
			minStack:   minStack(2, 0),
			maxStack:   maxStack(2, 0),
			writes:     true,
		},
		vm.JUMP: {
			execute:     opJump,
			constantGas: vm.GasMidStep,
			minStack:    minStack(1, 0),
			maxStack:    maxStack(1, 0),
			jumps:       true,
		},
		vm.	JUMPI: {
			execute:     opJumpi,
			constantGas: vm.GasSlowStep,
			minStack:    minStack(2, 0),
			maxStack:    maxStack(2, 0),
			jumps:       true,
		},
		vm.	PC: {
			execute:     opPc,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	MSIZE: {
			execute:     opMsize,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	GAS: {
			execute:     opGas,
			constantGas: vm.GasQuickStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	JUMPDEST: {
			execute:     opJumpdest,
			constantGas: params.JumpdestGas,
			minStack:    minStack(0, 0),
			maxStack:    maxStack(0, 0),
		},
		vm.	PUSH1: {
			execute:     opPush1,
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH2: {
			execute:     makePush(2, 2),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH3: {
			execute:     makePush(3, 3),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH4: {
			execute:     makePush(4, 4),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH5: {
			execute:     makePush(5, 5),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	PUSH6: {
			execute:     makePush(6, 6),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	PUSH7: {
			execute:     makePush(7, 7),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	PUSH8: {
			execute:     makePush(8, 8),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH9: {
			execute:     makePush(9, 9),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	PUSH10: {
			execute:     makePush(10, 10),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH11: {
			execute:     makePush(11, 11),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH12: {
			execute:     makePush(12, 12),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	PUSH13: {
			execute:     makePush(13, 13),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	PUSH14: {
			execute:     makePush(14, 14),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH15: {
			execute:     makePush(15, 15),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH16: {
			execute:     makePush(16, 16),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	PUSH17: {
			execute:     makePush(17, 17),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	PUSH18: {
			execute:     makePush(18, 18),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	PUSH19: {
			execute:     makePush(19, 19),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH20: {
			execute:     makePush(20, 20),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH21: {
			execute:     makePush(21, 21),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH22: {
			execute:     makePush(22, 22),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	PUSH23: {
			execute:     makePush(23, 23),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH24: {
			execute:     makePush(24, 24),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	PUSH25: {
			execute:     makePush(25, 25),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	PUSH26: {
			execute:     makePush(26, 26),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH27: {
			execute:     makePush(27, 27),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH28: {
			execute:     makePush(28, 28),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH29: {
			execute:     makePush(29, 29),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH30: {
			execute:     makePush(30, 30),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.PUSH31: {
			execute:     makePush(31, 31),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	PUSH32: {
			execute:     makePush(32, 32),
			constantGas: vm.GasFastestStep,
			minStack:    minStack(0, 1),
			maxStack:    maxStack(0, 1),
		},
		vm.	DUP1: {
			execute:     makeDup(1),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(1),
			maxStack:    maxDupStack(1),
		},
		vm.	DUP2: {
			execute:     makeDup(2),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(2),
			maxStack:    maxDupStack(2),
		},
		vm.	DUP3: {
			execute:     makeDup(3),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(3),
			maxStack:    maxDupStack(3),
		},
		vm.	DUP4: {
			execute:     makeDup(4),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(4),
			maxStack:    maxDupStack(4),
		},
		vm.DUP5: {
			execute:     makeDup(5),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(5),
			maxStack:    maxDupStack(5),
		},
		vm.DUP6: {
			execute:     makeDup(6),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(6),
			maxStack:    maxDupStack(6),
		},
		vm.	DUP7: {
			execute:     makeDup(7),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(7),
			maxStack:    maxDupStack(7),
		},
		vm.DUP8: {
			execute:     makeDup(8),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(8),
			maxStack:    maxDupStack(8),
		},
		vm.DUP9: {
			execute:     makeDup(9),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(9),
			maxStack:    maxDupStack(9),
		},
		vm.	DUP10: {
			execute:     makeDup(10),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(10),
			maxStack:    maxDupStack(10),
		},
		vm.DUP11: {
			execute:     makeDup(11),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(11),
			maxStack:    maxDupStack(11),
		},
		vm.DUP12: {
			execute:     makeDup(12),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(12),
			maxStack:    maxDupStack(12),
		},
		vm.DUP13: {
			execute:     makeDup(13),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(13),
			maxStack:    maxDupStack(13),
		},
		vm.DUP14: {
			execute:     makeDup(14),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(14),
			maxStack:    maxDupStack(14),
		},
		vm.DUP15: {
			execute:     makeDup(15),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(15),
			maxStack:    maxDupStack(15),
		},
		vm.DUP16: {
			execute:     makeDup(16),
			constantGas: vm.GasFastestStep,
			minStack:    minDupStack(16),
			maxStack:    maxDupStack(16),
		},
		vm.SWAP1: {
			execute:     makeSwap(1),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(2),
			maxStack:    maxSwapStack(2),
		},
		vm.SWAP2: {
			execute:     makeSwap(2),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(3),
			maxStack:    maxSwapStack(3),
		},
		vm.SWAP3: {
			execute:     makeSwap(3),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(4),
			maxStack:    maxSwapStack(4),
		},
		vm.SWAP4: {
			execute:     makeSwap(4),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(5),
			maxStack:    maxSwapStack(5),
		},
		vm.SWAP5: {
			execute:     makeSwap(5),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(6),
			maxStack:    maxSwapStack(6),
		},
		vm.SWAP6: {
			execute:     makeSwap(6),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(7),
			maxStack:    maxSwapStack(7),
		},
		vm.SWAP7: {
			execute:     makeSwap(7),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(8),
			maxStack:    maxSwapStack(8),
		},
		vm.SWAP8: {
			execute:     makeSwap(8),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(9),
			maxStack:    maxSwapStack(9),
		},
		vm.SWAP9: {
			execute:     makeSwap(9),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(10),
			maxStack:    maxSwapStack(10),
		},
		vm.SWAP10: {
			execute:     makeSwap(10),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(11),
			maxStack:    maxSwapStack(11),
		},
		vm.SWAP11: {
			execute:     makeSwap(11),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(12),
			maxStack:    maxSwapStack(12),
		},
		vm.SWAP12: {
			execute:     makeSwap(12),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(13),
			maxStack:    maxSwapStack(13),
		},
		vm.SWAP13: {
			execute:     makeSwap(13),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(14),
			maxStack:    maxSwapStack(14),
		},
		vm.SWAP14: {
			execute:     makeSwap(14),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(15),
			maxStack:    maxSwapStack(15),
		},
		vm.SWAP15: {
			execute:     makeSwap(15),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(16),
			maxStack:    maxSwapStack(16),
		},
		vm.SWAP16: {
			execute:     makeSwap(16),
			constantGas: vm.GasFastestStep,
			minStack:    minSwapStack(17),
			maxStack:    maxSwapStack(17),
		},
		vm.LOG0: {
			execute:    makeLog(0),
			dynamicGas: makeGasLog(0),
			minStack:   minStack(2, 0),
			maxStack:   maxStack(2, 0),
			memorySize: memoryLog,
			writes:     true,
		},
		vm.LOG1: {
			execute:    makeLog(1),
			dynamicGas: makeGasLog(1),
			minStack:   minStack(3, 0),
			maxStack:   maxStack(3, 0),
			memorySize: memoryLog,
			writes:     true,
		},
		vm.LOG2: {
			execute:    makeLog(2),
			dynamicGas: makeGasLog(2),
			minStack:   minStack(4, 0),
			maxStack:   maxStack(4, 0),
			memorySize: memoryLog,
			writes:     true,
		},
		vm.LOG3: {
			execute:    makeLog(3),
			dynamicGas: makeGasLog(3),
			minStack:   minStack(5, 0),
			maxStack:   maxStack(5, 0),
			memorySize: memoryLog,
			writes:     true,
		},
		vm.LOG4: {
			execute:    makeLog(4),
			dynamicGas: makeGasLog(4),
			minStack:   minStack(6, 0),
			maxStack:   maxStack(6, 0),
			memorySize: memoryLog,
			writes:     true,
		},
		vm.CREATE: {
			execute:     opCreate,
			constantGas: params.CreateGas,
			dynamicGas:  gasCreate,
			minStack:    minStack(3, 1),
			maxStack:    maxStack(3, 1),
			memorySize:  memoryCreate,
			writes:      true,
			returns:     true,
		},
		vm.CALL: {
			execute:     opCall,
			constantGas: params.CallGasFrontier,
			dynamicGas:  gasCall,
			minStack:    minStack(7, 1),
			maxStack:    maxStack(7, 1),
			memorySize:  memoryCall,
			returns:     true,
		},
		vm.CALLCODE: {
			execute:     opCallCode,
			constantGas: params.CallGasFrontier,
			dynamicGas:  gasCallCode,
			minStack:    minStack(7, 1),
			maxStack:    maxStack(7, 1),
			memorySize:  memoryCall,
			returns:     true,
		},
		vm.RETURN: {
			execute:    opReturn,
			dynamicGas: gasReturn,
			minStack:   minStack(2, 0),
			maxStack:   maxStack(2, 0),
			memorySize: memoryReturn,
			halts:      true,
		},
		vm.SELFDESTRUCT: {
			execute:    opSuicide,
			dynamicGas: gasSelfdestruct,
			minStack:   minStack(1, 0),
			maxStack:   maxStack(1, 0),
			halts:      true,
			writes:     true,
		},
	}
}