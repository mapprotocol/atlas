package bn256_keep_network

import (
	"crypto/rand"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"github.com/keep-network/keep-core/pkg/altbn128"
)

var (
	zeroBigInt = *big.NewInt(0)
	oneBigInt  = *big.NewInt(1)
	g2         = *new(bn256.G2).ScalarBaseMult(&oneBigInt)  // Generator point of G2 group
	zeroG1     = *new(bn256.G1).ScalarBaseMult(&zeroBigInt) // Zero point in G2 group
	zeroG2     = *new(bn256.G2).ScalarBaseMult(&zeroBigInt) // Zero point in G2 group
)

// GenerateRandomKey creates a random private and its corresponding public keys
func GenerateRandomKey() (PrivateKey, PublicKey) {
	priv, pub, _ := bn256.RandomG2(rand.Reader)
	return PrivateKey{p: priv}, PublicKey{p: pub}
}

// EmptyMultisigMask returns zero bitmask
func EmptyMultisigMask() big.Int {
	return *big.NewInt(0)
}

// hashToPointMsg performs "message augmentation": hashes the message and the
// point to the point of G1 curve (a signature)
func hashToPointMsg(p *bn256.G2, message []byte) *bn256.G1 {
	var data []byte
	data = append(data, p.Marshal()...)
	data = append(data, message...)
	return altbn128.G1HashToPoint(data)
}

// hashToPointIndex hashes the aggregated public key (G2 point) and the given
// index (of the signer within a group of signers) to the point in G1 curve (a
// signature)
func hashToPointIndex(pub *bn256.G2, index byte) *bn256.G1 {
	data := make([]byte, 32)
	data[31] = index
	return hashToPointMsg(pub, data)
}

func HashToPointIndex(pub PublicKey, index byte) Signature {
	return Signature{p: hashToPointIndex(pub.p, index)}
}
