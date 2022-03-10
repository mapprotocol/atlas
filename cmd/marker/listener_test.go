package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/mapprotocol/atlas/atlas"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/core/state"
	"io/ioutil"
	"math/big"
	"testing"
)

func Test_dumpStateDb(t *testing.T) {
	client, err := rpc.Dial("https://poc2-rpc.maplabs.io")
	if err != nil {
		fmt.Println(err)
	}
	number := "0x173050"
	start := common.Hash{}.Bytes()
	genesisAlloc := make(map[common.Address]chain.GenesisAccount)
	for i := 0; i < 100; i++ {
		t.Log("index", i)
		var IteratorDump state.IteratorDump
		//blockNrOrHash rpc.BlockNumberOrHash, start []byte, maxResults int, nocode, nostorage, incompletes bool
		if err = client.Call(&IteratorDump, "debug_accountRange",
			number, start, uint64(atlas.AccountRangeMaxResults), false, false, false); err != nil {
			t.Log("err==>", err)
		}
		start = IteratorDump.Next
		if start == nil {
			t.Errorf("its over")
			break
		}

		for acc, dumpAcc := range IteratorDump.Accounts {
			if _, duplicate := genesisAlloc[acc]; duplicate {
				t.Fatalf("pagination test failed:  results should not overlap")
			}
			var account chain.GenesisAccount

			if dumpAcc.Balance != "" {
				account.Balance, _ = new(big.Int).SetString(dumpAcc.Balance, 10)
			}

			if dumpAcc.Code != nil {
				account.Code = dumpAcc.Code
			}
			if len(dumpAcc.Storage) > 0 {
				account.Storage = make(map[common.Hash]common.Hash)
				for k, v := range dumpAcc.Storage {
					account.Storage[k] = common.HexToHash(v)
				}
			}
			genesisAlloc[acc] = account
		}
	}

	if err := WriteJson(genesisAlloc, "D:/work/root/atlasEnv/dumpStateDb.json"); err != nil {
		t.Fatalf("err==>%s", err)
	}
}
func WriteJson(in interface{}, filepath string) error {
	byteValue, err := json.MarshalIndent(in, " ", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath, byteValue, 0644)
}

func Test_proxy(t *testing.T) {
	client, err := rpc.Dial("localhost:8549")
	if err != nil {
		fmt.Println(err)
	}
	var ret interface{}
	//blockNrOrHash rpc.BlockNumberOrHash, start []byte, maxResults int, nocode, nostorage, incompletes bool
	if err = client.Call(&ret, "istanbul_addProxy"); err != nil {
		t.Log("err==>", err)
	}
}
func Test_ProxiedValidator(t *testing.T) {
	client, err := rpc.Dial("http://localhost:8545")
	if err != nil {
		fmt.Println(err)
	}
	var ret interface{}
	//blockNrOrHash rpc.BlockNumberOrHash, start []byte, maxResults int, nocode, nostorage, incompletes bool
	url := "enode://290ef09419dc28a367a93a4266c646e379ba4dd0bd2fae7f86277d3d4c330179ee2d70b282de4a5d0d8cc1130c36a88b8fe61baa1726dc41f16e192a3d6af8e4@127.0.0.1:31004"
	externalUrl := "enode://99ea9aab0498007f662ca5122e39e7353db3f69b9f1aebd96fcd33bd1a098c4cdb41b97c479d7eecd9d5def59ce7e9f0c6534ccca95811b480e39db37f424215@127.0.0.1:31005"
	if err = client.Call(&ret, "istanbul_addProxy", url, externalUrl); err != nil {
		t.Log("err==>", err)
	}
	fmt.Println(ret)
}
