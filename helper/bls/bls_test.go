package bls

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/require"
	"log"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)
func Test01(t *testing.T) {

}
func TestECDSAToBLS(t *testing.T) {
	privateKeyECDSA, _ := crypto.HexToECDSA("4f837096cd8578c1f14c9644692c444bbb61426297ff9e8a78a1e7242f541fb3")
	key := BN256{}
	blskey, _ := key.ECDSAToBLS(privateKeyECDSA)
	log.Printf("private key:%d %x", len(blskey), blskey)
	pubkey, _ := PrivateToPublic(blskey)
	log.Printf("public key:%d %x", len(pubkey), pubkey)

}

func TestECDSAToBLS2(t *testing.T) {
	c := BN256{}
	privateKeyECDSA, _ := crypto.HexToECDSA("4f837096cd8578c1f14c9644692c444bbb61426297ff9e8a78a1e7242f541fb3")
	privateKeyBLSBytes, _ := c.ECDSAToBLS(privateKeyECDSA)
	t.Logf("private key: %x", privateKeyBLSBytes)
	privateKeyBLS, _ := DeserializePrivateKey(privateKeyBLSBytes)
	publicKeyBLS := privateKeyBLS.ToPublic()
	publicKeyBLSBytes := publicKeyBLS.Marshal()
	t.Logf("public key: %x", publicKeyBLSBytes)

	address, _ := hex.DecodeString("4f837096cd8578c1f14c9644692c444bbb614262")
	pop, _ := Sign(privateKeyBLS,publicKeyBLS,address)
	popBytes := pop.Marshal()
	t.Logf("pop: %x", popBytes)

	err := Verify(NewApk(publicKeyBLS),address,pop)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("finish")
}
func TestECDSAToBLS3(t *testing.T) {
	c := BN256{}
	privateKeyECDSA, _ := crypto.HexToECDSA("de858b9c8a3502d3fc6a74e558078d606ad6d5b6444f43ac69d2ee83adb6baca")
	privateKeyBLSBytes, _ := c.ECDSAToBLS(privateKeyECDSA)
	t.Logf("private key: %x", privateKeyBLSBytes)
	privateKeyBLS := &SecretKey{ privateKeyECDSA.D}
	//privateKeyBLS, _ := DeserializePrivateKey(privateKeyBLSBytes)
	publicKeyBLS := privateKeyBLS.ToPublic()
	publicKeyBLSBytes := publicKeyBLS.Marshal()
	t.Logf("public key: %x", publicKeyBLSBytes)

	address, _ := hex.DecodeString("4f837096cd8578c1f14c9644692c444bbb614262")
	pop, _ := Sign(privateKeyBLS,publicKeyBLS,address)
	popBytes := pop.Marshal()
	t.Logf("pop: %x", popBytes)

	err := Verify(NewApk(publicKeyBLS),address,pop)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("finish")
}

func TestECDSAToBLS4(t *testing.T) {
	pub, priv, err1 := GenKeyPair(rand.Reader)
	if err1 != nil {
		fmt.Println("gen failed",err1)
		return
	}
	privBytes,_ := priv.Serialize()
	t.Logf("private key: %x", privBytes)
	publicKeyBLS := priv.ToPublic()
	t.Logf("public key1: %x", publicKeyBLS.Marshal())
	t.Logf("public key2: %x", pub.Marshal())

	address, _ := hex.DecodeString("4f837096cd8578c1f14c9644692c444bbb614262")
	pop, _ := Sign(priv,pub,address)
	popBytes := pop.Marshal()
	t.Logf("pop: %x", popBytes)

	err := Verify(NewApk(pub),address,pop)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("finish")
}
func TestECDSAToBLS5(t *testing.T) {
	pub, priv, err1 := GenKeyPair(rand.Reader)
	if err1 != nil {
		fmt.Println("gen failed",err1)
		return
	}
	privBytes,_ := priv.Serialize()
	t.Logf("private key: %x", privBytes)
	publicKeyBLS := priv.ToPublic()
	t.Logf("public key1: %x", publicKeyBLS.Marshal())
	t.Logf("public key2: %x", pub.Marshal())

	address, _ := hex.DecodeString("4f837096cd8578c1f14c9644692c444bbb614262")
	pop, _ := UnsafeSign(priv,address)
	popBytes := pop.Marshal()
	t.Logf("pop: %x", popBytes)

	err := VerifyUnsafe(pub,address,pop)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("finish")
}
// TestSignVerify
func TestSignVerify(t *testing.T) {
	msg := randomMessage()
	pub, priv, err := GenKeyPair(rand.Reader)
	require.NoError(t, err)

	sig, err := UnsafeSign(priv, msg)
	require.NoError(t, err)
	require.NoError(t, VerifyUnsafe(pub, msg, sig))

	// Testing that changing the message, the signature is no longer valid
	require.NotNil(t, VerifyUnsafe(pub, randomMessage(), sig))

	// Testing that using a random PK, the signature cannot be verified
	pub2, _, err := GenKeyPair(rand.Reader)
	require.NoError(t, err)
	require.NotNil(t, VerifyUnsafe(pub2, msg, sig))
	fmt.Printf("sign len:%d\n", len(sig.Marshal()))
}

func randomMessage() []byte {
	msg := make([]byte, 32)
	_, _ = rand.Read(msg)
	return msg
}
