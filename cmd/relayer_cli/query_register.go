package main

import (
	"gopkg.in/urfave/cli.v1"
	"math/big"
)

var AppendCommand = cli.Command{
	Name:   "append",
	Usage:  "Append validator registered fund ",
	Action: MigrateFlags(Append),
	Flags:  RegisterFlags,
}

func Append(ctx *cli.Context) error {
	loadPrivate(ctx)

	conn, url := dialConn(ctx)
	printBaseInfo(conn, url)

	value := ethToWei(ctx, false)

	input := packInput("append", from, value)
	txHash := sendContractTransaction(conn, from, RelayerAddress, nil, priKey, input)

	getResult(conn, txHash, true)

	return nil
}

var withdrawCommand = cli.Command{
	Name:   "withdraw",
	Usage:  "Call this will instant receive your registered fund ",
	Action: MigrateFlags(withdraw),
	Flags:  RegisterFlags,
}

func withdraw(ctx *cli.Context) error {
	loadPrivate(ctx)
	conn, url := dialConn(ctx)
	printBaseInfo(conn, url)
	PrintBalance(conn, from)

	value := ethToWei(ctx, false)

	input := packInput("withdraw", from, value)

	txHash := sendContractTransaction(conn, from, RelayerAddress, new(big.Int).SetInt64(0), priKey, input)

	getResult(conn, txHash, true)
	PrintBalance(conn, from)
	return nil
}

var queryRegisterCommand = cli.Command{
	Name:   "queryRelayer",
	Usage:  "Query relayer info, get transaction result",
	Action: MigrateFlags(queryRegister),
	Flags:  RegisterFlags,
}

func queryRegister(ctx *cli.Context) error {
	loadPrivate(ctx)
	conn, url := dialConn(ctx)
	printBaseInfo(conn, url)

	queryRegisterInfo(conn)
	return nil
}

var queryBalanceCommand = cli.Command{
	Name:   "queryBalance",
	Usage:  "Query reward info, contain reward,fine,unlocked,locked and registered",
	Action: MigrateFlags(queryBalance),
	Flags:  RegisterFlags,
}

func queryBalance(ctx *cli.Context) error {
	loadPrivate(ctx)
	conn, url := dialConn(ctx)

	printBaseInfo(conn, url)
	queryAccountBalance(conn)
	return nil
}

var queryEpochCommand = cli.Command{
	Name:   "queryEpoch",
	Usage:  "Query Epoch, get transaction result",
	Action: MigrateFlags(queryEpoch),
	Flags:  RegisterFlags,
}

func queryEpoch(ctx *cli.Context) error {
	loadPrivate(ctx)
	conn, url := dialConn(ctx)

	printBaseInfo(conn, url)
	queryRelayerEpoch(conn)
	return nil
}
