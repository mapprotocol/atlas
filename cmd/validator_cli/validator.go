package main

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/validator_cli/env"
	"github.com/mapprotocol/atlas/params"

	"gopkg.in/urfave/cli.v1"
	"math/big"
	"os"
)

var registerValidatorCommand = cli.Command{
	Name:   "registerValidator",
	Usage:  "register validator ",
	Action: MigrateFlags(registerValidator),
	Flags:  ValidatorFlags,
}
var createAccountCommand = cli.Command{
	Name:   "createAccount",
	Usage:  "creat validator account",
	Action: MigrateFlags(createAccount1),
	Flags:  ValidatorFlags,
}
var deregisterValidatorCommand = cli.Command{
	Name:   "deregisterValidator",
	Usage:  "deregister validator",
	Action: MigrateFlags(deregisterValidator),
	Flags:  ValidatorFlags,
}
var lockedMAPCommand = cli.Command{
	Name:   "lockedMAP",
	Usage:  "locked MAP",
	Action: MigrateFlags(lockedMAP),
	Flags:  ValidatorFlags,
}

const (
	datadirPrivateKey      = "key"
	datadirDefaultKeyStore = "keystore"
	RegisterAmount         = 100000
	RewardInterval         = 14
	commission             = 80
)

var (
	abiValidators  *abi.ABI
	abiLocaledGold *abi.ABI
	abiAccounts    *abi.ABI
	abiElection    *abi.ABI

	ValidatorAddress  = MustProxyAddressFor("Validators")
	LockedGoldAddress = MustProxyAddressFor("LockedGold")
	AccountsAddress   = MustProxyAddressFor("Accounts")
	ElectionAddress   = MustProxyAddressFor("Election")

	priKey   *ecdsa.PrivateKey
	from     common.Address
	password string
	Value    uint64
	fee      uint64

	Base = new(big.Int).SetUint64(10000)
)

func init() {
	abiValidators = AbiFor("Validators")
	abiLocaledGold = AbiFor("LockedGold")
	abiAccounts = AbiFor("Accounts")
	abiElection = AbiFor("Election")

	ValidatorAddress = MustProxyAddressFor("Validators")
	LockedGoldAddress = MustProxyAddressFor("LockedGold")
	AccountsAddress = MustProxyAddressFor("Accounts")
	ElectionAddress = MustProxyAddressFor("Election")

	Base = new(big.Int).SetUint64(10000)
	password = ""

	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.LvlInfo)
	log.Root().SetHandler(glogger)
}
func registerValidator(ctx *cli.Context) error {
	//------------------ pre set --------------------------
	path := ""
	password = "111111"
	commission := int64(80)
	lesser := params.ZeroAddress
	greater := params.ZeroAddress
	//-----------------------------------------------------

	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(CommissionFlag.Name) {
		commission = ctx.GlobalInt64(CommissionFlag.Name)
	}
	if ctx.IsSet(lesserFlag.Name) {
		lesser = common.HexToAddress(ctx.GlobalString(lesserFlag.Name))
	}
	if ctx.IsSet(greaterFlag.Name) {
		greater = common.HexToAddress(ctx.GlobalString(greaterFlag.Name))
	}
	validator := loadAccount(path, password)

	blsPub, err := validator.BLSPublicKey()
	if err != nil {
		return err
	}

	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	//---------------------------- create account ----------------------------------
	createAccount(conn, validator, "validator")

	//---------------------------- lock ----------------------------------
	groupRequiredGold := params.MustBigInt("10000000000000000000000") // 10k Atlas per validator,
	log.Info("=== Lock validator gold ===")
	log.Info("Lock group gold", "amount", groupRequiredGold)
	input := packInput(abiLocaledGold, "lock")
	txHash := sendContractTransaction(conn, validator.Address, LockedGoldAddress, groupRequiredGold, priKey, input)
	getResult(conn, txHash, true)

	//----------------------------- registerValidator ---------------------------------
	log.Info("=== Register validator ===")
	pubKey := validator.PublicKey()[1:]
	input = packInput(abiValidators, "registerValidator", big.NewInt(commission), lesser, greater, pubKey, blsPub[:], validator.MustBLSProofOfPossession())
	txHash = sendContractTransaction(conn, validator.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)
	return nil
}
func lockedMAP(ctx *cli.Context) error {
	//------------------ pre set --------------------------
	path := ""
	password = "111111"
	//-----------------------------------------------------

	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	loadPrivateKey(path)
	account := loadAccount(path, password)
	conn, _ := dialConn(ctx)
	groupRequiredGold := params.MustBigInt("10000000000000000000000") // 10k Atlas per groupAccount,
	log.Info("=== Lock group gold ===")
	log.Info("Lock group gold", "amount", groupRequiredGold)
	input := packInput(abiLocaledGold, "lock")
	txHash := sendContractTransaction(conn, account.Address, LockedGoldAddress, groupRequiredGold, priKey, input)
	getResult(conn, txHash, true)
	return nil

}
func createAccount1(ctx *cli.Context) error {
	//------------------ pre set --------------------------
	path := ""
	password = "111111"
	namePrefix := "validator"
	//-----------------------------------------------------

	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(NamePrefixFlag.Name) {
		namePrefix = ctx.GlobalString(NamePrefixFlag.Name)
	}
	loadPrivateKey(path)
	validator := loadAccount(path, password)
	conn, _ := dialConn(ctx)
	createAccount(conn, validator, namePrefix)
	return nil
}
func createAccount(conn *ethclient.Client, account env.Account, namePrefix string) {
	logger := log.New("func", "createAccount")
	logger.Info("Create account", "address", account.Address, "name", namePrefix)

	log.Info("=== create Account ===")
	input := packInput(abiAccounts, "createAccount")
	txHash := sendContractTransaction(conn, account.Address, AccountsAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	log.Info("=== setName name ===")
	input = packInput(abiAccounts, "setName", namePrefix)
	txHash = sendContractTransaction(conn, account.Address, AccountsAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	log.Info("=== setAccountDataEncryptionKey ===")
	input = packInput(abiAccounts, "setAccountDataEncryptionKey", account.PublicKey())
	txHash = sendContractTransaction(conn, account.Address, AccountsAddress, nil, priKey, input)
	getResult(conn, txHash, true)
}
func deregisterValidator(ctx *cli.Context) error {
	//------------------------pre set ------------------------------------------------
	path := ""
	password = "111111"
	idx := big.NewInt(int64(4)) // index in registeredValidators
	//--------------------------------------------------------------------------------

	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	if ctx.IsSet(IdxFlag.Name) {
		idx = big.NewInt(ctx.GlobalInt64(IdxFlag.Name))
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	validator := loadAccount(path, password)
	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	//----------------------------- deregisterValidator --------------------------------
	log.Info("====== deregisterValidator ======")
	input := packInput(abiValidators, "deregisterValidator", idx)
	txHash := sendContractTransaction(conn, validator.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)
	return nil
}
