package main

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/keystore"
	"io/ioutil"
	"math/big"
	"testing"
)

func TestJsonTransferKey(t *testing.T) {
	//9 "111111"
	//53ee6ae610b7478404ae2fd07501cfd7688af191e22b553afafa293fbe364980
	keyfile := "D:/root/data_ibft1/keystore/UTC--2021-07-19T02-09-17.552426700Z--81f02fd21657df80783755874a92c996749777bf"

	password := "111111"
	keyjson, err := ioutil.ReadFile(keyfile)
	if err != nil {
		fmt.Printf("failed to read the keyfile at '%s': %v", keyfile, err)
	}
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		log.Error("", fmt.Errorf("error decrypting key: %v", err))
	}
	priKey := key.PrivateKey
	privHex := hex.EncodeToString(crypto.FromECDSA(priKey))
	fmt.Println("private key:", privHex)
	pkHash := common.Bytes2Hex(crypto.FromECDSAPub(&priKey.PublicKey))
	fmt.Println("public key:", pkHash)
	from := crypto.PubkeyToAddress(priKey.PublicKey)
	fmt.Println("address:", from)
}

func TestBigInt(t *testing.T) {
	c, _ := new(big.Int).SetString("2999999999999999999999", 10)
	c.Mul(c, big.NewInt(3))
	fmt.Println(c)
}
