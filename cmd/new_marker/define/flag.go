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
		Name:  "relayer",
		Usage: "Register validator param",
	}
	IndexFlag = cli.Int64Flag{
		Name:  "index",
		Usage: "Index number",
	}
	VerbosityFlag = cli.Int64Flag{
		Name:  "verbosity",
		Usage: "Verbosity of log level",
	}
	RPCAddrFlag = cli.StringFlag{
		Name:  "rpcaddr",
		Usage: "HTTP-RPC server listening interface",
		Value: "localhost",
	}
	ValueFlag = cli.Uint64Flag{
		Name:  "value",
		Usage: "A specific command to be executed requires a specified numerical value in eth.",
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
	AddressFlag = cli.StringFlag{
		Name:  "address",
		Usage: "Account address",
		Value: "",
	}
	PrivateKeyFlag = cli.StringFlag{
		Name:  "privatekey",
		Usage: "Hex string of private key",
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
	AccountFlag = cli.StringFlag{ // todo
		Name:  "account",
		Usage: "The address corresponding to the keystore",
		Value: "",
	}
	BuildpathFlag = cli.StringFlag{
		Name:  "buildpath",
		Usage: "Directory where smart contract truffle build file live",
	}
	NewEnvFlag = cli.StringFlag{
		Name:  "newenv",
		Usage: "Creates a new env in desired folder",
	}
	MarkerCfgFlag = cli.StringFlag{
		Name:  "markercfg",
		Usage: "Marker config path",
	}
	P2pAddressFlag = cli.StringFlag{
		Name:  "p2pAddress",
		Usage: "IP address used for P2P communication",
	}
	TssRegisterPkFlag = cli.StringFlag{
		Name:  "registerPk",
		Usage: "TSS register public key, Please use the uncompressed pk, prefixed with 04. default is use keystore",
	}
	TestnetFlag = cli.BoolFlag{
		Name:  "testnet",
		Usage: "use testnet network, default is mainnet",
	}
	PreMainFlag = cli.BoolFlag{
		Name:  "premain",
		Usage: "use pre main network, default is mainnet",
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
	AddressFlag,
	AccountFlag,
	TestnetFlag,
}

var MustFlagCombination = []cli.Flag{
	RPCAddrFlag,
	KeyStoreFlag,
	GasLimitFlag,
	TestnetFlag,
}
