package bls12381

import (
	"errors"
	"github.com/phoreproject/bls/g2pubs"
)

var ErrSigMismatch = errors.New("signature mismatch")
var ErrInvalidSig = errors.New("invalid signature")

type secret struct {
	sk *g2pubs.SecretKey
}

type public struct {
	pk *g2pubs.PublicKey
}

type signature struct {
	sig *g2pubs.Signature
	cb  *CompressedSignature //for optimization reason
}

func (s *secret) Sign(m Message) Signature {
	sig := g2pubs.Sign(m, s.sk)
	return &signature{sig: sig}
}

// PubKey returns the corresponding public key.
func (s *secret) PubKey() (PublicKey, error) {
	pk := g2pubs.PrivToPub(s.sk)
	return &public{pk: pk}, nil
}

// Compress compresses the secret key to a byte slice.
func (s *secret) Compress() CompressedSecret {
	return s.sk.Serialize()
}

// Verify verifies a signature against a message and the public key.
func (p *public) Verify(m Message, sig Signature) error {
	osig, ok := sig.(*signature)
	if !ok {
		return ErrInvalidSig
	}
	if ok := g2pubs.Verify(m, p.pk, osig.sig); ok {
		return nil
	}
	return ErrSigMismatch
}

// Aggregate adds an other public key to the current.
func (p *public) Aggregate(other PublicKey) error {
	op, ok := other.(*public)
	if ok {
		p.pk.Aggregate(op.pk)
		return nil
	} else {
		return errors.New("invalid public key")
	}
}

// Compress compresses the public key to a byte slice.
func (p *public) Compress() CompressedPublic {
	return p.pk.Serialize()
}

// Aggregate adds an other signature to the current.
func (s *signature) Aggregate(other Signature) error {
	osig, ok := other.(*signature)
	if ok {
		s.sig.Aggregate(osig.sig)
		s.cb = nil
		return nil
	}
	return ErrInvalidSig
}

// Compress compresses the signature to a byte slice.
func (s *signature) Compress() CompressedSignature {
	if s.cb == nil {
		cb := CompressedSignature(s.sig.Serialize())
		s.cb = &cb
	}
	var copyBytes CompressedSignature
	copy(copyBytes[:], s.cb[:])
	return copyBytes
}

// serialize just for compare with Compress
func (s *signature) serialize() CompressedSignature {
	return s.sig.Serialize()
}
