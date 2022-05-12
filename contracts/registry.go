package contracts

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/contracts/abis"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
)

var getAddressMethod = NewBoundMethod(params.RegistrySmartContractAddress, abis.Registry, "getAddressFor", params.MaxGasForGetAddressFor)

// TODO(kevjue) - Re-Enable caching of the retrieved registered address

// GetRegisteredAddress returns the address on the registry for a given id
func GetRegisteredAddress(vmRunner vm.EVMRunner, registryId common.Hash) (common.Address, error) {

	//vmRunner.StopGasMetering()
	//defer vmRunner.StartGasMetering()

	var contractAddress common.Address
	err := getAddressMethod.Query(vmRunner, &contractAddress, registryId)

	// TODO (mcortesi) Remove ErrEmptyArguments check after we change Proxy to fail on unset impl
	// TODO(asa): Why was this change necessary?
	if err == abi.ErrEmptyArguments || err == vm.ErrExecutionReverted {
		return common.BytesToAddress([]byte{}), ErrRegistryContractNotDeployed
	} else if err != nil {
		return common.BytesToAddress([]byte{}), err
	}

	if contractAddress == common.BytesToAddress([]byte{}) {
		return common.BytesToAddress([]byte{}), ErrSmartContractNotDeployed
	}

	return contractAddress, nil
}
