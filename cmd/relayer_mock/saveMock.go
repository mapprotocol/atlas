package main

import (
	"context"
	"encoding/json"
	"fmt"
	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/chains/headers/ethereum"
	"github.com/mapprotocol/atlas/cmd/ethclient"
	"github.com/mapprotocol/atlas/core/rawdb"
	"gopkg.in/urfave/cli.v1"
	"log"
	"math/big"
)

func saveMock(ctx *cli.Context) error {
	debugInfo := debugInfo{}
	debugInfo.relayerData = []*relayerInfo{
		//{url: keystore2},
		//{url: keystore3},
		//{url: keystore4},
		//{url: keystore5},
	}
	debugInfo.preWork(ctx, true)
	debugInfo.saveForkBlock(ctx) //change this
	return nil
}

func (d *debugInfo) saveMock(ctx *cli.Context) {
	go d.atlasBackend()
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println("CURRENT EPOCH ========>", currentEpoch)
			currentEpoch1 := int(currentEpoch)
			for i := 0; i < len(d.step); i++ {
				if d.step[i] == currentEpoch1 {
					currentEpoch1 = i + 1
					break
				}
			}
			switch currentEpoch1 {
			case 1:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.queryDebuginfo(REWARD)
				d.doSave(d.ethData[:10])
				d.atlasBackendCh <- NEXT_STEP
			case 2:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.queryDebuginfo(REWARD)
				d.doSave(d.ethData[:10])
				d.atlasBackendCh <- NEXT_STEP
			case 3:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
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
		d.relayerData[k].realSave(conn, ChainTypeETH, marshal)
	}
}
func (r *relayerInfo) realSave(conn *ethclient.Client, chainType rawdb.ChainType, marshal []byte) bool {
	//header, err := conn.HeaderByNumber(context.Background(), nil)
	//if err != nil {
	//	log.Fatal(err)
	//	return false
	//}
	input := packInputStore("save", big.NewInt(int64(chainType)), big.NewInt(int64(ChainTypeMAP)), marshal)
	sendContractTransaction(conn, r.from, HeaderStoreAddress, nil, r.priKey, input)

	//input := packInputStore("save", chainType, "MAP", marshal)
	//msg := ethchain.CallMsg{From: r.from, To: &HeaderStoreAddress, Data: input}
	//_, err = conn.CallContract(context.Background(), msg, header.Number)
	//if err != nil {
	//	//log.Fatal("method CallContract error (realSave) :", err)
	//	fmt.Println("save false")
	//	return false
	//}
	//fmt.Println("save success")
	return true
}
func (d *debugInfo) saveByDifferentAccounts(ctx *cli.Context) {
	go d.atlasBackend()
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println("CURRENT EPOCH ========>", currentEpoch)
			currentEpoch1 := int(currentEpoch)
			for i := 0; i < len(d.step); i++ {
				if d.step[i] == currentEpoch1 {
					currentEpoch1 = i + 1
					break
				}
			}
			switch currentEpoch1 {
			case 1:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				//d.queryDebuginfo(QUERY_RELAYERINFO)
				//d.queryDebuginfo(BALANCE)
				//d.queryDebuginfo(REGISTER_BALANCE)
				//d.query_debugInfo(REWARD)
				d.doSave(d.ethData[:10])
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.doSave(d.ethData[:9])
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.doSave(d.ethData[:2])
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.doSave(d.ethData[:1])
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.atlasBackendCh <- NEXT_STEP
			case 2:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				//d.query_debugInfo(REWARD)
				d.doSave(d.ethData[10:20])
				d.atlasBackendCh <- NEXT_STEP
			case 3:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				//d.query_debugInfo(REWARD)
				d.doSave(d.ethData[10:20])
				d.atlasBackendCh <- NEXT_STEP
			case 4:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.atlasBackendCh <- NEXT_STEP
				return
			default:
				fmt.Println("over")
			}
		}
	}
}

func (d *debugInfo) saveForkBlock(ctx *cli.Context) {
	go d.atlasBackend()
	A, B := getForkBlock()
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println("CURRENT EPOCH ========>", currentEpoch)
			currentEpoch1 := int(currentEpoch)
			for i := 0; i < len(d.step); i++ {
				if d.step[i] == currentEpoch1 {
					currentEpoch1 = i + 1
					break
				}
			}
			switch currentEpoch1 {
			case 1:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.queryDebuginfo(REWARD)
				d.doSave(A[:10])
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.queryDebuginfo(REWARD)
				d.doSave(B[:8])
				d.atlasBackendCh <- NEXT_STEP
			case 2:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.queryDebuginfo(REWARD)
				d.doSave(d.ethData[:10])
				d.atlasBackendCh <- NEXT_STEP
			case 3:
				d.queryDebuginfo(CHAINTYPE_HEIGHT)
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.queryDebuginfo(REWARD)
				d.atlasBackendCh <- NEXT_STEP
				return
			default:
				fmt.Println("over")
			}
		}
	}
}

//  getCurrent type chain number by abi
func getCurrentNumberAbi(conn *ethclient.Client, chainType rawdb.ChainType, from common.Address) uint64 {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	input := packInputStore(CurNbrAndHash, big.NewInt(int64(chainType)))
	msg := ethchain.CallMsg{From: from, To: &HeaderStoreAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Fatal("method CallContract error", err)
	}
	method, _ := abiHeaderStore.Methods[CurNbrAndHash]
	ret, err := method.Outputs.Unpack(output)
	ret1 := ret[0].(*big.Int).Uint64()
	//ret2 := common.BytesToHash(ret[1].([]byte))
	//fmt.Println(ret2)
	return ret1
}

func packInputStore(abiMethod string, params ...interface{}) []byte {
	input, err := abiHeaderStore.Pack(abiMethod, params...)
	if err != nil {
		log.Fatal(abiMethod, " error ", err)
	}
	return input
}
