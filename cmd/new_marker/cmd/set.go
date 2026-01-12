package cmd

import (
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"

	"github.com/mapprotocol/atlas/cmd/new_marker/define"
)

var (
	ToolSet      []cli.Command
	AccountSet   = cli.Command{}
	ValidatorSet = cli.Command{}
	VoterSet     = cli.Command{}
	ViewerSet    = cli.Command{}
	OwnerSet     = cli.Command{}
	TssSet       = cli.Command{}
)

func init() {
	account := NewAccount()
	AccountSet = cli.Command{
		Name:  "account",
		Usage: "Commands related to account",
		Subcommands: []cli.Command{
			{
				Name:   "createAccount",
				Usage:  "Create an account and set the account name",
				Action: MigrateFlags(account.CreateAccount),
				Flags:  append([]cli.Flag{}, define.KeyStoreFlag, define.GasLimitFlag, define.NameFlag, define.TestnetFlag),
			},
			{
				Name:   "getAccountName",
				Usage:  "Get name of account",
				Action: MigrateFlags(account.GetAccountName),
				Flags:  define.BaseFlagCombination,
			},
			{
				Name:   "setAccountName",
				Usage:  "Set name of account",
				Action: MigrateFlags(account.SetAccountName),
				Flags:  append([]cli.Flag{}, define.KeyStoreFlag, define.GasLimitFlag, define.NameFlag, define.TestnetFlag),
			},
			{
				Name:   "getAccountMetadataURL",
				Usage:  "Get metadata url of account",
				Action: MigrateFlags(account.GetAccountMetadataURL),
				Flags:  define.BaseFlagCombination,
			},
			{
				Name:   "setAccountMetadataURL",
				Usage:  "Set metadata url of account",
				Action: MigrateFlags(account.SetAccountMetadataURL),
				Flags:  append(define.BaseFlagCombination, define.URLFlag),
			},
			{
				Name:   "getAccountTotalLockedGold",
				Usage:  "Returns the total amount of locked gold for an account",
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
				Usage:  "Returns the pending votes for validator made by account",
				Action: MigrateFlags(account.GetPendingVotesForValidatorByAccount),
				Flags:  define.BaseFlagCombination,
			},
			{
				Name:   "getActiveVotesForValidatorByAccount",
				Usage:  "Returns the active votes for validator made by account",
				Action: MigrateFlags(account.GetActiveVotesForValidatorByAccount),
				Flags:  define.BaseFlagCombination,
			},
			{
				Name:   "getValidatorsVotedForByAccount",
				Usage:  "Returns the validators that account has voted for",
				Action: MigrateFlags(account.GetValidatorsVotedForByAccount),
				Flags:  define.BaseFlagCombination,
			},
			{
				Name:   "signerToAccount",
				Usage:  "Returns the account associated with signer",
				Action: MigrateFlags(account.SignerToAccount),
				Flags:  define.BaseFlagCombination,
			},
		},
	}

	validator := NewValidator()
	ValidatorSet = cli.Command{
		Name:  "validator",
		Usage: "Commands related to validator",
		Subcommands: []cli.Command{
			{
				Name:   "register",
				Usage:  "Registers a validator",
				Action: MigrateFlags(validator.RegisterValidator),
				Flags:  append([]cli.Flag{}, define.RPCAddrFlag, define.KeyStoreFlag, define.CommissionFlag, define.PrivateKeyFlag, define.TestnetFlag),
			},
			{
				Name:   "quicklyRegister",
				Usage:  "Quickly register validator",
				Action: MigrateFlags(validator.QuicklyRegisterValidator),
				Flags:  append(define.MustFlagCombination, define.CommissionFlag, define.ValueFlag, define.NameFlag, define.PrivateKeyFlag),
			},
			{
				Name:   "generateSignerProof",
				Usage:  "Generate proof of signer",
				Action: MigrateFlags(validator.GenerateSignerProof),
				Flags:  append([]cli.Flag{}, define.KeyStoreFlag, define.AddressFlag, define.PrivateKeyFlag, define.TestnetFlag),
			},
			{
				Name:   "registerByProof",
				Usage:  "Registers a validator by signer proof",
				Action: MigrateFlags(validator.RegisterValidatorByProof),
				Flags:  append(define.MustFlagCombination, define.ProofFlag, define.CommissionFlag),
			},
			{
				Name:   "deregister",
				Usage:  "De-registers a validator",
				Action: MigrateFlags(validator.DeregisterValidator),
				Flags:  define.MustFlagCombination,
			},
			{
				Name:   "revertRegister",
				Usage:  "Restore your validator identity",
				Action: MigrateFlags(validator.RevertRegisterValidator),
				Flags:  define.MustFlagCombination,
			},
			{
				Name:   "authorizeValidatorSigner",
				Usage:  "Authorize a specific address to use a private key to sign consensus messages on behalf of an account",
				Action: MigrateFlags(validator.AuthorizeValidatorSigner),
				Flags:  append(define.MustFlagCombination, define.PrivateKeyFlag),
			},
			{
				Name:   "authorizeValidatorSignerBySignature",
				Usage:  "Authorize a specific address to use a signature to sign consensus messages on behalf of an account",
				Action: MigrateFlags(validator.AuthorizeValidatorSignerBySignature),
				Flags:  append(define.MustFlagCombination, define.SignatureFlag, define.AddressFlag),
			},
			{
				Name:   "makeECDSASignatureFromSigner",
				Usage:  "Print a ECDSASignature that signer sign the account(validator)",
				Action: MigrateFlags(validator.MakeECDSASignatureFromSigner),
				Flags:  append([]cli.Flag{}, define.KeyStoreFlag, define.PrivateKeyFlag, define.AddressFlag, define.TestnetFlag),
			},
			{
				Name:   "makeBLSProofOfPossessionFromSigner",
				Usage:  "Print a BLSProofOfPossession that signer BLSSign the account(validator)",
				Action: MigrateFlags(validator.MakeBLSProofOfPossessionFromsigner),
				Flags:  append([]cli.Flag{}, define.KeyStoreFlag, define.PrivateKeyFlag, define.AddressFlag, define.TestnetFlag),
			},
			{
				Name:   "updateBlsPublicKey",
				Usage:  "Updates a validator's BLS public key",
				Action: MigrateFlags(validator.updateBlsPublicKey),
				Flags:  define.MustFlagCombination,
			},
			{
				Name:   "setNextCommissionUpdate",
				Usage:  "Queues an update to a validator's commission.",
				Action: MigrateFlags(validator.setNextCommissionUpdate),
				Flags:  append(define.MustFlagCombination, define.CommissionFlag),
			},
			{
				Name:   "updateCommission",
				Usage:  "Updates a validator's commission based on the previously queued update",
				Action: MigrateFlags(validator.updateCommission),
				Flags:  append(define.MustFlagCombination, define.CommissionFlag),
			},
			{
				Name:   "getValidatorRewardInfo",
				Usage:  "Get validator reward information",
				Action: MigrateFlags(validator.GetRewardInfo),
				Flags:  define.MustFlagCombination,
			},
		},
	}

	voter := NewVoter()
	VoterSet = cli.Command{
		Name:  "voter",
		Usage: "Commands related to voter",
		Subcommands: []cli.Command{
			{
				Name:   "vote",
				Usage:  "Increments the number of total and pending votes for validator",
				Action: MigrateFlags(voter.Vote),
				Flags:  append(define.BaseFlagCombination, define.ValueFlag),
			},
			{
				Name:   "quicklyVote",
				Usage:  "Create an account, lock tokens, and increments the number of total and pending votes for validator",
				Action: MigrateFlags(voter.QuicklyVote),
				Flags:  append(define.BaseFlagCombination, define.NameFlag, define.ValueFlag, define.ValueFlag),
			},
			{
				Name:   "activate",
				Usage:  "Converts account's pending votes for validator to active votes",
				Action: MigrateFlags(voter.Activate),
				Flags:  define.BaseFlagCombination,
			},
			{
				Name:   "revokePending",
				Usage:  "Revokes value pending votes for validator",
				Action: MigrateFlags(voter.RevokePending),
				Flags:  append(define.BaseFlagCombination, define.ValueFlag),
			},
			{
				Name:   "revokeActive",
				Usage:  "Revokes value active votes for validator",
				Action: MigrateFlags(voter.RevokeActive),
				Flags:  append(define.BaseFlagCombination, define.ValueFlag),
			},
			{
				Name:   "locked",
				Usage:  "Locks MAP to be used for voting",
				Action: MigrateFlags(voter.LockedMAP),
				Flags:  append(define.MustFlagCombination, define.ValueFlag),
			},
			{
				Name:   "unlock",
				Usage:  "Unlock the locked MAP",
				Action: MigrateFlags(voter.UnlockedMAP),
				Flags:  append(define.MustFlagCombination, define.ValueFlag),
			},
			{
				Name:   "relock",
				Usage:  "Relocks MAP that has been unlocked but not withdrawn",
				Action: MigrateFlags(voter.RelockMAP),
				Flags:  append(define.MustFlagCombination, define.ValueFlag, define.IndexFlag),
			},
			{
				Name:   "withdraw",
				Usage:  "Withdraws MAP that has been unlocked after the unlocking period has passed",
				Action: MigrateFlags(voter.Withdraw),
				Flags:  append(define.MustFlagCombination, define.IndexFlag),
			},
			{
				Name:   "getTotalVotes",
				Usage:  "返回所有验证者所获得的总票数",
				Action: MigrateFlags(voter.getTotalVotes),
				//Flags:  define.MustFlagCombination,
				Flags: []cli.Flag{define.TestnetFlag},
			},
			{
				Name:   "getVoterRewardInfo",
				Usage:  "Get voter reward information",
				Action: MigrateFlags(voter.getVoterRewardInfo),
				Flags:  define.MustFlagCombination,
			},
		},
	}

	viewer := NewViewer()
	ViewerSet = cli.Command{
		Name:  "viewer",
		Usage: "Commands related to viewing public information",
		Subcommands: []cli.Command{
			{
				Name:   "getActiveVotesForValidator",
				Usage:  "Returns the total active vote units made for `validator`.",
				Action: MigrateFlags(viewer.GetActiveVotesForValidator),
				Flags:  define.BaseFlagCombination,
			},
			{
				Name:   "getPendingVotersForValidator",
				Usage:  "Returns the total pending voters vote for target `validator`.",
				Action: MigrateFlags(viewer.GetPendingVotersForValidator),
				Flags:  define.BaseFlagCombination,
			},
			{
				Name:   "getPendingInfoForValidator",
				Usage:  "Returns the pending Info voters vote And Epoch for target `validator`.",
				Action: MigrateFlags(viewer.GetPendingInfoForValidator),
				Flags:  define.BaseFlagCombination,
			},
			{
				Name:   "getTotalVotesForEligibleValidators",
				Usage:  "Returns lists of all validator validators and the number of votes they've received",
				Action: MigrateFlags(viewer.GetTotalVotesForEligibleValidators),
				Flags:  define.MustFlagCombination,
			},
			{
				Name:   "getRegisteredValidatorSigners",
				Usage:  "Returns the list of signers for the registered validator accounts",
				Action: MigrateFlags(viewer.GetRegisteredValidatorSigners),
				Flags:  define.MustFlagCombination,
			},
			{
				Name:   "getValidator",
				Usage:  "Returns validator information",
				Action: MigrateFlags(viewer.GetValidator),
				Flags:  define.BaseFlagCombination,
			},
			{
				Name:   "getNumRegisteredValidators",
				Usage:  "Get Num RegisteredValidators",
				Action: MigrateFlags(viewer.getNumRegisteredValidators),
				Flags:  define.MustFlagCombination,
			},
			{
				Name:   "getTopValidators",
				Usage:  "Get Top Validators",
				Action: MigrateFlags(viewer.getTopValidators),
				Flags:  append(define.MustFlagCombination, define.ValueFlag),
			},
			{
				Name:   "getValidatorEligibility",
				Usage:  "Judge whether the verifier`s Eligibility",
				Action: MigrateFlags(viewer.getValidatorEligibility),
				Flags:  define.BaseFlagCombination,
			},
			{
				Name:   "getPendingWithdrawals",
				Usage:  "Returns the pending withdrawals from unlocked gold for an account",
				Action: MigrateFlags(viewer.getPendingWithdrawals),
				Flags:  define.BaseFlagCombination,
			},
			{
				Name:   "balanceOf",
				Usage:  "Gets the balance of the specified address",
				Action: MigrateFlags(viewer.balanceOf),
				Flags:  define.BaseFlagCombination,
			},
		},
	}

	owner := NewOwner()
	OwnerSet = cli.Command{
		Name:  "owner",
		Usage: "Commands related to owner",
		Subcommands: []cli.Command{
			{
				Name:   "setValidatorLockedGoldRequirements",
				Usage:  "Updates the Locked Gold requirements for Validators.",
				Action: MigrateFlags(owner.setValidatorLockedGoldRequirements),
				Flags:  append(define.MustFlagCombination, define.DurationFlag, define.ValueFlag),
			},
			{
				Name:   "setImplementation",
				Usage:  "Sets the address of the implementation contract.",
				Action: MigrateFlags(owner.setImplementation),
				Flags:  append(define.MustFlagCombination, define.AddressFlag, define.ImplementationAddressFlag),
			},
			{
				Name:   "getContractOwner",
				Usage:  "Transfers ownership of the contract to a new account (`newOwner`).",
				Action: MigrateFlags(owner.getContractOwner),
				Flags:  append(define.MustFlagCombination, define.AddressFlag),
			},
			{
				Name:   "setContractOwner",
				Usage:  "Transfers ownership of the contract to a new account (`newOwner`).",
				Action: MigrateFlags(owner.setContractOwner),
				Flags:  append(define.BaseFlagCombination, define.AddressFlag),
			},
			{
				Name:   "getProxyContractOwner",
				Usage:  "Transfers ownership of the contract to a new account (`newOwner`)",
				Action: MigrateFlags(owner.getProxyContractOwner),
				Flags:  append(define.MustFlagCombination, define.AddressFlag),
			},
			{
				Name:   "setProxyContractOwner",
				Usage:  "Transfers ownership of the contract to a new account (`newOwner`).",
				Action: MigrateFlags(owner.setProxyContractOwner),
				Flags:  append(define.MustFlagCombination, define.AddressFlag),
			},
			{
				Name:   "setEpochMaintainerPaymentFraction",
				Usage:  "Set Epoch Maintainer PaymentFraction",
				Action: MigrateFlags(owner.setEpochMaintainerPaymentFraction),
				Flags:  append(define.MustFlagCombination, define.RelayerFlag),
			},
			{
				Name:   "getMgrMaintainerAddress",
				Usage:  "Set manager maintainer address",
				Action: MigrateFlags(owner.getMgrMaintainerAddress),
				Flags:  define.MustFlagCombination,
			},
			{
				Name:   "setMgrMaintainerAddress",
				Usage:  "Set manager maintainer address",
				Action: MigrateFlags(owner.setMgrMaintainerAddress),
				Flags:  define.BaseFlagCombination,
			},
			{
				Name:   "setValidatorEpochPayment",
				Usage:  "Sets the target per-epoch payment in MAP  for validators",
				Action: MigrateFlags(owner.setTargetValidatorEpochPayment),
				Flags:  append(define.MustFlagCombination, define.ValueFlag),
			},
			{
				Name:   "setEpochMaintainerPaymentFraction",
				Usage:  "Set Epoch Maintainer PaymentFraction",
				Action: MigrateFlags(owner.setEpochMaintainerPaymentFraction),
				Flags:  append(define.MustFlagCombination, define.RelayerFlag),
			},
		},
	}

	tool := NewTool()
	ToolSet = append(ToolSet, []cli.Command{
		{
			Name:      "genesis",
			Usage:     "Creates genesis.json from a template and overrides",
			Action:    tool.createGenesis,
			ArgsUsage: "",
			Flags: append(
				[]cli.Flag{
					define.BuildpathFlag,
					define.NewEnvFlag,
					define.MarkerCfgFlag,
					define.TestnetFlag,
				},
				define.TemplateFlags...),
		},
		{
			Name:   "transfer",
			Usage:  "Transfer",
			Action: MigrateFlags(tool.transfer),
			Flags:  append(define.MustFlagCombination, define.AmountFlag, define.AddressFlag),
		},
		//{
		//	Name:   "voterMonitor",
		//	Usage:  "Monitor the revenue of voter to a validator",
		//	Action: MigrateFlags(tool.voterMonitor),
		//	Flags:  define.MustFlagCombination,
		//},
	}...)

	tss := NewTss()
	TssSet = cli.Command{
		Name:  "maintainer",
		Usage: "Commands related to maintainer",
		Subcommands: []cli.Command{
			{
				Name:      "register",
				Usage:     "Register tss maintainers",
				Action:    MigrateFlags(tss.Register),
				ArgsUsage: "",
				Flags: append([]cli.Flag{
					define.KeyStoreFlag,
					define.TssRegisterPkFlag,
					define.P2pAddressFlag,
					define.TestnetFlag,
				}),
			},
			{
				Name:      "update",
				Usage:     "Update tss maintainers",
				Action:    MigrateFlags(tss.Update),
				ArgsUsage: "",
				Flags: append([]cli.Flag{
					define.KeyStoreFlag,
					define.TssRegisterPkFlag,
					define.P2pAddressFlag,
					define.TestnetFlag,
				}),
			},
			{
				Name:      "active",
				Usage:     "Active tss maintainers",
				Action:    MigrateFlags(tss.Activate),
				ArgsUsage: "",
				Flags: append([]cli.Flag{
					define.KeyStoreFlag,
					define.TestnetFlag,
				}),
			},
			{
				Name:      "revoke",
				Usage:     "Revoke tss maintainers",
				Action:    MigrateFlags(tss.Revoke),
				ArgsUsage: "",
				Flags: append([]cli.Flag{
					define.KeyStoreFlag,
					define.TestnetFlag,
				}),
			},
		},
	}
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
