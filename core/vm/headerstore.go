package vm

import (
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/chainsdb"
	"github.com/mapprotocol/atlas/chains/headers/ethereum"
	ve "github.com/mapprotocol/atlas/chains/validates/ethereum"
)

const ABI_JSON = `[
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "height",
				"type": "uint256"
			}
		],
		"name": "getAbnormalMsg",
		"outputs": [
			{
				"internalType": "bytes",
				"name": "abnormalMsg",
				"type": "bytes"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
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
				"name": "headers",
				"type": "bytes"
			}
		],
		"name": "save",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

const (
	Save           = "save"
	GetAbnormalMsg = "getAbnormalMsg"
)

const TimesLimit = 3

// Sync contract ABI
var (
	abiHeaderStore, _  = abi.JSON(strings.NewReader(ABI_JSON))
	HeaderStoreAddress = common.BytesToAddress([]byte("headerstore"))
)

var (
	syncLimit = "sync times limit exceeded"
)

// SyncGas defines all method gas
var SyncGas = map[string]uint64{
	Save:           0,
	GetAbnormalMsg: 0,
}

// RunHeaderStore execute atlas header store contract
func RunHeaderStore(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	method, err := abiHeaderStore.MethodById(input)
	if err != nil {
		log.Error("No method found")
		return nil, ErrExecutionReverted
	}

	data := input[4:]
	switch method.Name {
	case Save:
		ret, err = save(evm, contract, data)
	default:
		log.Warn("sync contract failed, invalid method name", "methodName", method.Name)
		err = ErrSyncInvalidInput
	}

	if err != nil {
		log.Warn("sync contract failed", "error", err)
		err = ErrExecutionReverted
	}

	return ret, err
}

func save(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	// check if the current epoch is registered
	if !IsInCurrentEpoch(evm.StateDB, contract.CallerAddress) {
		return nil, errors.New("invalid work epoch, please register first")
	}

	// decode
	args := struct {
		From    string
		To      string
		Headers []byte
	}{}

	method, _ := abiHeaderStore.Methods[Save]
	unpack, err := method.Inputs.Unpack(input)
	if err != nil {
		return nil, err
	}
	if err := method.Inputs.Copy(&args, unpack); err != nil {
		return nil, err
	}

	var hs []*ethereum.Header
	err = json.Unmarshal(args.Headers, &hs)
	if err != nil {
		log.Error("args.Header json unmarshal failed.", "args.Header", args.Headers, "err", err)
		return nil, ErrJSONUnmarshal
	}

	// validate header
	header := new(ve.Validate)
	start := time.Now()
	if _, err := header.ValidateHeaderChain(hs); err != nil {
		log.Error("ValidateHeaderChain failed.", "err", err)
		return nil, err
	}

	// store synchronization information
	headerStore := NewHeaderStore()
	err = headerStore.Load(evm.StateDB, HeaderStoreAddress)
	if err != nil {
		log.Error("header store load error", "error", err)
		return nil, err
	}

	var total uint64
	for _, h := range hs {
		if headerStore.GetReceiveTimes(h.Number.Uint64()) >= TimesLimit {
			headerStore.StoreAbnormalMsg(contract.CallerAddress, h.Number, syncLimit)
			continue
		}
		total++
		headerStore.IncrReceiveTimes(h.Number.Uint64())
	}
	epochID, err := GetCurrentEpochID(evm)
	if err != nil {
		return nil, err
	}
	headerStore.AddSyncTimes(epochID, total, contract.CallerAddress)

	err = headerStore.Store(evm.StateDB, HeaderStoreAddress)
	if err != nil {
		log.Error("sync save state error", "error", err)
		return nil, err
	}

	// store block header
	store, err := chainsdb.GetStoreMgr(chains.ChainTypeETH)
	if err != nil {
		return nil, err
	}
	if _, err := store.InsertHeaderChain(hs, start); err != nil {
		log.Error("InsertHeaderChain failed.", "err", err)
		return nil, err
	}
	return nil, nil
}

func getAbnormalMsg(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	args := struct {
		height *big.Int
	}{}

	err = abiHeaderStore.UnpackIntoInterface(args, GetAbnormalMsg, input)
	if err != nil {
		log.Error("save Unpack error", "err", err)
		return nil, ErrSyncInvalidInput
	}

	headerStore := NewHeaderStore()
	err = headerStore.Load(evm.StateDB, HeaderStoreAddress)
	if err != nil {
		log.Error("header store load error", "error", err)
		return nil, err
	}

	msg := headerStore.LoadAbnormalMsg(contract.CallerAddress, args.height)
	if msg == "" {
		msg = "not found abnormal msg"
	}

	return abiHeaderStore.Methods[GetAbnormalMsg].Outputs.Pack(msg)
}
