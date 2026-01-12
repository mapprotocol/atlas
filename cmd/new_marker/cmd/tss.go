package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/new_marker/define"
	"github.com/mapprotocol/atlas/cmd/new_marker/mapprotocol"
	"gopkg.in/urfave/cli.v1"
)

var (
	testnet = common.HexToAddress("0x60c2e5bd5b785910424C48098292Ab410884B5ad")
	mainnet = common.HexToAddress("0x7e22B9FC15054546028Df928eB7560AEd8F0eF48")
)

type Tss struct {
	*base
	abi *abi.ABI
}

func NewTss() *Tss {
	return &Tss{
		base: newBase(),
		abi:  mapprotocol.AbiFor("Maintainer"),
	}
}

func (s *Tss) Register(ctx *cli.Context, cfg *define.Config) error {
	p2pAddr := ""
	if ctx.IsSet(define.P2pAddressFlag.Name) {
		p2pAddr = ctx.String(define.P2pAddressFlag.Name)
	}
	if p2pAddr == "" {
		return fmt.Errorf("p2p address is required")
	}
	to := mainnet
	if ctx.Bool(define.TestnetFlag.Name) {
		to = testnet
	}
	registerFrom := cfg.From
	pkBytes := cfg.PublicKey // remove pk prefix 0x04
	if ctx.IsSet(define.TssRegisterPkFlag.Name) {
		pkStr := ctx.String(define.TssRegisterPkFlag.Name)
		if len(pkStr) != 130 || pkStr[:2] != "04" {
			return fmt.Errorf("invalid registerPk format, should be uncompressed pk prefixed with 04")
		}
		var err error
		pkBytes = common.Hex2Bytes(pkStr)
		if err != nil {
			return fmt.Errorf("invalid registerPk hex string: %v", err)
		}

		hash := crypto.Keccak256(pkBytes[1:])
		registerFrom = common.BytesToAddress(hash[12:])
	}

	s.handleType1Msg(cfg, to, nil, s.abi, "register", registerFrom, pkBytes[1:], pkBytes[1:], p2pAddr)

	return nil
}

func (s *Tss) Update(ctx *cli.Context, cfg *define.Config) error {
	p2pAddr := ""
	if ctx.IsSet(define.P2pAddressFlag.Name) {
		p2pAddr = ctx.String(define.P2pAddressFlag.Name)
	}
	if p2pAddr == "" {
		return fmt.Errorf("p2p address is required")
	}
	to := mainnet
	if ctx.Bool(define.TestnetFlag.Name) {
		to = testnet
	}
	var err error
	registerFrom := cfg.From
	pkBytes := cfg.PublicKey // remove pk prefix 0x04
	if ctx.IsSet(define.TssRegisterPkFlag.Name) {
		pkStr := ctx.String(define.TssRegisterPkFlag.Name)
		if len(pkStr) != 130 || pkStr[:2] != "04" {
			return fmt.Errorf("invalid registerPk format, should be uncompressed pk prefixed with 04")
		}
		pkBytes, err = hex.DecodeString(pkStr)
		if err != nil {
			return fmt.Errorf("invalid registerPk hex string: %v", err)
		}

		hash := crypto.Keccak256(pkBytes[1:])
		registerFrom = common.BytesToAddress(hash[12:])
	}

	s.handleType1Msg(cfg, to, nil, s.abi, "update", registerFrom, pkBytes[1:], pkBytes[1:], p2pAddr)

	return nil
}

func (s *Tss) Activate(ctx *cli.Context, cfg *define.Config) error {
	to := mainnet
	if ctx.Bool(define.TestnetFlag.Name) {
		to = testnet
	}
	s.handleType1Msg(cfg, to, nil, s.abi, "activate")

	return nil
}

func (s *Tss) Revoke(ctx *cli.Context, cfg *define.Config) error {
	to := mainnet
	if ctx.Bool(define.TestnetFlag.Name) {
		to = testnet
	}
	s.handleType1Msg(cfg, to, nil, s.abi, "revoke")

	return nil
}
