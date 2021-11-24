package main

import (
	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"
)

var addValidatorToGroupCommand = cli.Command{
	Name:   "addValidatorToGroups",
	Usage:  "add Validator to Groups ",
	Action: MigrateFlags(addValidatorToGroup),
	Flags:  ValidatorFlags,
}

func addValidatorToGroup(ctx *cli.Context) error {

	conn, url := dialConn(ctx)
	printBaseInfo(conn, url)

	path := ""
	password = ""
	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	validator := loadAccount(path, password)
	groupAddress := loadAccount("", "password")
	loadPrivateKey(path)
	log.Info("Add validator to group", "validator", validator.Address, "groupAddress", groupAddress.Address)

	input := packInput(abiValidators, "affiliate", groupAddress)
	txHash := sendContractTransaction(conn, validator.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	loadPrivateKey("")
	log.Info("Register validator")
	input = packInput(abiValidators, "addMember", validator.Address)
	txHash = sendContractTransaction(conn, groupAddress.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	return nil
}
