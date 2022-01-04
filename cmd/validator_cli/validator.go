package main

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/params"

	"gopkg.in/urfave/cli.v1"
	"math/big"
	"os"
)

var registerValidatorCommand = cli.Command{
	Name:   "register",
	Usage:  "register validator",
	Action: MigrateFlags(registerValidator),
	Flags:  ValidatorFlags,
}
var createAccountCommand = cli.Command{
	Name:   "createAccount",
	Usage:  "creat validator account",
	Action: MigrateFlags(createAccount1),
	Flags:  ValidatorFlags,
}

var lockedMAPCommand = cli.Command{
	Name:   "lockedMAP",
	Usage:  "locked MAP",
	Action: MigrateFlags(lockedMAP),
	Flags:  ValidatorFlags,
}
var unlockedMAPCommand = cli.Command{
	Name:   "unlockedMAP",
	Usage:  "unlocked MAP",
	Action: MigrateFlags(unlockedMAP),
	Flags:  ValidatorFlags,
}
var relockMAPCommand = cli.Command{
	Name:   "relockMAP",
	Usage:  "unlocked MAP",
	Action: MigrateFlags(relockMAP),
	Flags:  ValidatorFlags,
}
var withdrawCommand = cli.Command{
	Name:   "withdraw",
	Usage:  "withdraw MAP",
	Action: MigrateFlags(withdraw),
	Flags:  ValidatorFlags,
}

var validatorCommand = cli.Command{
	Name:  "validator",
	Usage: "validator commands",
	Subcommands: []cli.Command{
		createAccountCommand,
		lockedMAPCommand,
		registerValidatorCommand,
		unlockedMAPCommand,
		relockMAPCommand,
		withdrawCommand,

		queryRegisteredValidatorSignersCommand,
		queryTopValidatorsCommand,
	},
}

var (
	abiValidators *abi.ABI
	abiLockedGold *abi.ABI
	abiAccounts   *abi.ABI
	abiElection   *abi.ABI
	abiGoldToken  *abi.ABI

	ValidatorAddress  = MustProxyAddressFor("Validators")
	LockedGoldAddress = MustProxyAddressFor("LockedGold")
	AccountsAddress   = MustProxyAddressFor("Accounts")
	ElectionAddress   = MustProxyAddressFor("Election")
	GoldTokenAddress  = MustProxyAddressFor("StableToken")

	priKey   *ecdsa.PrivateKey
	from     common.Address
	password string
	Value    uint64
	fee      uint64

	Base = new(big.Int).SetUint64(10000)
)

func init() {
	abiValidators = AbiFor("Validators")
	abiLockedGold = AbiFor("LockedGold")
	abiAccounts = AbiFor("Accounts")
	abiElection = AbiFor("Election")
	abiGoldToken = AbiFor("GoldToken")

	ValidatorAddress = MustProxyAddressFor("Validators")
	LockedGoldAddress = MustProxyAddressFor("LockedGold")
	AccountsAddress = MustProxyAddressFor("Accounts")
	ElectionAddress = MustProxyAddressFor("Election")
	GoldTokenAddress = MustProxyAddressFor("GoldToken")

	Base = new(big.Int).SetUint64(10000)
	password = ""

	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.LvlInfo)
	log.Root().SetHandler(glogger)
}

func registerValidator(ctx *cli.Context, config *Config) error {
	//---------------------------- create account ----------------------------------
	createAccount(config.conn, config, "validator")
	//---------------------------- lock ----------------------------------
	groupRequiredGold := params.MustBigInt("10000000000000000000000") // 10k Atlas per validator,
	log.Info("=== Lock validator gold ===")
	log.Info("Lock group gold", "amount", groupRequiredGold)
	input := packInput(abiLockedGold, "lock")
	txHash := sendContractTransaction(config.conn, config.from, LockedGoldAddress, groupRequiredGold, priKey, input)
	getResult(config.conn, txHash, true)

	//----------------------------- registerValidator ---------------------------------
	log.Info("=== Register validator ===")
	input = packInput(abiValidators, "registerValidator", big.NewInt(config.Commission), config.lesser, config.greater, config.PublicKey, config.BlsPub[:], config.BLSProof)
	txHash = sendContractTransaction(config.conn, config.from, ValidatorAddress, nil, priKey, input)
	getResult(config.conn, txHash, true)
	return nil
}
func lockedMAP(ctx *cli.Context, config *Config) error {
	groupRequiredGold := params.MustBigInt("10000000000000000000000") // 10k Atlas per groupAccount,
	log.Info("=== Lock  gold ===")
	log.Info("Lock  gold", "amount", groupRequiredGold)
	input := packInput(abiLockedGold, "lock")
	txHash := sendContractTransaction(config.conn, config.from, LockedGoldAddress, groupRequiredGold, priKey, input)
	getResult(config.conn, txHash, true)
	return nil
}

func unlockedMAP(ctx *cli.Context, config *Config) error {
	groupRequiredGold := params.MustBigInt("10000000000000000000000") // 10k Map
	log.Info("=== unLock validator gold ===")
	log.Info("unLock validator gold", "amount", groupRequiredGold)
	input := packInput(abiLockedGold, "unlock", groupRequiredGold)
	txHash := sendContractTransaction(config.conn, config.from, LockedGoldAddress, nil, priKey, input)
	getResult(config.conn, txHash, true)
	return nil
}
func relockMAP(ctx *cli.Context, config *Config) error {
	groupRequiredGold := params.MustBigInt("10000000000000000000000") // 10k Map
	log.Info("=== relockMAP validator gold ===")
	log.Info("relockMAP validator gold", "amount", groupRequiredGold)
	input := packInput(abiLockedGold, "relock", big.NewInt(0), groupRequiredGold)
	txHash := sendContractTransaction(config.conn, config.from, LockedGoldAddress, nil, priKey, input)
	getResult(config.conn, txHash, true)
	return nil
}

func withdraw(ctx *cli.Context, config *Config) error {
	log.Info("=== withdraw validator gold ===")
	input := packInput(abiLockedGold, "withdraw", big.NewInt(0))
	txHash := sendContractTransaction(config.conn, config.from, LockedGoldAddress, nil, priKey, input)
	getResult(config.conn, txHash, true)
	return nil
}
func createAccount1(ctx *cli.Context, config *Config) error {
	createAccount(config.conn, config, "validator")
	return nil
}
func createAccount(conn *ethclient.Client, cfg *Config, namePrefix string) {
	logger := log.New("func", "createAccount")
	logger.Info("Create account", "address", cfg.from, "name", namePrefix)

	log.Info("=== create Account ===")
	input := packInput(abiAccounts, "createAccount")
	txHash := sendContractTransaction(conn, cfg.from, AccountsAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	log.Info("=== setName name ===")
	input = packInput(abiAccounts, "setName", namePrefix)
	txHash = sendContractTransaction(conn, cfg.from, AccountsAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	log.Info("=== setAccountDataEncryptionKey ===")
	input = packInput(abiAccounts, "setAccountDataEncryptionKey", cfg.PublicKey)
	txHash = sendContractTransaction(conn, cfg.from, AccountsAddress, nil, priKey, input)
	getResult(conn, txHash, true)
}
