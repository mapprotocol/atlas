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
	CommissionFlag = cli.Int64Flag{
		Name:  "commission",
		Usage: "register validator param",
	}
	LesserFlag = cli.StringFlag{
		Name: "lesser",
		Usage: "The validator receiving fewer votes than the validator for which the vote was revoked," +
			"or 0 if that validator has the fewest votes of any validator validator",
	}
	GreaterFlag = cli.StringFlag{
		Name: "greater",
		Usage: "Greater The validator receiving more votes than the validator for which the vote was revoked," +
			"or 0 if that validator has the most votes of any validator validator.",
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
	ValidatorIndexFlag = cli.Int64Flag{
		Name:  "validatorIndex",
		Usage: "use for revokePending or revokeActive",
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
		Usage: "HTTP-RPC server listening port",
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
)
