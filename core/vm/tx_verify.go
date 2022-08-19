package vm

import (
	"bytes"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/interfaces"
	"github.com/mapprotocol/atlas/params"
)

const (
	VerifyProof = "verifyProofData"
)

// TxVerify contract ABI
var (
	abiTxVerify, _ = abi.JSON(strings.NewReader(params.TxVerifyABIJSON))
)

var TxVerifyGas = map[string]uint64{
	VerifyProof: 42000,
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
	case VerifyProof:
		ret, err = verifyProofData(evm, contract, data)
	default:
		log.Warn("run tx verify contract failed, invalid method", "method", method.Name)
		return ret, errors.New("invalid method name")
	}

	if err != nil {
		log.Error("run tx verify contract failed", "method", method.Name, "error", err)
	} else {
		log.Info("run tx verify contract succeed", "method", method.Name)
	}
	return ret, err
}

func verifyProofData(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	var (
		success      = true
		message      = ""
		logs         []byte
		receiptProof []byte
	)
	args := struct {
		Router   common.Address
		Coin     common.Address
		SrcChain *big.Int
		DstChain *big.Int
		TxProve  []byte
	}{}

	verifyProof := abiTxVerify.Methods[VerifyProof]
	defer func() {
		var packErr error

		if err != nil {
			success, message, logs = false, err.Error(), []byte{}
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
		ret, packErr = verifyProof.Outputs.Pack(success, message, logs)
		if packErr != nil {
			log.Error("verify proof outputs pack failed", "error", packErr.Error())
		}
	}()

	unpack, err := verifyProof.Inputs.Unpack(input)
	if err != nil {
		return nil, err
	}
	if err = verifyProof.Inputs.Copy(&receiptProof, unpack); err != nil {
		return nil, err
	}
	if err := rlp.DecodeBytes(receiptProof, &args); err != nil {
		log.Error("rlp decode receiptProof failed", "err", err)
		return nil, err
	}
	log.Info("verifyProofData args", "router", args.Router, "coin", args.Coin, "srcChain", args.SrcChain, "dstChain", args.DstChain)

	// params check
	if bytes.Equal(args.Router.Bytes(), common.Address{}.Bytes()) {
		return nil, errors.New("router address is empty")
	}
	//if bytes.Equal(args.Coin.Bytes(), common.Address{}.Bytes()) {
	//	return nil, errors.New("coin address is empty")
	//}
	if !chains.IsSupportedChain(chains.ChainType(args.SrcChain.Uint64())) {
		return nil, ErrNotSupportChain
	}
	group, err := chains.ChainType2ChainGroup(chains.ChainType(args.SrcChain.Uint64()))
	if err != nil {
		return nil, err
	}

	v, err := interfaces.VerifyFactory(group)
	if err != nil {
		return nil, err
	}
	logs, err = v.Verify(evm.StateDB, args.Router, args.TxProve)
	if err != nil {
		log.Error("verify proof failed", "err", err.Error())
		return nil, err
	}
	return nil, nil
}
