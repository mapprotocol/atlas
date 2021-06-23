package sync

import (
	"strings"

	"github.com/abeychain/go-abey/accounts/abi"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
)

const ABI_JSON = `[
	{
		"inputs": [
			{
				"internalType": "string",
				"name": "from",
				"type": "string"
			},
			{
				"internalType": "string",
				"name": "to",
				"type": "string"
			},
			{
				"internalType": "bytes",
				"name": "header",
				"type": "bytes"
			}
		],
		"name": "sam",
		"outputs": [
			{
				"internalType": "bool",
				"name": "success",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

const (
	SAVE = "save"
)

// Sync contract ABI
var abiSync, _ = abi.JSON(strings.NewReader(ABI_JSON))

// Gas defines all method gas
var Gas = map[string]uint64{
	"save": 0,
}

// RunSync execute atlas sync contract
func RunSync(evm *vm.EVM, contract *vm.Contract, input []byte) (ret []byte, err error) {
	method, err := abiSync.MethodById(input)
	if err != nil {
		log.Error("No method found")
		return nil, ErrExecutionReverted
	}

	data := input[4:]
	switch method.Name {
	case SAVE:
		ret, err = save(evm, contract, data)
	default:
		log.Warn("sync call fallback function")
		err = ErrSyncInvalidInput
	}

	if err != nil {
		log.Warn("Sync error code", "code", err)
		err = ErrExecutionReverted
	}

	return ret, err
}

func save(evm *vm.EVM, contract *vm.Contract, input []byte) (ret []byte, err error) {
	// RLP decode
	args := struct {
		from   string
		to     string
		header []*ETHHeader
	}{}

	err = abiSync.Methods[SAVE].Inputs.Unpack(args, input)
	if err != nil {
		log.Error("save Unpack error", "err", err)
		return nil, ErrSyncInvalidInput
	}

	// validate header
	header := new(ETHHeader)
	if _, err := header.ValidateHeaderChain(args.header); err != nil {
		return nil, err
	}

	// save
	return nil, nil
}
