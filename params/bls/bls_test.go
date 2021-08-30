package bls

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
)

func generateTestData(numEpochs int, numValidators int, composite, cip22 bool) []*SignedBlockHeader {
	var headers []*SignedBlockHeader
	for i := 0; i < numEpochs; i++ {
		message := []byte(fmt.Sprintf("msg_%d", i))
		extraData := []byte(fmt.Sprintf("extra_%d", i))
		var epoch_sigs []*Signature
		var epoch_pubkeys []*PublicKey
		for j := 0; j < numValidators; j++ {
			// generate a private key
			privateKey, _ := GeneratePrivateKey()

			// sign each message
			signature, _ := privateKey.SignMessage(message, extraData, composite, cip22)
			// save the sig to generate the epoch's asig
			epoch_sigs = append(epoch_sigs, signature)

			// save the pubkey to generate the epoch's apubkey
			publicKey, _ := privateKey.ToPublic()
			epoch_pubkeys = append(epoch_pubkeys, publicKey)
		}

		epoch_asig, _ := AggregateSignatures(epoch_sigs)
		epoch_apubkey, _ := AggregatePublicKeys(epoch_pubkeys)

		header := &SignedBlockHeader{
			Data:   message,
			Extra:  extraData,
			Pubkey: epoch_apubkey,
			Sig:    epoch_asig,
		}
		headers = append(headers, header)
	}

	return headers
}

func BenchmarkBlsBatch(b *testing.B) {
	composite := false
	cip22 := false
	headers := generateTestData(10, 10, composite, cip22)

	b.Run("individual", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			for _, h := range headers {
				err := h.Pubkey.VerifySignature(h.Data, h.Extra, h.Sig, composite, cip22)
				if err != nil {
					panic("sig verification should not fail")
				}
			}
		}
	})

	b.Run("batched", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			err := BatchVerifyEpochs(headers, composite, cip22)
			if err != nil {
				panic("sig verification should not fail")
			}
		}
	})
}

// this test is a copy of the `bls-crypto::keys::test_batch_verify` Rust test
func TestBatchVerify(t *testing.T) {
	InitBLSCrypto()

	testBatchVerify(t, true, true)
	testBatchVerify(t, true, false)
	testBatchVerify(t, false, false)
}

func testBatchVerify(t *testing.T, composite, cip22 bool) {
	msgs := generateTestData(10, 7, composite, cip22)

	err := BatchVerifyEpochs(msgs, composite, cip22)
	if err != nil {
		t.Fatalf("batch verification failed, err: %s", err)
	}
}

func TestAggregatedSig(t *testing.T) {
	InitBLSCrypto()
	privateKey, _ := GeneratePrivateKey()
	defer privateKey.Destroy()
	publicKey, _ := privateKey.ToPublic()
	message := []byte("test")
	extraData := []byte("extra")
	for _, cip22 := range []bool{false, true} {
		signature, _ := privateKey.SignMessage(message, extraData, true, cip22)
		err := publicKey.VerifySignature(message, extraData, signature, true, cip22)
		if err != nil {
			t.Fatalf("failed verifying signature for pk 1, error was: %s", err)
		}

		privateKey2, _ := GeneratePrivateKey()
		defer privateKey2.Destroy()
		publicKey2, _ := privateKey2.ToPublic()
		signature2, _ := privateKey2.SignMessage(message, extraData, true, cip22)
		err = publicKey2.VerifySignature(message, extraData, signature2, true, cip22)
		if err != nil {
			t.Fatalf("failed verifying signature for pk 2, error was: %s", err)
		}

		aggergatedPublicKey, _ := AggregatePublicKeys([]*PublicKey{publicKey, publicKey2})
		aggergatedSignature, _ := AggregateSignatures([]*Signature{signature, signature2})
		err = aggergatedPublicKey.VerifySignature(message, extraData, aggergatedSignature, true, cip22)
		if err != nil {
			t.Fatalf("failed verifying signature for aggregated pk, error was: %s", err)
		}
		err = publicKey.VerifySignature(message, extraData, aggergatedSignature, true, cip22)
		if err == nil {
			t.Fatalf("succeeded verifying signature for wrong pk, shouldn't have!")
		}

		subtractedPublicKey, _ := AggregatePublicKeysSubtract(aggergatedPublicKey, []*PublicKey{publicKey2})
		err = subtractedPublicKey.VerifySignature(message, extraData, signature, true, cip22)
		if err != nil {
			t.Fatalf("failed verifying signature for subtractedPublicKey pk, error was: %s", err)
		}
	}
}

