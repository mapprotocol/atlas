package main

import (
	"context"
	"fmt"
	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"
	"math/big"
)

var queryRegisteredValidatorSignersCommand = cli.Command{
	Name:   "getRegisteredValidatorSigners",
	Usage:  "Registered Validator Signers",
	Action: MigrateFlags(getRegisteredValidatorSigners),
	Flags:  ValidatorFlags,
}
var queryTopValidatorsCommand = cli.Command{
	Name:   "getTopValidators",
	Usage:  "get Top Group Validators",
	Action: MigrateFlags(getTopValidators),
	Flags:  ValidatorFlags,
}

func getRegisteredValidatorSigners(ctx *cli.Context) error {
	//---------------------------- pre set ----------------------
	path := ""
	password = ""
	//-----------------------------------------------------------
	log.Info("==== getRegisteredValidatorSigners ===")
	methodName := "getRegisteredValidatorSigners"
	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	header, err := conn.HeaderByNumber(context.Background(), nil)
	input := packInput(abiValidators, methodName)
	msg := ethchain.CallMsg{From: from, To: &ValidatorAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "error", err)
	}
	ValidatorSigners := new([]common.Address)
	err = abiValidators.UnpackIntoInterface(&ValidatorSigners, methodName, output)
	if err != nil {
		log.Error("method UnpackIntoInterface error", err)
		return nil
	}
	if len(*ValidatorSigners) == 0 {
		log.Info("ValidatorSigners:", "obj", "[]")
		return nil
	}
	for i, v := range *ValidatorSigners {
		fmt.Println("getRegisteredValidatorSigners:", v.String(), "  index:", i)
	}
	return nil
}
func getTopValidators(ctx *cli.Context) error {
	//--------------------------- pre set -------------------------------------------
	path := ""
	n := big.NewInt(5) // top number
	password = ""
	//-------------------------------------------------------------------------------
	if ctx.IsSet(TopNumFlag.Name) {
		n = big.NewInt(ctx.GlobalInt64(TopNumFlag.Name))
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	methodName := "getTopValidators"
	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	header, err := conn.HeaderByNumber(context.Background(), nil)

	log.Info("=== getTopValidators admin", "obj", from)
	input := packInput(abiElection, methodName, n)
	msg := ethchain.CallMsg{From: from, To: &ElectionAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "error", err)
	}
	TopValidators := new([]common.Address)
	err = abiElection.UnpackIntoInterface(&TopValidators, methodName, output)
	if err != nil {
		log.Error("method UnpackIntoInterface", "error", err)
		return nil
	}
	if len(*TopValidators) == 0 {
		log.Info("TopValidators:", "obj", "[]")
		return nil
	}
	for _, v := range *TopValidators {
		log.Info("Address:", "obj", v.String())
	}
	return nil
}
func getNumberValidators(ctx *cli.Context) error {
	//--------------------------- pre set -------------------------------------------
	path := ""
	n := big.NewInt(5) // top number
	password = ""
	//-------------------------------------------------------------------------------
	if ctx.IsSet(TopNumFlag.Name) {
		n = big.NewInt(ctx.GlobalInt64(TopNumFlag.Name))
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	methodName := "getTopValidators"
	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	header, err := conn.HeaderByNumber(context.Background(), nil)

	log.Info("=== getTopValidators admin", "obj", from)
	input := packInput(abiElection, methodName, n)
	msg := ethchain.CallMsg{From: from, To: &ElectionAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "error", err)
	}
	TopValidators := new([]common.Address)
	err = abiElection.UnpackIntoInterface(&TopValidators, methodName, output)
	if err != nil {
		log.Error("method UnpackIntoInterface", "error", err)
		return nil
	}
	if len(*TopValidators) == 0 {
		log.Info("TopValidators:", "obj", "[]")
		return nil
	}
	for _, v := range *TopValidators {
		log.Info("Address:", "obj", v.String())
	}
	return nil
}
