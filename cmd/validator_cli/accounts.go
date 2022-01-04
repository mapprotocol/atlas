package main

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/celo-org/celo-bls-go/bls"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	blscrypto "github.com/mapprotocol/atlas/helper/bls"
	"github.com/tyler-smith/go-bip39"
)

// MustNewMnemonic creates a new mnemonic (panics on error)
func MustNewMnemonic() string {
	res, err := NewMnemonic()
	if err != nil {
		panic(err)
	}
	return res
}

// NewMnemonic creates a new mnemonic
func NewMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

// Account represents a atlas Account
type Account struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
}

// MustBLSProofOfPossession variant of BLSProofOfPossession that panics on error
func (a *Account) MustBLSProofOfPossession() []byte {
	pop, err := a.BLSProofOfPossession()
	if err != nil {
		panic(err)
	}
	return pop
}

// BLSProofOfPossession generates bls proof of possession
func (a *Account) BLSProofOfPossession() ([]byte, error) {
	privateKeyBytes, err := blscrypto.ECDSAToBLS(a.PrivateKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := bls.DeserializePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	defer privateKey.Destroy()

	signature, err := privateKey.SignPoP(a.Address.Bytes())
	if err != nil {
		return nil, err
	}
	defer signature.Destroy()

	signatureBytes, err := signature.Serialize()
	if err != nil {
		return nil, err
	}
	return signatureBytes, nil
}

// BLSPublicKey returns the bls public key
func (a *Account) BLSPublicKey() (blscrypto.SerializedPublicKey, error) {
	privateKey, err := blscrypto.ECDSAToBLS(a.PrivateKey)
	if err != nil {
		return blscrypto.SerializedPublicKey{}, err
	}

	return blscrypto.PrivateToPublic(privateKey)
}

// PublicKeyHex hex representation of the public key
func (a *Account) PublicKey() []byte {
	return crypto.FromECDSAPub(&a.PrivateKey.PublicKey)
}

// PrivateKeyHex hex representation of the private key
func (a *Account) PrivateKeyHex() string {
	return common.Bytes2Hex(crypto.FromECDSA(a.PrivateKey))
}

func (a *Account) String() string {
	return fmt.Sprintf("{ address: %s\tprivateKey: %s }",
		a.Address.Hex(),
		a.PrivateKeyHex(),
	)
}
