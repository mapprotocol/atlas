package define

import (
	"gopkg.in/urfave/cli.v1"
)

var (
	KeyStoreFlag = cli.StringFlag{
		Name:  "keystore",
		Usage: "Keystore file path",
	}
	NameFlag = cli.StringFlag{
		Name:  "name",
		Usage: "Name of account",
	}
	URLFlag = cli.StringFlag{
		Name:  "url",
		Usage: "Metadata url of account",
	}
	CommissionFlag = cli.Uint64Flag{
		Name:  "commission",
		Usage: "Register validator param",
	}
	RelayerFlag = cli.StringFlag{
		Name:  "relayerf",
		Usage: "Register validator param",
	}
	VoteNumFlag = cli.Int64Flag{
		Name:  "voteNum",
		Usage: "The amount of gold to use to vote",
	}
	TopNumFlag = cli.Int64Flag{
		Name:  "topNum",
		Usage: "TopNum of validator",
	}
	LockedNumFlag = cli.Int64Flag{
		Name:  "lockedNum",
		Usage: "The amount of map to lock 、unlock、relock、withdraw ",
	}
	WithdrawIndexFlag = cli.Int64Flag{
		Name:  "withdrawIndex",
		Usage: "Use for withdraw",
	}
	ReLockIndexFlag = cli.Int64Flag{
		Name:  "relockIndex",
		Usage: "Use for relock",
	}
	VerbosityFlag = cli.Int64Flag{
		Name:  "Verbosity",
		Usage: "Verbosity of log level",
	}
	RPCAddrFlag = cli.StringFlag{
		Name:  "rpcaddr",
		Usage: "HTTP-RPC server listening interface",
		Value: "localhost",
	}
	ValueFlag = cli.Uint64Flag{
		Name:  "value",
		Usage: "Value units one eth",
		Value: 0,
	}
	AmountFlag = cli.StringFlag{
		Name:  "amount",
		Usage: "Transfer amount, unit (wei)",
		Value: "0",
	}
	DurationFlag = cli.Int64Flag{
		Name:  "duration",
		Usage: "Duration The time (in seconds) that these requirements persist for.",
		Value: 0,
	}
	TargetAddressFlag = cli.StringFlag{
		Name:  "target",
		Usage: "Target query address",
		Value: "",
	}
	ValidatorAddressFlag = cli.StringFlag{
		Name:  "validator",
		Usage: "Validator address",
		Value: "",
	}
	SignerPriFlag = cli.StringFlag{
		Name:  "signerPriv",
		Usage: "Signer private",
		Value: "",
	}
	SignerFlag = cli.StringFlag{
		Name:  "signer",
		Usage: "Signer address",
		Value: "",
	}
	SignatureFlag = cli.StringFlag{
		Name:  "signature",
		Usage: "ECDSA Signature",
		Value: "",
	}
	ProofFlag = cli.StringFlag{
		Name:  "proof",
		Usage: "Signer proof",
		Value: "",
	}
	AccountAddressFlag = cli.StringFlag{
		Name:  "accountAddress",
		Usage: "Account address",
		Value: "",
	}
	ContractAddressFlag = cli.StringFlag{
		Name:  "contractAddress",
		Usage: "Set contract Address",
		Value: "",
	}
	ImplementationAddressFlag = cli.StringFlag{
		Name:  "implementationAddress",
		Usage: "Set implementation Address",
		Value: "",
	}
	GasLimitFlag = cli.Int64Flag{
		Name:  "gasLimit",
		Usage: "Use for sendContractTransaction gasLimit",
		Value: 0,
	}
	KeystoreAddressFlag = cli.StringFlag{
		Name:  "keystoreAddress",
		Usage: "The address corresponding to the keystore",
		Value: "",
	}
	BuildpathFlag = cli.StringFlag{
		Name:  "buildpath",
		Usage: "Directory where smartcontract truffle build file live",
	}
	NewEnvFlag = cli.StringFlag{
		Name:  "newenv",
		Usage: "Creates a new env in desired folder",
	}
	MarkerCfgFlag = cli.StringFlag{
		Name:  "markercfg",
		Usage: "Marker config path",
	}
)

var TemplateFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "template",
		Usage: "Optional template to use (default: local)",
	},
	cli.IntFlag{
		Name:  "validators",
		Usage: "Number of Validators",
	},
	cli.IntFlag{
		Name:  "dev.accounts",
		Usage: "Number of developer accounts",
	},
	cli.Uint64Flag{
		Name:  "blockperiod",
		Usage: "Seconds between each block",
	},
	cli.Uint64Flag{
		Name:  "epoch",
		Usage: "Epoch size",
	},
	cli.Int64Flag{
		Name:  "blockgaslimit",
		Usage: "Block gas limit",
	},
	cli.StringFlag{
		Name:  "mnemonic",
		Usage: "Mnemonic to generate accounts",
	},
}

var BaseFlagCombination = []cli.Flag{
	RPCAddrFlag,
	KeyStoreFlag,
	GasLimitFlag,
	TargetAddressFlag,
	KeystoreAddressFlag,
}

var MustFlagCombination = []cli.Flag{
	RPCAddrFlag,
	KeyStoreFlag,
	GasLimitFlag,
}
