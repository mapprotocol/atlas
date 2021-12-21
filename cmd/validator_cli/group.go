package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
	"math/big"
)

var (
	//------------------------------- pre set -------------------------------------
	pathValidator01 = "D:/root/keystore/UTC--2021-09-08T08-00-15.473724074Z--1c0edab88dbb72b119039c4d14b1663525b3ac15"
	pathValidator02 = "D:/root/keystore/UTC--2021-09-08T10-12-17.687481942Z--16fdbcac4d4cc24dca47b9b80f58155a551ca2af"
	pathValidator03 = "D:/root/keystore/UTC--2021-09-08T10-16-18.520295371Z--2dc45799000ab08e60b7441c36fcc74060ccbe11"
	pathValidator04 = "D:/root/keystore/UTC--2021-09-08T10-16-35.698273293Z--6c5938b49bacde73a8db7c3a7da208846898bff5"

	pathValidator1 = "D:/root/keystore/UTC--2021-07-19T02-09-17.552426700Z--81f02fd21657df80783755874a92c996749777bf"
	pathValidator2 = "D:/root/keystore/UTC--2021-07-19T02-04-57.993791200Z--df945e6ffd840ed5787d367708307bd1fa3d40f4"
	pathValidator3 = "D:/root/keystore/UTC--2021-07-19T02-05-14.453062600Z--32cd75ca677e9c37fd989272afa8504cb8f6eb52"
	pathValidator4 = "D:/root/keystore/UTC--2021-07-19T02-07-11.808701800Z--3e3429f72450a39ce227026e8ddef331e9973e4d"

	pathGroup = "D:/root/keystore/UTC--2021-11-11T13-28-01.812954600Z--ce90710a4673b87a6881b0907358119baf0304a5"
)
var addFirstMemberCommand = cli.Command{
	Name:   "addFirstMember",
	Usage:  "add first member validator to Validators ",
	Action: MigrateFlags(addFirstMemberToGroup),
	Flags:  ValidatorFlags,
}

var addToGroupCommand = cli.Command{
	Name:   "addToGroup",
	Usage:  "add Validator to Validators ",
	Action: MigrateFlags(addValidatorToGroup),
	Flags:  ValidatorFlags,
}

var removeMemberCommand = cli.Command{
	Name:   "removeMember",
	Usage:  "add Validator to Validators ",
	Action: MigrateFlags(removeMember),
	Flags:  ValidatorFlags,
}

var deregisterValidatorGroupCommand = cli.Command{
	Name:   "deregisterValidatorGroup",
	Usage:  "deregister validator group",
	Action: MigrateFlags(deregisterValidatorGroup),
	Flags:  ValidatorFlags,
}

var affiliateCommand = cli.Command{
	Name:   "affiliate",
	Usage:  "affiliate validator group",
	Action: MigrateFlags(affiliate),
	Flags:  ValidatorFlags,
}

