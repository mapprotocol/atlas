package cmd

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/new_marker/config"
	"github.com/mapprotocol/atlas/cmd/new_marker/mapprotocol"
	"github.com/mapprotocol/atlas/cmd/new_marker/writer"
	"gopkg.in/urfave/cli.v1"
	"math/big"
)

type Account struct {
	*base
	*writer.Writer
	to, lockGoldTo     common.Address
	abi, lockedGoldAbi *abi.ABI
}

func NewAccount() *Account {
	return &Account{
		base:          newBase(),
		abi:           mapprotocol.AbiFor("Accounts"),
		lockedGoldAbi: mapprotocol.AbiFor("LockedGold"),
		to:            mapprotocol.MustProxyAddressFor("Accounts"),
		lockGoldTo:    mapprotocol.MustProxyAddressFor("LockedGold"),
	}
}

func (a *Account) GetAccountMetadataURL(_ *cli.Context, cfg *config.Config) error {
	var (
		ret interface{}
	)
	a.handleType3Msg(cfg, &ret, a.to, nil, a.abi, "getMetadataURL", cfg.TargetAddress)
	log.Info("get account metadata url", "address", cfg.TargetAddress, "url", ret)
	return nil
}

func (a *Account) GetAccountName(_ *cli.Context, cfg *config.Config) error {
	var (
		ret interface{}
	)
	a.handleType3Msg(cfg, &ret, a.to, nil, a.abi, "getName", cfg.TargetAddress)
	log.Info("get name", "address", cfg.TargetAddress, "name", ret)
	return nil
}

func (a *Account) GetAccountTotalLockedGold(_ *cli.Context, cfg *config.Config) error {
	var (
		ret interface{}
	)
	log.Info("=== getAccountTotalLockedGold ===", "admin", cfg.From, "target", cfg.TargetAddress.String())
	a.handleType3Msg(cfg, &ret, a.lockGoldTo, nil, a.lockedGoldAbi, "getAccountTotalLockedGold", cfg.TargetAddress)
	result := ret.(*big.Int)
	log.Info("result", "lockedGold", result)
	return nil
}

func (a *Account) GetAccountNonvotingLockedGold(_ *cli.Context, cfg *config.Config) error {
	var (
		ret interface{}
	)
	log.Info("=== getAccountNonvotingLockedGold ===", "admin", cfg.From, "target", cfg.TargetAddress.String())
	a.handleType3Msg(cfg, &ret, a.lockGoldTo, nil, a.lockedGoldAbi, "getAccountNonvotingLockedGold", cfg.TargetAddress)
	result := ret.(*big.Int)
	log.Info("result", "lockedGold", result)
	return nil
}
