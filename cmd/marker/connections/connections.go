package connections

import (
	"context"
	"fmt"
	"net"
	"strings"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/mapprotocol/atlas/cmd/marker/config"
)

const (
	httpScheme  = "http"
	httpsScheme = "https"
)

func DialConn(ctx *cli.Context, config *config.Config) (*ethclient.Client, string) {
	var (
		url    string
		host   string
		scheme string
	)
	ip := config.Ip
	port := config.Port
	parts := strings.Split(ip, "://")
	addr := parts[len(parts)-1]

	if len(parts) > 1 {
		scheme = parts[0]
	}
	if port != 0 {
		host = fmt.Sprintf("%s:%d", addr, port)
	} else {
		host = addr
	}
	if net.ParseIP(ip) != nil {
		if scheme == "" {
			scheme = httpScheme
		}
	} else {
		if scheme == "" {
			scheme = httpsScheme
		}
	}
	url = fmt.Sprintf("%s://%s", scheme, host)
	conn, err := ethclient.Dial(url)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to the map chain, error: %v", err))
	}

	_, err = conn.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	return conn, url
}

func DialRpc(config *config.Config) (*rpc.Client, string) {
	logger := log.New("func", "dialConn")
	ip := config.Ip     //utils.RPCListenAddrFlag.Name)
	port := config.Port //utils.RPCPortFlag.Name)
	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	//url := "https://poc2-rpc.maplabs.io"
	conn, err := rpc.Dial(url)
	if err != nil {
		logger.Error("Failed to connect to the Atlaschain client: %v", err)
	}
	return conn, url
}