func addFirstMemberToGroup(ctx *cli.Context) error {
	//--------------- pre set ----------------------------------
	passwordGroup := ""
	//----------------------------------------------------------
	if ctx.IsSet(PasswordFlag.Name) {
		passwordGroup = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(KeyStoreFlag.Name) {
		pathGroup = ctx.GlobalString(KeyStoreFlag.Name)
	}
	validator := "0x81f02fd21657df80783755874a92c996749777bf"
	if ctx.IsSet(AddressFlag.Name) {
		validator = ctx.GlobalString(AddressFlag.Name)
	}

	conn, _ := dialConn(ctx)
	group := loadAccount(pathGroup, passwordGroup)
	password = passwordGroup
	loadPrivateKey(pathGroup)
	log.Info("Add validator to group", "validator", validator, "groupAddress", group.Address)

	//--------------------- affiliate -------------------------------
	//log.Info("Validator affiliate to group")
	//input := packInput(abiValidators, "affiliate", group.Address)
	//txHash := sendContractTransaction(conn,validator, ValidatorAddress, nil, priKey, input)
	//getResult(conn, txHash, true)

	//--------------------- addFirstMember --------------------------
	log.Info("addMember validator to group")
	input := packInput(abiValidators, "addFirstMember", common.HexToAddress(validator), params.ZeroAddress, params.ZeroAddress)
	txHash := sendContractTransaction(conn, group.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	return nil
}

func addValidatorToGroup(ctx *cli.Context) error {
	//--------- pre set ------------
	passwordGroup := ""
	//------------------------------
	if ctx.IsSet(PasswordFlag.Name) {
		passwordGroup = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(KeyStoreFlag.Name) {
		pathGroup = ctx.GlobalString(KeyStoreFlag.Name)
	}
	addMemberFunc := func(validatorAccount string) {
		conn, _ := dialConn(ctx)
		group := loadAccount(pathGroup, passwordGroup)
		log.Info("Add validator to group", "validator", validatorAccount, "groupAddress", group.Address)

		//--------------------- addMember --------------------------
		password = passwordGroup
		loadPrivateKey(pathGroup)
		log.Info("=== addMember validator to group ===")
		input := packInput(abiValidators, "addMember", common.HexToAddress(validatorAccount))
		txHash := sendContractTransaction(conn, group.Address, ValidatorAddress, nil, priKey, input)
		getResult(conn, txHash, true)
	}
	var Validatorlist = []string{
		"0x81f02fd21657df80783755874a92c996749777bf",
		"0xdf945e6ffd840ed5787d367708307bd1fa3d40f4",
		"0x32cd75ca677e9c37fd989272afa8504cb8f6eb52",
		"0x3e3429f72450a39ce227026e8ddef331e9973e4d",
	}

	if ctx.IsSet(AddressFlag.Name) {
		address := ctx.GlobalString(AddressFlag.Name)
		addMemberFunc(address)
		return nil
	}

	for _, v := range Validatorlist {
		addMemberFunc(v)
	}

	return nil
}

func removeMember(ctx *cli.Context) error {
	//-------------- pre set ------------------------
	passwordGroup := ""
	//----------------------------------------------
	if ctx.IsSet(PasswordFlag.Name) {
		passwordGroup = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(KeyStoreFlag.Name) {
		pathGroup = ctx.GlobalString(KeyStoreFlag.Name)
	}

	removeMemberFunc := func(validatorAccount string) {
		conn, _ := dialConn(ctx)
		group := loadAccount(pathGroup, passwordGroup)
		log.Info("remove Member", "validator", validatorAccount, "groupAddress", group.Address)
		//--------------------- removeMember --------------------------
		password = passwordGroup
		loadPrivateKey(pathGroup)
		log.Info("=== removeMember ===")
		input := packInput(abiValidators, "removeMember", common.HexToAddress(validatorAccount))
		txHash := sendContractTransaction(conn, group.Address, ValidatorAddress, nil, priKey, input)
		getResult(conn, txHash, true)
	}
	var Validatorlist = []string{
		"0x1c0edab88dbb72b119039c4d14b1663525b3ac15",
		"0x16fdbcac4d4cc24dca47b9b80f58155a551ca2af",
		"0x2dc45799000ab08e60b7441c36fcc74060ccbe11",
		"0x6c5938b49bacde73a8db7c3a7da208846898bff5",

		"0x81f02fd21657df80783755874a92c996749777bf",
		"0xdf945e6ffd840ed5787d367708307bd1fa3d40f4",
		"0x32cd75ca677e9c37fd989272afa8504cb8f6eb52",
		"0x3e3429f72450a39ce227026e8ddef331e9973e4d",
	}

	if ctx.IsSet(AddressFlag.Name) {
		address := ctx.GlobalString(AddressFlag.Name)
		removeMemberFunc(address)
		return nil
	}

	for _, v := range Validatorlist {
		removeMemberFunc(v)
	}

	return nil
}

func deregisterValidatorGroup(ctx *cli.Context) error {
	//------------------------pre set ------------------------------------------------
	path := ""
	password = "111111"
	n := int64(1) // index in group
	//--------------------------------------------------------------------------------

	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	validator := loadAccount(path, password)
	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	//----------------------------- deregisterValidatorGroup --------------------------------
	log.Info("====== deregisterValidator ======")
	input := packInput(abiValidators, "deregisterValidatorGroup", big.NewInt(n))
	txHash := sendContractTransaction(conn, validator.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)
	return nil
}

func affiliate(ctx *cli.Context) error {
	//--------------- pre set ----------------------------------
	path := pathValidator1
	password = ""
	groupAddress := ""
	//----------------------------------------------------------
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(GroupAddressFlag.Name) {
		groupAddress = ctx.GlobalString(GroupAddressFlag.Name)
	}
	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	loadPrivateKey(path)
	account := loadAccount(path, password)
	conn, _ := dialConn(ctx)
	log.Info("=== Validator affiliate to group ===")
	input := packInput(abiValidators, "affiliate", common.HexToAddress(groupAddress))
	txHash := sendContractTransaction(conn, account.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)
	return nil
}
