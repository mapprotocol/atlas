package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"math/big"

	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/urfave/cli.v1"
)

var queryGroupsCommand = cli.Command{
	Name:   "queryGroups",
	Usage:  "query Groups",
	Action: MigrateFlags(queryGroups),
	Flags:  ValidatorFlags,
}
var queryRegisteredValidatorSignersCommand = cli.Command{
	Name:   "getRegisteredValidatorSigners",
	Usage:  "Registered Validator Signers",
	Action: MigrateFlags(getRegisteredValidatorSigners),
	Flags:  ValidatorFlags,
}
var queryTopGroupValidatorsCommand = cli.Command{
	Name:   "getTopGroupValidators",
	Usage:  "get Top Group Validators",
	Action: MigrateFlags(getTopGroupValidators),
	Flags:  ValidatorFlags,
}

func queryGroups(ctx *cli.Context) error {
	//---------------- pre set ---------------------------------------
	path := ""
	password = ""
	//----------------------------------------------------------------
	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	header, err := conn.HeaderByNumber(context.Background(), nil)
	input := packInput(abiValidators, "getRegisteredValidatorGroups")
	msg := ethchain.CallMsg{From: from, To: &ValidatorAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "error", err)
	}
	groups := new([]common.Address)
	err = abiValidators.UnpackIntoInterface(&groups, "getRegisteredValidatorGroups", output)
	if err != nil {
		log.Error("method UnpackIntoInterface error", "error", err)
	}
	if len(*groups) == 0 {
		log.Info("groups:", "obj", "[]")
		return nil
	}
	for _, v := range *groups {
		log.Info("getRegisteredValidatorGroups:", "obj", v.String())
	}
	return nil
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
	}
	if len(*ValidatorSigners) == 0 {
		log.Info("ValidatorSigners:", "obj", "[]")
		return nil
	}
	for _, v := range *ValidatorSigners {
		fmt.Println("getRegisteredValidatorSigners:", v.String())
	}
	return nil
}

func getTopGroupValidators(ctx *cli.Context) error {
	//--------------------------- pre set -------------------------------------------
	path := pathGroup
	n := big.NewInt(5) // top number
	password = ""
	//-------------------------------------------------------------------------------
	if ctx.IsSet(TopNumFlag.Name) {
		n = big.NewInt(ctx.GlobalInt64(TopNumFlag.Name))
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	methodName := "getTopGroupValidators"
	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	header, err := conn.HeaderByNumber(context.Background(), nil)

	log.Info("=== getTopGroupValidators Group", "obj", from, " ===")
	input := packInput(abiValidators, methodName, from, n)
	msg := ethchain.CallMsg{From: from, To: &ValidatorAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "error", err)
	}
	TopValidators := new([]common.Address)
	err = abiValidators.UnpackIntoInterface(&TopValidators, methodName, output)
	if err != nil {
		log.Error("method UnpackIntoInterface", "error", err)
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
