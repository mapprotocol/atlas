package cmd

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/new_marker/define"
	"github.com/mapprotocol/atlas/cmd/new_marker/mapprotocol"
	"github.com/mapprotocol/atlas/cmd/new_marker/writer"
	"gopkg.in/urfave/cli.v1"
	"math/big"
)

type Account struct {
	*base
	*writer.Writer
	to, lockGoldTo, electionTo      common.Address
	abi, lockedGoldAbi, electionAbi *abi.ABI
}

func NewAccount() *Account {
	return &Account{
		base:          newBase(),
		to:            mapprotocol.MustProxyAddressFor("Accounts"),
		abi:           mapprotocol.AbiFor("Accounts"),
		lockGoldTo:    mapprotocol.MustProxyAddressFor("LockedGold"),
		lockedGoldAbi: mapprotocol.AbiFor("LockedGold"),
		electionTo:    mapprotocol.MustProxyAddressFor("Election"),
		electionAbi:   mapprotocol.AbiFor("Election"),
	}
}

func (a *Account) GetAccountMetadataURL(_ *cli.Context, cfg *define.Config) error {
	var (
		ret interface{}
	)
	a.handleType3Msg(cfg, &ret, a.to, nil, a.abi, "getMetadataURL", cfg.TargetAddress)
	log.Info("get account metadata url", "address", cfg.TargetAddress, "url", ret)
	return nil
}

func (a *Account) GetAccountName(_ *cli.Context, cfg *define.Config) error {
	var (
		ret interface{}
	)
	a.handleType3Msg(cfg, &ret, a.to, nil, a.abi, "getName", cfg.TargetAddress)
	log.Info("get name", "address", cfg.TargetAddress, "name", ret)
	return nil
}

func (a *Account) GetAccountTotalLockedGold(_ *cli.Context, cfg *define.Config) error {
	var (
		ret interface{}
	)
	log.Info("=== getAccountTotalLockedGold ===", "admin", cfg.From, "target", cfg.TargetAddress.String())
	a.handleType3Msg(cfg, &ret, a.lockGoldTo, nil, a.lockedGoldAbi, "getAccountTotalLockedGold", cfg.TargetAddress)
	result := ret.(*big.Int)
	log.Info("result", "lockedGold", result)
	return nil
}

func (a *Account) GetAccountNonvotingLockedGold(_ *cli.Context, cfg *define.Config) error {
	var (
		ret interface{}
	)
	log.Info("=== getAccountNonvotingLockedGold ===", "admin", cfg.From, "target", cfg.TargetAddress.String())
	a.handleType3Msg(cfg, &ret, a.lockGoldTo, nil, a.lockedGoldAbi, "getAccountNonvotingLockedGold", cfg.TargetAddress)
	result := ret.(*big.Int)
	log.Info("result", "lockedGold", result)
	return nil
}

func (a *Account) GetPendingVotesForValidatorByAccount(_ *cli.Context, cfg *define.Config) error {
	var (
		ret interface{}
	)
	log.Info("=== getPendingVotesForValidatorByAccount ===", "admin", cfg.From)
	a.handleType3Msg(cfg, &ret, a.electionTo, nil, a.electionAbi, "getPendingVotesForValidatorByAccount", cfg.TargetAddress, cfg.From)
	log.Info("PendingVotes", "balance", ret.(*big.Int))
	return nil
}

func (a *Account) GetActiveVotesForValidatorByAccount(_ *cli.Context, cfg *define.Config) error {
	var (
		ret interface{}
	)
	log.Info("=== getActiveVotesForValidatorByAccount ===", "admin", cfg.From)
	a.handleType3Msg(cfg, &ret, a.electionTo, nil, a.electionAbi, "getActiveVotesForValidatorByAccount",
		cfg.TargetAddress, cfg.From)
	log.Info("ActiveVotes", "balance", ret.(*big.Int))
	return nil
}

func (a *Account) GetValidatorsVotedForByAccount(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== getValidatorsVotedForByAccount ===", "admin", cfg.From)
	var (
		ret interface{}
	)
	a.handleType3Msg(cfg, &ret, a.electionTo, nil, a.electionAbi, "getValidatorsVotedForByAccount", cfg.TargetAddress)
	result := ret.([]common.Address)
	if len(result) == 0 {
		log.Info("nil")
	}
	for i := 0; i < len(result); i++ {
		log.Info("validator", "Address", result[i])
	}
	return nil
}

func (a *Account) SetAccountMetadataURL(_ *cli.Context, cfg *define.Config) error {
	a.handleType1Msg(cfg, a.to, nil, a.abi, "setMetadataURL", cfg.MetadataURL)
	log.Info("set account metadata url", "address", cfg.From, "url", cfg.MetadataURL)
	return nil
}

func (a *Account) SetAccountName(_ *cli.Context, cfg *define.Config) error {
	log.Info("set name", "address", cfg.From, "name", cfg.Name)
	a.handleType1Msg(cfg, a.to, nil, a.abi, "setName", cfg.Name)
	return nil
}

func (a *Account) CreateAccount(_ *cli.Context, cfg *define.Config) error {
	logger := log.New("func", "createAccount")
	logger.Info("Create account", "address", cfg.From, "name", cfg.Name)
	log.Info("=== create Account ===")
	// step 1
	a.handleType1Msg(cfg, a.to, nil, a.abi, "createAccount")
	// step 2
	log.Info("=== setName name ===")
	a.handleType1Msg(cfg, a.to, nil, a.abi, "setName", cfg.Name)
	// step 3
	log.Info("=== setAccountDataEncryptionKey ===")
	a.handleType1Msg(cfg, a.to, nil, a.abi, "setAccountDataEncryptionKey", cfg.PublicKey)
	return nil
}

// SignerToAccount : Query the account of a target signer
func (a *Account) SignerToAccount(_ *cli.Context, cfg *define.Config) error {
	//----------------------------- signerToAccount ---------------------------------
	logger := log.New("func", "signerToAccount")
	var ret common.Address
	a.handleType3Msg(cfg, &ret, a.to, nil, a.abi, "signerToAccount", cfg.TargetAddress)
	logger.Info("signerToAccount", "authorizingAccount", ret)
	return nil
}
