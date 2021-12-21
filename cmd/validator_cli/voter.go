package main

import (
	"fmt"
	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"context"
	"github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
	"math/big"
)

var voteValidatorCommand = cli.Command{
	Name:   "voteValidator",
	Usage:  "vote validator ",
	Action: MigrateFlags(vote),
	Flags:  ValidatorFlags,
}

var getValidatorEligibilityCommand = cli.Command{
	Name:   "getValidatorEligibility",
	Usage:  "Judge whether the verifier`s Eligibility",
	Action: MigrateFlags(getValidatorEligibility),
	Flags:  ValidatorFlags,
}
var getTotalVotesForVCommand = cli.Command{
	Name:   "getTotalVotesForV",
	Usage:  "vote validator ",
	Action: MigrateFlags(getTotalVotesForEligibleValidators),
	Flags:  ValidatorFlags,
}

func vote(ctx *cli.Context) error {
	//------------------ pre set --------------------------
	path := ""
	password = "111111"
	voteNum := big.NewInt(int64(100))
	lesser := params.ZeroAddress
	greater := params.ZeroAddress
	validatorAddr := params.ZeroAddress
	//-----------------------------------------------------

	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(AddressFlag.Name) {
		validatorAddr = common.HexToAddress(ctx.GlobalString(AddressFlag.Name))
	}
	if ctx.IsSet(lesserFlag.Name) {
		lesser = common.HexToAddress(ctx.GlobalString(lesserFlag.Name))
	}
	if ctx.IsSet(greaterFlag.Name) {
		greater = common.HexToAddress(ctx.GlobalString(greaterFlag.Name))
	}
	if ctx.IsSet(voteNumFlag.Name) {
		voteNum = big.NewInt(ctx.Int64(voteNumFlag.Name))
	}
	validator := loadAccount(path, password)
	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	//----------------------------- vote Validator ---------------------------------
	log.Info("=== vote Validator ===")
	amount := new(big.Int).Mul(voteNum, big.NewInt(1e18))
	input := packInput(abiElection, "vote", validatorAddr, amount, lesser, greater)
	txHash := sendContractTransaction(conn, validator.Address, ElectionAddress, nil, priKey, input)
	getResult(conn, txHash, true)
	return nil
}

func getTotalVotesForEligibleValidators(ctx *cli.Context) error {
	//--------------------------- pre set -------------------------------------------
	path := ""
	password = "111111"
	//-------------------------------------------------------------------------------
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	methodName := "getTotalVotesForEligibleValidators"
	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	header, err := conn.HeaderByNumber(context.Background(), nil)

	log.Info("=== getTotalVotesForEligibleValidators admin", "obj", from)
	input := packInput(abiElection, methodName)
	msg := ethchain.CallMsg{From: from, To: &ElectionAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "error", err)
	}

	type TestStruct struct {
		Validators interface{} // indexed
		Values     interface{}
	}
	var t TestStruct
	err = abiElection.UnpackIntoInterface(&t, methodName, output)
	if err != nil {
		log.Error("method UnpackIntoInterface", "error", err)
		return nil
	}

	fmt.Println((t.Validators).([]common.Address))
	fmt.Println((t.Values).([]*big.Int))

	return nil
}

//getValidatorEligibility
func getValidatorEligibility(ctx *cli.Context) error {
	//--------------------------- pre set -------------------------------------------
	path := ""
	validatorAddr := params.ZeroAddress
	password = "111111"
	//-------------------------------------------------------------------------------

	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}
	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	if ctx.IsSet(AddressFlag.Name) {
		validatorAddr = common.HexToAddress(ctx.GlobalString(AddressFlag.Name))
	}
	methodName := "getValidatorEligibility"
	loadPrivateKey(path)
	conn, _ := dialConn(ctx)
	header, err := conn.HeaderByNumber(context.Background(), nil)

	log.Info("=== getValidatorEligibility admin", "obj", from)
	input := packInput(abiElection, methodName, validatorAddr)
	msg := ethchain.CallMsg{From: from, To: &ElectionAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "error", err)
	}

	var ret bool
	err = abiElection.UnpackIntoInterface(&ret, methodName, output)
	if err != nil {
		log.Error("method UnpackIntoInterface", "error", err)
		return nil
	}

	fmt.Println(ret)

	return nil
}
