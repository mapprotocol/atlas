package bn256_dusk_network

import (
	"crypto/rand"
	"io"
	"math/big"
	"reflect"
	"testing"

	"github.com/dusk-network/bn256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func randomMessage() []byte {
	msg := make([]byte, 32)
	_, _ = rand.Read(msg)
	return msg
}

func TestCopyApk(t *testing.T) {
	require := require.New(t)
	pub, _, err := GenKeyPair(rand.Reader)
	require.NoError(err)
	apk := NewApk(pub)

	cpy := apk.Copy()
	require.True(reflect.DeepEqual(apk, cpy))
}

func TestCopySig(t *testing.T) {
	require := require.New(t)
	pub, priv, err := GenKeyPair(rand.Reader)
	require.NoError(err)

	sigma, err := Sign(priv, pub, []byte("ciao!!"))
	require.NoError(err)

	cpy := sigma.Copy()
	require.True(reflect.DeepEqual(sigma, cpy))
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
	//k := []byte(priv.x.String())
	//fmt.Println(len(k), len(pub.Marshal()), len(sig.Marshal()))
}

// TestCombine checks for the Batched form of the BLS signature
func TestCombine(t *testing.T) {
	reader := rand.Reader
	msg1 := []byte("Get Funky Tonight")
	msg2 := []byte("Gonna Get Funky Tonight")

	pub1, priv1, err := GenKeyPair(reader)
	require.NoError(t, err)

	pub2, priv2, err := GenKeyPair(reader)
	require.NoError(t, err)

	str1, err := pub1.MarshalText()
	require.NoError(t, err)

	str2, err := pub2.MarshalText()
	require.NoError(t, err)

	require.NotEqual(t, str1, str2)

	sig1, err := UnsafeSign(priv1, msg1)
	require.NoError(t, err)
	require.NoError(t, VerifyUnsafe(pub1, msg1, sig1))

	sig2, err := UnsafeSign(priv2, msg2)
	require.NoError(t, err)
	require.NoError(t, VerifyUnsafe(pub2, msg2, sig2))

	sig3 := UnsafeAggregate(sig1, sig2)
	pkeys := []*PublicKey{pub1, pub2}
	require.NoError(t, VerifyUnsafeBatch(pkeys, [][]byte{msg1, msg2}, sig3))
}

func TestHashToPoint(t *testing.T) {
	msg := []byte("test data")
	g1, err := h0(msg)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, g1)
}

func randomInt(r io.Reader) *big.Int {
	for {
		k, _ := rand.Int(r, bn256.Order)
		if k.Sign() > 0 {
			return k
		}
	}
}
func TestRogueKey(t *testing.T) {
	reader := rand.Reader
	pub, _, err := GenKeyPair(reader)
	require.NoError(t, err)
	// α is the pseudo-secret key of the attacker
	alpha := randomInt(reader)
	// g₂ᵅ
	g1Alpha := newG1().ScalarBaseMult(alpha)

	// pk⁻¹
	rogueGx := newG1()
	rogueGx.Neg(pub.gx)

	pRogue := newG1()
	pRogue.Add(g1Alpha, rogueGx)

	sk, pk := &SecretKey{alpha}, &PublicKey{pRogue}

	msg := []byte("test data")
	rogueSignature, err := UnsafeSign(sk, msg)
	require.NoError(t, err)

	require.NoError(t, verifyBatch([]*bn256.G1{pub.gx, pk.gx}, [][]byte{msg, msg}, rogueSignature.e, true))
}

func TestMarshalPk(t *testing.T) {
	reader := rand.Reader
	pub, _, err := GenKeyPair(reader)
	require.NoError(t, err)

	pkByteRepr := pub.Marshal()

	g1 := newG1()
	_, _ = g1.Unmarshal(pkByteRepr)

	g1ByteRepr := g1.Marshal()
	require.Equal(t, pkByteRepr, g1ByteRepr)

	pkInt := new(big.Int).SetBytes(pkByteRepr)
	g1Int := new(big.Int).SetBytes(g1ByteRepr)
	require.Equal(t, pkInt, g1Int)

	pk, err := UnmarshalPk(pkByteRepr)
	require.NoError(t, err)
	require.Equal(t, pub, pk)
}