func TestProofOfPossession(t *testing.T) {
	privateKey, _ := GeneratePrivateKey()
	defer privateKey.Destroy()
	publicKey, _ := privateKey.ToPublic()
	message := []byte("just some message")
	signature, _ := privateKey.SignPoP(message)
	err := publicKey.VerifyPoP(message, signature)
	if err != nil {
		t.Fatalf("failed verifying PoP for pk 1, error was: %s", err)
	}

	privateKey2, _ := GeneratePrivateKey()
	defer privateKey2.Destroy()
	publicKey2, _ := privateKey2.ToPublic()
	if err != nil {
		t.Fatalf("failed verifying PoP for pk 2, error was: %s", err)
	}

	err = publicKey2.VerifyPoP(message, signature)
	if err == nil {
		t.Fatalf("succeeded verifying PoP for wrong pk, shouldn't have!")
	}
}

func TestNonCompositeSig(t *testing.T) {
	privateKey, _ := GeneratePrivateKey()
	defer privateKey.Destroy()
	publicKey, _ := privateKey.ToPublic()
	message := []byte("test")
	extraData := []byte("extra")
	signature, _ := privateKey.SignMessage(message, extraData, false, false)
	err := publicKey.VerifySignature(message, extraData, signature, false, false)
	if err != nil {
		t.Fatalf("failed verifying signature for pk 1, error was: %s", err)
	}

	privateKey2, _ := GeneratePrivateKey()
	defer privateKey2.Destroy()
	publicKey2, _ := privateKey2.ToPublic()
	signature2, _ := privateKey2.SignMessage(message, extraData, false, false)
	err = publicKey2.VerifySignature(message, extraData, signature2, false, false)
	if err != nil {
		t.Fatalf("failed verifying signature for pk 2, error was: %s", err)
	}

	aggergatedPublicKey, _ := AggregatePublicKeys([]*PublicKey{publicKey, publicKey2})
	aggergatedSignature, _ := AggregateSignatures([]*Signature{signature, signature2})
	err = aggergatedPublicKey.VerifySignature(message, extraData, aggergatedSignature, false, false)
	if err != nil {
		t.Fatalf("failed verifying signature for aggregated pk, error was: %s", err)
	}
	err = publicKey.VerifySignature(message, extraData, aggergatedSignature, false, false)
	if err == nil {
		t.Fatalf("succeeded verifying signature for wrong pk, shouldn't have!")
	}
}

func TestEncodingCIP22(t *testing.T) {
	InitBLSCrypto()
	privateKey, _ := GeneratePrivateKey()
	defer privateKey.Destroy()
	publicKey, _ := privateKey.ToPublic()

	privateKey2, _ := GeneratePrivateKey()
	defer privateKey2.Destroy()
	publicKey2, _ := privateKey2.ToPublic()

	var blockHash, parentHash EpochEntropy
	copy(blockHash[:], "foo")
	copy(parentHash[:], "bar")

	bytes, extraData, err := EncodeEpochToBytesCIP22(10, 5, blockHash, parentHash, 20, 2, []*PublicKey{publicKey, publicKey2})
	if err != nil {
		t.Fatalf("failed encoding epoch bytes")
	}
	if len(bytes) != 221 || len(extraData) != 7 {
		t.Fatalf("wrong length for bytes (221 != %v) or for extra data (7 != %v)", len(bytes), len(extraData))
	}
	t.Logf("encoding: %s, %s\n", hex.EncodeToString(bytes), hex.EncodeToString(extraData))
}

func TestEncoding(t *testing.T) {
	InitBLSCrypto()
	privateKey, _ := GeneratePrivateKey()
	defer privateKey.Destroy()
	publicKey, _ := privateKey.ToPublic()

	privateKey2, _ := GeneratePrivateKey()
	defer privateKey2.Destroy()
	publicKey2, _ := privateKey2.ToPublic()

	_, err := EncodeEpochToBytes(10, 20, []*PublicKey{publicKey, publicKey2})
	if err != nil {
		t.Fatalf("failed encoding epoch bytes")
	}
}

func TestAggregatePublicKeysErrors(t *testing.T) {
	InitBLSCrypto()
	privateKey, _ := GeneratePrivateKey()
	defer privateKey.Destroy()
	publicKey, _ := privateKey.ToPublic()

	_, err := AggregatePublicKeys([]*PublicKey{publicKey, nil})
	if err != NilPointerError {
		t.Fatalf("should have been a nil pointer")
	}
	_, err = AggregatePublicKeys([]*PublicKey{})
	if err != EmptySliceError {
		t.Fatalf("should have been an empty slice")
	}
	_, err = AggregatePublicKeys(nil)
	if err != EmptySliceError {
		t.Fatalf("should have been an empty slice")
	}
}

