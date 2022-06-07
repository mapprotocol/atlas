package config

import (
	"crypto/ecdsa"
	"github.com/mapprotocol/atlas/cmd/marker/mapprotocol"
	"gopkg.in/urfave/cli.v1"
	"math/big"

	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/marker/account"
	blscrypto "github.com/mapprotocol/atlas/helper/bls"
	"github.com/mapprotocol/atlas/params"

	"github.com/ethereum/go-ethereum/common"
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
	Value      uint64
	Duration   int64
	Commission uint64
	Fixed      string

	VoteNum       *big.Int
	TopNum        *big.Int
	LockedNum     *big.Int
	WithdrawIndex *big.Int
	RelockIndex   *big.Int

	TargetAddress         common.Address
	ContractAddress       common.Address
	SignerPriv            string
	AccountAddress        common.Address //validator
	ImplementationAddress common.Address
	Ip                    string
	Port                  int
	GasLimit              int64
	Verbosity             string
	NamePrefix            string
	LockedGoldParameters  LockedGoldParameters
	AccountsParameters    AccountsParameters
	ValidatorParameters   ValidatorParameters
	EpochRewardParameters EpochRewardsParameters
	TestPoc2Parameters    TestPoc2
	ElectionParameters    ElectionParameters
	GoldTokenParameters   GoldTokenParameters
}

func AssemblyConfig(ctx *cli.Context) (*Config, error) {
	config := Config{}
	//------------------ pre set --------------------------
	path := ""
	password := "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	config.VoteNum = big.NewInt(int64(0))
	config.TargetAddress = params.ZeroAddress
	config.Commission = 1000000 //default 1  be relative to 1000,000
	config.Verbosity = "3"
	config.NamePrefix = "validator"

	//-----------------------------------------------------
	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.String(KeyStoreFlag.Name)
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.String(PasswordFlag.Name)
	}
	if ctx.IsSet(CommissionFlag.Name) {
		config.Commission = ctx.Uint64(CommissionFlag.Name)
	}
	if ctx.IsSet(RelayerfFlag.Name) {
		config.Fixed = ctx.String(RelayerfFlag.Name)
	}
	if ctx.IsSet(VoteNumFlag.Name) {
		config.VoteNum = big.NewInt(ctx.Int64(VoteNumFlag.Name))
	}
	if ctx.IsSet(TargetAddressFlag.Name) {
		config.TargetAddress = common.HexToAddress(ctx.String(TargetAddressFlag.Name))
	}
	if ctx.IsSet(ValidatorAddressFlag.Name) {
		config.TargetAddress = common.HexToAddress(ctx.String(ValidatorAddressFlag.Name))
	}
	if ctx.IsSet(ValidatorAddressFlag.Name) {
		config.AccountAddress = common.HexToAddress(ctx.String(ValidatorAddressFlag.Name))
	}
	if ctx.IsSet(SignerPrivFlag.Name) {
		config.SignerPriv = ctx.String(SignerPrivFlag.Name)
	}
	if ctx.IsSet(ImplementationAddressFlag.Name) {
		config.ImplementationAddress = common.HexToAddress(ctx.String(ImplementationAddressFlag.Name))
	}
	if ctx.IsSet(ContractAddressFlag.Name) {
		config.ContractAddress = common.HexToAddress(ctx.String(ContractAddressFlag.Name))
	}
	if ctx.IsSet(ValueFlag.Name) {
		config.Value = ctx.Uint64(ValueFlag.Name)
	}
	if ctx.IsSet(DurationFlag.Name) {
		config.Duration = ctx.Int64(DurationFlag.Name)
	}
	if ctx.IsSet(TopNumFlag.Name) {
		config.TopNum = big.NewInt(ctx.Int64(TopNumFlag.Name))
	}
	if ctx.IsSet(LockedNumFlag.Name) {
		config.LockedNum = big.NewInt(ctx.Int64(LockedNumFlag.Name))
	}
	if ctx.IsSet(MAPValueFlag.Name) {
		config.LockedNum = big.NewInt(ctx.Int64(MAPValueFlag.Name))
	}
	if ctx.IsSet(WithdrawIndexFlag.Name) {
		config.WithdrawIndex = big.NewInt(ctx.Int64(WithdrawIndexFlag.Name))
	}
	if ctx.IsSet(RelockIndexFlag.Name) {
		config.RelockIndex = big.NewInt(ctx.Int64(RelockIndexFlag.Name))
	}
	if ctx.IsSet(VerbosityFlag.Name) {
		config.Verbosity = ctx.String(VerbosityFlag.Name)
	}
	if ctx.IsSet(NamePrefixFlag.Name) {
		config.NamePrefix = ctx.String(NamePrefixFlag.Name)
	}
	if ctx.IsSet(RPCListenAddrFlag.Name) {
		config.Ip = ctx.String(RPCListenAddrFlag.Name)
	}
	if ctx.IsSet(RPCPortFlag.Name) {
		config.Port = ctx.Int(RPCPortFlag.Name)
	}
	if ctx.IsSet(GasLimitFlag.Name) {
		config.GasLimit = ctx.Int64(GasLimitFlag.Name)
	}
	if path != "" {
		_account, err := account.LoadAccount(path, password)
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
	AccountsAddress := mapprotocol.MustProxyAddressFor("Accounts")
	ElectionAddress := mapprotocol.MustProxyAddressFor("Election")
	GoldTokenAddress := mapprotocol.MustProxyAddressFor("GoldToken")
	EpochRewardsAddress := mapprotocol.MustProxyAddressFor("EpochRewards")
	config.ValidatorParameters.ValidatorAddress = ValidatorAddress
	config.EpochRewardParameters.EpochRewardsAddress = EpochRewardsAddress
	config.TestPoc2Parameters.Address = common.HexToAddress("0xb586DC60e9e39F87c9CB8B7D7E30b2f04D40D14c")
	config.LockedGoldParameters.LockedGoldAddress = LockedGoldAddress
	config.AccountsParameters.AccountsAddress = AccountsAddress
	config.ElectionParameters.ElectionAddress = ElectionAddress
	config.GoldTokenParameters.GoldTokenAddress = GoldTokenAddress

	abiValidators := mapprotocol.AbiFor("Validators")
	abiLockedGold := mapprotocol.AbiFor("LockedGold")
	abiAccounts := mapprotocol.AbiFor("Accounts")
	abiElection := mapprotocol.AbiFor("Election")
	abiGoldToken := mapprotocol.AbiFor("GoldToken")
	abiEpochRewards := mapprotocol.AbiFor("EpochRewards")

	config.ValidatorParameters.ValidatorABI = abiValidators
	config.EpochRewardParameters.EpochRewardsABI = abiEpochRewards
	config.TestPoc2Parameters.ABI = mapprotocol.AbiFor("TestPoc2")
	config.LockedGoldParameters.LockedGoldABI = abiLockedGold
	config.AccountsParameters.AccountsABI = abiAccounts
	config.ElectionParameters.ElectionABI = abiElection
	config.GoldTokenParameters.GoldTokenABI = abiGoldToken

	return &config, nil
}
