package account

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts"
	"github.com/mapprotocol/atlas/helper/bls"
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
		log.Error("MustBLSProofOfPossession", "err", err)
	}
	return pop
}

// BLSProofOfPossession generates bls proof of possession
func (a *Account) BLSProofOfPossession() ([]byte, error) {
	privateKey, err := bls.CryptoType().ECDSAToBLS(a.PrivateKey)
	if err != nil {
		privdata := crypto.FromECDSA(a.PrivateKey)
		log.Error("ECDSAToBLS", "privdata", hexutil.Encode(privdata))
		return nil, err
	}
	key, err := bls.DeserializePrivateKey(privateKey)
	if err != nil {
		log.Error("DeserializePrivateKey", "err", err)
		return nil, err
	}
	//pkbytes, err := bls.CryptoType().PrivateToPublic(privateKey)
	//if err != nil {
	//	privdata := crypto.FromECDSA(a.PrivateKey)
	//	log.Error("PrivateToPublic", "err", err, "privdata", hexutil.Encode(privdata), "address", a.Address.String())
	//	return nil, err
	//}
	//pubkey, err := bls.UnmarshalPk(pkbytes[:])
	//if err != nil {
	//	log.Error("bn256.UnmarshalPk", "err", err)
	//	return nil, err
	//}

	signature, err := bls.UnsafeSign(key, a.Address.Bytes())
	if err != nil {
		log.Error("bn256.Sign", "err", err)
		return nil, err
	}
	return signature.Marshal(), nil
}

// ECDSASignature generates ECDSASignature proof of account
func (a *Account) ECDSASignature() []byte {
	privateKey, err := bls.CryptoType().ECDSAToBLS(a.PrivateKey)
	if err != nil {
		privdata := crypto.FromECDSA(a.PrivateKey)
		log.Error("ECDSAToBLS", "privdata", hexutil.Encode(privdata))
		panic(err)
	}
	priv, err := crypto.ToECDSA(privateKey)
	if err != nil {
		panic(err)
	}
	account_ := crypto.PubkeyToAddress(priv.PublicKey)
	hash := accounts.TextHash(crypto.Keccak256(account_[:]))
	sig, err := crypto.Sign(hash, priv)
	if err != nil {
		panic(err)
	}
	//for test
	//recoverPubKey, err := crypto.SigToPub(hash, sig)
	//if err != nil {
	//	panic(err)
	//}
	//log.Info("=== singer  ===", "account", crypto.PubkeyToAddress(*recoverPubKey))
	//log.Info("ECDSASignature", "result", hexutil.Encode(sig))
	return sig
}

// BLSPublicKey returns the bls public key
func (a *Account) BLSPublicKey() (bls.SerializedPublicKey, error) {
	privateKey, err := bls.CryptoType().ECDSAToBLS(a.PrivateKey)
	if err != nil {
		return bls.SerializedPublicKey{}, err
	}

	return bls.CryptoType().PrivateToPublic(privateKey)
}

// BLSG1PublicKey returns the bls G1 public key
func (a *Account) BLSG1PublicKey() (bls.SerializedG1PublicKey, error) {
	privateKey, err := bls.CryptoType().ECDSAToBLS(a.PrivateKey)
	if err != nil {
		return bls.SerializedG1PublicKey{}, err
	}

	return bls.CryptoType().PrivateToG1Public(privateKey)
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