func TestApkVerificationSingleKey(t *testing.T) {
	reader := rand.Reader
	msg := []byte("Get Funky Tonight")

	pub1, priv1, err := GenKeyPair(reader)
	require.NoError(t, err)

	apk := NewApk(pub1)

	signature, err := Sign(priv1, pub1, msg)
	require.NoError(t, err)
	require.NoError(t, Verify(apk, msg, signature))

	// testing unmarshalling
	bApk := apk.Marshal()
	uApk, err := UnmarshalApk(bApk)
	assert.NoError(t, err)
	require.NoError(t, Verify(uApk, msg, signature))
}

func TestApkVerification(t *testing.T) {
	reader := rand.Reader
	msg := []byte("Get Funky Tonight")

	pub1, priv1, err := GenKeyPair(reader)
	require.NoError(t, err)

	pub2, priv2, err := GenKeyPair(reader)
	require.NoError(t, err)

	apk := NewApk(pub1)
	assert.NoError(t, apk.Aggregate(pub2))

	signature, err := Sign(priv1, pub1, msg)
	require.NoError(t, err)
	sig2, err := Sign(priv2, pub2, msg)
	require.NoError(t, err)

	signature.Aggregate(sig2)
	require.NoError(t, Verify(apk, msg, signature))
}

func TestApkBatchVerification(t *testing.T) {
	reader := rand.Reader
	msg := []byte("Get Funky Tonight")

	pub1, priv1, err := GenKeyPair(reader)
	require.NoError(t, err)

	pub2, priv2, err := GenKeyPair(reader)
	require.NoError(t, err)

	apk := NewApk(pub1)
	assert.NoError(t, apk.Aggregate(pub2))

	sigma, err := Sign(priv1, pub1, msg)
	require.NoError(t, err)
	sig2_1, err := Sign(priv2, pub2, msg)
	require.NoError(t, err)
	sig := sigma.Aggregate(sig2_1)
	require.NoError(t, Verify(apk, msg, sig))

	msg2 := []byte("Gonna get Shwifty tonight")
	pub3, priv3, err := GenKeyPair(reader)
	require.NoError(t, err)
	apk2 := NewApk(pub2)
	require.NoError(t, apk2.Aggregate(pub3))
	sig2_2, err := Sign(priv2, pub2, msg2)
	require.NoError(t, err)
	sig3_2, err := Sign(priv3, pub3, msg2)
	require.NoError(t, err)
	sig2 := sig2_2.Aggregate(sig3_2)
	require.NoError(t, Verify(apk2, msg2, sig2))

	sigma.Aggregate(sig2)
	require.NoError(t, VerifyBatch(
		[]*Apk{apk, apk2},
		[][]byte{msg, msg2},
		sigma,
	))
}

//func TestSafeCompress(t *testing.T) {
//	msg := randomMessage()
//	pub, priv, err := GenKeyPair(rand.Reader)
//	require.NoError(t, err)
//
//	sig, err := Sign(priv, pub, msg)
//	require.NoError(t, err)
//	require.NoError(t, Verify(NewApk(pub), msg, sig))
//
//	sigb := sig.Compress()
//	sigTest := &Signature{e: newG1()}
//	require.NoError(t, sigTest.Decompress(sigb))
//
//	require.Equal(t, sig.Marshal(), sigTest.Marshal())
//}
//
//func TestUnsafeCompress(t *testing.T) {
//	msg := randomMessage()
//	pub, priv, err := GenKeyPair(rand.Reader)
//	require.NoError(t, err)
//
//	sig, err := UnsafeSign(priv, msg)
//	require.NoError(t, err)
//	require.NoError(t, VerifyUnsafe(pub, msg, sig))
//
//	sigb := sig.Compress()
//	sigTest := &UnsafeSignature{e: newG1()}
//	require.NoError(t, sigTest.Decompress(sigb))
//
//	sigM := sig.e.Marshal()
//	require.NotEmpty(t, sigM)
//	require.Equal(t, sigM, sigTest.e.Marshal())
//}

