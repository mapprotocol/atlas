package bn256_keep_network

import (
	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

type PublicKey struct {
	p *bn256.G2
}

// ZeroPublicKey returns zero public key (point at infinity)
func ZeroPublicKey() PublicKey {
	return PublicKey{p: new(bn256.G2).Set(&zeroG2)}
}

// Aggregate adds the given public keys
func (pub PublicKey) Aggregate(onemore PublicKey) PublicKey {
	p := new(bn256.G2).Set(pub.p)
	p.Add(p, onemore.p)
	return PublicKey{p: p}
}

func (pub PublicKey) Marshal() []byte {
	if pub.p == nil {
		return nil
	}
	return pub.p.Marshal()
}

func UnmarshalPublicKey(raw []byte) (PublicKey, error) {
	if raw == nil || len(raw) == 0 {
		return PublicKey{}, nil
	}
	p := new(bn256.G2)
	_, err := p.Unmarshal(raw)
	return PublicKey{p: p}, err
}
