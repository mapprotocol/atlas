package cmd

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/cmd/new_marker/config"
	"gopkg.in/urfave/cli.v1"
	"os"
	"strconv"
)

var Set []cli.Command

func init() {
	account := NewAccount()
	Set = append(Set, []cli.Command{
		{
			Name:   "getAccountMetadataURL",
			Usage:  "get metadata url of account",
			Action: Warp(account.GetAccountMetadataURL),
			Flags:  config.BaseFlagCombination,
		},
		{
			Name:   "getAccountName",
			Usage:  "get name of account",
			Action: Warp(account.GetAccountName),
			Flags:  config.BaseFlagCombination,
		},
		{
			Name:   "getAccountTotalLockedGold",
			Usage:  "Returns the total amount of locked gold for an account.",
			Action: Warp(account.GetAccountTotalLockedGold),
			Flags:  config.BaseFlagCombination,
		},
		{
			Name:   "getAccountNonvotingLockedGold",
			Usage:  "Returns the total amount of non-voting locked gold for an account",
			Action: Warp(account.GetAccountNonvotingLockedGold),
			Flags:  config.BaseFlagCombination,
		},
	}...)
}

func Warp(hdl func(ctx *cli.Context, cfg *config.Config) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		for _, name := range ctx.FlagNames() {
			if ctx.IsSet(name) {
				err := ctx.Set(name, ctx.String(name))
				if err != nil {
					log.Error("MigrateFlags", "=== err ===", err, ctx.IsSet(name))
				}
			}
		}
		_config, err := config.AssemblyConfig(ctx)
		if err != nil {
			cli.ShowAppHelpAndExit(ctx, 1)
			panic(err)
		}
		err = startLogger(ctx, _config)
		if err != nil {
			cli.ShowAppHelpAndExit(ctx, 1)
			panic(err)
		}
		return hdl(ctx, _config)
	}
}

func startLogger(_ *cli.Context, config *config.Config) error {
	logger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	var lvl log.Lvl
	if lvlToInt, err := strconv.Atoi(config.Verbosity); err == nil {
		lvl = log.Lvl(lvlToInt)
	} else if lvl, err = log.LvlFromString(config.Verbosity); err != nil {
		return err
	}
	logger.Verbosity(lvl)
	log.Root().SetHandler(log.LvlFilterHandler(lvl, logger))
	return nil
}
