package main

import (
	"fmt"
	"log"
	"testing"
)

// test register
func Test_register(t *testing.T) {
	conn, _ := dialConnCommon()
	aBalance, txResult, from := registerCommon(conn, keystore1)
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
