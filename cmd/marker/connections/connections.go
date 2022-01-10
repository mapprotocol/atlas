package connections

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

func DialConn(ctx *cli.Context) (*ethclient.Client, string) {
	logger := log.New("func", "dialConn")
	ip := ctx.GlobalString("rpcaddr") //utils.RPCListenAddrFlag.Name)
	port := ctx.GlobalInt("rpcport")  //utils.RPCPortFlag.Name)
	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	conn, err := ethclient.Dial(url)
	if err != nil {
		logger.Error("Failed to connect to the Atlaschain client: %v", err)
	}
	return conn, url
}
