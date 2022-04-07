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

	err := Verify(&Apk{publicKeyBLS},address,pop)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("finish")
}
func TestECDSAToBLS3(t *testing.T) {
	c := BN256{}
	privateKeyECDSA, _ := crypto.HexToECDSA("7f54814b7d0576b474d36853b6797175c41b64900b1531e8437d3724dc90d332")
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

	err := Verify(&Apk{publicKeyBLS},address,pop)
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

	err := Verify(&Apk{pub},address,pop)
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
