package main

import (
	"context"
	"fmt"
	ethchain "github.com/ethereum/go-ethereum"
	"gopkg.in/urfave/cli.v1"
)

var queryGroupsCommand = cli.Command{
	Name:   "queryGroups",
	Usage:  "query Groups",
	Action: MigrateFlags(queryGroups),
	Flags:  ValidatorFlags,
}

func queryGroups(ctx *cli.Context) error {
	conn, _ := dialConn(ctx)
	header, err := conn.HeaderByNumber(context.Background(), nil)
	input := packInput("getRegisteredValidatorGroups")
	msg := ethchain.CallMsg{From: from, To: &ValidatorAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		printError("method CallContract error", err)
	}
	fmt.Println(output)
	return nil
}