func TestAggregateSignaturesErrors(t *testing.T) {
	InitBLSCrypto()
	privateKey, _ := GeneratePrivateKey()
	defer privateKey.Destroy()
	message := []byte("test")
	extraData := []byte("extra")
	for _, cip22 := range []bool{false, true} {
		signature, _ := privateKey.SignMessage(message, extraData, true, cip22)

		_, err := AggregateSignatures([]*Signature{signature, nil})
		if err != NilPointerError {
			t.Fatalf("should have been a nil pointer")
		}
		_, err = AggregateSignatures([]*Signature{})
		if err != EmptySliceError {
			t.Fatalf("should have been an empty slice")
		}
		_, err = AggregateSignatures(nil)
		if err != EmptySliceError {
			t.Fatalf("should have been an empty slice")
		}
	}
}

func TestEncodeErrors(t *testing.T) {
	InitBLSCrypto()

	_, _, err := EncodeEpochToBytesCIP22(0, 5, EpochEntropy{}, EpochEntropy{}, 0, 2, []*PublicKey{})
	if err != EmptySliceError {
		t.Fatalf("should have been an empty slice")
	}
	_, _, err = EncodeEpochToBytesCIP22(0, 5, EpochEntropy{}, EpochEntropy{}, 0, 2, nil)
	if err != EmptySliceError {
		t.Fatalf("should have been an empty slice")
	}
}

func TestVerifyPoPErrors(t *testing.T) {
	InitBLSCrypto()
	privateKey, _ := GeneratePrivateKey()
	defer privateKey.Destroy()
	publicKey, _ := privateKey.ToPublic()
	message := []byte("test")
	err := publicKey.VerifyPoP(message, nil)
	if err != NilPointerError {
		t.Fatalf("should have been a nil pointer")
	}
}

func TestVerifySignatureErrors(t *testing.T) {
	InitBLSCrypto()
	privateKey, _ := GeneratePrivateKey()
	defer privateKey.Destroy()
	publicKey, _ := privateKey.ToPublic()
	message := []byte("test")
	extraData := []byte("extra")
	for _, cip22 := range []bool{false, true} {
		err := publicKey.VerifySignature(message, extraData, nil, false, cip22)
		if err != NilPointerError {
			t.Fatalf("should have been a nil pointer")
		}

		err = publicKey.VerifySignature(message, extraData, nil, true, cip22)
		if err != NilPointerError {
			t.Fatalf("should have been a nil pointer")
		}
	}
}

func TestPublicKeySerialization(t *testing.T) {
	InitBLSCrypto()
	privateKey, _ := GeneratePrivateKey()
	defer privateKey.Destroy()
	publicKey, _ := privateKey.ToPublic()
	publicKeyBytes, _ := publicKey.Serialize()

	deserializedKey, _ := DeserializePublicKey(publicKeyBytes)
	deserializedKey2, _ := DeserializePublicKeyCached(publicKeyBytes)

	serializedKey, _ := deserializedKey.Serialize()
	serializedKey2, _ := deserializedKey2.Serialize()
	if !bytes.Equal(serializedKey, serializedKey2) {
		t.Fatalf("public keys should have been equal")
	}
}

func TestHashCRHTestVector(t *testing.T) {
	InitBLSCrypto()
	expectedHash := "955d45ca56cae9d4868f5ebb6921f5212f53ea795a8b0f9490452bf508cd8ab44cbb3fbf046fce68a69fc4f346e83e01"
	decodedExpectedHash, _ := hex.DecodeString(expectedHash)
	v, _ := HashCRH([]byte("hello"), 32)
	if !bytes.Equal(v, decodedExpectedHash) {
		t.Fatalf("v is different from expected: v=%v", hex.EncodeToString(v))
	}
}

func TestHashCompositeCIP22TestVector(t *testing.T) {
	InitBLSCrypto()
	expectedHash := "02c4eab618a78cabd7a0718fd2433e5e1eb7e9119238186f24cefa9fe550ce093c134bd0d8c3bf36ac64226fbd551500c7551b353aba99ec3c2ec53378b5f8cc7abe3472669e472cbb93a51a5378df9f3dc1bf90f39e49a370e1101c39cd0d01804ed25d7b4f31c511ff8be190c1fc794b73671c1d2fb56aaa98157ff99b35fe8db4d311fa343abd8db5ae8b52538601"
	decodedExpectedHash, _ := hex.DecodeString(expectedHash)
	expectedCounter := uint8(1)
	v, c, _ := HashCompositeCIP22([]byte{}, []byte{})
	if c != expectedCounter || !bytes.Equal(v, decodedExpectedHash) {
		t.Fatalf("v or c are different from expected: v=%v, c=%v", hex.EncodeToString(v), c)
	}
}
