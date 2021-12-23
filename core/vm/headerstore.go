package vm

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/ethereum"
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
	//Save:          0,
	CurNbrAndHash: 42000,
}

// RunHeaderStore execute atlas header store contract
func RunHeaderStore(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	method, err := abiHeaderStore.MethodById(input)
	if err != nil {
		log.Error("get header store ABI method failed", "error", err)
		return nil, err
	}

	data := input[4:]
	switch method.Name {
	case Save:
		ret, err = save(evm, contract, data)
	case CurNbrAndHash:
		ret, err = currentNumberAndHash(evm, contract, data)
	default:
		log.Warn("run contract failed, invalid method name", "method.name", method.Name)
		return ret, errors.New("invalid method name")
	}

	if err != nil {
		log.Error("run contract failed", "method.name", method.Name, "error", err)
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
		From    *big.Int
		To      *big.Int
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
	fromChain := rawdb.ChainType(args.From.Uint64())
	toChain := rawdb.ChainType(args.To.Uint64())
	if !(chains.IsSupportedChain(fromChain) || chains.IsSupportedChain(toChain)) {
		return nil, ErrNotSupportChain
	}

	//group, err := chains.ChainType2ChainGroup(rawdb.ChainType(args.From.Uint64()))
	//if err != nil {
	//	return nil, err
	//}

	var hs []*ethereum.Header
	if err := rlp.DecodeBytes(args.Headers, &hs); err != nil {
		log.Error("rlp decode failed.", "err", err)
		return nil, ErrRLPDecode
	}

	// validate header
	header := new(ethereum.Validate)
	//start := time.Now()
	if _, err := header.ValidateHeaderChain(evm.StateDB, fromChain, hs); err != nil {
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

	// todo
	// store block header
	//store, err := chainsdb.GetStoreMgr(fromChain)
	//if err != nil {
	//	return nil, err
	//}
	//if _, err := store.InsertHeaderChain(hs, start); err != nil {
	//	log.Error("InsertHeaderChain failed.", "err", err)
	//	return nil, err
	//}
	//
	//_, err = chains.HeaderStoreFactory(group)
	//if err != nil {
	//	return nil, err
	//}

	// store synchronization information
	err = headerStore.Store(evm.StateDB, params.HeaderStoreAddress)
	if err != nil {
		log.Error("store state error", "error", err)
		return nil, err
	}
	log.Info("save contract execution complete")
	return nil, nil
}

func currentNumberAndHash(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	args := struct {
		ChainID *big.Int
	}{}
	method, _ := abiHeaderStore.Methods[CurNbrAndHash]
	unpack, err := method.Inputs.Unpack(input)
	if err != nil {
		return nil, err
	}
	if err := method.Inputs.Copy(&args, unpack); err != nil {
		return nil, err
	}

	//v := new(ethereum.Validate)
	//c := rawdb.ChainType(args.ChainID.Uint64())
	//number, err := v.GetCurrentHeaderNumber(evm.StateDB, c)
	//if err != nil {
	//	return nil, err
	//}
	//hash, err := v.GetHashByNumber(evm.StateDB, number)
	//if err != nil {
	//	return nil, err
	//}
	//return method.Outputs.Pack(new(big.Int).SetUint64(number), hash.Bytes())
	return []byte{}, nil
}
