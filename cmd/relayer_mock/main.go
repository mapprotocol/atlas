package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/cmd/ethclient"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
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
	RPCListenAddrFlag = cli.StringFlag{
		Name:  "rpcaddr",
		Usage: "HTTP-RPC server listening interface",
		Value: "localhost",
	}
	RPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening port",
		Value: 7445,
	}

	FeeFlag = cli.Uint64Flag{
		Name:  "fee",
		Usage: "work fee",
		Value: 0,
	}

	PublicAdressFlag = cli.StringFlag{
		Name:  "PkAdress",
		Usage: "Relayer bft key for BFT (no 0x prefix)",
		Value: "",
	}

	relayerflags = []cli.Flag{
		KeyFlag,
		KeyStoreFlag,
		RPCListenAddrFlag,
		RPCPortFlag,
		FeeFlag,
	}

	withdrawMockCommand = cli.Command{
		Name:   "withdrawMock",
		Usage:  "relayer to withdraw ",
		Action: MigrateFlags(withdrawMock),
		Flags:  relayerflags,
	}
	appendMockCommand = cli.Command{
		Name:   "appendMock",
		Usage:  "relayer to append ",
		Action: MigrateFlags(appendMock),
		Flags:  relayerflags,
	}
	registerMockCommand = cli.Command{
		Name:   "registerMock",
		Usage:  "relayer to register ",
		Action: MigrateFlags(registerMock),
		Flags:  relayerflags,
	}
	saveMockCommand = cli.Command{
		Name:   "saveMock",
		Usage:  "relayer to save ",
		Action: MigrateFlags(saveMock),
		Flags:  relayerflags,
	}
)

func init() {
	app = cli.NewApp()
	app.Usage = "Atlas Register Tool"
	app.Name = filepath.Base(os.Args[0])
	app.Version = "1.0.0"
	app.Copyright = "Copyright 2020-2021 The Atlas Authors"
	app.Flags = []cli.Flag{
		KeyFlag,
		KeyStoreFlag,
		RPCListenAddrFlag,
		RPCPortFlag,
		FeeFlag,
		PublicAdressFlag,
	}
	app.Action = MigrateFlags(start)
	app.CommandNotFound = func(ctx *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "No such command: %s\n", cmd)
		os.Exit(1)
	}
	// Add subcommands.
	app.Commands = []cli.Command{
		withdrawMockCommand,
		appendMockCommand,
		registerMockCommand,
		saveMockCommand,
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

func start(ctx *cli.Context) error {
	conn := getConn11(ctx)
	from, priKey := register(ctx, conn)
	syncloop(ctx, conn, from, priKey)
	return nil
}

//single node start
func syncloop(ctx *cli.Context, conn *ethclient.Client, from common.Address, priKey *ecdsa.PrivateKey) {
	chains := getEthChains()
	for {
		for {
			time.Sleep(time.Second * 1)
			//1.now Number
			num, err := conn.BlockNumber(context.Background())
			if err != nil {
				log.Fatal("BlockNumber err")
			}
			//2. isrelayers
			isrelayers := queryIsRegister(conn, from)
			if !isrelayers {
				log.Fatal("not Relayers")
				time.Sleep(time.Second)
				continue
			}
			//3.judge number at range
			if !queryRelayerEpoch(conn, num, from) {
				log.Fatal("wrong range !")
				continue
			}
			break
		}
		// 1.get current num
		chainNum := getCurrentNumberAbi(conn, ChainTypeETH, from)
		// 2.store
		marshal, err2 := json.Marshal(chains[chainNum : chainNum+10])

		if err2 != nil {
			log.Fatal("marshal err")
		}
		//ret, _ := rlp.EncodeToBytes(chains)
		input := packInputStore("save", "ETH", "ETH", marshal)
		sendContractTransaction(conn, from, HeaderStoreAddress, nil, priKey, input)
		Append(conn, from, priKey)
		withdraw(conn, from, priKey)
	}
}
