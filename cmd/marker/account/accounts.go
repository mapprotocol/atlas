package account

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	blscrypto "github.com/mapprotocol/atlas/helper/bls"
	bn256 "github.com/mapprotocol/atlas/helper/bls"
	"io/ioutil"
)

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
	key := bn256.NewKey(a.PrivateKey.D)
	keybytes := crypto.FromECDSA(a.PrivateKey)
	pkbytes, err := blscrypto.CryptoType().PrivateToPublic(keybytes)
	if err != nil {
		return nil, err
	}
	pubkey, err := bn256.UnmarshalPk(pkbytes[:])
	if err != nil {
		return nil, err
	}
	signature, err := bn256.Sign(&key, pubkey, a.Address.Bytes())
	if err != nil {
		return nil, err
	}
	return signature.Marshal(), nil
}

// BLSPublicKey returns the bls public key
func (a *Account) BLSPublicKey() (blscrypto.SerializedPublicKey, error) {
	privateKey, err := blscrypto.CryptoType().ECDSAToBLS(a.PrivateKey)
	if err != nil {
		return blscrypto.SerializedPublicKey{}, err
	}

	return blscrypto.CryptoType().PrivateToPublic(privateKey)
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
func LoadAccount(path string, password string) (*Account, error) {
	keyjson, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error("LoadAccount", "msg", fmt.Errorf("failed to read the keyfile at '%s': %v", path, err))
		return nil, err
	}
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		log.Error("LoadAccount", "msg", fmt.Errorf("error decrypting key: %v", err))
		return nil, err
	}
	priKey1 := key.PrivateKey
	publicAddr := crypto.PubkeyToAddress(priKey1.PublicKey)
	var addr common.Address
	addr.SetBytes(publicAddr.Bytes())

	return &Account{
		Address:    addr,
		PrivateKey: priKey1,
	}, nil
}
