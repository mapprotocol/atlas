package tools

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/accounts/keystore"
	"github.com/mapprotocol/atlas/core/types"
	blscrypto "github.com/mapprotocol/atlas/helper/bls"
	"io/ioutil"
	"math/big"
	"testing"
)

func TestMakeValidator(t *testing.T) {

	// There are at least four addresses
	// In here, you should use private key created by yourself, else the data is not valid
	ads := make([]common.Address, 0)
	sliceOfAddressHexStr := []string{
		"0x1c0eDab88dbb72B119039c4d14b1663525b3aC15",
		"0x16FdBcAC4D4Cc24DCa47B9b80f58155a551ca2aF",
		"0x2dC45799000ab08E60b7441c36fCC74060Ccbe11",
		"0x6C5938B49bACDe73a8Db7C3A7DA208846898BFf5",
	}

	for _, s := range sliceOfAddressHexStr {
		ads = append(ads, common.HexToAddress(s))
	}

	// account publick key, use your public key to instead of them
	apks := make([]blscrypto.SerializedPublicKey, 0)
	sliceOfBlsPubKeyHexStr := []string{
		"32071fff6599fcdefb78d8048abf7d32165e4dad0a00d7667ba4e1933a6f1bff00",
		"66b74fbfc9c23963a9a21e12d79422fc288b7598b58f23d4ec04ea2657a05a9901",
		"40cdae9b90b80179ac73341dd83974fa6dd85f921080770241df8b4f3eb2244e01",
		"82b9df317d21429c6f0b74c96c21a610483be1d234c2815c50be454a689c35ae01",
	}

	for _, s := range sliceOfBlsPubKeyHexStr {
		pk1 := blscrypto.SerializedPublicKey{}
		pk1.UnmarshalText([]byte(s))
		apks = append(apks, pk1)
	}

	ist := types.IstanbulExtra{
		AddedValidators:           ads,
		AddedValidatorsPublicKeys: apks,
		RemovedValidators:         big.NewInt(0),
		Seal:                      []byte(""),
		AggregatedSeal:            types.IstanbulAggregatedSeal{},
		ParentAggregatedSeal:      types.IstanbulAggregatedSeal{},
	}

	payload, err := rlp.EncodeToBytes(&ist)
	if err != nil {
		fmt.Printf("encode failed: %#v", err.Error())
		return
	}

	finalExtra := append(bytes.Repeat([]byte{0x00}, types.IstanbulExtraVanity), payload...)

	fmt.Printf("finalExtra :%s\n", hexutil.Encode(finalExtra))
}

func TestMakeKey(t *testing.T) {
	privateKey, _ := crypto.GenerateKey()
	blsPrivateKey, _ := blscrypto.CryptoType().ECDSAToBLS(privateKey)
	blsPublicKey, _ := blscrypto.CryptoType().PrivateToPublic(blsPrivateKey)
	bp, _ := blsPublicKey.MarshalText()
	from := crypto.PubkeyToAddress(privateKey.PublicKey)
	privHex := hex.EncodeToString(crypto.FromECDSA(privateKey))
	fmt.Println("address:", from)
	fmt.Println("private key:    ", privHex)
	fmt.Printf("bls public key: %s\n", bp)

	privKeyHex := hex.EncodeToString(blsPrivateKey)
	pubkeyHash := common.Bytes2Hex(crypto.FromECDSAPub(&privateKey.PublicKey))
	fmt.Println("bls private key:", privKeyHex)
	fmt.Println("public key:     ", pubkeyHash)
}

func TestMakeKeyFromJson(t *testing.T) {
	keyfile := "../datenode/keystore/UTC--2021-12-02T07-27-59.482004300Z--1f39d97a8f697502884fe01cf23dba4eb66e0481"
	password := ""
	keyjson, err := ioutil.ReadFile(keyfile)
	if err != nil {
		fmt.Printf("failed to read the keyfile at '%s': %v", keyfile, err)
	}
	key, _ := keystore.DecryptKey(keyjson, password)
	privateKey := key.PrivateKey
	blsPrivateKey, _ := blscrypto.CryptoType().ECDSAToBLS(privateKey)
	blsPublicKey, _ := blscrypto.CryptoType().PrivateToPublic(blsPrivateKey)
	bp, _ := blsPublicKey.MarshalText()
	from := crypto.PubkeyToAddress(privateKey.PublicKey)
	privHex := hex.EncodeToString(crypto.FromECDSA(privateKey))
	fmt.Println("address:", from)
	fmt.Println("private key:    ", privHex)
	fmt.Printf("bls public key: %s\n", bp)

	privKeyHex := hex.EncodeToString(blsPrivateKey)
	pubkeyHash := common.Bytes2Hex(crypto.FromECDSAPub(&privateKey.PublicKey))
	fmt.Println("bls private key:", privKeyHex)
	fmt.Println("public key:     ", pubkeyHash)
}
