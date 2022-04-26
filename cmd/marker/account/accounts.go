package account

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
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
		panic(err)
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

	signature, err := bls.Sign(key, key.ToPublic(), a.Address.Bytes())
	if err != nil {
		log.Error("bn256.Sign", "err", err)
		return nil, err
	}
	return signature.Marshal(), nil
}

// BLSPublicKey returns the bls public key
func (a *Account) BLSPublicKey() (bls.SerializedPublicKey, error) {
	privateKey, err := bls.CryptoType().ECDSAToBLS(a.PrivateKey)
	if err != nil {
		return bls.SerializedPublicKey{}, err
	}

	return bls.CryptoType().PrivateToPublic(privateKey)
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
