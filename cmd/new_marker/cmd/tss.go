package cmd

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/new_marker/define"
	"github.com/mapprotocol/atlas/cmd/new_marker/mapprotocol"
	"gopkg.in/urfave/cli.v1"
)

var (
	testnet = common.HexToAddress("0x606A45f78D3A706F7a621fF03Dd62C513fa13b2c")
	mainnet = common.HexToAddress("")
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
	if ctx.IsSet(define.P2pAddress.Name) {
		p2pAddr = ctx.String(define.P2pAddress.Name)
	}
	if p2pAddr == "" {
		return fmt.Errorf("p2p address is required")
	}
	to := mainnet
	if ctx.Bool(define.Testnet.Name) {
		to = testnet
	}

	s.handleType1Msg(cfg, to, nil, s.abi, "register",
		cfg.From, cfg.PublicKey[1:], cfg.PublicKey[1:], p2pAddr) // remove pk prefix 0x04

	return nil
}
