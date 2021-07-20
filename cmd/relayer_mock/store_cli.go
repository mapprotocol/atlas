package main

import (
	"context"
	"github.com/mapprotocol/atlas/cmd/ethclient"
)

// 1. getCurrent type chain number by abi
func getCurrentNumberAbi(conn *ethclient.Client, chainType string) {
	input := packInputStore("currentHeaderNumber", chainType)
	txHash := sendContractTransaction(conn, from, HeaderStoreAddress, nil, priKey, input)
	getResult(conn, txHash, true, false)
}

//  2. getCurrent type chain number by rpc
func getCurrentNumberRpc(conn *ethclient.Client, chainType string) (uint64, error) {
	return conn.BlockNumberChains(context.Background(), chainType)
}

func packInputStore(abiMethod string, params ...interface{}) []byte {
	input, err := abiHeaderStore.Pack(abiMethod, params...)
	if err != nil {
		printError(abiMethod, " error ", err)
	}
	return input
}
