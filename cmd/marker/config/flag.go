package config

import (
	"gopkg.in/urfave/cli.v1"
)

var (
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
	PasswordFlag = cli.StringFlag{
		Name:  "password",
		Usage: "Keystore file`s password",
	}

	NamePrefixFlag = cli.StringFlag{
		Name:  "namePrefix",
		Usage: "namePrefix",
	}
	CommissionFlag = cli.StringFlag{
		Name:  "commission",
		Usage: "register validator param",
	}

	VoteNumFlag = cli.Int64Flag{
		Name:  "voteNum",
		Usage: "The amount of gold to use to vote",
	}
	TopNumFlag = cli.Int64Flag{
		Name:  "topNum",
		Usage: "topNum of validator",
	}
	LockedNumFlag = cli.Int64Flag{
		Name:  "lockedNum",
		Usage: "The amount of map to lock 、unlock、relock、withdraw ",
	}
	WithdrawIndexFlag = cli.Int64Flag{
		Name:  "withdrawIndex",
		Usage: "use for withdraw",
	}
	RelockIndexFlag = cli.Int64Flag{
		Name:  "relockIndex",
		Usage: "use for relock",
	}

	VerbosityFlag = cli.Int64Flag{
		Name:  "Verbosity",
		Usage: "Verbosity of log level",
	}

	RPCListenAddrFlag = cli.StringFlag{
		Name:  "rpcaddr",
		Usage: "HTTP-RPC server listening interface",
		Value: "localhost",
	}
	RPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening Port",
		Value: 8545,
	}
	ValueFlag = cli.Uint64Flag{
		Name:  "value",
		Usage: "value units one eth",
		Value: 0,
	}
	DurationFlag = cli.Int64Flag{
		Name:  "duration",
		Usage: "duration The time (in seconds) that these requirements persist for.",
		Value: 0,
	}
	TargetAddressFlag = cli.StringFlag{
		Name:  "target",
		Usage: "Transfer address",
		Value: "",
	}
	GasLimitFlag = cli.Int64Flag{
		Name:  "gasLimit",
		Usage: "use for sendContractTransaction gasLimit",
		Value: 0,
	}
)
