package vm

import (
	"bytes"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/interfaces"
	"github.com/mapprotocol/atlas/params"
)

const (
	Save          = "save"
	Reset         = "reset"
	CurNbrAndHash = "currentNumberAndHash"
	SetRelayer    = "setRelayer"
	GetRelayer    = "getRelayer"
)

// HeaderStore contract ABI
var (
	abiHeaderStore, _ = abi.JSON(strings.NewReader(params.HeaderStoreABIJSON))
)

// SyncGas defines all method gas
var SyncGas = map[string]uint64{
	CurNbrAndHash: 42000,
	SetRelayer:    2100,
	GetRelayer:    0,
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
	case Reset:
		ret, err = reset(evm, contract, data)
	case CurNbrAndHash:
		ret, err = currentNumberAndHash(evm, contract, data)
	case SetRelayer:
		ret, err = setRelayer(evm, contract, data)
	case GetRelayer:
		ret, err = getRelayer(evm)
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

	if err := validateRelayer(evm, contract.CallerAddress); err != nil {
		return nil, err
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

	group, err := chains.ChainType2ChainGroup(fromChain)
	if err != nil {
		return nil, err
	}

	chain, err := interfaces.ChainFactory(group)
	if err != nil {
		return nil, err
	}
	if _, err := chain.ValidateHeaderChain(evm.StateDB, args.Headers, fromChain); err != nil {
		log.Error("failed to validate header chain", "error", err)
		return nil, err
	}

	_, err = chain.InsertHeaders(evm.StateDB, args.Headers)
	if err != nil {
		log.Error("failed to write headers", "error", err)
		return nil, err
	}
	return nil, nil
}

func reset(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	args := struct {
		From   *big.Int
		Td     *big.Int
		Header []byte
	}{}

	adminHash := evm.StateDB.GetState(params.RegistryProxyAddress, params.ProxyOwnerStorageLocation)
	if !bytes.Equal(contract.CallerAddress.Bytes(), adminHash[12:]) {
		return nil, errors.New("forbidden")
	}

	method, _ := abiHeaderStore.Methods[Reset]
	unpack, err := method.Inputs.Unpack(input)
	if err != nil {
		return nil, err
	}
	if err := method.Inputs.Copy(&args, unpack); err != nil {
		return nil, err
	}

	from := chains.ChainType(args.From.Uint64())
	chainID, err := chains.ChainType2ChainID(from)
	if err != nil {
		return nil, err
	}
	if evm.chainConfig.ChainID.Cmp(new(big.Int).SetUint64(chainID)) != 0 {
		return nil, errors.New("current chainID does not match the from parameter")
	}

	group, err := chains.ChainType2ChainGroup(from)
	if err != nil {
		return nil, err
	}
	hs, err := interfaces.HeaderStoreFactory(group)
	if err != nil {
		return nil, err
	}
	if err := hs.ResetHeaderStore(evm.StateDB, args.Header, args.Td); err != nil {
		log.Error("failed to reset header store", "error", err)
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

func setRelayer(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	adminHash := evm.StateDB.GetState(params.RegistryProxyAddress, params.ProxyOwnerStorageLocation)
	if !bytes.Equal(contract.CallerAddress.Bytes(), adminHash[12:]) {
		return nil, errors.New("forbidden")
	}

	args := struct {
		Relayer common.Address
	}{}
	method := abiHeaderStore.Methods[SetRelayer]
	unpack, err := method.Inputs.Unpack(input)
	if err != nil {
		return nil, err
	}
	if err := method.Inputs.Copy(&args, unpack); err != nil {
		return nil, err
	}
	evm.StateDB.SetPOWState(params.NewRelayerAddress, common.BytesToHash(params.NewRelayerAddress[:]), args.Relayer.Bytes())
	return nil, nil
}

func getRelayer(evm *EVM) (ret []byte, err error) {
	method := abiHeaderStore.Methods[GetRelayer]
	relayerBytes := evm.StateDB.GetPOWState(params.NewRelayerAddress, common.BytesToHash(params.NewRelayerAddress[:]))
	return method.Outputs.Pack(common.BytesToAddress(relayerBytes))
}

func validateRelayer(evm *EVM, caller common.Address) error {
	adminAddrBytes := evm.StateDB.GetPOWState(params.NewRelayerAddress, common.BytesToHash(params.NewRelayerAddress[:]))
	if !bytes.Equal(caller.Bytes(), adminAddrBytes) {
		return errors.New("invalid relayer")
	}
	return nil
}
