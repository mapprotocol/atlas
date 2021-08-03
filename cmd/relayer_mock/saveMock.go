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
)

func saveMock(ctx *cli.Context) error {
	debugInfo := debugInfo{}
	debugInfo.relayerData = []*relayerInfo{
		{url: keystore2},
		{url: keystore3},
		{url: keystore4},
		{url: keystore5},
	}
	debugInfo.preWork(ctx, []int{1, 2, 3, 4}, true)
	debugInfo.saveByDifferentAccounts(ctx) //change this
	return nil
}

func (d *debugInfo) saveMock(ctx *cli.Context) {
	go d.atlasBackend()
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println("CURRENT EPOCH ========>", currentEpoch)
			switch currentEpoch {
			case 1:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(IMPAWN_BALANCE)
				d.queryDebuginfo(REWARD)
				d.doSave(d.ethData[:10])
				d.atlasBackendCh <- NEXT_STEP
			case 2:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(IMPAWN_BALANCE)
				d.queryDebuginfo(REWARD)
				d.doSave(d.ethData[:10])
				d.atlasBackendCh <- NEXT_STEP
			case 3:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(IMPAWN_BALANCE)
				d.queryDebuginfo(REWARD)
				d.atlasBackendCh <- NEXT_STEP
				return
			default:
				fmt.Println("over")
			}
		}
	}
}

func (d *debugInfo) doSave(chains []ethereum.Header) {
	fmt.Println("=================DO SAVE========================")
	marshal, _ := json.Marshal(chains)
	conn := d.client
	for k, _ := range d.relayerData {
		fmt.Println("ADDRESS:", d.relayerData[k].from)
		d.relayerData[k].realSave(conn, "ETH", marshal)
	}
}
func (r *relayerInfo) realSave(conn *ethclient.Client, chainType string, marshal []byte) bool {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
		return false
	}
	input := packInputStore("save", chainType, "MAP", marshal)
	msg := ethchain.CallMsg{From: r.from, To: &HeaderStoreAddress, Data: input}
	_, err = conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		//log.Fatal("method CallContract error (realSave) :", err)
		fmt.Println("save false")
		return false
	}
	fmt.Println("save success")
	return true
}
func (d *debugInfo) saveByDifferentAccounts(ctx *cli.Context) {
	go d.atlasBackend()
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println("CURRENT EPOCH ========>", currentEpoch)
			switch currentEpoch {
			case 1:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(IMPAWN_BALANCE)
				//d.query_debugInfo(REWARD)
				d.doSave(d.ethData[:10])
				d.atlasBackendCh <- NEXT_STEP
			case 2:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(IMPAWN_BALANCE)
				//d.query_debugInfo(REWARD)
				d.doSave(d.ethData[10:20])
				d.atlasBackendCh <- NEXT_STEP
			case 3:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(IMPAWN_BALANCE)
				//d.query_debugInfo(REWARD)
				d.doSave(d.ethData[10:20])
				d.atlasBackendCh <- NEXT_STEP
			case 4:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(IMPAWN_BALANCE)
				d.atlasBackendCh <- NEXT_STEP
				return
			default:
				fmt.Println("over")
			}
		}
	}
}

//  getCurrent type chain number by abi
func getCurrentNumberAbi(conn *ethclient.Client, chainType string, from common.Address) uint64 {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	input := packInputStore("currentHeaderNumber", chainType)
	msg := ethchain.CallMsg{From: from, To: &HeaderStoreAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Fatal("method CallContract error", err)
	}
	method, _ := abiHeaderStore.Methods["currentHeaderNumber"]
	ret, err := method.Outputs.Unpack(output)
	ret1 := ret[0].(*big.Int).Uint64()
	return ret1
}

func packInputStore(abiMethod string, params ...interface{}) []byte {
	input, err := abiHeaderStore.Pack(abiMethod, params...)
	if err != nil {
		log.Fatal(abiMethod, " error ", err)
	}
	return input
}
