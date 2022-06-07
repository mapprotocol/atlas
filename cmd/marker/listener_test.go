package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/mapprotocol/atlas/accounts"
	"github.com/mapprotocol/atlas/helper/bls"
	"github.com/mapprotocol/atlas/helper/fileutils"
	"github.com/mapprotocol/atlas/marker/genesis"

	"github.com/mapprotocol/atlas/accounts/keystore"
	"github.com/mapprotocol/atlas/atlas"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/marker/env"
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

	if err := fileutils.WriteJson(genesisAlloc, "D:/work/root/atlasEnv/dumpStateDb.json"); err != nil {
		t.Fatalf("err==>%s", err)
	}
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

// simulation Account new
//0x1a7559d3ca2e6d4ee76bf97e816c21319e31a8ff58368c747fd8909bf37b48db0fc69bc7c6fffc0665ff1801fb17afe79ae31042411d3eb600ab73bb02f3e8b32fb4218ab8a7018b6300ea6500ef438817eed6c986901eb212d4c6de093ef63020b20eb8e1ed4365d33519cdc1c290ae7a386ff9743c57b1b420be20b94698b7
func Test_autoGenerateMarkerCfg(t *testing.T) {
	account := []string{"0xB16561A66B6439944DAf0388b5E6a2D3D0a49e12", "0xdC757c8e3b8800d977a34f802131FAA870d264c4", "0x257Cc34FB139A2db4Da96496Be03358d89e52d95", "0xac146d6629F8C3B8F2e830275B583C5402032472"}
	path1 := "/Users/zhangwei/work/atlasEnv/data/data_ibft1/keystore/UTC--2021-09-08T08-00-15.473724074Z--1c0edab88dbb72b119039c4d14b1663525b3ac15"
	path2 := "/Users/zhangwei/work/atlasEnv/data/data_ibft1/keystore/UTC--2021-09-08T10-12-17.687481942Z--16fdbcac4d4cc24dca47b9b80f58155a551ca2af"
	path3 := "/Users/zhangwei/work/atlasEnv/data/data_ibft1/keystore/UTC--2021-09-08T10-16-18.520295371Z--2dc45799000ab08e60b7441c36fcc74060ccbe11"
	path4 := "/Users/zhangwei/work/atlasEnv/data/data_ibft1/keystore/UTC--2021-09-08T10-16-35.698273293Z--6c5938b49bacde73a8db7c3a7da208846898bff5"
	//path5 := "/Users/zhangwei/work/atlasEnv/keystore/UTC--2022-05-27T12-24-07.345965000Z--05d0cfd882185deb9b3e0ea7872ad332acb9e31d"
	marker := genesis.MarkerInfo{}
	marker.AdminAddress = "0x1c0edab88dbb72b119039c4d14b1663525b3ac15"
	for i, path := range []string{path1, path2, path3, path4} {
		Password := ""
		keyjson, err := ioutil.ReadFile(path)
		if err != nil {
			t.Error("loadPrivate ReadFile", fmt.Errorf("failed to read the keyfile at '%s': %v", path, err))
		}
		key, err := keystore.DecryptKey(keyjson, Password)
		if err != nil {
			t.Error("loadPrivate DecryptKey", fmt.Errorf("error decrypting key: %v", err))
		}
		priKey1 := key.PrivateKey
		publicAddr := crypto.PubkeyToAddress(priKey1.PublicKey)
		var addr common.Address
		addr.SetBytes(publicAddr.Bytes())
		accountBls := env.Account{
			Address:    addr,
			PrivateKey: priKey1,
		}

		if err != nil {
			t.Error("Failed to create account: ", err)
		}
		//blsProofOfPossession := accountBls.MustBLSProofOfPossession()
		blsPubKey, err := accountBls.BLSPublicKey()
		if err != nil {
			t.Error("Failed to create account: ", err)
		}
		blsPubKeyText, err := blsPubKey.MarshalText()
		if err != nil {
			t.Error("Failed to create account: ", err)
		}
		blsG1PubKey, err := accountBls.BLSG1PublicKey()
		if err != nil {
			t.Error("Failed to create account: ", err)
		}
		blsG1PubKeyText, err := blsG1PubKey.MarshalText()
		if err != nil {
			t.Error("Failed to create account", err)
		}
		fmt.Printf("\nYour new key was generated\n\n")
		fmt.Printf("Address:   %s\n", accountBls.Address.Hex())
		fmt.Printf("PublicKey:   %s\n", hexutil.Encode(accountBls.PublicKey()))
		fmt.Printf("BLS Public key:%d   %s\n", len(blsPubKey), blsPubKeyText)
		fmt.Printf("BLS G1 Public key:%d   %s\n", len(blsG1PubKey), blsG1PubKeyText)

		// -------------------------- ECDSASignature  ---------------------------------
		hash := accounts.TextHash(crypto.Keccak256(common.HexToAddress(account[i]).Bytes()))
		sig, err := crypto.Sign(hash, priKey1)
		if err != nil {
			panic(err)
		}
		// --------------------------  blsProofOfPossession -----------------------------
		blsPrivateKey, _ := bls.CryptoType().ECDSAToBLS(priKey1)
		privateKey, _ := bls.DeserializePrivateKey(blsPrivateKey)
		signature, err := bls.UnsafeSign(privateKey, common.HexToAddress(account[i]).Bytes())
		blsProofOfPossession := signature.Marshal()
		if err != nil {
			panic(err)
		}
		serializedPrivateKey, _ := privateKey.Serialize()
		publicKey, _ := bls.CryptoType().PrivateToPublic(serializedPrivateKey)
		pk, err := bls.UnmarshalPk(publicKey[:])
		//test
		if err := bls.VerifyUnsafe(pk, common.HexToAddress(account[i]).Bytes(), signature); err != nil {
			panic(err)
		}
		fmt.Printf("BLSProofOfPossession: %d  %s\n", len(blsProofOfPossession), hexutil.Encode(blsProofOfPossession))
		marker.Validators = append(marker.Validators, genesis.AccoutInfo{
			Address:              account[i],
			SignerAddress:        accountBls.Address.Hex(),
			ECDSASignature:       hexutil.Encode(sig),
			PublicKeyHex:         hexutil.Encode(accountBls.PublicKey()),
			BLSPubKey:            hexutil.Encode(blsPubKey[:]),
			BLSG1PubKey:          hexutil.Encode(blsG1PubKey[:]),
			BLSProofOfPossession: hexutil.Encode(blsProofOfPossession),
		})
	}
	fileutils.WriteJson(marker, "/Users/zhangwei/work/atlas/marker/config/markerConfig.json")
}

func Test_sign(T *testing.T) {
	account := common.HexToAddress("0x6621F2b6Da2BEd64b5fFBD6C5b2138547f44C8f9")
	singerPriv := "564e1166e9c1d51f00e01b230f8a33a944c4c742fc839add8daada2cffc0e022"
	privECDSA, err := crypto.ToECDSA(common.FromHex(singerPriv))
	fmt.Println("===singer ===", crypto.PubkeyToAddress(privECDSA.PublicKey))
	priv, err := bls.DeserializePrivateKey(common.FromHex(singerPriv))
	if err != nil {
		panic(err)
	}
	pub := priv.ToPublic()
	if err != nil {
		panic(err)
	}
	pop, _ := bls.UnsafeSign(priv, account.Bytes())
	popBytes := pop.Marshal()
	T.Log(":", "pop:", hexutil.Encode(popBytes))
	// test
	err = bls.VerifyUnsafe(pub, account.Bytes(), pop)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("finish")
}
