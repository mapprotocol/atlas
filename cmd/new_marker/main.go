package main

import (
	"fmt"
	"github.com/mapprotocol/atlas/cmd/new_marker/cmd"
	"gopkg.in/urfave/cli.v1"
	"os"
	"sort"
)

var (
	// The app that holds all commands and flags.
	app *cli.App
)

func init() {
	app = cli.NewApp()
	app.Usage = "Atlas Marker Tool"
	app.Name = "marker"
	app.Version = "3.0.1"
	app.Copyright = "Copyright 2020-2021 The Atlas Authors"
	app.CommandNotFound = func(ctx *cli.Context, cmd string) {
		_, _ = fmt.Fprintf(os.Stderr, "No such command: %s\n", cmd)
		os.Exit(1)
	}
	app.Commands = append(app.Commands, cmd.AccountSet...)
	app.Commands = append(app.Commands, cmd.ValidatorSet...)
	app.Commands = append(app.Commands, cmd.VoterSet...)
	app.Commands = append(app.Commands, cmd.ToolSet...)
	sort.Sort(cli.CommandsByName(app.Commands))
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
