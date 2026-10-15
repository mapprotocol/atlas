package define

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/new_marker/mapprotocol"
	blscrypto "github.com/mapprotocol/atlas/helper/bls"
	"github.com/mapprotocol/atlas/params"
	"golang.org/x/term"
	"gopkg.in/urfave/cli.v1"
)

const (
	TestnetRPC = "https://testnet-rpc.maplabs.io"
	MainnetRPC = "https://rpc.maplabs.io"
)

type LockedGoldParameters struct {
	LockedGoldABI     *abi.ABI
	LockedGoldAddress common.Address
}
type AccountsParameters struct {
	AccountsABI     *abi.ABI
	AccountsAddress common.Address
}
type ValidatorParameters struct {
	ValidatorABI     *abi.ABI
	ValidatorAddress common.Address
}
type EpochRewardsParameters struct {
	EpochRewardsABI     *abi.ABI
	EpochRewardsAddress common.Address
}
type ElectionParameters struct {
	ElectionABI     *abi.ABI
	ElectionAddress common.Address
}
type GoldTokenParameters struct {
	GoldTokenABI     *abi.ABI
	GoldTokenAddress common.Address
}
type TestPoc2 struct {
	ABI     *abi.ABI
	Address common.Address
}
type Config struct {
	From       common.Address
	PublicKey  []byte
	PrivateKey *ecdsa.PrivateKey
	BlsPub     blscrypto.SerializedPublicKey
	BlsG1Pub   blscrypto.SerializedG1PublicKey
	BLSProof   []byte
	Amount     string
	Duration   int64
	Commission uint64
	Fixed      string

	Value *big.Int
	Index *big.Int

	PrivateKeyHex         string
	Signature             string
	Proof                 string
	Address               common.Address
	ImplementationAddress common.Address
	RPCAddr               string
	GasLimit              int64
	Verbosity             string
	Name                  string
	URL                   string

	LockedGoldParameters  LockedGoldParameters
	ValidatorParameters   ValidatorParameters
	EpochRewardParameters EpochRewardsParameters
	TestPoc2Parameters    TestPoc2
	ElectionParameters    ElectionParameters
	GoldTokenParameters   GoldTokenParameters
}

