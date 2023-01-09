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
	VoterSet     []cli.Command
	ToolSet      []cli.Command
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
			Name:   "generateSignerProof",
			Usage:  "generate proof of signer",
			Action: MigrateFlags(validator.GenerateSignerProof),
			Flags:  append([]cli.Flag{}, define.KeyStoreFlag, define.ValidatorAddressFlag, define.SignerPriFlag),
		},
		{
			Name:   "registerByProof",
			Usage:  "register validator by signer proof",
			Action: MigrateFlags(validator.RegisterValidatorByProof),
			Flags:  append(define.MustFlagCombination, define.ProofFlag, define.CommissionFlag),
		},
		{
			Name:   "revertRegister",
			Usage:  "register validator",
			Action: MigrateFlags(validator.RevertRegisterValidator),
			Flags:  define.MustFlagCombination,
		},
		{
			Name:   "deregister",
			Usage:  "deregister Validator",
			Action: MigrateFlags(validator.DeregisterValidator),
			Flags:  define.MustFlagCombination,
		},
		{
			Name:   "quicklyRegister",
			Usage:  "register validator",
			Action: MigrateFlags(validator.QuicklyRegisterValidator),
			Flags:  append(define.MustFlagCombination, define.CommissionFlag, define.LockedNumFlag, define.NameFlag),
		},
		{
			Name:   "authorizeValidatorSigner",
			Usage:  "Finish the process of authorizing an address to sign on behalf of the account.",
			Action: MigrateFlags(validator.AuthorizeValidatorSigner),
			Flags:  append(define.MustFlagCombination, define.SignerPriFlag),
		},
		{
			Name:   "authorizeValidatorSignerBySignature",
			Usage:  "Finish the process of authorizing an address to sign on behalf of the account.",
			Action: MigrateFlags(validator.AuthorizeValidatorSignerBySignature),
			Flags:  append(define.MustFlagCombination, define.SignatureFlag, define.SignerFlag),
		},
		{
			Name:   "makeECDSASignatureFromSigner",
			Usage:  "print a ECDSASignature that signer sign the account(validator)",
			Action: MigrateFlags(validator.MakeECDSASignatureFromSigner),
			Flags:  append([]cli.Flag{}, define.RPCAddrFlag, define.KeyStoreFlag, define.SignerPriFlag, define.TargetAddressFlag),
		},
		{
			Name:   "makeBLSProofOfPossessionFromSigner",
			Usage:  "print a BLSProofOfPossession that signer BLSSign the account(validator)",
			Action: MigrateFlags(validator.MakeBLSProofOfPossessionFromsigner),
			Flags:  append([]cli.Flag{}, define.RPCAddrFlag, define.KeyStoreFlag, define.SignerPriFlag, define.TargetAddressFlag),
		},
	}...)
	voter := NewVoter()
	VoterSet = append(VoterSet, []cli.Command{
		{
			Name:   "vote",
			Usage:  "vote validator ",
			Action: MigrateFlags(voter.Vote),
			Flags:  append(define.BaseFlagCombination, define.VoteNumFlag),
		},
		{
			Name:   "quicklyVote",
			Usage:  "vote validator ",
			Action: MigrateFlags(voter.QuicklyVote),
			Flags:  append(define.BaseFlagCombination, define.NameFlag, define.LockedNumFlag, define.TargetAddressFlag, define.VoteNumFlag),
		},
		{
			Name:   "activate",
			Usage:  "Converts `account`'s pending votes for `validator` to active votes.",
			Action: MigrateFlags(voter.Activate),
			Flags:  define.BaseFlagCombination,
		},
		{
			Name:   "getActiveVotesForValidator",
			Usage:  "Returns the total active vote units made for `validator`.",
			Action: MigrateFlags(voter.GetActiveVotesForValidator),
			Flags:  define.BaseFlagCombination,
		},
		{
			Name:   "getPendingVotersForValidator",
			Usage:  "Returns the total pending voters vote for target `validator`.",
			Action: MigrateFlags(voter.GetPendingVotersForValidator),
			Flags:  define.BaseFlagCombination,
		},
		{
			Name:   "getPendingInfoForValidator",
			Usage:  "Returns the  pending Info voters vote And Epoch for target `validator`.",
			Action: MigrateFlags(voter.GetPendingInfoForValidator),
			Flags:  define.BaseFlagCombination,
		},
		{
			Name:   "revokePending",
			Usage:  "Revokes `value` pending votes for `validator`",
			Action: MigrateFlags(voter.RevokePending),
			Flags:  append(define.BaseFlagCombination, define.LockedNumFlag),
		},
		{
			Name:   "revokeActive",
			Usage:  "Revokes `value` active votes for `validator`",
			Action: MigrateFlags(voter.RevokeActive),
			Flags:  append(define.BaseFlagCombination, define.LockedNumFlag),
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
