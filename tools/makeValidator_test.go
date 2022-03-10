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
	bn256 "github.com/mapprotocol/bn256/bls"
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
		"0xbe77f945929d5dd3fe99aa825df0f5b1e8ea11786333b4492a8624a4d08dcee0e89df327359e8ec3f2d8ae01e938b7003414aa2d6523ffa02fde42b278cbae311fd39f1fbcad8e3188442ea31dee662389599751f8e73b99215cefc2e0003f81",
		"0x4f38a71fb13ab20f7bbfc2749ab15d775b7729842d967ca4f4115d1fcb3f378c892d073344f84e2abd8995a16eeee8004f4e588c30261e08a5dae70c581f904ea86b574bfe279222cf6b7913bebb0d3bd6c2bbe2e2ea1d338f145c4d95b99201",
		"0x8cf3bfcbfc76e9a99b70cad65ae51f8a8972e3e230445a55c8cf6b96dea7a2d0d970e3545e1316554d5d3b0a53582800ad4de92e3b06b62aa6f7677fdc2885a90b75fd80e2db2775512d8f3d3900aabae5b0525786d65615994b07afe7f69481",
		"0x1bbb8eb14a7f5dddc9de3356ce4247dab8e554fa83cd33e663db148b5d2dd14485f090978c84074154b450329de06b018eac04113ede1eedadf891ee862877af92a648c162be62182db90e8c83f8fd154fc14f13676bcb1fe3503260b6261a01",
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

func TestMakeKeyFromJson1(t *testing.T) {
	keyfile := "../node4/keystore/UTC--2021-09-08T08-00-15.473724074Z--1c0edab88dbb72b119039c4d14b1663525b3ac15"
	password := ""
	keyjson, err := ioutil.ReadFile(keyfile)
	if err != nil {
		fmt.Printf("failed to read the keyfile at '%s': %v", keyfile, err)
	}
	key, _ := keystore.DecryptKey(keyjson, password)
	privateKey := key.PrivateKey

	blsPublicKey, _ := bn256.PrivateToPublic(crypto.FromECDSA(privateKey))

	from := crypto.PubkeyToAddress(privateKey.PublicKey)
	privHex := hex.EncodeToString(crypto.FromECDSA(privateKey))
	fmt.Println("address:", from)
	fmt.Println("private key:    ", privHex)
	fmt.Printf("bls public key:  %x\n", blsPublicKey)

	pubkeyHash := common.Bytes2Hex(crypto.FromECDSAPub(&privateKey.PublicKey))
	fmt.Println("public key:     ", pubkeyHash)
	///////////////
	pubKeyBytesFixed := blscrypto.SerializedPublicKey{}
	copy(pubKeyBytesFixed[:], blsPublicKey)

	var pk bn256.PublicKey
	err = pk.Decompress(blsPublicKey[:])
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("pk", len(blsPublicKey), pk.Marshal())
}

//pubkey
//32071fff6599fcdefb78d8048abf7d32165e4dad0a00d7667ba4e1933a6f1bff00
//66b74fbfc9c23963a9a21e12d79422fc288b7598b58f23d4ec04ea2657a05a9901
//40cdae9b90b80179ac73341dd83974fa6dd85f921080770241df8b4f3eb2244e01
//82b9df317d21429c6f0b74c96c21a610483be1d234c2815c50be454a689c35ae01

func TestMakeValidator1(t *testing.T) {

	// There are at least four addresses
	// In here, you should use private key created by yourself, else the data is not valid
	ads := make([]common.Address, 0)
	sliceOfAddressHexStr := []string{
		"0x16FdBcAC4D4Cc24DCa47B9b80f58155a551ca2aF",
		"0x2dC45799000ab08E60b7441c36fCC74060Ccbe11",
		"0x6C5938B49bACDe73a8Db7C3A7DA208846898BFf5",
		"0x1c0eDab88dbb72B119039c4d14b1663525b3aC15",
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
		p := common.Hex2Bytes(s)
		copy(pk1[:], p)
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

	var istanbulExtra *types.IstanbulExtra
	err = rlp.DecodeBytes(finalExtra[32:], &istanbulExtra)
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Printf("pk0:%x\n", istanbulExtra.AddedValidatorsPublicKeys[0])
}
