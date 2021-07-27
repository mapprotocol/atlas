package main

import (
	"context"
	"encoding/json"
	"fmt"
	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/chains/headers/ethereum"
	"github.com/mapprotocol/atlas/cmd/ethclient"
	"gopkg.in/urfave/cli.v1"
	"log"
	"math/big"
	"time"
)

var (
	saveManyTimesCommand = cli.Command{
		Name:   "saveManyTimes",
		Usage:  "saveManyTimes ",
		Action: MigrateFlags(saveManyTimes),
		Flags:  relayerflags,
	}
)

// 1. getCurrent type chain number by abi
func getCurrentNumberAbi(conn *ethclient.Client, chainType string) uint64 {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	input := packInputStore("currentHeaderNumber", chainType)
	msg := ethchain.CallMsg{From: from, To: &HeaderStoreAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		printError("method CallContract error", err)
	}
	method, _ := abiHeaderStore.Methods["currentHeaderNumber"]
	ret, err := method.Outputs.Unpack(output)
	ret1 := ret[0].(*big.Int).Uint64()
	return ret1
}

//  2. getCurrent type chain number by rpc
func getCurrentNumberRpc(conn *ethclient.Client, chainType string) (uint64, error) {
	return conn.BlockNumberChains(context.Background(), chainType)
}

// 3. test save
func save(_ *cli.Context) error {
	_, from1 := loadprivateCommon(keystore1)
	connEth, _ := dialEthConn()
	chains := getChainsCommon(connEth)
	marshal, _ := json.Marshal(chains[:10])
	conn, _ := dialConnCommon()
	bool := realSave(conn, "ETH", marshal, from1)
	fmt.Printf("save %v\n", bool)
	queryRegisterInfo(conn, from1, "1")
	return nil
}

//Real storage
func realSave(conn *ethclient.Client, chainType string, marshal []byte, from common.Address) bool {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
		return false
	}
	input := packInputStore("save", chainType, "MAP", marshal)
	msg := ethchain.CallMsg{From: from, To: &HeaderStoreAddress, Data: input}
	_, err = conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Fatal("method CallContract error (realSave) :", err)
		return false
	}
	return true
}

func SaveByNum(conn *ethclient.Client, number int, from common.Address, chains []ethereum.Header) {
	marshal, _ := json.Marshal(chains[:number])
	bool := realSave(conn, "ETH", marshal, from)
	fmt.Printf("save %v\n", bool)
}

func packInputStore(abiMethod string, params ...interface{}) []byte {
	input, err := abiHeaderStore.Pack(abiMethod, params...)
	if err != nil {
		printError(abiMethod, " error ", err)
	}
	return input
}

//Store many times at kinds Epoch
func saveManyTimes(ctx *cli.Context) error {
	conn := getConn(ctx)
	priKey, from = loadprivateCommon(keystore1)
	register(ctx, conn, from)
	boolPrint = false
	_, _, curEpoch, err := queryRegisterInfo(conn, from, "myAccount")
	if err != nil {
		log.Fatal(err)
	}
	connEth, _ := dialEthConn()

	chains := getChainsCommon(connEth)

	curEpoch2 := big.NewInt(curEpoch.Int64())
	save := func() {
		aBalance := PrintBalance(conn, from)

		marshal, _ := json.Marshal(chains[:10])
		bool := realSave(conn, "ETH", marshal, from)
		fmt.Printf("save %v\n", bool)
		bBalance := PrintBalance(conn, from)
		printChangeBalance(*aBalance, *bBalance)

		aBalance = PrintBalance(conn, from)
		marshal, _ = json.Marshal(chains[:10])
		bool = realSave(conn, "ETH", marshal, from)
		fmt.Printf("save %v\n", bool)
		bBalance = PrintBalance(conn, from)
		printChangeBalance(*aBalance, *bBalance)

		aBalance = PrintBalance(conn, from)
		marshal, _ = json.Marshal(chains[10:20])
		bool = realSave(conn, "ETH", marshal, from)
		fmt.Printf("save %v\n", bool)
		bBalance = PrintBalance(conn, from)
		printChangeBalance(*aBalance, *bBalance)

		aBalance = PrintBalance(conn, from)
		marshal, _ = json.Marshal(chains[15:25])
		bool = realSave(conn, "ETH", marshal, from)
		fmt.Printf("save %v\n", bool)
		bBalance = PrintBalance(conn, from)
		printChangeBalance(*aBalance, *bBalance)

		aBalance = PrintBalance(conn, from)
		marshal, _ = json.Marshal(chains[25:50])
		bool = realSave(conn, "ETH", marshal, from)
		fmt.Printf("save %v\n", bool)
		bBalance = PrintBalance(conn, from)
		printChangeBalance(*aBalance, *bBalance)
	}
	count := 0
	oldbalance := PrintBalance(conn, from)
	if curEpoch2.CmpAbs(curEpoch) == 0 {
		fmt.Println("================== save ============================curEpoch: ", curEpoch)
		save()
		curEpoch2.Add(curEpoch2, common.Big1)
	}
	for {
		_, _, curEpoch, err = queryRegisterInfo(conn, from, "001:")
		if curEpoch2.Cmp(curEpoch) == 0 {
			fmt.Println("================== query ================curEpoch:", curEpoch)
			queryAccountBalance(conn, from)
			curBalance := PrintBalance(conn, from)
			printChangeBalance(*oldbalance, *curBalance)
			fmt.Println("================== save ================curEpoch:", curEpoch)
			a := PrintBalance(conn, from)
			save()
			b := PrintBalance(conn, from)
			printChangeBalance(*a, *b)
			queryAccountBalance(conn, from)
			oldbalance = curBalance
			count++
			curEpoch2.Add(curEpoch2, common.Big1)
			if count > 3 {
				break
			}
		}
		time.Sleep(time.Second)
	}

	return nil
}