//func TestAmbiguousCompress(t *testing.T) {
//	msg := randomMessage()
//	pub, priv, err := GenKeyPair(rand.Reader)
//	require.NoError(t, err)
//
//	sig, err := UnsafeSign(priv, msg)
//	require.NoError(t, err)
//	require.NoError(t, VerifyUnsafe(pub, msg, sig))
//
//	sigb := sig.Compress()
//	require.Equal(t, len(sigb), 33)
//
//	xy1, xy2, err := bn256.DecompressAmbiguous(sigb)
//	require.NoError(t, err)
//
//	if xy1 == nil && xy2 == nil {
//		fmt.Printf("Original signature: %v\n", new(big.Int).SetBytes(sig.Marshal()).String())
//		fmt.Printf("Compressed signature: %v\n", new(big.Int).SetBytes(sigb).String())
//		require.Fail(t, "Orcoddue")
//	}
//
//	sigM := sig.e.Marshal()
//	require.NotEmpty(t, sigM)
//
//	if xy1 != nil {
//		sig1 := &UnsafeSignature{xy1}
//		if bytes.Equal(sigM, sig1.e.Marshal()) {
//			return
//		}
//	}
//	if xy2 != nil {
//		sig2 := &UnsafeSignature{xy2}
//		if bytes.Equal(sigM, sig2.e.Marshal()) {
//			return
//		}
//	}
//
//	require.Fail(t, "Decompression failed both xy1 and xy2 are nil")
//}

func BenchmarkSign(b *testing.B) {
	msg := randomMessage()
	pk, sk, _ := GenKeyPair(rand.Reader)

	b.ReportAllocs()
	b.StopTimer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		_, _ = Sign(sk, pk, msg)
		b.StopTimer()
	}
}

func aggregateXSignatures(b *testing.B, nr int) {
	var sigmas = make([]*Signature, nr)
	reader := rand.Reader
	msg := []byte("Get Funky Tonight")

	for i := 0; i < nr; i++ {
		pk, sk, _ := GenKeyPair(reader)
		sigmas[i], _ = Sign(sk, pk, msg)
	}

	s, sigmas := sigmas[0], sigmas[1:]

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, sig := range sigmas {
			s = s.Aggregate(sig)
		}
	}
}

func BenchmarkAggregate10Signatures(b *testing.B) {
	aggregateXSignatures(b, 10)
}

func BenchmarkAggregate100Signatures(b *testing.B) {
	aggregateXSignatures(b, 100)
}

func BenchmarkAggregate1000Signatures(b *testing.B) {
	aggregateXSignatures(b, 1000)
}

func BenchmarkVerifySingleSignature(b *testing.B) {
	msg := randomMessage()
	pk, sk, _ := GenKeyPair(rand.Reader)
	b.StopTimer()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		signature, _ := Sign(sk, pk, msg)

		b.StartTimer()
		_ = Verify(NewApk(pk), msg, signature)
		b.StopTimer()
	}
}

func BenchmarkVerifyUnsafeSingleSignature(b *testing.B) {
	msg := randomMessage()
	b.StopTimer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pk, sk, _ := GenKeyPair(rand.Reader)
		signature, _ := UnsafeSign(sk, msg)

		b.StartTimer()
		_ = VerifyUnsafe(pk, msg, signature)
		b.StopTimer()
	}
}

//func BenchmarkVerifySingleCompressedSignature(b *testing.B) {
//	msg := randomMessage()
//	b.StopTimer()
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		pk, sk, _ := GenKeyPair(rand.Reader)
//		signature, _ := Sign(sk, pk, msg)
//		sigb := signature.Compress()
//		apk := NewApk(pk)
//
//		b.StartTimer()
//		signature = &Signature{}
//		_ = signature.Decompress(sigb)
//		_ = Verify(apk, msg, signature)
//		b.StopTimer()
//	}
//}
//
//func BenchmarkVerifySingleCompressedUnsafeSignature(b *testing.B) {
//	msg := randomMessage()
//	b.StopTimer()
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		pk, sk, _ := GenKeyPair(rand.Reader)
//		signature, _ := UnsafeSign(sk, msg)
//		sigb := signature.Compress()
//
//		b.StartTimer()
//		signature = &UnsafeSignature{}
//		_ = signature.Decompress(sigb)
//		_ = VerifyUnsafe(pk, msg, signature)
//		b.StopTimer()
//	}
//}
