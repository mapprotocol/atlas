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
	keyfile := "../../data3/keystore/UTC--2021-08-06T03-32-55.462419725Z--d0d471aaea6bc0321e9c7f7696aac6c8626d1420"
	password := ""
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
