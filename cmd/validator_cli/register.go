package main

import (
	"crypto/ecdsa"
	"fmt"

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
	path := ""
	password = "111111"
	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
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
	log.Info("Lock group gold", "amount", groupRequiredGold)
	input := packInput(abiLocaledGold, "lock")
	txHash := sendContractTransaction(conn, validator.Address, LockedGoldAddress, groupRequiredGold, priKey, input)
	getResult(conn, txHash, true)

	//----------------------------- registerValidator ---------------------------------
	log.Info("Register validator")
	pubKey := validator.PublicKey()[1:]
	input = packInput(abiValidators, "registerValidator", pubKey, blsPub[:], validator.MustBLSProofOfPossession())
	txHash = sendContractTransaction(conn, validator.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)
	if err != nil {
		return err
	}
	return nil
}
func registerGroup(ctx *cli.Context) error {
	validator := loadAccount("", "password")

	blsPub, err := validator.BLSPublicKey()
	if err != nil {
		return err
	}
	fmt.Println(abiValidators.Methods)

	loadPrivate(ctx)
	conn, url := dialConn(ctx)
	printBaseInfo(conn, url)
	// remove the 0x04 prefix from the pub key (we need the 64 bytes variant)
	pubKey := validator.PublicKey()[1:]
	groupRequiredGold := new(big.Int).Mul(
		params.MustBigInt("10000000000000000000000"), // 10k Atlas per validator,
		big.NewInt(4),
	)

	log.Info("Lock group gold", "amount", groupRequiredGold)
	input := packInput(abiLocaledGold, "lock")
	txHash := sendContractTransaction(conn, validator.Address, LockedGoldAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	log.Info("Register validator")
	input = packInput(abiValidators, "registerValidator", pubKey, blsPub[:], validator.MustBLSProofOfPossession())
	txHash = sendContractTransaction(conn, validator.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	return nil
}

func createAccount(conn *ethclient.Client, account env.Account, namePrefix string) {
	logger := log.New("func", "createAccount")
	logger.Info("Create account", "address", account.Address, "name", namePrefix)

	log.Info("create Account")
	input := packInput(abiAccounts, "createAccount")
	txHash := sendContractTransaction(conn, account.Address, AccountsAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	log.Info("setName name")
	input = packInput(abiAccounts, "setName", namePrefix)
	txHash = sendContractTransaction(conn, account.Address, AccountsAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	log.Info("setAccountDataEncryptionKey")
	input = packInput(abiAccounts, "setAccountDataEncryptionKey", account.PublicKey())
	txHash = sendContractTransaction(conn, account.Address, AccountsAddress, nil, priKey, input)
	getResult(conn, txHash, true)
}
