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
	Action: MigrateFlags(getTopGroupValidators),
	Flags:  ValidatorFlags,
}

func queryGroups(ctx *cli.Context) error {
	path := ""
	password = ""
	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	header, err := conn.HeaderByNumber(context.Background(), nil)
	input := packInput(abiValidators, "getRegisteredValidatorGroups")
	msg := ethchain.CallMsg{From: from, To: &ValidatorAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "err", err)
	}
	groups := new([]common.Address)
	err = abiValidators.UnpackIntoInterface(&groups, "getRegisteredValidatorGroups", output)
	if err != nil {
		log.Error("method UnpackIntoInterface error", "err", err)
	}
	for _, v := range *groups {
		fmt.Println("getRegisteredValidatorGroups:", v.String())
	}
	return nil
}

func getRegisteredValidatorSigners(ctx *cli.Context) error {
	methodName := "getRegisteredValidatorSigners"
	path := ""
	password = ""
	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	header, err := conn.HeaderByNumber(context.Background(), nil)
	input := packInput(abiValidators, methodName)
	msg := ethchain.CallMsg{From: from, To: &ValidatorAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "err", err)
	}
	ValidatorSigners := new([]common.Address)
	err = abiValidators.UnpackIntoInterface(&ValidatorSigners, methodName, output)
	if err != nil {
		log.Error("method UnpackIntoInterface error", err)
	}
	for _, v := range *ValidatorSigners {
		fmt.Println("getRegisteredValidatorSigners:", v.String())
	}
	return nil
}

func getTopGroupValidators(ctx *cli.Context) error {
	methodName := "getTopGroupValidators"
	path := pathGroup
	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	header, err := conn.HeaderByNumber(context.Background(), nil)
	n := big.NewInt(1) // top 5
	log.Info("getTopGroupValidators Group", "address", from)
	input := packInput(abiValidators, methodName, from, n)
	msg := ethchain.CallMsg{From: from, To: &ValidatorAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "err", err)
	}
	TopValidators := new([]common.Address)
	err = abiValidators.UnpackIntoInterface(&TopValidators, methodName, output)
	if err != nil {
		log.Error("method UnpackIntoInterface", " error", err)
	}
	for _, v := range *TopValidators {
		log.Info("Address:", "address", v.String())
	}
	return nil
}
