package bls

import (
	"encoding/hex"
	"testing"

	"github.com/celo-org/celo-bls-go/bls"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestECDSAToBLS(t *testing.T) {
	privateKeyECDSA, _ := crypto.HexToECDSA("4f837096cd8578c1f14c9644692c444bbb61426297ff9e8a78a1e7242f541fb3")
	privateKeyBLSBytes, _ := CryptoType().ECDSAToBLS(privateKeyECDSA)
	t.Logf("private key: %x", privateKeyBLSBytes)
	privateKeyBLS, _ := bls.DeserializePrivateKey(privateKeyBLSBytes)
	publicKeyBLS, _ := privateKeyBLS.ToPublic()
	publicKeyBLSBytes, _ := publicKeyBLS.Serialize()
	t.Logf("public key: %x", publicKeyBLSBytes)

	address, _ := hex.DecodeString("4f837096cd8578c1f14c9644692c444bbb614262")
	pop, _ := privateKeyBLS.SignPoP(address)
	popBytes, _ := pop.Serialize()
	t.Logf("pop: %x", popBytes)
}

func split(buf []byte, lim int) []SerializedPublicKey {
	chunks := make([]SerializedPublicKey, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		pubKeyBytesFixed := SerializedPublicKey{}
		copy(pubKeyBytesFixed[:], buf[:lim])
		buf = buf[lim:]
		chunks = append(chunks, pubKeyBytesFixed)
	}
	if len(buf) > 0 {
		pubKeyBytesFixed := SerializedPublicKey{}
		copy(pubKeyBytesFixed[:], buf)
		chunks = append(chunks, pubKeyBytesFixed)
	}
	return chunks
}
