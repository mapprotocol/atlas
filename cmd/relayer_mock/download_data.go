package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mapprotocol/atlas/chains/headers/ethereum"
	"github.com/mapprotocol/atlas/cmd/ethclient"
	"log"
	"math/big"
)

const (
	syncNumber       = 10
	EthRPCListenAddr = "localhost"
	EthRPCPortFlag   = 8083
)

func getChains(startNum uint64) ([]ethereum.Header, []bytes.Buffer) {
	conn, _ := dialEthConn()
	currentNum_, _ := conn.BlockNumber(context.Background())
	//if startNum == currentNum_ {
	//	log.Fatalf("startNum == currentNum ")
	//}
	//if startNum > currentNum_ {
	//	log.Fatalf("startNum %v > currentNum %v ",startNum,currentNum_)
	//}
	currentNum := int(currentNum_)
	if currentNum == 0 {
		fmt.Println("currentNum ==0")
	}
	startNum01 := int(startNum)
	end := Min(startNum01+syncNumber, currentNum)

	Headers := make([]ethereum.Header, end-startNum01+1)
	HeaderBytes := make([]bytes.Buffer, end-startNum01+1)
	j := 0
	for i := startNum01; i <= end; i++ {
		Header, _ := conn.HeaderByNumber(context.Background(), big.NewInt(int64(i)))
		convertChain(&Headers[j], &HeaderBytes[j], Header)
		j++
	}
	return Headers, HeaderBytes
}
func dialEthConn() (*ethclient.Client, string) {
	ip = EthRPCListenAddr //utils.RPCListenAddrFlag.Name)
	port = EthRPCPortFlag //utils.RPCPortFlag.Name)
	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	conn, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the Abeychain client: %v", err)
	}
	return conn, url
}
