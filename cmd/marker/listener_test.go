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
	var (
		adminAddress = "0xef021f15d188ad28625517a8d73cd20ce743a32d"
		account      = []string{
			"0xef021f15d188ad28625517a8d73cd20ce743a32d",
			"0x72610a6c8066e5735a29fe17b30c7482660517c0",
			"0x4a437f62f9ef8f82df6828cae0b30e8f3f8f6de4",
			"0xd571c16af4997ef10cc14be503af5d892a2e0722",
		}
		signersPath = []string{
			"/Users/t/data/atlas-1/keystore/UTC--2022-08-26T10-58-25.448097000Z--85e6574b9d2a169111ae0b38e1b31bf1094ef1e4",
			"/Users/t/data/atlas-1/keystore/UTC--2022-08-26T10-58-38.317733000Z--622f6ca33bb4ea1d48cf5c067ef6826f4ac18a75",
			"/Users/t/data/atlas-1/keystore/UTC--2022-08-26T10-58-42.772412000Z--9d71e0f25ba36f82d70dae874f99ee3629f8b015",
			"/Users/t/data/atlas-1/keystore/UTC--2022-08-26T10-58-47.879731000Z--d1414da995fd4f04ceb33c4366be5bfae5a01523",
		}
	)

	accountInfos := make([]genesis.AccoutInfo, 0, len(account))
	for i, path := range signersPath {
		Password := ""
		signerFile, err := ioutil.ReadFile(path)
		if err != nil {
			t.Error("loadPrivate ReadFile", fmt.Errorf("failed to read the keyfile at '%s': %v", path, err))
		}
		signerKey, err := keystore.DecryptKey(signerFile, Password)
		if err != nil {
			t.Error("loadPrivate DecryptKey", fmt.Errorf("error decrypting key: %v", err))
		}
		signerPrivateKey := signerKey.PrivateKey
		signerAddr := crypto.PubkeyToAddress(signerPrivateKey.PublicKey)
		var addr common.Address
		addr.SetBytes(signerAddr.Bytes())
		singerAccount := env.Account{
			Address:    addr,
			PrivateKey: signerPrivateKey,
		}

		if err != nil {
			t.Error("Failed to create account: ", err)
		}
		//blsProofOfPossession := singerAccount.MustBLSProofOfPossession()
		singerBlsPubKey, err := singerAccount.BLSPublicKey()
		if err != nil {
			t.Error("Failed to create account: ", err)
		}
		signerBlsPubKeyText, err := singerBlsPubKey.MarshalText()
		if err != nil {
			t.Error("Failed to create account: ", err)
		}
		signerBlsG1PubKey, err := singerAccount.BLSG1PublicKey()
		if err != nil {
			t.Error("Failed to create account: ", err)
		}
		signerBlsG1PubKeyText, err := signerBlsG1PubKey.MarshalText()
		if err != nil {
			t.Error("Failed to create account", err)
		}
		fmt.Printf("\nYour new key was generated\n\n")
		fmt.Printf("Address:   %s\n", singerAccount.Address.Hex())
		fmt.Printf("PublicKey:   %s\n", hexutil.Encode(singerAccount.PublicKey()))
		fmt.Printf("BLS Public key:%d   %s\n", len(singerBlsPubKey), signerBlsPubKeyText)
		fmt.Printf("BLS G1 Public key:%d   %s\n", len(signerBlsG1PubKey), signerBlsG1PubKeyText)

		// -------------------------- ECDSASignature  ---------------------------------
		hash := accounts.TextHash(crypto.Keccak256(common.HexToAddress(account[i]).Bytes()))
		sig, err := crypto.Sign(hash, signerPrivateKey)
		if err != nil {
			t.Fatal(err.Error())
		}
		// --------------------------  blsProofOfPossession -----------------------------
		signerBlsPrivateKey, err := bls.CryptoType().ECDSAToBLS(signerPrivateKey)
		if err != nil {
			t.Fatal(err.Error())
		}
		secretKey, err := bls.DeserializePrivateKey(signerBlsPrivateKey)
		if err != nil {
			t.Fatal(err.Error())
		}
		signature, err := bls.UnsafeSign(secretKey, common.HexToAddress(account[i]).Bytes())
		if err != nil {
			t.Fatal(err.Error())
		}
		blsProofOfPossession := signature.Marshal()
		serializedPrivateKey, err := secretKey.Serialize()
		if err != nil {
			t.Fatal(err.Error())
		}
		publicKey, err := bls.CryptoType().PrivateToPublic(serializedPrivateKey)
		if err != nil {
			t.Fatal(err.Error())
		}
		pk, err := bls.UnmarshalPk(publicKey[:])
		if err != nil {
			t.Fatal(err.Error())
		}
		//test
		if err := bls.VerifyUnsafe(pk, common.HexToAddress(account[i]).Bytes(), signature); err != nil {
			t.Fatal(err.Error())
		}
		fmt.Printf("BLSProofOfPossession: %d  %s\n", len(blsProofOfPossession), hexutil.Encode(blsProofOfPossession))
		accountInfos = append(accountInfos, genesis.AccoutInfo{
			Address:              account[i],
			SignerAddress:        singerAccount.Address.Hex(),
			ECDSASignature:       hexutil.Encode(sig), //
			PublicKeyHex:         hexutil.Encode(singerAccount.PublicKey()),
			BLSPubKey:            hexutil.Encode(singerBlsPubKey[:]),
			BLSG1PubKey:          hexutil.Encode(signerBlsG1PubKey[:]),
			BLSProofOfPossession: hexutil.Encode(blsProofOfPossession), //
		})
	}

	marker := genesis.MarkerInfo{
		AdminAddress: adminAddress,
		Validators:   accountInfos,
	}

	if err := fileutils.WriteJson(marker, "marker_config.json"); err != nil {
		t.Fatal(err.Error())
	}
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
