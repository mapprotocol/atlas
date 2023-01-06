package cmd

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/cmd/new_marker/define"
	"gopkg.in/urfave/cli.v1"
	"os"
	"strconv"
)

var (
	AccountSet   []cli.Command
	ValidatorSet []cli.Command
)

func init() {
	account := NewAccount()
	AccountSet = append(AccountSet, []cli.Command{
		{
			Name:   "getAccountMetadataURL",
			Usage:  "get metadata url of account",
			Action: MigrateFlags(account.GetAccountMetadataURL),
			Flags:  define.BaseFlagCombination,
		},
		{
			Name:   "getAccountName",
			Usage:  "get name of account",
			Action: MigrateFlags(account.GetAccountName),
			Flags:  define.BaseFlagCombination,
		},
		{
			Name:   "getAccountTotalLockedGold",
			Usage:  "Returns the total amount of locked gold for an account.",
			Action: MigrateFlags(account.GetAccountTotalLockedGold),
			Flags:  define.BaseFlagCombination,
		},
		{
			Name:   "getAccountNonvotingLockedGold",
			Usage:  "Returns the total amount of non-voting locked gold for an account",
			Action: MigrateFlags(account.GetAccountNonvotingLockedGold),
			Flags:  define.BaseFlagCombination,
		},
		{
			Name:   "getPendingVotesForValidatorByAccount",
			Usage:  "Returns the pending votes for `validator` made by `account`",
			Action: MigrateFlags(account.GetPendingVotesForValidatorByAccount),
			Flags:  define.BaseFlagCombination,
		},
		{
			Name:   "getActiveVotesForValidatorByAccount",
			Usage:  "Returns the active votes for `validator` made by `account`",
			Action: MigrateFlags(account.GetActiveVotesForValidatorByAccount),
			Flags:  define.BaseFlagCombination,
		},
		{
			Name:   "getValidatorsVotedForByAccount",
			Usage:  "Returns the validators that `account` has voted for.",
			Action: MigrateFlags(account.GetValidatorsVotedForByAccount),
			Flags:  define.BaseFlagCombination,
		},
		{
			Name:   "setAccountMetadataURL",
			Usage:  "set metadata url of account",
			Action: MigrateFlags(account.SetAccountMetadataURL),
			Flags:  append(define.BaseFlagCombination, define.URLFlag),
		},
		{
			Name:   "setAccountName",
			Usage:  "set name of account",
			Action: MigrateFlags(account.SetAccountName),
			Flags:  append([]cli.Flag{}, define.RPCAddrFlag, define.KeyStoreFlag, define.GasLimitFlag, define.NameFlag),
		},
		{
			Name:   "createAccount",
			Usage:  "creat validator account",
			Action: MigrateFlags(account.CreateAccount),
			Flags:  append([]cli.Flag{}, define.RPCAddrFlag, define.KeyStoreFlag, define.GasLimitFlag, define.NameFlag),
		},
		{
			Name:   "signerToAccount",
			Usage:  "Returns the account associated with `signer`.",
			Action: MigrateFlags(account.SignerToAccount),
			Flags:  define.BaseFlagCombination,
		},
	}...)
	validator := NewValidator()
	ValidatorSet = append(ValidatorSet, []cli.Command{
		{
			Name:   "register",
			Usage:  "register validator",
			Action: MigrateFlags(validator.RegisterValidator),
			Flags:  append([]cli.Flag{}, define.RPCAddrFlag, define.KeyStoreFlag, define.CommissionFlag, define.SignerPriFlag),
		},
		{
			Name:   "registerByProof",
			Usage:  "register validator by signer proof",
			Action: MigrateFlags(validator.RegisterValidatorByProof),
			Flags:  append(define.MustFlagCombination, define.ProofFlag, define.CommissionFlag),
		},
	}...)
}

func MigrateFlags(hdl func(ctx *cli.Context, cfg *define.Config) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		for _, name := range ctx.FlagNames() {
			if ctx.IsSet(name) {
				err := ctx.Set(name, ctx.String(name))
				if err != nil {
					log.Error("MigrateFlags", "=== err ===", err, ctx.IsSet(name))
				}
			}
		}
		_config, err := define.AssemblyConfig(ctx)
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

func startLogger(_ *cli.Context, config *define.Config) error {
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
