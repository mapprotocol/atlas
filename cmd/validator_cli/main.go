package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path/filepath"
	"sort"
)

var (
	// The app that holds all commands and flags.
	app *cli.App

	// Flags needed by abigen
	KeyFlag = cli.StringFlag{
		Name:  "key",
		Usage: "Private key file path",
		Value: "",
	}
	KeyStoreFlag = cli.StringFlag{
		Name:  "keystore",
		Usage: "Keystore file path",
	}
	PasswordFlag = cli.StringFlag{
		Name:  "password",
		Usage: "Keystore file`s password",
	}
	GroupAddressFlag = cli.StringFlag{
		Name:  "groupAddress",
		Usage: "group hex address",
	}
	NamePrefixFlag = cli.StringFlag{
		Name:  "namePrefix",
		Usage: "Keystore file`s password",
	}
	CommissionFlag = cli.Int64Flag{
		Name:  "commission",
		Usage: "register group param",
	}
	maxSizeFlag = cli.Int64Flag{
		Name:  "maxSize",
		Usage: "set the max group size",
	}
	TopNumFlag = cli.Int64Flag{
		Name:  "topNum",
		Usage: "topNum of group`s member",
	}
	ReadConfigFlag = cli.BoolFlag{
		Name:  "readConfig",
		Usage: "read Config to get validators",
	}
	RPCListenAddrFlag = cli.StringFlag{
		Name:  "rpcaddr",
		Usage: "HTTP-RPC server listening interface",
		Value: "localhost",
	}
	RPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening port",
		Value: 8545,
	}
	ValueFlag = cli.Uint64Flag{
		Name:  "value",
		Usage: "value units one eth",
		Value: 0,
	}
	FeeFlag = cli.Uint64Flag{
		Name:  "fee",
		Usage: "work fee",
		Value: 0,
	}
	AddressFlag = cli.StringFlag{
		Name:  "address",
		Usage: "Transfer address",
		Value: "",
	}
	BFTKeyKeyFlag = cli.StringFlag{
		Name:  "bftkey",
		Usage: "Validator bft key for BFT (no 0x prefix)",
		Value: "",
	}

	ValidatorFlags = []cli.Flag{
		KeyFlag,
		KeyStoreFlag,
		RPCListenAddrFlag,
		RPCPortFlag,
		ValueFlag,
		FeeFlag,
		BFTKeyKeyFlag,
		PasswordFlag,
		GroupAddressFlag,
		CommissionFlag,
		ReadConfigFlag,
		TopNumFlag,
		maxSizeFlag,
		AddressFlag,
	}
)

func init() {
	app = cli.NewApp()
	app.Usage = "Atlas Marker Tool"
	app.Name = filepath.Base(os.Args[0])
	app.Version = "1.0.0"
	app.Copyright = "Copyright 2020-2021 The Atlas Authors"
	app.Flags = []cli.Flag{
		KeyFlag,
		KeyStoreFlag,
		RPCListenAddrFlag,
		RPCPortFlag,
		ValueFlag,
		FeeFlag,
		BFTKeyKeyFlag,
		PasswordFlag,
		GroupAddressFlag,
		CommissionFlag,
		ReadConfigFlag,
		TopNumFlag,
		maxSizeFlag,
		AddressFlag,
	}
	app.Action = MigrateFlags(registerValidator)
	app.CommandNotFound = func(ctx *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "No such command: %s\n", cmd)
		os.Exit(1)
	}
	// Add subcommands.
	app.Commands = []cli.Command{
		registerGroupCommand,
		registerValidatorCommand,

		queryGroupsCommand,
		queryRegisteredValidatorSignersCommand,
		queryTopGroupValidatorsCommand,

		addFirstMemberCommand,
		addToGroupCommand,
		removeMemberCommand,
		deregisterValidatorGroupCommand,
		deregisterValidatorCommand,
		createAccountCommand,
		lockedMAPCommand,
		affiliateCommand,

		setMaxGroupSizeCommand,
	}
	cli.CommandHelpTemplate = OriginCommandHelpTemplate
	sort.Sort(cli.CommandsByName(app.Commands))
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var OriginCommandHelpTemplate string = `{{.Name}}{{if .Subcommands}} command{{end}}{{if .Flags}} [command options]{{end}} [arguments...] {{if .Description}}{{.Description}} {{end}}{{if .Subcommands}} SUBCOMMANDS:     {{range .Subcommands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}     {{end}}{{end}}{{if .Flags}} OPTIONS: {{range $.Flags}}{{"\t"}}{{.}} {{end}} {{end}}`

func MigrateFlags(action func(ctx *cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		for _, name := range ctx.FlagNames() {
			if ctx.IsSet(name) {
				ctx.GlobalSet(name, ctx.String(name))
			}
		}
		return action(ctx)
	}
}
