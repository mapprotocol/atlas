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

var registerGroupCommand = cli.Command{
	Name:   "registerGroup",
	Usage:  "register group ",
	Action: MigrateFlags(registerGroup),
	Flags:  ValidatorFlags,
}
var registerValidatorCommand = cli.Command{
	Name:   "registerValidator",
	Usage:  "register validator ",
	Action: MigrateFlags(registerValidator),
	Flags:  ValidatorFlags,
}
var setMaxGroupSizeCommand = cli.Command{
	Name:   "setMaxGroupSize",
	Usage:  "set Max Group Size",
	Action: MigrateFlags(setMaxGroupSize),
	Flags:  ValidatorFlags,
}

var deregisterValidatorCommand = cli.Command{
	Name:   "deregisterValidator",
	Usage:  "deregister validator",
	Action: MigrateFlags(deregisterValidator),
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

	ValidatorAddress  = MustProxyAddressFor("Validators")
	LockedGoldAddress = MustProxyAddressFor("LockedGold")
	AccountsAddress   = MustProxyAddressFor("Accounts")

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

	ValidatorAddress = MustProxyAddressFor("Validators")
	LockedGoldAddress = MustProxyAddressFor("LockedGold")
	AccountsAddress = MustProxyAddressFor("Accounts")

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
	//-----------------------------------------------------

	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
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
	input = packInput(abiValidators, "registerValidator", pubKey, blsPub[:], validator.MustBLSProofOfPossession())
	txHash = sendContractTransaction(conn, validator.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)
	return nil
}

func registerGroup(ctx *cli.Context) error {
	//------------------ pre set --------------------------------------------------
	path := ""
	password = "111111"
	commission := int64(80)
	//------------------------------------------------------------------------------

	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(CommissionFlag.Name) {
		commission = ctx.GlobalInt64(CommissionFlag.Name)
	}
	groupAccount := loadAccount(path, password)

	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	//---------------------------- create account ----------------------------------
	createAccount(conn, groupAccount, "group")

	//---------------------------- lock --------------------------------------------
	groupRequiredGold := params.MustBigInt("10000000000000000000000") // 10k Atlas per groupAccount,
	log.Info("=== Lock group gold ===")
	log.Info("Lock group gold", "amount", groupRequiredGold)
	input := packInput(abiLocaledGold, "lock")
	txHash := sendContractTransaction(conn, groupAccount.Address, LockedGoldAddress, groupRequiredGold, priKey, input)
	getResult(conn, txHash, true)

	//----------------------------- registerValidator -----------------------------
	log.Info("=== Register group ===")
	input = packInput(abiValidators, "registerValidatorGroup", big.NewInt(commission))
	txHash = sendContractTransaction(conn, groupAccount.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)
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

func setMaxGroupSize(ctx *cli.Context) error {
	//------------------------pre set ------------------------------------------------
	path := ""
	password = ""
	maxSize := int64(100)
	//--------------------------------------------------------------------------------

	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(maxSizeFlag.Name) {
		maxSize = ctx.GlobalInt64(maxSizeFlag.Name)
	}
	validator := loadAccount(path, password)
	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	//----------------------------- registerValidator --------------------------------
	log.Info("====== set Max Group Size ======")
	input := packInput(abiValidators, "setMaxGroupSize", big.NewInt(maxSize))
	txHash := sendContractTransaction(conn, validator.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)
	return nil
}

func deregisterValidator(ctx *cli.Context) error {
	//------------------------pre set ------------------------------------------------
	path := ""
	password = "111111"
	n := big.NewInt(int64(4)) // index in registeredValidators
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
	//----------------------------- deregisterValidator --------------------------------
	log.Info("====== deregisterValidator ======")
	input := packInput(abiValidators, "deregisterValidator", n)
	txHash := sendContractTransaction(conn, validator.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)
	return nil
}
