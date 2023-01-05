package connections

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/mapprotocol/atlas/cmd/new_marker/config"
)

const (
	httpScheme  = "http"
	httpsScheme = "https"
	localHost   = "localhost"
)

func DialConn(addr string) *ethclient.Client {
	conn, err := ethclient.Dial(addr)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to the map chain, addr: %s, error: %v", addr, err))
	}

	_, err = conn.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	return conn
}

func DialRpc(config *config.Config) (*rpc.Client, string) {
	logger := log.New("func", "dialConn")
	conn, err := rpc.Dial(config.RPCAddr)
	if err != nil {
		logger.Error("Failed to connect to the Atlaschain client: %v", err)
	}
	return conn, config.RPCAddr
}
