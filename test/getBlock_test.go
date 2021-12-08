package test

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/mapprotocol/atlas/core/types"
	"log"
	"math/big"
)

func getChains(startNum, endNum int64) []*types.Header {
	conn, _ := dialEthConn()
	Headers := make([]*types.Header, 0, endNum-startNum+1)
	for i := startNum; i <= endNum; i++ {
		Header, _ := conn.HeaderByNumber(context.Background(), big.NewInt(i))
		if Header != nil {
			Headers = append(Headers, Header)
		}
	}
	return Headers
}

func dialEthConn() (*Client, string) {
	ip := "127.0.0.1" //utils.RPCListenAddrFlag.Name)
	port := 7415      //utils.RPCPortFlag.Name)
	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	conn, err := Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the Atlaschain client: %v", err)
	}
	return conn, url
}

// Client defines typed wrappers for the Ethereum RPC API.
type Client struct {
	c *rpc.Client
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (ec *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	var head *types.Header
	err := ec.c.CallContext(ctx, &head, "eth_getBlockByNumber", toBlockNumArg(number), false)
	if err == nil && head == nil {
		err = errors.New("err")
	}
	return head, err
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	pending := big.NewInt(-1)
	if number.Cmp(pending) == 0 {
		return "pending"
	}
	return hexutil.EncodeBig(number)
}

// Dial connects a client to the given URL.
func Dial(rawurl string) (*Client, error) {
	return DialContext(context.Background(), rawurl)
}

func DialContext(ctx context.Context, rawurl string) (*Client, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

// NewClient creates a client that uses the given RPC client.
func NewClient(c *rpc.Client) *Client {
	return &Client{c}
}
