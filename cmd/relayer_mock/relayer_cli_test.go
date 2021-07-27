package main

import (
	"fmt"
	"log"
	"math/big"
	"testing"
)

// test register
func Test_register(t *testing.T) {
	fee := uint64(0)
	value := ethToWei(false)
	priKey, from = loadprivateCommon(keystore1)
	conn, _ := dialConnCommon()
	pkey, pk, _ := getPubKey(priKey)
	aBalance := PrintBalance(conn, from)
	fmt.Printf("Fee: %v \nPub key:%v\nvalue:%v\n \n", fee, pkey, value)
	input := packInput("register", pk, new(big.Int).SetUint64(fee), value)
	txResult := sendContractTransaction(conn, from, RelayerAddress, nil, priKey, input)
	getResult(conn, txResult, true, from)
	relayerBool := queryIsRegister(conn, from)
	fmt.Printf("isrelayers:%v  \n", relayerBool)
	bBalance := PrintBalance(conn, from)
	printChangeBalance(*aBalance, *bBalance)
}

func Test_withdraw(t *testing.T) {
	priKey, from = loadprivateCommon(keystore1)
	conn, _ := dialConnCommon()
	a := PrintBalance(conn, from)
	withdraw(conn, from, priKey)
	b := PrintBalance(conn, from)
	printChangeBalance(*a, *b)
}
func Test_append(t *testing.T) {
	priKey, from = loadprivateCommon(keystore1)
	conn, _ := dialConnCommon()
	a := PrintBalance(conn, from)
	err := Append(conn, from, priKey)
	if err != nil {
		log.Fatal(err)
	}
	b := PrintBalance(conn, from)
	printChangeBalance(*a, *b)
}
