package main

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
	"math/big"
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
	abiValidators     = AbiFor("Validators")
	abiLocaledGold    = AbiFor("LockedGold")
	priKey            *ecdsa.PrivateKey
	from              common.Address
	Value             uint64
	fee               uint64
	ValidatorAddress  common.Address = MustProxyAddressFor("Validators")
	LockedGoldAddress common.Address = MustProxyAddressFor("LockedGold")
	Base                             = new(big.Int).SetUint64(10000)
)

func registerValidator(ctx *cli.Context) error {
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
	groupRequiredGold := params.MustBigInt("10000000000000000000000") // 10k Atlas per validator,
	log.Info("Lock group gold", "amount", groupRequiredGold)
	input := packInput("lock", groupRequiredGold)
	txHash := sendContractTransaction(conn, validator.Address, LockedGoldAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	log.Info("Register validator")
	input = packInput("registerValidator", pubKey, blsPub[:], validator.MustBLSProofOfPossession())
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
	input := packInput("lock", groupRequiredGold)
	txHash := sendContractTransaction(conn, validator.Address, LockedGoldAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	log.Info("Register validator")
	input = packInput("registerValidator", pubKey, blsPub[:], validator.MustBLSProofOfPossession())
	txHash = sendContractTransaction(conn, validator.Address, ValidatorAddress, nil, priKey, input)
	getResult(conn, txHash, true)

	return nil
}
