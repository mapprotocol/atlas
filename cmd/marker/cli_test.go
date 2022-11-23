package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/keystore"
)

func TestJsonTransferKey(t *testing.T) {
	//9 "111111"
	//53ee6ae610b7478404ae2fd07501cfd7688af191e22b553afafa293fbe364980
	//keyfile := "/Users/user0/work/atlasEnv/data/data_ibft1/keystore/UTC--2022-05-11T03-31-08.562982000Z--6621f2b6da2bed64b5ffbd6c5b2138547f44c8f9"
	//path1 := "/Users/user0/work/atlasEnv/data/data_ibft1/keystore/UTC--2021-09-08T08-00-15.473724074Z--1c0edab88dbb72b119039c4d14b1663525b3ac15"
	//path2 := "/Users/user0/work/atlasEnv/data/data_ibft1/keystore/UTC--2021-09-08T10-12-17.687481942Z--16fdbcac4d4cc24dca47b9b80f58155a551ca2af"
	//path3 := "/Users/user0/work/atlasEnv/data/data_ibft1/keystore/UTC--2021-09-08T10-16-18.520295371Z--2dc45799000ab08e60b7441c36fcc74060ccbe11"
	//path4 := "/Users/user0/work/atlasEnv/data/data_ibft1/keystore/UTC--2021-09-08T10-16-35.698273293Z--6c5938b49bacde73a8db7c3a7da208846898bff5"
	path5 := "/Users/user0/work/atlasEnv/keystore/UTC--2022-05-27T12-24-07.345965000Z--05d0cfd882185deb9b3e0ea7872ad332acb9e31d"
	password := "111111"
	for _, keyfile := range []string{path5} {
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
}

func TestBigInt(t *testing.T) {
	privHex := hexutil.Encode([]byte{1, 2, 3, 4, 5, 6, 7})
	fmt.Println("private key:", privHex)
	c, _ := new(big.Int).SetString("2999999999999999999999", 10)
	c.Mul(c, big.NewInt(3))
	fmt.Println(c)
}
