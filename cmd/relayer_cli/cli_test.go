package main

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mapprotocol/atlas/accounts/keystore"
	"io/ioutil"
	"testing"
)

func TestJsonTransferKey(t *testing.T) {
	keyfile := "../../data/keystore/UTC--2021-07-21T10-18-19.019189000Z--b5ac31a4a887e9f773b5fd0aba3fc0fe95c2a750"
	password := "111111"
	keyjson, err := ioutil.ReadFile(keyfile)
	if err != nil {
		fmt.Printf("failed to read the keyfile at '%s': %v", keyfile, err)
	}
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		printError(fmt.Errorf("error decrypting key: %v", err))
	}
	priKey = key.PrivateKey
	privHex := hex.EncodeToString(crypto.FromECDSA(priKey))
	fmt.Println("private key:", privHex)
	pkHash := common.Bytes2Hex(crypto.FromECDSAPub(&priKey.PublicKey))
	fmt.Println("public key:", pkHash)
	from = crypto.PubkeyToAddress(priKey.PublicKey)
	fmt.Println("address:", from)
}
