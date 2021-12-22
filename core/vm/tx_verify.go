package vm

import (
	"bytes"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/params"
)

const (
	TxVerify = "txVerify"
)

// TxVerify contract ABI
var (
	abiTxVerify, _ = abi.JSON(strings.NewReader(params.TxVerifyABIJSON))
)

var TxVerifyGas = map[string]uint64{
	TxVerify: 42000,
}

// RunTxVerify execute atlas tx verify contract
func RunTxVerify(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	method, err := abiTxVerify.MethodById(input)
	if err != nil {
		log.Error("get tx verify ABI method failed", "error", err)
		return nil, err
	}

	data := input[4:]
	switch method.Name {
	// input[:4] 0x7df27c0b
	case TxVerify:
		ret, err = txVerify(evm, contract, data)
	default:
		log.Warn("run tx verify contract failed, invalid method name", "method.name", method.Name)
		return ret, errors.New("invalid method name")
	}

	if err != nil {
		log.Error("run tx verify contract failed", "method.name", method.Name, "error", err)
	} else {
		log.Info("run tx verify contract succeed", "method.name", method.Name)
	}
	return ret, err
}

func txVerify(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	args := struct {
		Router   common.Address
		Coin     common.Address
		SrcChain *big.Int
		DstChain *big.Int
		TxProve  []byte
	}{}

	method := abiTxVerify.Methods[TxVerify]
	defer func() {
		var (
			packErr error
			message string
			success = true
		)

		if err != nil {
			success, message = false, err.Error()
		}
		// In general, the Pack operation will not fail. Here we can choose to ignore Pack operation error.
		// This is not absolute, so use a new value to receive the Pack operation error,
		// and record the error in the log when the error is not nil.
		/*
			1. ret, packErr = method.Outputs.Pack(success, message)
				packErr == nil
					ret == {true, ""}, err == nil
					ret == {false, "... error"}, err == errors.New("... error")
				packErr != nil
					ret == nil, err == nil  // unexpected
					ret == nil, err == errors.New("... error")

			2. ret, err` = method.Outputs.Pack(success, message)
				err` == nil
					ret == {true, ""}, err == nil
					ret == {false, "... error"}, err == nil  // unexpected
				err` != nil
					ret == nil, err == errors.New("pack error ...")
		*/
		ret, packErr = method.Outputs.Pack(success, message)
		if packErr != nil {
			log.Error("txVerify outputs pack failed", "error", packErr.Error())
		}
	}()

	unpack, err := method.Inputs.Unpack(input)
	if err != nil {
		return nil, err
	}
	if err := method.Inputs.Copy(&args, unpack); err != nil {
		return nil, err
	}
	log.Info("txVerify input params", "router", args.Router, "coin", args.Coin, "srcChain", args.SrcChain, "dstChain", args.DstChain)

	// params check
	if bytes.Equal(args.Router.Bytes(), common.Address{}.Bytes()) {
		return nil, errors.New("router address is empty")
	}
	if bytes.Equal(args.Coin.Bytes(), common.Address{}.Bytes()) {
		return nil, errors.New("router address is empty")
	}
	if !(chains.IsSupportedChain(rawdb.ChainType(args.SrcChain.Uint64())) || chains.IsSupportedChain(rawdb.ChainType(args.DstChain.Uint64()))) {
		return nil, ErrNotSupportChain
	}
	group, err := chains.ChainType2ChainGroup(rawdb.ChainType(args.SrcChain.Uint64()))
	if err != nil {
		return nil, err
	}

	v, err := chains.VerifyFactory(group)
	if err != nil {
		return nil, err
	}
	return nil, v.Verify(args.Router, args.SrcChain, args.DstChain, args.TxProve)
}
