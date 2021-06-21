package sync

import "github.com/ethereum/go-ethereum/core/vm"

type sync struct{}

func (s *sync) RequiredGas(evm *vm.EVM, input []byte) uint64 {
	var (
		baseGas uint64 = 0 // todo sync
	)

	method, err := abiSync.MethodById(input)
	if err != nil {
		return baseGas
	}

	if gas, ok := Gas[method.Name]; ok {
		return gas
	}
	return baseGas
}

func (s *sync) Run(evm *vm.EVM, contract *vm.Contract, input []byte) (ret []byte, err error) {
	return RunSync(evm, contract, input)
}
