package connections

import (
	"context"
	"fmt"
	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/mapprotocol/atlas/cmd/marker/config"
)

const (
	httpScheme  = "http"
	httpsScheme = "https"
	localHost   = "localhost"
)

func DialConn(ctx *cli.Context, config *config.Config) (*ethclient.Client, string) {
	conn, err := ethclient.Dial(config.RPCAddr)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to the map chain, addr: %s, error: %v", config.RPCAddr, err))
	}

	_, err = conn.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	return conn, config.RPCAddr
}

func DialRpc(config *config.Config) (*rpc.Client, string) {
	logger := log.New("func", "dialConn")
	conn, err := rpc.Dial(config.RPCAddr)
	if err != nil {
		logger.Error("Failed to connect to the Atlaschain client: %v", err)
	}
	return conn, config.RPCAddr
}
