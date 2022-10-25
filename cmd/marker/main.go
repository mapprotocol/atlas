package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/log"

	"github.com/mapprotocol/atlas/cmd/marker/config"
	"github.com/mapprotocol/atlas/cmd/marker/genesis"
)

var (
	// The app that holds all commands and flags.
	app   *cli.App
	Flags = []cli.Flag{
		config.KeyFlag,
		config.KeyStoreFlag,
		config.RPCListenAddrFlag,
		config.ValueFlag,
		config.AmountFlag,
		config.DurationFlag,
		config.CommissionFlag,
		config.RelayerfFlag,
		config.NameFlag,
		config.URLFlag,
		config.VoteNumFlag,
		config.TopNumFlag,
		config.LockedNumFlag,
		config.WithdrawIndexFlag,
		config.RelockIndexFlag,
		config.TargetAddressFlag,
		config.ValidatorAddressFlag,
		config.AccountAddressFlag,
		config.SignerPrivFlag,
		config.SignerFlag,
		config.SignatureFlag,
		config.ProofFlag,
		config.ContractAddressFlag,
		config.MAPValueFlag,
		config.GasLimitFlag,
		config.ImplementationAddressFlag,
	}
)

func init() {
	app = cli.NewApp()
	app.Usage = "Atlas Marker Tool"
	app.Name = "marker"
	app.Version = "1.2.1"
	app.Copyright = "Copyright 2020-2021 The Atlas Authors"
	app.Action = MigrateFlags(registerValidator)
	app.CommandNotFound = func(ctx *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "No such command: %s\n", cmd)
		os.Exit(1)
	}
	// Add subcommands.
	app.Commands = []cli.Command{
		//------ validator -----
		registerValidatorCommand,
		generateSignerProofCommand,
		registerValidatorByProofCommand,
		revertRegisterValidatorCommand,
		quicklyRegisterValidatorCommand,
		authorizeValidatorSignerCommand,
		authorizeValidatorSignerSignatureCommand,
		signerToAccountCommand,
		makeECDSASignatureFromSignerCommand,
		makeBLSProofOfPossessionFromSignerCommand,
		deregisterValidatorCommand,
		//------ voter -----
		voteValidatorCommand,
		quicklyVoteValidatorCommand,
		activateCommand,
		getPendingVotesForValidatorByAccountCommand,
		getActiveVotesForValidatorByAccountCommand,
		getActiveVotesForValidatorCommand,
		getPendingVotersForValidatorCommand,
		getPendingInfoForValidatorCommand,

		revokePendingCommand,
		revokeActiveCommand,
		createAccountCommand,
		lockedMAPCommand,
		unlockedMAPCommand,
		relockMAPCommand,
		withdrawCommand,
		queryTotalVotesForEligibleValidatorsCommand,
		queryRegisteredValidatorSignersCommand,
		getValidatorCommand,
		getRewardInfoCommand,
		getVoterRewardInfoCommand,
		queryNumRegisteredValidatorsCommand,
		queryTopValidatorsCommand,
		queryValidatorEligibilityCommand,
		getBalanceCommand,
		getValidatorsVotedForByAccountCommand,
		getTotalVotesCommand,
		getAccountTotalLockedGoldCommand,
		getAccountNonvotingLockedGoldCommand,
		getAccountLockedGoldRequirementCommand,
		getPendingWithdrawalsCommand,
		setValidatorLockedGoldRequirementsCommand,
		setImplementationCommand,
		setOwnerCommand,
		setProxyContractOwnerCommand,
		getProxyContractOwnerCommand,
		getContractOwnerCommand,
		updateBlsPublicKeyCommand,
		setNextCommissionUpdateCommand,
		updateCommissionCommand,
		setTargetValidatorEpochPaymentCommand,
		setEpochMaintainerPaymentFractionCommand,
		setMgrMaintainerAddressCommand,
		getMgrMaintainerAddressCommand,
		transferCommand,
		getAccountMetadataURLCommand,
		setAccountMetadataURLCommand,
		getAccountNameCommand,
		setAccountNameCommand,
		//---------- CreateGenesis --------
		genesis.CreateGenesisCommand,

		//---------------------------------
		voterMonitorCommand,
	}
	app.Flags = Flags
	cli.CommandHelpTemplate = OriginCommandHelpTemplate
	sort.Sort(cli.CommandsByName(app.Commands))
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var OriginCommandHelpTemplate string = `{{.Name}}{{if .Subcommands}} command{{end}}{{if .Flags}} [command options]{{end}} [arguments...] {{if .Description}}{{.Description}} {{end}}{{if .Subcommands}} SUBCOMMANDS:     {{range .Subcommands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}     {{end}}{{end}}{{if .Flags}} {{"\n"}}OPTIONS: {{range $.Flags}}{{"\n\t"}}{{.}} {{end}} {{end}}`

func MigrateFlags(hdl func(ctx *cli.Context, config *listener) error) func(*cli.Context) error {
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
		core := NewListener(ctx, _config)
		writer := NewWriter(ctx, _config)
		core.setWriter(writer)
		return hdl(ctx, core)
	}
}

func MigrateFlagsOfLocalCommand(hdl func(ctx *cli.Context, config *listener) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		for _, name := range ctx.FlagNames() {
			if ctx.IsSet(name) {
				err := ctx.Set(name, ctx.String(name))
				if err != nil {
					log.Error("MigrateFlags", "=== err ===", err, ctx.IsSet(name))
				}
			}
		}
		cfg, err := config.AssemblyConfig(ctx)
		if err != nil {
			cli.ShowAppHelpAndExit(ctx, 1)
			panic(err)
		}
		err = startLogger(ctx, cfg)
		if err != nil {
			cli.ShowAppHelpAndExit(ctx, 1)
			panic(err)
		}
		core := NewListenerNotConn(cfg)
		writer := NewWriterNotConn(cfg)
		core.setWriter(writer)
		return hdl(ctx, core)
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
