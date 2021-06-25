package sync

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
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
		"name": "save",
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
	// decode
	args := struct {
		From   string
		To     string
		Header []byte
	}{}

	err = abiSync.UnpackIntoInterface(args, SAVE, input)
	if err != nil {
		log.Error("save Unpack error", "err", err)
		return nil, ErrSyncInvalidInput
	}

	var hs []*ETHHeader
	err = json.Unmarshal(args.Header, &hs)
	if err != nil {
		// todo
		log.Error(fmt.Sprintf("args.header json unmarshal failed, args.header: %+v, error: %v", err, args.Header))
		return nil, ErrJSONUnmarshal
	}

	// validate header
	header := new(ETHHeader)
	if _, err := header.ValidateHeaderChain(hs); err != nil {
		return nil, err
	}

	// reward

	// store

	return nil, nil
}
