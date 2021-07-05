package vm

import (
	"encoding/json"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mapprotocol/atlas/multiChain"
	"github.com/mapprotocol/atlas/multiChain/chainDB"
	"github.com/mapprotocol/atlas/multiChain/ethereum"
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
	abiSync, _  = abi.JSON(strings.NewReader(ABI_JSON))
	SyncAddress = common.BytesToAddress([]byte("atlas_sync"))
)

var (
	syncLimit = "sync limit exceeded"
)

// SyncGas defines all method gas
var SyncGas = map[string]uint64{
	"save": 0,
}

// RunSync execute atlas header store contract
func RunSync(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	method, err := abiSync.MethodById(input)
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
	// decode
	args := struct {
		From   string
		To     string
		Header []byte
	}{}

	err = abiSync.UnpackIntoInterface(args, Save, input)
	if err != nil {
		log.Error("save Unpack error", "err", err)
		return nil, ErrSyncInvalidInput
	}

	var hs []*ethereum.Header
	err = json.Unmarshal(args.Header, &hs)
	if err != nil {
		log.Error("args.Header json unmarshal failed.", "args.Header", args.Header, "err", err)
		return nil, ErrJSONUnmarshal
	}

	// validate header
	header := new(ethereum.Header)
	start := time.Now()
	if _, err := header.ValidateHeaderChain(hs); err != nil {
		log.Error("ValidateHeaderChain failed.", "err", err)
		return nil, err
	}

	headerStore := NewHeaderStore()
	err = headerStore.Load(evm.StateDB, SyncAddress)
	if err != nil {
		log.Error("header store load error", "error", err)
		return nil, err
	}

	var total uint64
	for _, h := range hs {
		// todo 查询 relayer 有效工作区间

		if headerStore.GetReceiveTimes(h.Number.Uint64()) >= TimesLimit {
			headerStore.StoreAbnormalMsg(h.Number.Uint64(), syncLimit)
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

	err = headerStore.Store(evm.StateDB, SyncAddress)
	if err != nil {
		log.Error("Staking save state error", "error", err)
		return nil, err
	}

	// reward

	// store
	store, err := chainDB.GetStoreMgr(multiChain.ChainTypeETH)
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
		height big.Int
	}{}

	err = abiSync.UnpackIntoInterface(args, GetAbnormalMsg, input)
	if err != nil {
		log.Error("save Unpack error", "err", err)
		return nil, ErrSyncInvalidInput
	}

	headerStore := NewHeaderStore()
	err = headerStore.Load(evm.StateDB, SyncAddress)
	if err != nil {
		log.Error("header store load error", "error", err)
		return nil, err
	}

	msg := headerStore.LoadAbnormalMsg(args.height.Uint64())
	if msg == "" {
		msg = "not found abnormal msg"
	}

	return abiSync.Methods[GetAbnormalMsg].Outputs.Pack(msg)
}
