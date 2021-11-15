package tools

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/core/types"
	blscrypto "github.com/mapprotocol/atlas/params/bls"
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
	blsPrivateKey, _ := blscrypto.ECDSAToBLS(privateKey)
	blsPublicKey, _ := blscrypto.PrivateToPublic(blsPrivateKey)
	bp, _ := blsPublicKey.MarshalText()
	from := crypto.PubkeyToAddress(privateKey.PublicKey)
	fmt.Println("address:", from)
	privHex := hex.EncodeToString(crypto.FromECDSA(privateKey))
	fmt.Println("private key:    ", privHex)
	pkHash := common.Bytes2Hex(bp)
	fmt.Println("bls public key: ", pkHash)

	//keyfile := "../../data3/keystore/UTC--2021-08-06T03-32-55.462419725Z--d0d471aaea6bc0321e9c7f7696aac6c8626d1420"
	//password := ""
	//keyjson, err := ioutil.ReadFile(keyfile)
	//if err != nil {
	//	fmt.Printf("failed to read the keyfile at '%s': %v", keyfile, err)
	//}
	//key, _ := keystore.DecryptKey(keyjson, password)
	//priKey := key.PrivateKey
	//privKeyHex := hex.EncodeToString(blsPrivateKey)
	//fmt.Println("bls private key:", privKeyHex)
	//pubkeyHash := common.Bytes2Hex(crypto.FromECDSAPub(&privateKey.PublicKey))
	//fmt.Println("public key:     ", pubkeyHash)
}
func TestMakeBLSPublicKey(t *testing.T) {

}
