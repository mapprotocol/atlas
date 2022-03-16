package connections

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/cmd/marker/config"
)

func DialConn(ctx *cli.Context, config *config.Config) (*ethclient.Client, string) {
	logger := log.New("func", "dialConn")
	ip := config.Ip     //utils.RPCListenAddrFlag.Name)
	port := config.Port //utils.RPCPortFlag.Name)
	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	//url := "https://poc2-rpc.maplabs.io"
	conn, err := ethclient.Dial(url)
	if err != nil {
		logger.Error("Failed to connect to the Atlaschain client: %v", err)
	}
	return conn, url
}
