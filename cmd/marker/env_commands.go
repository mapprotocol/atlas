package main

import (
	"fmt"
	"github.com/mapprotocol/atlas/cmd/utils"
	"github.com/mapprotocol/atlas/marker/env"
	"gopkg.in/urfave/cli.v1"
)



var (
	idxFlag = cli.IntFlag{
		Name:  "idx",
		Usage: "account index",
		Value: 0,
	}
	accountTypeFlag = utils.TextMarshalerFlag{
		Name:  "type",
		Usage: `Account type (validator, developer, txNode, faucet, attestation, priceOracle, proxy, attestationBot, votingBot, txNodePrivate, validatorGroup, admin)`,
		Value: &env.DeveloperAT,
	}
)

func getAccount(ctx *cli.Context) error {
	myatlasEnv, err := readEnv(ctx)
	if err != nil {
		return err
	}

	idx := ctx.Int(idxFlag.Name)
	accountType := *utils.LocalTextMarshaler(ctx, accountTypeFlag.Name).(*env.AccountType)

	account, err := myatlasEnv.Accounts().Account(accountType, idx)
	if err != nil {
		return err
	}

	fmt.Printf("AccountType: %s\nIndex:%d\nAddress: %s\nPrivateKey: %s\n", accountType, idx, account.Address.Hex(), account.PrivateKeyHex())
	return nil
}
