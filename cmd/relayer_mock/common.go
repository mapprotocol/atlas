package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mapprotocol/atlas/chains/headers/ethereum"
	"github.com/mapprotocol/atlas/cmd/ethclient"
	"io/ioutil"
	"log"
	"math/big"
)

var (
	keystore1 = "D:/BaiduNetdiskDownload/test015/atlas/data555/keystore/UTC--2021-07-09T06-27-06.967129500Z--c971f9cec4310cf001ca55078b43a568aaa0366d"
	keystore2 = "D:/BaiduNetdiskDownload/test015/atlas/data555/keystore/UTC--2021-07-09T06-26-32.960000300Z--78c5285c42572677d3f9dcc27b9ac7b1ff49843c"
	keystore3 = "D:/BaiduNetdiskDownload/test015/atlas/data555/keystore/UTC--2021-07-11T06-35-36.635750800Z--70bf8d9de50713101992649a4f0d7fa505ebb334"
	keystore4 = "D:/BaiduNetdiskDownload/test015/atlas/data555/keystore/UTC--2021-07-19T11-51-51.704095400Z--4e0449459f73341f8e9339cb9e49dae3115ec80f"
	keystore5 = "D:/BaiduNetdiskDownload/test015/atlas/data555/keystore/UTC--2021-07-21T10-26-12.236878500Z--8becddb5fbe6f3d6b08450e2d33e48e63d6c4b29"
	boolPrint = true
)

func loadprivateCommon(keyfile string) (*ecdsa.PrivateKey, common.Address) {
	keyjson, err := ioutil.ReadFile(keyfile)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to read the keyfile at '%s': %v", keyfile, err))
	}
	password := "123456"
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		log.Fatal(fmt.Errorf("error decrypting key: %v", err))
	}
	priKey1 := key.PrivateKey
	return priKey1, crypto.PubkeyToAddress(priKey1.PublicKey)
}

func dialConnCommon() (*ethclient.Client, string) {
	ip := "localhost"
	port := 7445
	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	// Create an IPC based RPC connection to a remote node
	// "http://39.100.97.129:8545"
	conn, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the Abeychain client: %v", err)
	}
	return conn, url
}

func getChainsCommon(conn *ethclient.Client) []ethereum.Header {
	startNum := 1
	endNum := 100
	Headers := make([]ethereum.Header, 100)
	HeaderBytes := make([]bytes.Buffer, 100)
	for i := startNum; i <= endNum; i++ {
		Header, err := conn.HeaderByNumber(context.Background(), big.NewInt(int64(i)))
		if err != nil {
			log.Fatal(err)
		}
		convertChain(&Headers[i-1], &HeaderBytes[i-1], Header)
	}
	return Headers
}
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func convertChain(header *ethereum.Header, headerbyte *bytes.Buffer, e *types.Header) (*ethereum.Header, *bytes.Buffer) {
	if header == nil || e == nil {
		fmt.Println("header:", header, "e:", e)
		return header, headerbyte
	}
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
func queryAccountBalance(conn *ethclient.Client, from common.Address) {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	input := packInput("getBalance", from)
	msg := ethchain.CallMsg{From: from, To: &RelayerAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Fatal("method CallContract error", err)
	}

	PrintBalance(conn, from)
	fmt.Println()

	method, _ := abiRelayer.Methods["getBalance"]
	ret, err := method.Outputs.Unpack(output)
	if len(ret) != 0 {
		args := struct {
			register *big.Int
			locked   *big.Int
			unlocked *big.Int
			reward   *big.Int
			fine     *big.Int
		}{
			ret[0].(*big.Int),
			ret[1].(*big.Int),
			ret[2].(*big.Int),
			ret[3].(*big.Int),
			ret[4].(*big.Int),
		}
		fmt.Println("query successfully,your account:")
		fmt.Println("register amount: ", args.register)
		fmt.Println("locked amount:", args.locked)
		fmt.Println("unlocked amount:", args.unlocked)
		fmt.Println("reward amount:", args.reward)
		fmt.Println("fine amount:", args.fine)
	} else {
		fmt.Println("Contract query failed result len == 0")
	}
}

func printChangeBalance(old, new big.Float) {
	f, _ := old.Float64()
	old1 := big.NewFloat(f)
	f2, _ := new.Float64()
	new1 := big.NewFloat(f2)
	f3, _ := old1.Float64()
	c := big.NewFloat(f3)
	fmt.Printf("old balance:%v  new balance %v  change %v\n",
		old1, new1, c.Abs(c.Sub(c, new1)))
}
