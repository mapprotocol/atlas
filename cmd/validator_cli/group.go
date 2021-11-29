package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
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
	Usage:  "add first member validator to Groups ",
	Action: MigrateFlags(addFirstMemberToGroup),
	Flags:  ValidatorFlags,
}

var addToGroupCommand = cli.Command{
	Name:   "addToGroup",
	Usage:  "add Validator to Groups ",
	Action: MigrateFlags(addValidatorToGroup),
	Flags:  ValidatorFlags,
}

var removeMemberCommand = cli.Command{
	Name:   "removeMember",
	Usage:  "add Validator to Groups ",
	Action: MigrateFlags(removeMember),
	Flags:  ValidatorFlags,
}

var deregisterValidatorGroupCommand = cli.Command{
	Name:   "deregisterValidatorGroup",
	Usage:  "deregister validator group",
	Action: MigrateFlags(deregisterValidatorGroup),
	Flags:  ValidatorFlags,
}

func addFirstMemberToGroup(ctx *cli.Context) error {
	//--------------- pre set ----------------------------------
	path := pathValidator1
	passwordValidator := "111111"
	passwordGroup := ""
	//----------------------------------------------------------
	if ctx.IsSet(PasswordFlag.Name) {
		passwordGroup = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(KeyStoreFlag.Name) {
		pathGroup = ctx.GlobalString(KeyStoreFlag.Name)
	}

	if ctx.IsSet(ReadConfigFlag.Name) {
		type AccoutInfo struct {
			Account  string
			Password string
		}
		type ValidatorsInfo struct {
			Validators []AccoutInfo
		}
		keyDir := fmt.Sprintf("./config/validatorCfg.json")
		data, err := ioutil.ReadFile(keyDir)
		if err != nil {
			log.Crit(" readFile Err:", "err:", err.Error())
		}

		ValidatorsInfoCfg := &ValidatorsInfo{}
		_ = json.Unmarshal(data, ValidatorsInfoCfg)

		passwordValidator = ValidatorsInfoCfg.Validators[0].Password
		path = ValidatorsInfoCfg.Validators[0].Account
	}

	conn, _ := dialConn(ctx)
	validator := loadAccount(path, passwordValidator)
	group := loadAccount(pathGroup, passwordGroup)
	password = passwordValidator
	loadPrivateKey(path)
	log.Info("Add validator to group", "validator", validator.Address, "groupAddress", group.Address)

	//--------------------- affiliate -------------------------------
	log.Info("Validator affiliate to group")
	input := packInput(abiValidators, "affiliate", group.Address)
	txHash := sendContractTransaction(conn, validator.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	//--------------------- addFirstMember --------------------------
	password = passwordGroup
	loadPrivateKey(pathGroup)
	log.Info("addMember validator to group")
	input = packInput(abiValidators, "addFirstMember", validator.Address, params.ZeroAddress, params.ZeroAddress)
	txHash = sendContractTransaction(conn, group.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	return nil
}

func addValidatorToGroup(ctx *cli.Context) error {
	//--------- pre set ------------
	passwordGroup := ""
	passwordValidator := "111111"
	//------------------------------
	if ctx.IsSet(PasswordFlag.Name) {
		passwordGroup = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(KeyStoreFlag.Name) {
		pathGroup = ctx.GlobalString(KeyStoreFlag.Name)
	}

	addMemberFunc := func(path string, _password string) {
		conn, _ := dialConn(ctx)

		validator := loadAccount(path, _password)
		group := loadAccount(pathGroup, passwordGroup)
		password = _password
		loadPrivateKey(path)
		log.Info("Add validator to group", "validator", validator.Address, "groupAddress", group.Address)

		//--------------------- affiliate --------------------------
		log.Info("=== Validator affiliate to group ===")
		input := packInput(abiValidators, "affiliate", group.Address)
		txHash := sendContractTransaction(conn, validator.Address, ValidatorAddress, nil, priKey, input)
		getResult(conn, txHash, true)

		//--------------------- addMember --------------------------
		password = passwordGroup
		loadPrivateKey(pathGroup)
		log.Info("=== addMember validator to group ===")
		input = packInput(abiValidators, "addMember", validator.Address)
		txHash = sendContractTransaction(conn, group.Address, ValidatorAddress, nil, priKey, input)
		getResult(conn, txHash, true)
	}
	var Validatorlist = []struct {
		a string
		b string
	}{
		{pathValidator1, passwordValidator},
		{pathValidator2, passwordValidator},
		{pathValidator3, passwordValidator},
		{pathValidator4, passwordValidator},
	}
	if ctx.IsSet(ReadConfigFlag.Name) {
		type AccoutInfo struct {
			Account  string
			Password string
		}
		type ValidatorsInfo struct {
			Validators []AccoutInfo
		}
		keyDir := fmt.Sprintf("./config/validatorCfg.json")
		data, err := ioutil.ReadFile(keyDir)
		if err != nil {
			log.Crit(" readFile Err:", "err:", err.Error())
		}

		ValidatorsInfoCfg := &ValidatorsInfo{}
		_ = json.Unmarshal(data, ValidatorsInfoCfg)

		for _, v := range ValidatorsInfoCfg.Validators {
			addMemberFunc(v.Account, v.Password)
		}
		return nil
	}
	for _, v := range Validatorlist {
		addMemberFunc(v.a, v.b)
	}

	return nil
}

func removeMember(ctx *cli.Context) error {
	//-------------- pre set ------------------------
	passwordValidator := "111111"
	passwordGroup := ""
	//----------------------------------------------
	if ctx.IsSet(PasswordFlag.Name) {
		passwordGroup = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(KeyStoreFlag.Name) {
		pathGroup = ctx.GlobalString(KeyStoreFlag.Name)
	}

	removeMemberFunc := func(pathValidator string, _password string) {
		conn, _ := dialConn(ctx)

		path := pathValidator
		validator := loadAccount(path, _password)
		group := loadAccount(pathGroup, passwordGroup)
		password = _password
		loadPrivateKey(path)
		log.Info("remove Member", "validator", validator.Address, "groupAddress", group.Address)
		//--------------------- removeMember --------------------------
		password = passwordGroup
		loadPrivateKey(pathGroup)
		log.Info("=== removeMember ===")
		input := packInput(abiValidators, "removeMember", validator.Address)
		txHash := sendContractTransaction(conn, group.Address, ValidatorAddress, nil, priKey, input)
		getResult(conn, txHash, true)
	}
	var Validatorlist = []struct {
		a string
		b string
	}{
		{pathValidator01, passwordValidator},
		{pathValidator02, passwordValidator},
		{pathValidator03, passwordValidator},
		{pathValidator04, passwordValidator},
		{pathValidator1, passwordValidator},
		{pathValidator2, passwordValidator},
		{pathValidator3, passwordValidator},
		{pathValidator4, passwordValidator},
	}

	if ctx.IsSet(ReadConfigFlag.Name) {
		type AccoutInfo struct {
			Account  string
			Password string
		}
		type ValidatorsInfo struct {
			Validators []AccoutInfo
		}
		keyDir := fmt.Sprintf("./config/validatorCfg.json")
		data, err := ioutil.ReadFile(keyDir)
		if err != nil {
			log.Crit(" readFile Err:", "err:", err.Error())
		}

		ValidatorsInfoCfg := &ValidatorsInfo{}
		_ = json.Unmarshal(data, ValidatorsInfoCfg)

		for _, v := range ValidatorsInfoCfg.Validators {
			removeMemberFunc(v.Account, v.Password)
		}
		return nil
	}

	for _, v := range Validatorlist {
		removeMemberFunc(v.a, v.b)
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
