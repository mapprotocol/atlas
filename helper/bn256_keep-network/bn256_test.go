package bn256_keep_network

import (
	"crypto/rand"
	"fmt"
	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

const (
	MESSAGE_SIZE        = 32
	PARTICIPANTS_NUMBER = 64
)

var (
	msg         = GenRandomBytes(MESSAGE_SIZE)
	privs, pubs = GenerateRandomKeys(PARTICIPANTS_NUMBER)
	as          = CalculateAntiRogueCoefficients(pubs)
	aggPub      = AggregatePublicKeys(pubs, as)

	secretKey, publicKey = GenerateRandomKey()
	signature            = secretKey.Sign(msg)
)

func Test_VerifySignature(t *testing.T) {
	require.True(t, signature.Verify(publicKey, msg))
}

func Test_AggregatedSignature(t *testing.T) {
	priv1, pub1 := GenerateRandomKey()
	priv2, pub2 := GenerateRandomKey()
	priv3, pub3 := GenerateRandomKey()
	sig1 := priv1.Sign(msg)
	sig2 := priv2.Sign(msg)
	sig3 := priv3.Sign(msg)
	//fmt.Println(len(priv1.Marshal()), len(pub1.Marshal()), len(sig1.Marshal()))

	require.True(t, sig1.Aggregate(sig2).Aggregate(sig3).Verify(pub1.Aggregate(pub2).Aggregate(pub3), msg))
}

func Test_VerifyMultisigDemo(t *testing.T) {
	priv0, pub0 := bls.GenerateRandomKey()
	priv1, pub1 := bls.GenerateRandomKey()
	priv2, pub2 := bls.GenerateRandomKey()
	Simple := *big.NewInt(1) // in real life use coeficients against anti rogue key attack

	allPub_ := pub0.Aggregate(pub1)
	// Aggregated public key of all participants
	allPub := pub0.Aggregate(pub1).Aggregate(pub2)

	fmt.Printf("pub0:%x\n", pub0.Marshal())
	fmt.Printf("pub1:%x\n", pub1.Marshal())
	fmt.Printf("pub2:%x\n", pub2.Marshal())
	fmt.Printf("pub_:%x\n", allPub_.Marshal())
	fmt.Printf("pub:%x\n", allPub.Marshal())

	// Setup phase - generate membership keys
	mk0 := priv0.GenerateMembershipKeyPart(0, allPub, Simple).
		Aggregate(priv1.GenerateMembershipKeyPart(0, allPub, Simple)).
		Aggregate(priv2.GenerateMembershipKeyPart(0, allPub, Simple))
	mk2 := priv0.GenerateMembershipKeyPart(2, allPub, Simple).
		Aggregate(priv1.GenerateMembershipKeyPart(2, allPub, Simple)).
		Aggregate(priv2.GenerateMembershipKeyPart(2, allPub, Simple))

	// Sign only by #0 and #2
	mask := big.NewInt(0b101)
	sig0 := priv0.Multisign(msg, allPub, mk0)
	sig2 := priv2.Multisign(msg, allPub, mk2)
	subSig := sig0.Aggregate(sig2)
	subPub := pub0.Aggregate(pub2)

	// Verify in Golang
	require.True(t, subSig.VerifyMultisig(allPub, subPub, msg, mask))

	// Verify in EVM
	//_, err := blsSignatureTest.VerifyMultisignature(owner, allPub.Marshal(), subPub.Marshal(), msg, subSig.Marshal(), mask)
	//require.NoError(t, err)
	//backend.Commit()
	//verifiedSol, err := blsSignatureTest.Verified(&bind.CallOpts{})
	//require.True(t, verifiedSol)
}

// GenRandomBytes generates byte array with random data
func GenRandomBytes(size int) (blk []byte) {
	blk = make([]byte, size)
	_, _ = rand.Reader.Read(blk)
	return
}

// GenerateRandomKeys creates an array of random private and their corresponding public keys
func GenerateRandomKeys(total int) ([]PrivateKey, []PublicKey) {
	privs, pubs := make([]PrivateKey, total), make([]PublicKey, total)
	for i := 0; i < total; i++ {
		privs[i], pubs[i] = GenerateRandomKey()
	}
	return privs, pubs
}
