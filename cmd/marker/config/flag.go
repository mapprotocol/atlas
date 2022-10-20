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
	//PasswordFlag = cli.StringFlag{
	//	Name:  "password",
	//	Usage: "Keystore file`s password",
	//}

	NameFlag = cli.StringFlag{
		Name:  "name",
		Usage: "name of account",
	}
	URLFlag = cli.StringFlag{
		Name:  "url",
		Usage: "metadata url of account",
	}
	CommissionFlag = cli.Uint64Flag{
		Name:  "commission",
		Usage: "register validator param",
	}
	RelayerfFlag = cli.StringFlag{
		Name:  "relayerf",
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
	MAPValueFlag = cli.Int64Flag{
		Name:  "mapValue",
		Usage: "validator address",
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
	ValueFlag = cli.Uint64Flag{
		Name:  "value",
		Usage: "value units one eth",
		Value: 0,
	}
	AmountFlag = cli.StringFlag{
		Name:  "amount",
		Usage: "transfer amount, unit (wei)",
		Value: "0",
	}
	DurationFlag = cli.Int64Flag{
		Name:  "duration",
		Usage: "duration The time (in seconds) that these requirements persist for.",
		Value: 0,
	}
	TargetAddressFlag = cli.StringFlag{
		Name:  "target",
		Usage: "Target query address",
		Value: "",
	}

	ValidatorAddressFlag = cli.StringFlag{
		Name:  "validator",
		Usage: "validator address",
		Value: "",
	}
	SignerPrivFlag = cli.StringFlag{
		Name:  "signerPriv",
		Usage: "signer private",
		Value: "",
	}
	SignerFlag = cli.StringFlag{
		Name:  "signer",
		Usage: "signer address",
		Value: "",
	}
	SignatureFlag = cli.StringFlag{
		Name:  "signature",
		Usage: "ECDSA Signature",
		Value: "",
	}
	ProofFlag = cli.StringFlag{
		Name:  "proof",
		Usage: "signer proof",
		Value: "",
	}
	AccountAddressFlag = cli.StringFlag{
		Name:  "accountAddress",
		Usage: "account address",
		Value: "",
	}
	ContractAddressFlag = cli.StringFlag{
		Name:  "contractAddress",
		Usage: "set contract Address",
		Value: "",
	}
	ImplementationAddressFlag = cli.StringFlag{
		Name:  "implementationAddress",
		Usage: "set implementation Address",
		Value: "",
	}
	GasLimitFlag = cli.Int64Flag{
		Name:  "gasLimit",
		Usage: "use for sendContractTransaction gasLimit",
		Value: 0,
	}
)
