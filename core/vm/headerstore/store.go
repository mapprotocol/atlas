package headerstore

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/core"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/core/chain/eth"
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

// RunSync execute atlas header store contract
func RunSync(evm *vm.EVM, contract *vm.Contract, input []byte) (ret []byte, err error) {
	method, err := abiSync.MethodById(input)
	if err != nil {
		log.Error("No method found")
		return nil, eth.ErrExecutionReverted
	}

	data := input[4:]
	switch method.Name {
	case SAVE:
		ret, err = save(evm, contract, data)
	default:
		log.Warn("header call fallback function")
		err = eth.ErrSyncInvalidInput
	}

	if err != nil {
		log.Warn("Sync error code", "code", err)
		err = eth.ErrExecutionReverted
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
		return nil, eth.ErrSyncInvalidInput
	}

	var hs []*eth.Header
	err = json.Unmarshal(args.Header, &hs)
	if err != nil {
		log.Error("args.Header json unmarshal failed.", "args.Header", args.Header, "err", err)
		return nil, eth.ErrJSONUnmarshal
	}

	// validate header
	header := new(eth.Header)
	start := time.Now()
	if _, err := header.ValidateHeaderChain(hs); err != nil {
		log.Error("ValidateHeaderChain failed.", "err", err)
		return nil, err
	}

	// reward

	// store
	if _, err = core.GetStoreMgr(chain.ChainTypeETH).InsertHeaderChain(hs, start); err != nil {
		log.Error("InsertHeaderChain failed.", "err", err)
	}
	return nil, nil
}
