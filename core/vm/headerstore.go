package vm

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/chainsdb"
	"github.com/mapprotocol/atlas/chains/headers/ethereum"
	ve "github.com/mapprotocol/atlas/chains/validates/ethereum"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/params"
)

const (
	Save          = "save"
	CurNbrAndHash = "currentNumberAndHash"
)

const TimesLimit = 3

// HeaderStore contract ABI
var (
	abiHeaderStore, _ = abi.JSON(strings.NewReader(params.HeaderStoreABIJSON))
)

// SyncGas defines all method gas
var SyncGas = map[string]uint64{
	Save:          0,
	CurNbrAndHash: 0,
}

// RunHeaderStore execute atlas header store contract
func RunHeaderStore(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	method, err := abiHeaderStore.MethodById(input)
	if err != nil {
		log.Error(fmt.Sprintf("header store contract method(%s) not found", method))
		return nil, ErrExecutionReverted
	}

	data := input[4:]
	switch method.Name {
	case Save:
		ret, err = save(evm, contract, data)
	case CurNbrAndHash:
		ret, err = currentNumberAndHash(evm, contract, data)
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
	// check if the relayer is registered in the current epoch
	if !IsInCurrentEpoch(evm.StateDB, contract.CallerAddress) {
		return nil, errors.New("invalid work epoch, please register first")
	}

	// decode
	args := struct {
		From    uint64
		To      uint64
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

	// check if it is a supported chain
	if !(chains.IsSupportedChain(rawdb.ChainType(args.From)) || chains.IsSupportedChain(rawdb.ChainType(args.To))) {
		return nil, ErrNotSupportChain
	}

	var hs []*ethereum.Header
	err = json.Unmarshal(args.Headers, &hs)
	if err != nil {
		log.Error("args.Header json unmarshal failed.", "err", err, "args.Header", string(args.Headers))
		return nil, ErrJSONUnmarshal
	}

	// validate header
	header := new(ve.Validate)
	start := time.Now()
	if _, err := header.ValidateHeaderChain(hs); err != nil {
		log.Error("ValidateHeaderChain failed.", "err", err)
		return nil, err
	}

	// calc synchronization information
	headerStore := NewHeaderStore()
	err = headerStore.Load(evm.StateDB, params.HeaderStoreAddress)
	if err != nil {
		log.Error("header store load error", "error", err)
		return nil, err
	}

	var total uint64
	for _, h := range hs {
		if headerStore.GetReceiveTimes(h.Number.Uint64()) >= TimesLimit {
			return nil, fmt.Errorf("the number of synchronizations has reached the limit(%d)", TimesLimit)
		}
		total++
		headerStore.IncrReceiveTimes(h.Number.Uint64())
	}
	epochID, err := GetCurrentEpochID(evm)
	if err != nil {
		return nil, err
	}
	headerStore.AddSyncTimes(epochID, total, contract.CallerAddress)

	// store block header
	store, err := chainsdb.GetStoreMgr(rawdb.ChainType(args.From))
	if err != nil {
		return nil, err
	}
	if _, err := store.InsertHeaderChain(hs, start); err != nil {
		log.Error("InsertHeaderChain failed.", "err", err)
		return nil, err
	}

	// store synchronization information
	err = headerStore.Store(evm.StateDB, params.HeaderStoreAddress)
	if err != nil {
		log.Error("sync save state error", "error", err)
		return nil, err
	}
	return nil, nil
}

func currentNumberAndHash(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	args := struct {
		ChainID uint64
	}{}
	method, _ := abiHeaderStore.Methods[CurNbrAndHash]
	unpack, err := method.Inputs.Unpack(input)
	if err != nil {
		return nil, err
	}
	if err := method.Inputs.Copy(&args, unpack); err != nil {
		return nil, err
	}

	v := new(ve.Validate)
	c := rawdb.ChainType(args.ChainID)
	number, err := v.GetCurrentHeaderNumber(c)
	if err != nil {
		return nil, err
	}
	hash, err := v.GetHashByNumber(c, number)
	if err != nil {
		return nil, err
	}
	return method.Outputs.Pack(new(big.Int).SetUint64(number), hash)
}
