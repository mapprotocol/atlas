package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/mapprotocol/atlas/chains/ethereum"
	"github.com/mapprotocol/atlas/cmd/ethclient"
	"gopkg.in/urfave/cli.v1"
	"log"
	"math/big"
)

const (
	syncNumber       = 10
	EthRPCListenAddr = "localhost"
	EthRPCPortFlag   = 8082
)

func getChains(ctx *cli.Context, startNum uint64) ([]ethereum.Header, []bytes.Buffer) {
	conn, _ := dialEthConn(ctx)
	currentNum_, _ := conn.BlockNumber(context.Background())
	if startNum == currentNum_ {
		log.Fatalf("startNum == currentNum ")
	}
	if startNum > currentNum_ {
		log.Fatalf("startNum > currentNum ")
	}
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
func dialEthConn(ctx *cli.Context) (*ethclient.Client, string) {
	ip = EthRPCListenAddr //utils.RPCListenAddrFlag.Name)
	port = EthRPCPortFlag //utils.RPCPortFlag.Name)
	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	// Create an IPC based RPC connection to a remote node
	// "http://39.100.97.129:8545"
	conn, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the Abeychain client: %v", err)
	}
	return conn, url
}
func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
func convertChain(header *ethereum.Header, headerbyte *bytes.Buffer, e *types.Header) (*ethereum.Header, *bytes.Buffer) {
	header.ParentHash = e.ParentHash
	header.UncleHash = e.UncleHash
	header.Coinbase = e.Coinbase
	header.Root = e.Root
	header.TxHash = e.TxHash
	header.ReceiptHash = e.ReceiptHash
	header.GasLimit = e.GasLimit
	header.GasUsed = e.GasUsed
	header.Time = e.Time
	header.MixDigest = e.MixDigest
	header.Nonce = types.EncodeNonce(e.Nonce.Uint64())
	header.Bloom.SetBytes(e.Bloom.Bytes())
	if header.Difficulty = new(big.Int); e.Difficulty != nil {
		header.Difficulty.Set(e.Difficulty)
	}
	if header.Number = new(big.Int); e.Number != nil {
		header.Number.Set(e.Number)
	}
	if len(e.Extra) > 0 {
		header.Extra = make([]byte, len(e.Extra))
		copy(header.Extra, e.Extra)
	}
	binary.Write(headerbyte, binary.BigEndian, header)
	// test rlp
	//fmt.Println(e.Hash(), "/n", header.Hash())
	return header, headerbyte
}
