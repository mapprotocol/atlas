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
	conn, _ := connections.DialConn(ctx)
	return &writer{
		config: config,
		conn:   conn,
	}
}

func (w *writer) ResolveMessage(m Message) bool {
	switch m.messageType {
	case SolveType1:
		txHash := sendContractTransaction(w.conn, m.from, m.to, nil, m.priKey, m.input)
		getResult(w.conn, txHash, true)
		m.DoneCh <- struct{}{}
	case SolveType2:
		txHash := sendContractTransaction(w.conn, m.from, m.to, m.value, m.priKey, m.input)
		getResult(w.conn, txHash, true)
		m.DoneCh <- struct{}{}
	case SolveType3:
		w.handleUnpackMethod(m)
		m.DoneCh <- struct{}{}
	default:
	}
	return true
}
