package main

import (
	"fmt"
	"log"
	"math/big"
	"testing"
)

// test register
func Test_register(t *testing.T) {
	conn, _ := dialConn()
	aBalance, from := registerCommon(conn, keystore1)
	relayerBool := queryIsRegister(conn, from)
	fmt.Printf("isrelayers:%v  \n", relayerBool)
	bBalance := getBalance(conn, from)
	printChangeBalance(*aBalance, *bBalance)
}

func Test_withdraw(t *testing.T) {
	priKey, from := loadprivateCommon(keystore1)
	conn, _ := dialConn()
	a := getBalance(conn, from)
	err := withdraw(conn, from, priKey)
	if err != nil {
		log.Fatal(err)
	}
	b := getBalance(conn, from)
	printChangeBalance(*a, *b)
}

func Test_append(t *testing.T) {
	priKey, from := loadprivateCommon(keystore1)
	conn, _ := dialConn()
	a := getBalance(conn, from)
	err := Append(conn, from, priKey)
	if err != nil {
		log.Fatal(err)
	}
	b := getBalance(conn, from)
	printChangeBalance(*a, *b)
}

func Test_moneyChange(t *testing.T) {
	a := new(big.Int).SetInt64(100000)
	b := new(big.Int).SetInt64(100001)
	fmt.Println(a.Abs(a.Sub(a, b)))
	a = new(big.Int).SetInt64(100001)
	b = new(big.Int).SetInt64(100000)
	fmt.Println(a.Abs(a.Sub(a, b)))
	a = new(big.Int).SetInt64(100000)
	b = new(big.Int).SetInt64(100000)
	fmt.Println(a.Abs(a.Sub(a, b)))
}

func Test_getEthChains(t *testing.T) {
	getEthChains()
}
