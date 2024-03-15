package bls

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	cfbn256 "github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"
	"log"
	"math/big"
	"testing"
)

func Test01(t *testing.T) {
	privateKeyECDSA, _ := crypto.HexToECDSA("")
	privateKey, err := CryptoType().ECDSAToBLS(privateKeyECDSA)
	if err != nil {
		fmt.Println(err)
		return
	}
	s1, err := CryptoType().PrivateToPublic(privateKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	b1, err := s1.MarshalText()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("bls pubkey: %s\n", b1)

	s2, err := CryptoType().PrivateToG1Public(privateKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	b2, err := s2.MarshalText()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("bls g1 pubkey: %s\n", b2)
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
	pop, _ := Sign(privateKeyBLS, publicKeyBLS, address)
	popBytes := pop.Marshal()
	t.Logf("pop: %x", popBytes)

	err := Verify(NewApk(publicKeyBLS), address, pop)
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
	privateKeyBLS := &SecretKey{privateKeyECDSA.D}
	//privateKeyBLS, _ := DeserializePrivateKey(privateKeyBLSBytes)
	publicKeyBLS := privateKeyBLS.ToPublic()
	publicKeyBLSBytes := publicKeyBLS.Marshal()
	t.Logf("public key: %x", publicKeyBLSBytes)

	address, _ := hex.DecodeString("4f837096cd8578c1f14c9644692c444bbb614262")
	pop, _ := Sign(privateKeyBLS, publicKeyBLS, address)
	popBytes := pop.Marshal()
	t.Logf("pop: %x", popBytes)

	err := Verify(NewApk(publicKeyBLS), address, pop)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("finish")
}

func TestECDSAToBLS4(t *testing.T) {
	pub, priv, err1 := GenKeyPair(rand.Reader)
	if err1 != nil {
		fmt.Println("gen failed", err1)
		return
	}
	privBytes, _ := priv.Serialize()
	t.Logf("private key: %x", privBytes)
	publicKeyBLS := priv.ToPublic()
	t.Logf("public key1: %x", publicKeyBLS.Marshal())
	t.Logf("public key2: %x", pub.Marshal())

	address, _ := hex.DecodeString("4f837096cd8578c1f14c9644692c444bbb614262")
	pop, _ := Sign(priv, pub, address)
	popBytes := pop.Marshal()
	t.Logf("pop: %x", popBytes)

	err := Verify(NewApk(pub), address, pop)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("finish")
}
func TestECDSAToBLS5(t *testing.T) {
	pub, priv, err1 := GenKeyPair(rand.Reader)
	if err1 != nil {
		fmt.Println("gen failed", err1)
		return
	}
	privBytes, _ := priv.Serialize()
	t.Logf("private key: %x", privBytes)
	publicKeyBLS := priv.ToPublic()
	t.Logf("public key1: %x", publicKeyBLS.Marshal())
	t.Logf("public key2: %x", pub.Marshal())

	address, _ := hex.DecodeString("4f837096cd8578c1f14c9644692c444bbb614262")
	pop, _ := UnsafeSign(priv, address)
	popBytes := pop.Marshal()
	t.Logf("pop: %x", popBytes)

	err := VerifyUnsafe(pub, address, pop)
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

func TestG1pubkeyVerify01(t *testing.T) {
	msg := randomMessage()
	pub, priv, err := GenKeyPair(rand.Reader)
	require.NoError(t, err)

	sig, err := UnsafeSign(priv, msg)
	require.NoError(t, err)
	require.NoError(t, VerifyUnsafe(pub, msg, sig))

	g1puk := priv.ToG1Public()

	err = VerifyG1Pk(g1puk, pub.Marshal())
	require.NoError(t, err)

	pub2, priv2, err2 := GenKeyPair(rand.Reader)
	require.NoError(t, err2)
	fmt.Println(hex.EncodeToString(pub2.Marshal()))
	g1puk2 := priv2.ToG1Public()

	err = VerifyG1Pk(g1puk2, pub.Marshal())
	if err != nil {
		fmt.Println(err)
	}
}
func TestG1pubkeyVerify02(t *testing.T) {
	for i := 0; i < 10000; i++ {
		pub, priv, err := GenKeyPair(rand.Reader)
		require.NoError(t, err)
		g1puk := priv.ToG1Public()
		err = VerifyG1Pk(g1puk, pub.Marshal())
		require.NoError(t, err)
	}
}
func randomMessage() []byte {
	msg := make([]byte, 32)
	_, _ = rand.Read(msg)
	return msg
}

func hash256(msg []byte) ([]byte, error) {
	H := sha3.New256()
	_, err := H.Write(msg)
	if err != nil {
		return nil, err
	}
	return H.Sum(nil), err
}
func hashLegacy256(msg []byte) ([]byte, error) {
	H := sha3.NewLegacyKeccak256()
	_, err := H.Write(msg)
	if err != nil {
		return nil, err
	}
	return H.Sum(nil), err
}
func TestHash01(t *testing.T) {
	secret1 := SecretKey{big.NewInt(10099)}
	pkey1, pkey2 := secret1.ToPublic(), secret1.ToG1Public()
	fmt.Println("pkey1", hex.EncodeToString(pkey1.Marshal()))
	fmt.Println("pkey2", hex.EncodeToString(pkey2))

	h0, _ := hash256([]byte{1})
	h1, _ := hashLegacy256([]byte{1})

	b0, e := hexutil.Decode("0x1234")
	if e != nil {
		fmt.Println(e)
	}

	h2, _ := hash256(b0)
	h3, _ := hashLegacy256(b0)

	fmt.Println("h0:", hex.EncodeToString(h0))
	fmt.Println("h1:", hex.EncodeToString(h1))

	fmt.Println("h2:", hex.EncodeToString(h2))
	fmt.Println("h3:", hex.EncodeToString(h3))
}

func Test_UnsafeVerify(t *testing.T) {
	big1, big2 := big.NewInt(1), big.NewInt(2)
	message := []byte{1, 2, 3}
	//g2pks := make([]*PublicKey, 0, 2)

	secret1, secret2 := SecretKey{big1}, SecretKey{big2}
	g2PublicKey1, g2PublicKey2 := secret1.ToPublic(), secret2.ToPublic()

	// agg pk
	aggrPubkey := g2PublicKey1.Aggregate(g2PublicKey2)

	// sign
	sign1, err := UnsafeSign(&secret1, message)
	if err != nil {
		panic(err)
	}
	sign2, err := UnsafeSign(&secret2, message)
	if err != nil {
		panic(err)
	}

	// agg sign
	aggSign := UnsafeAggregate(sign1, sign2)
	if err != nil {
		panic(err)
	}

	// verify
	err = VerifyUnsafe(aggrPubkey, message, aggSign)
	if err != nil {
		panic(err)
	}
}

func Test_Verify(t *testing.T) {
	big1, big2 := big.NewInt(1), big.NewInt(2)
	message := []byte{1, 2, 3}
	//g2pks := make([]*PublicKey, 0, 2)

	secret1, secret2 := SecretKey{big1}, SecretKey{big2}
	g2PublicKey1, g2PublicKey2 := secret1.ToPublic(), secret2.ToPublic()

	// agg pk
	aggrPubkey, err := AggregateApk([]*PublicKey{g2PublicKey1, g2PublicKey2})
	if err != nil {
		panic(err)
	}
	// sign
	sign1, err := Sign(&secret1, g2PublicKey1, message)
	if err != nil {
		panic(err)
	}
	sign2, err := Sign(&secret2, g2PublicKey2, message)
	if err != nil {
		panic(err)
	}

	// agg sign
	aggSign := sign1.Aggregate(sign2)
	if err != nil {
		panic(err)
	}

	// verify
	err = Verify(aggrPubkey, message, aggSign)
	if err != nil {
		panic(err)
	}
}

func Test02(t *testing.T) {
	big1 := big.NewInt(1)
	message := []byte{1}

	secret1 := SecretKey{big1}
	g2PublicKey1 := secret1.ToPublic()

	// sign
	sign1, err := UnsafeSign(&secret1, message)
	if err != nil {
		panic(err)
	}
	h, _ := hashLegacy256(message)
	fmt.Println(hex.EncodeToString(h))
	fmt.Println(hex.EncodeToString(sign1.Marshal()))
	// verify
	err = VerifyUnsafe(g2PublicKey1, message, sign1)
	if err != nil {
		panic(err)
	}
}

// /////////////////////////////////////////////////////////////////////
func TestBls2(t *testing.T) {
	pub, priv, err1 := GenKeyPair(rand.Reader)
	if err1 != nil {
		fmt.Println("gen failed", err1)
		return
	}
	privBytes, _ := priv.Serialize()
	t.Logf("private key: %x", privBytes)
	publicKeyBLS := priv.ToPublic()
	t.Logf("public key1: %x", publicKeyBLS.Marshal())
	t.Logf("public key2: %x", pub.Marshal())

	address, _ := hex.DecodeString("4f837096cd8578c1f14c9644692c444bbb614262")
	pop, _ := Sign2(priv, pub, address)
	popBytes := pop.Marshal()
	t.Logf("pop: %x", popBytes)

	err := Verify2(NewApk(pub), address, pop)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("finish")
}

func TestBls03(t *testing.T) {
	b0, err := hexutil.Decode("0x07b68ce4d3ba620d2c916b52a04331f9c1de66070ce9327c2b8ed57ad91ecaec")
	if err != nil {
		t.Error(err)
	}
	privateKeyBLS, err := DeserializePrivateKey(b0)
	if err != nil {
		t.Error(err)
	}
	publicKeyBLS := privateKeyBLS.ToPublic()
	publicKeyBLSBytes := publicKeyBLS.Marshal()
	t.Logf("public key: %x", publicKeyBLSBytes)

	//address, _ := hex.DecodeString("0x6162636566676869")
	msg := []byte{1}
	pop, _ := Sign2(privateKeyBLS, publicKeyBLS, msg)
	popBytes := pop.Marshal()
	t.Logf("signature2: %x", popBytes)

	pop2, _ := Sign(privateKeyBLS, publicKeyBLS, msg)
	popBytes2 := pop2.Marshal()
	t.Logf("signature1: %x", popBytes2)

	err = Verify2(NewApk(publicKeyBLS), msg, pop)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("finish")
}
func TestBls04(t *testing.T) {
	//msg := []byte{1}
	msg, err := hex.DecodeString("6162636566676869")
	if err != nil {
		t.Error(err)
	}
	g1, err := cfbn256.HashToG1(msg)
	if err != nil {
		t.Error(err)
	}
	b0 := g1.Marshal()
	t.Logf("hash: %x", b0)
}
