package bn256_keep_network

import (
	"crypto"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"github.com/keep-network/keep-core/pkg/altbn128"
)

type PrivateKey struct {
	p *big.Int
}

// PublicKey calculates public key corresponding to the given private key
func (priv PrivateKey) Public() crypto.PublicKey {
	return &PublicKey{p: new(bn256.G2).ScalarBaseMult(priv.p)}
}

func (priv PrivateKey) PublicKey() PublicKey {
	return PublicKey{p: new(bn256.G2).ScalarBaseMult(priv.p)}
}

// Sign generates a simple BLS signature of the given message
func (secretKey PrivateKey) Sign(message []byte) Signature {
	hashPoint := altbn128.G1HashToPoint(message)
	return Signature{p: new(bn256.G1).ScalarMult(hashPoint, secretKey.p)}
}

// Multisign generates BLS multi-signature of the given message, aggregated
// public key of all participants and the membership key of the signer
func (secretKey PrivateKey) Multisign(message []byte, aggPublicKey PublicKey, membershipKey Signature) Signature {
	s := new(bn256.G1).ScalarMult(hashToPointMsg(aggPublicKey.p, message), secretKey.p)
	s.Add(s, membershipKey.p)
	return Signature{p: s}
}

// GenerateMembershipKeyPart generates the participant signature to be aggregated into membership key
func (secretKey PrivateKey) GenerateMembershipKeyPart(index byte, aggPub PublicKey, anticoef big.Int) Signature {
	res := new(bn256.G1).ScalarMult(hashToPointIndex(aggPub.p, index), secretKey.p)
	res.ScalarMult(res, &anticoef)
	return Signature{p: res}
}

func (secretKey PrivateKey) Marshal() []byte {
	if secretKey.p == nil {
		return nil
	}
	return []byte(secretKey.p.String())
}

// UnmarshalBlsPrivateKey reads the private key from the given byte array
func UnmarshalPrivateKey(data []byte) (PrivateKey, error) {
	p := new(big.Int)
	err := p.UnmarshalJSON(data)
	return PrivateKey{p: p}, err
}
