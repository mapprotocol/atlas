package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"math/big"
)

var appendCommand = cli.Command{
	Name:   "append",
	Usage:  "Append validator registered fund ",
	Action: MigrateFlags(_append),
	Flags:  RegisterFlags,
}

func _append(ctx *cli.Context) error {
	loadPrivate(ctx)

	conn, url := dialConn(ctx)
	printBaseInfo(conn, url)

	value := ethToWei(ctx, false)

	input := packInput("append", value)
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

	value := ethToWei(ctx, false)

	input := packInput("withdraw", value)

	txHash := sendContractTransaction(conn, from, RelayerAddress, new(big.Int).SetInt64(0), priKey, input)

	getResult(conn, txHash, true)

	return nil
}

var unregisterCommand = cli.Command{
	Name:   "unregister",
	Usage:  "Call this will instant cancel your registered fund ",
	Action: MigrateFlags(unregister),
	Flags:  RegisterFlags,
}

func unregister(ctx *cli.Context) error {
	loadPrivate(ctx)
	conn, url := dialConn(ctx)
	printBaseInfo(conn, url)

	value := ethToWei(ctx, false)

	input := packInput("unregister", value)

	txHash := sendContractTransaction(conn, from, RelayerAddress, new(big.Int).SetInt64(0), priKey, input)

	getResult(conn, txHash, true)

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

var queryCommand = cli.Command{
	Name:   "query",
	Usage:  "Query Command include QueryRelayer Command, QueryBalance Command and QueryEpoch Command",
	Action: MigrateFlags(query),
	Flags:  RegisterFlags,
}

func query(ctx *cli.Context) error {
	loadPrivate(ctx)
	conn, url := dialConn(ctx)

	printBaseInfo(conn, url)
	fmt.Println()
	queryRegisterInfo(conn)
	fmt.Println()
	queryRelayerEpoch(conn)
	fmt.Println()
	queryAccountBalance(conn)
	fmt.Println()
	return nil
}
