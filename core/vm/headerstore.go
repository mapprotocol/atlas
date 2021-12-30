package vm

import (
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/interfaces"
	"github.com/mapprotocol/atlas/params"
)

const (
	Save          = "save"
	CurNbrAndHash = "currentNumberAndHash"
)

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
		log.Warn("run header store contract failed, invalid method name", "method.name", method.Name)
		return ret, errors.New("invalid method name")
	}

	if err != nil {
		log.Error("run header store contract failed", "method.name", method.Name, "error", err)
	} else {
		log.Info("run header store contract succeed", "method.name", method.Name)
	}

	return ret, err
}

func save(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	args := struct {
		From    *big.Int
		To      *big.Int
		Headers []byte
	}{}

	// check if the relayer is registered in the current epoch
	if !IsInCurrentEpoch(evm.StateDB, contract.CallerAddress) {
		return nil, errors.New("invalid work epoch, please register first")
	}

	method, _ := abiHeaderStore.Methods[Save]
	unpack, err := method.Inputs.Unpack(input)
	if err != nil {
		return nil, err
	}
	if err := method.Inputs.Copy(&args, unpack); err != nil {
		return nil, err
	}

	if len(args.Headers) == 0 {
		return nil, errors.New("headers cannot be empty")
	}
	// check if it is a supported chain
	fromChain := chains.ChainType(args.From.Uint64())
	toChain := chains.ChainType(args.To.Uint64())
	if !(chains.IsSupportedChain(fromChain) || chains.IsSupportedChain(toChain)) {
		return nil, ErrNotSupportChain
	}

	group, err := chains.ChainType2ChainGroup(chains.ChainType(args.From.Uint64()))
	if err != nil {
		return nil, err
	}

	chain, err := interfaces.ChainFactory(group)
	if err != nil {
		return nil, err
	}
	if _, err := chain.ValidateHeaderChain(evm.StateDB, args.Headers); err != nil {
		log.Error("failed to validate header chain", "error", err)
		return nil, err
	}

	inserted, err := chain.WriteHeaders(evm.StateDB, args.Headers)
	if err != nil {
		log.Error("failed to write headers", "error", err)
		return nil, err
	}

	epochID, err := GetCurrentEpochID(evm)
	if err != nil {
		return nil, err
	}
	if err := chain.StoreSyncTimes(evm.StateDB, epochID, contract.CallerAddress, inserted); err != nil {
		log.Error("failed to save sync times", "error", err)
		return nil, err
	}
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

	group, err := chains.ChainType2ChainGroup(chains.ChainType(args.ChainID.Uint64()))
	if err != nil {
		return nil, err
	}
	hs, err := interfaces.HeaderStoreFactory(group)
	if err != nil {
		return nil, err
	}
	number, hash, err := hs.GetCurrentNumberAndHash(evm.StateDB)
	if err != nil {
		return nil, err
	}
	return method.Outputs.Pack(new(big.Int).SetUint64(number), hash.Bytes())
}
