package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/urfave/cli.v1"
	"math/big"
)

var AppendCommand = cli.Command{
	Name:   "append",
	Usage:  "Append validator deposit staking count",
	Action: MigrateFlags(Append),
	Flags:  RegisterFlags,
}

func Append(ctx *cli.Context) error {
	loadPrivate(ctx)

	conn, url := dialConn(ctx)
	printBaseInfo(conn, url)

	value := ethToWei(ctx, false)

	input := packInput("append", value)
	txHash := sendContractTransaction(conn, from, RelayerAddress, nil, priKey, input)

	getResult(conn, txHash, true)

	return nil
}

var UpdatePKCommand = cli.Command{
	Name:   "updatepk",
	Usage:  "Update user pk will take effect in next epoch",
	Action: MigrateFlags(UpdatePKRegister),
	Flags:  RegisterFlags,
}

func UpdatePKRegister(ctx *cli.Context) error {
	loadPrivate(ctx)

	conn, url := dialConn(ctx)
	printBaseInfo(conn, url)

	pubkey, pk, _ := getPubKey(ctx)
	fmt.Println(" Pubkey ", pubkey)

	input := packInput("setPubkey", pk)
	txHash := sendContractTransaction(conn, from, RelayerAddress, new(big.Int).SetInt64(0), priKey, input)

	getResult(conn, txHash, true)
	return nil
}

var withdrawCommand = cli.Command{
	Name:   "withdraw",
	Usage:  "Call this will instant receive your deposit money",
	Action: MigrateFlags(withdraw),
	Flags:  RegisterFlags,
}

func withdraw(ctx *cli.Context) error {
	loadPrivate(ctx)
	conn, url := dialConn(ctx)
	printBaseInfo(conn, url)
	PrintBalance(conn, from)

	value := ethToWei(ctx, false)

	input := packInput("withdraw", value)

	txHash := sendContractTransaction(conn, from, RelayerAddress, new(big.Int).SetInt64(0), priKey, input)

	getResult(conn, txHash, true)
	PrintBalance(conn, from)
	return nil
}

var queryRegisterCommand = cli.Command{
	Name:   "queryrelayer",
	Usage:  "Query relayer info, can cancel info and can withdraw info",
	Action: MigrateFlags(queryRegister),
	Flags:  append(RegisterFlags, AddressFlag),
}

func queryRegister(ctx *cli.Context) error {
	loadPrivate(ctx)
	conn, url := dialConn(ctx)
	printBaseInfo(conn, url)

	queryRegisterInfo(conn, true)
	return nil
}

var queryBalanceCommand = cli.Command{
	Name:   "querybalance",
	Usage:  "Query reward info, contain deposit and delegate reward",
	Action: MigrateFlags(queryBalance),
	Flags:  append(RegisterFlags, AddressFlag),
}

func queryBalance(ctx *cli.Context) error {
	loadPrivate(ctx)
	conn, url := dialConn(ctx)

	printBaseInfo(conn, url)

	PrintBalance(conn, from)

	start := false
	snailNumber := uint64(0)
	if ctx.GlobalIsSet(NumberFlag.Name) {
		snailNumber = ctx.GlobalUint64(NumberFlag.Name)
		start = true
	}
	queryRewardInfo(conn, snailNumber, start)
	return nil
}

var sendCommand = cli.Command{
	Name:   "send",
	Usage:  "Send general transaction",
	Action: MigrateFlags(sendTX),
	Flags:  append(RegisterFlags, AddressFlag),
}

func sendTX(ctx *cli.Context) error {
	loadPrivate(ctx)
	conn, url := dialConn(ctx)
	printBaseInfo(conn, url)
	PrintBalance(conn, from)

	address := ctx.GlobalString(AddressFlag.Name)
	if !common.IsHexAddress(address) {
		printError("Must input correct address")
	}

	value := ethToWei(ctx, false)
	txHash := sendContractTransaction(conn, from, common.HexToAddress(address), value, priKey, nil)
	getResult(conn, txHash, false)
	return nil
}

var queryTxCommand = cli.Command{
	Name:   "querytx",
	Usage:  "Query tx hash, get transaction result",
	Action: MigrateFlags(queryTxRegister),
	Flags:  append(RegisterFlags, TxHashFlag),
}

func queryTxRegister(ctx *cli.Context) error {
	conn, url := dialConn(ctx)
	printBaseInfo(conn, url)

	txhash := ctx.GlobalString(TxHashFlag.Name)
	if txhash == "" {
		printError("Must input tx hash")
	}
	queryTx(conn, common.HexToHash(txhash), false, true)
	return nil
}

var queryEpochCommand = cli.Command{
	Name:   "queryEpoch",
	Usage:  "Query Epoch, get transaction result",
	Action: MigrateFlags(queryEpoch),
	Flags:  append(RegisterFlags, TxHashFlag),
}

func queryEpoch(ctx *cli.Context) error {
	conn, url := dialConn(ctx)
	printBaseInfo(conn, url)

	txhash := ctx.GlobalString(TxHashFlag.Name)
	if txhash == "" {
		printError("Must input tx hash")
	}
	queryTx(conn, common.HexToHash(txhash), false, true)
	return nil
}