func AssemblyConfig(ctx *cli.Context) (*Config, error) {
	config := Config{}
	//------------------ pre set --------------------------
	keyStorePath := ""
	config.Value = big.NewInt(0)
	config.Address = params.ZeroAddress
	config.Commission = 1000000 //default 1  be relative to 1000,000
	config.Verbosity = "info"
	config.Name = "validator"
	config.From = common.HexToAddress("0x0000000000000000000000000000000000000000") //  default

	//-----------------------------------------------------
	if ctx.IsSet(KeyStoreFlag.Name) {
		keyStorePath = ctx.String(KeyStoreFlag.Name)
	}
	if ctx.IsSet(CommissionFlag.Name) {
		config.Commission = ctx.Uint64(CommissionFlag.Name)
	}
	if ctx.IsSet(RelayerFlag.Name) {
		config.Fixed = ctx.String(RelayerFlag.Name)
	}
	if ctx.IsSet(AddressFlag.Name) {
		config.Address = common.HexToAddress(ctx.String(AddressFlag.Name))
	}
	if ctx.IsSet(PrivateKeyFlag.Name) {
		config.PrivateKeyHex = ctx.String(PrivateKeyFlag.Name)
	}
	if ctx.IsSet(SignatureFlag.Name) {
		config.Signature = ctx.String(SignatureFlag.Name)
	}
	if ctx.IsSet(ProofFlag.Name) {
		config.Proof = ctx.String(ProofFlag.Name)
	}
	if ctx.IsSet(ImplementationAddressFlag.Name) {
		config.ImplementationAddress = common.HexToAddress(ctx.String(ImplementationAddressFlag.Name))
	}
	if ctx.IsSet(AccountFlag.Name) {
		config.From = common.HexToAddress(ctx.String(AccountFlag.Name))
	}
	if ctx.IsSet(ValueFlag.Name) {
		config.Value = new(big.Int).SetUint64(ctx.Uint64(ValueFlag.Name))
	}
	if ctx.IsSet(AmountFlag.Name) {
		config.Amount = ctx.String(AmountFlag.Name)
	}
	if ctx.IsSet(DurationFlag.Name) {
		config.Duration = ctx.Int64(DurationFlag.Name)
	}
	if ctx.IsSet(IndexFlag.Name) {
		config.Index = big.NewInt(ctx.Int64(IndexFlag.Name))
	}
	if ctx.IsSet(IndexFlag.Name) {
		config.Index = big.NewInt(ctx.Int64(IndexFlag.Name))
	}
	if ctx.IsSet(VerbosityFlag.Name) {
		config.Verbosity = ctx.String(VerbosityFlag.Name)
	}
	if ctx.IsSet(NameFlag.Name) {
		config.Name = ctx.String(NameFlag.Name)
	}
	if ctx.IsSet(URLFlag.Name) {
		config.URL = ctx.String(URLFlag.Name)
	}
	if ctx.IsSet(RPCAddrFlag.Name) {
		config.RPCAddr = ctx.String(RPCAddrFlag.Name)
	} else {
		config.RPCAddr = MainnetRPC
		if ctx.Bool(TestnetFlag.Name) {
			config.RPCAddr = TestnetRPC
		}
	}
	if ctx.IsSet(GasLimitFlag.Name) {
		config.GasLimit = ctx.Int64(GasLimitFlag.Name)
	}
	if keyStorePath != "" {
		_account, err := LoadAccount(keyStorePath, string(GetPassword(fmt.Sprintf("Enter password for key %s:", keyStorePath))))
		if err != nil {
			return nil, err
		}
		blsPub, err := _account.BLSPublicKey()
		if err != nil {
			return nil, err
		}
		blsG1Pub, err := _account.BLSG1PublicKey()
		if err != nil {
			return nil, err
		}
		config.PublicKey = _account.PublicKey()

		config.From = _account.Address
		config.PrivateKey = _account.PrivateKey
		config.BlsPub = blsPub
		config.BlsG1Pub = blsG1Pub
		config.BLSProof = _account.MustBLSProofOfPossession()
	}

	ValidatorAddress := mapprotocol.MustProxyAddressFor("Validators")
	LockedGoldAddress := mapprotocol.MustProxyAddressFor("LockedGold")
	ElectionAddress := mapprotocol.MustProxyAddressFor("Election")
	GoldTokenAddress := mapprotocol.MustProxyAddressFor("GoldToken")
	EpochRewardsAddress := mapprotocol.MustProxyAddressFor("EpochRewards")
	config.ValidatorParameters.ValidatorAddress = ValidatorAddress
	config.EpochRewardParameters.EpochRewardsAddress = EpochRewardsAddress
	config.TestPoc2Parameters.Address = common.HexToAddress("0xb586DC60e9e39F87c9CB8B7D7E30b2f04D40D14c")
	config.LockedGoldParameters.LockedGoldAddress = LockedGoldAddress
	config.ElectionParameters.ElectionAddress = ElectionAddress
	config.GoldTokenParameters.GoldTokenAddress = GoldTokenAddress

	abiValidators := mapprotocol.AbiFor("Validators")
	abiLockedGold := mapprotocol.AbiFor("LockedGold")
	abiElection := mapprotocol.AbiFor("Election")
	abiGoldToken := mapprotocol.AbiFor("GoldToken")
	abiEpochRewards := mapprotocol.AbiFor("EpochRewards")

	config.ValidatorParameters.ValidatorABI = abiValidators
	config.EpochRewardParameters.EpochRewardsABI = abiEpochRewards
	config.TestPoc2Parameters.ABI = mapprotocol.AbiFor("TestPoc2")
	config.LockedGoldParameters.LockedGoldABI = abiLockedGold
	config.ElectionParameters.ElectionABI = abiElection
	config.GoldTokenParameters.GoldTokenABI = abiGoldToken

	return &config, nil
}

func GetPassword(msg string) []byte {
	for {
		fmt.Println(msg)
		fmt.Print("> ")
		password, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			fmt.Printf("invalid input: %s\n", err)
		} else {
			fmt.Printf("\n")
			return password
		}
	}
}
