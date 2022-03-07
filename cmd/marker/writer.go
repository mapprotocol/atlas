package main

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mapprotocol/atlas/cmd/marker/config"
	"github.com/mapprotocol/atlas/cmd/marker/connections"

	"gopkg.in/urfave/cli.v1"
)

type writer struct {
	config *config.Config
	conn   *ethclient.Client
}

func NewWriter(ctx *cli.Context, config *config.Config) *writer {
	conn, _ := connections.DialConn(ctx, config)
	return &writer{
		config: config,
		conn:   conn,
	}
}

func (w *writer) ResolveMessage(m Message) bool {
	switch m.messageType {
	case SolveSendTranstion1:
		txHash := sendContractTransaction(w.conn, m.from, m.to, nil, m.priKey, m.input, m.gasLimit)
		getResult(w.conn, txHash, true)
		m.DoneCh <- struct{}{}
	case SolveSendTranstion2:
		txHash := sendContractTransaction(w.conn, m.from, m.to, m.value, m.priKey, m.input, m.gasLimit)
		getResult(w.conn, txHash, true)
		m.DoneCh <- struct{}{}
	case SolveQueryResult3:
		w.handleUnpackMethodSolveType3(m)
		m.DoneCh <- struct{}{}
	case SolveQueryResult4:
		w.handleUnpackMethodSolveType4(m)
		m.DoneCh <- struct{}{}
	default:
	}
	return true
}
