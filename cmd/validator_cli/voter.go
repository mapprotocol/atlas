package main

import (
	"fmt"
	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"context"
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
var getBalanceCommand = cli.Command{
	Name:   "balanceOf",
	Usage:  "Gets the balance of the specified address using the presently stored inflation factor.",
	Action: MigrateFlags(balanceOf),
	Flags:  ValidatorFlags,
}
var activateCommand = cli.Command{
	Name:   "activate",
	Usage:  "Converts `account`'s pending votes for `validator` to active votes.",
	Action: MigrateFlags(activate),
	Flags:  ValidatorFlags,
}

var voterCommand = cli.Command{
	Name:  "voter",
	Usage: "voter commands",
	Subcommands: []cli.Command{
		voteValidatorCommand,
		getValidatorEligibilityCommand,
		getTotalVotesForVCommand,
		getBalanceCommand,
		activateCommand,
		queryRegisteredValidatorSignersCommand,
		queryTopValidatorsCommand,
	},
}

func vote(ctx *cli.Context, config *Config) error {
	//----------------------------- vote Validator ---------------------------------
	log.Info("=== vote Validator ===")
	amount := new(big.Int).Mul(config.voteNum, big.NewInt(1e18))
	input := packInput(abiElection, "vote", config.from, amount, config.lesser, config.greater)
	txHash := sendContractTransaction(config.conn, config.from, ElectionAddress, nil, priKey, input)
	getResult(config.conn, txHash, true)
	return nil
}

func getTotalVotesForEligibleValidators(ctx *cli.Context, config *Config) error {
	header, err := config.conn.HeaderByNumber(context.Background(), nil)
	log.Info("=== getTotalVotesForEligibleValidators admin", "obj", from)
	methodName := "getTotalVotesForEligibleValidators"
	input := packInput(abiElection, methodName)
	msg := ethchain.CallMsg{From: from, To: &ElectionAddress, Data: input}
	output, err := config.conn.CallContract(context.Background(), msg, header.Number)
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
func getValidatorEligibility(ctx *cli.Context, config *Config) error {
	methodName := "getValidatorEligibility"
	header, err := config.conn.HeaderByNumber(context.Background(), nil)
	log.Info("=== getValidatorEligibility admin", "obj", from)
	input := packInput(abiElection, methodName, config.targetAddress)
	msg := ethchain.CallMsg{From: from, To: &ElectionAddress, Data: input}
	output, err := config.conn.CallContract(context.Background(), msg, header.Number)
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

//StableToken
func balanceOf(ctx *cli.Context, config *Config) error {
	header, err := config.conn.HeaderByNumber(context.Background(), nil)
	methodName := "balanceOf"
	log.Info("=== balanceOf admin", "obj", from)
	input := packInput(abiGoldToken, methodName, config.targetAddress)
	msg := ethchain.CallMsg{From: from, To: &GoldTokenAddress, Data: input}
	output, err := config.conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "error", err)
	}
	var ret *big.Int
	err = abiGoldToken.UnpackIntoInterface(&ret, methodName, output)
	if err != nil {
		log.Error("method UnpackIntoInterface", "error", err)
		return nil
	}
	fmt.Println(ret)
	return nil
}

func activate(ctx *cli.Context, config *Config) error {
	log.Info("=== activate validator gold ===", "account.Address", config.from)
	input := packInput(abiElection, "activate", config.targetAddress)
	txHash := sendContractTransaction(config.conn, config.from, ElectionAddress, nil, priKey, input)
	getResult(config.conn, txHash, true)
	return nil
}

func getPendingVotesForValidatorByAccount(ctx *cli.Context, config *Config) error {
	header, _ := config.conn.HeaderByNumber(context.Background(), nil)
	log.Info("=== getPendingVotesForValidatorByAccount ===", "account.Address", config.from)
	methodName := "getPendingVotesForValidatorByAccount"
	input := packInput(abiElection, "getPendingVotesForValidatorByAccount", config.targetAddress, config.from)
	msg := ethchain.CallMsg{From: from, To: &ElectionAddress, Data: input}
	output, err := config.conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "error", err)
	}
	var ret big.Int
	err = abiElection.UnpackIntoInterface(&ret, methodName, output)
	fmt.Println(ret)
	return nil
}
func getValidatorsVotedForByAccount(ctx *cli.Context, config *Config) error {
	header, _ := config.conn.HeaderByNumber(context.Background(), nil)
	log.Info("=== getValidatorsVotedForByAccount ===", "account.Address", config.from)
	methodName := "getValidatorsVotedForByAccount"
	input := packInput(abiElection, "getValidatorsVotedForByAccount", config.targetAddress)
	msg := ethchain.CallMsg{From: from, To: &ElectionAddress, Data: input}
	output, err := config.conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "error", err)
	}
	var ret interface{}
	err = abiElection.UnpackIntoInterface(&ret, methodName, output)
	fmt.Println(ret.([]common.Address))
	return nil
}
