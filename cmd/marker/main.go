package main

import (
	"fmt"
	"github.com/mapprotocol/atlas/helper/flags"
	"os"
	"path/filepath"

	"github.com/mapprotocol/atlas/helper/debug"
	"github.com/mapprotocol/atlas/marker/env"
	"github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
)

var (
	// Git information set by linker when building with ci.go.
	gitCommit string
	gitDate   string
	app       = &cli.App{
		Name:                 filepath.Base(os.Args[0]),
		Usage:                "marker",
		Version:              params.VersionWithCommit(gitCommit, gitDate),
		Writer:               os.Stdout,
		HideVersion:          true,
		EnableBashCompletion: true,
	}
)

func init() {
	// Set up the CLI app.
	app.Flags = append(app.Flags, debug.Flags...)
	app.Before = func(ctx *cli.Context) error {
		return debug.Setup(ctx)
	}
	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		return nil
	}
	app.CommandNotFound = func(ctx *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "No such command: %s\n", cmd)
		os.Exit(1)
	}
	// Add subcommands.
	app.Commands = []cli.Command{
		createGenesisCommand,
	}
	cli.CommandHelpTemplate = flags.OriginCommandHelpTemplate
}

func main() {
	err := app.Run(os.Args)
	if err == nil {
		os.Exit(0)
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func readWorkdir(ctx *cli.Context) (string, error) {
	if ctx.NArg() != 1 {
		fmt.Println("Using current directory as workdir")
		dir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return dir, err
	}
	return ctx.Args().Get(0), nil
}

func readEnv(ctx *cli.Context) (*env.Environment, error) {
	workdir, err := readWorkdir(ctx)
	if err != nil {
		return nil, err
	}
	return env.Load(workdir)
}
