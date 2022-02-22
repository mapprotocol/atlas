package bls

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	bn256_dusk_network "github.com/mapprotocol/atlas/helper/bn256_dusk-network"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"

	"github.com/celo-org/celo-bls-go/bls"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	BLSCryptoType     = 1
	BN256Curve        = 1
	BLS12377Curve     = 2
	BLS12381Curve     = 3
	PUBLICKEYBYTES    = 33
	SIGNATUREBYTES    = 129
	EPOCHENTROPYBYTES = 16
)

var (
	serializedPublicKeyT = reflect.TypeOf(SerializedPublicKey{})
	serializedSignatureT = reflect.TypeOf(SerializedSignature{})
)

type SerializedPublicKey [PUBLICKEYBYTES]byte

// EpochEntropyFromHash truncates the given hash to the length of epoch SNARK entropy.
func EpochEntropyFromHash(hash common.Hash) bls.EpochEntropy {
	var entropy bls.EpochEntropy
	copy(entropy[:], hash[:EPOCHENTROPYBYTES])
	return entropy
}

// MarshalText returns the hex representation of pk.
func (pk SerializedPublicKey) MarshalText() ([]byte, error) {
	return hexutil.Bytes(pk[:]).MarshalText()
}

// UnmarshalText parses a BLS public key in hex syntax.
func (pk *SerializedPublicKey) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("SerializedPublicKey", input, pk[:])
}

// UnmarshalJSON parses a BLS public key in hex syntax.
func (pk *SerializedPublicKey) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(serializedPublicKeyT, input, pk[:])
}

type SerializedSignature [SIGNATUREBYTES]byte

// MarshalText returns the hex representation of sig.
func (sig SerializedSignature) MarshalText() ([]byte, error) {
	return hexutil.Bytes(sig[:]).MarshalText()
}

// UnmarshalText parses a BLS signature in hex syntax.
func (sig *SerializedSignature) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("SerializedSignature", input, sig[:])
}

// UnmarshalJSON parses a BLS signature in hex syntax.
func (sig *SerializedSignature) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(serializedSignatureT, input, sig[:])
}

func EncodeEpochSnarkData(newValSet []SerializedPublicKey, maximumNonSigners uint32, epochIndex uint16) ([]byte, []byte, error) {
	pubKeys := []*bls.PublicKey{}
	for _, pubKey := range newValSet {
		publicKeyObj, err := bls.DeserializePublicKeyCached(pubKey[:])
		if err != nil {
			return nil, nil, err
		}
		defer publicKeyObj.Destroy()

		pubKeys = append(pubKeys, publicKeyObj)
	}

	message, err := bls.EncodeEpochToBytes(epochIndex, maximumNonSigners, pubKeys)
	return message, nil, err
}

func SerializedSignatureFromBytes(serializedSignature []byte) (SerializedSignature, error) {
	fmt.Println("sl", len(serializedSignature))
	if len(serializedSignature) != SIGNATUREBYTES {
		return SerializedSignature{}, fmt.Errorf("wrong length for serialized signature: expected %d, got %d", SIGNATUREBYTES, len(serializedSignature))
	}
	signatureBytesFixed := SerializedSignature{}
	copy(signatureBytesFixed[:], serializedSignature)
	return signatureBytesFixed, nil
}

type BLSCryptoSelector interface {
	ECDSAToBLS(privateKeyECDSA *ecdsa.PrivateKey) ([]byte, error)
	PrivateToPublic(privateKeyBytes []byte) (SerializedPublicKey, error)
	VerifyAggregatedSignature(publicKeys []SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error
	AggregateSignatures(signatures [][]byte) ([]byte, error)
	VerifySignature(publicKey SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error
	EncodeEpochSnarkDataCIP22(newValSet []SerializedPublicKey, maximumNonSigners, maxValidators uint32, epochIndex uint16, round uint8, blockHash, parentHash bls.EpochEntropy) ([]byte, []byte, error)
	UncompressKey(serialized SerializedPublicKey) ([]byte, error)
}

func CryptoType() BLSCryptoSelector {
	switch BLSCryptoType {
	case BN256Curve:
		curve := BN256{}
		return curve
	case BLS12377Curve:
		//curve := BLS12377{}
		return nil //curve
	case BLS12381Curve:
		//curve := BLS12381{}
		return nil //curve
	default:
		// Programming error.
		panic(fmt.Sprintf("unknown bls crypto selection policy: %v", BLSCryptoType))
	}
}

type BLS12377 struct{}

func (BLS12377) ECDSAToBLS(privateKeyECDSA *ecdsa.PrivateKey) ([]byte, error) {
	for i := 0; i < 256; i++ {
		modulus := big.NewInt(0)
		modulus, ok := modulus.SetString(bls.MODULUS377, 10)
		if !ok {
			return nil, errors.New("can't parse modulus")
		}
		privateKeyECDSABytes := crypto.FromECDSA(privateKeyECDSA)

		keyBytes := []byte("ecdsatobls")
		keyBytes = append(keyBytes, uint8(i))
		keyBytes = append(keyBytes, privateKeyECDSABytes...)

		privateKeyBLSBytes := crypto.Keccak256(keyBytes)
		privateKeyBLSBytes[0] &= bls.MODULUSMASK
		privateKeyBLSBig := big.NewInt(0)
		privateKeyBLSBig.SetBytes(privateKeyBLSBytes)
		if privateKeyBLSBig.Cmp(modulus) >= 0 {
			continue
		}

		privateKeyBytes := privateKeyBLSBig.Bytes()
		for len(privateKeyBytes) < len(privateKeyBLSBytes) {
			privateKeyBytes = append([]byte{0x00}, privateKeyBytes...)
		}
		if !bytes.Equal(privateKeyBLSBytes, privateKeyBytes) {
			return nil, fmt.Errorf("private key bytes should have been the same: %s, %s", hex.EncodeToString(privateKeyBLSBytes), hex.EncodeToString(privateKeyBytes))
		}
		// reverse order, as the BLS library expects little endian
		for i := len(privateKeyBytes)/2 - 1; i >= 0; i-- {
			opp := len(privateKeyBytes) - 1 - i
			privateKeyBytes[i], privateKeyBytes[opp] = privateKeyBytes[opp], privateKeyBytes[i]
		}

		privateKeyBLS, err := bls.DeserializePrivateKey(privateKeyBytes)
		if err != nil {
			return nil, err
		}
		defer privateKeyBLS.Destroy()
		privateKeyBLSBytesFromLib, err := privateKeyBLS.Serialize()
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(privateKeyBytes, privateKeyBLSBytesFromLib) {
			return nil, errors.New("private key bytes from library should have been the same")
		}

		return privateKeyBLSBytesFromLib, nil
	}

	return nil, errors.New("couldn't derive a BLS key from an ECDSA key")
}

func (BLS12377) PrivateToPublic(privateKeyBytes []byte) (SerializedPublicKey, error) {
	privateKey, err := bls.DeserializePrivateKey(privateKeyBytes)
	if err != nil {
		return SerializedPublicKey{}, err
	}
	defer privateKey.Destroy()

	publicKey, err := privateKey.ToPublic()
	if err != nil {
		return SerializedPublicKey{}, err
	}
	defer publicKey.Destroy()

	pubKeyBytes, err := publicKey.Serialize()
	if err != nil {
		return SerializedPublicKey{}, err
	}

	pubKeyBytesFixed := SerializedPublicKey{}
	copy(pubKeyBytesFixed[:], pubKeyBytes)

	return pubKeyBytesFixed, nil
}

func (BLS12377) VerifyAggregatedSignature(publicKeys []SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
	publicKeyObjs := []*bls.PublicKey{}
	for _, publicKey := range publicKeys {
		publicKeyObj, err := bls.DeserializePublicKeyCached(publicKey[:])
		if err != nil {
			return err
		}
		defer publicKeyObj.Destroy()
		publicKeyObjs = append(publicKeyObjs, publicKeyObj)
	}
	apk, err := bls.AggregatePublicKeys(publicKeyObjs)
	if err != nil {
		return err
	}
	defer apk.Destroy()

	signatureObj, err := bls.DeserializeSignature(signature)
	if err != nil {
		return err
	}
	defer signatureObj.Destroy()

	err = apk.VerifySignature(message, extraData, signatureObj, shouldUseCompositeHasher, cip22)
	return err
}

func (BLS12377) AggregateSignatures(signatures [][]byte) ([]byte, error) {
	signatureObjs := []*bls.Signature{}
	for _, signature := range signatures {
		signatureObj, err := bls.DeserializeSignature(signature)
		if err != nil {
			return nil, err
		}
		defer signatureObj.Destroy()
		signatureObjs = append(signatureObjs, signatureObj)
	}

	asig, err := bls.AggregateSignatures(signatureObjs)
	if err != nil {
		return nil, err
	}
	defer asig.Destroy()

	asigBytes, err := asig.Serialize()
	if err != nil {
		return nil, err
	}

	return asigBytes, nil
}

func (BLS12377) VerifySignature(publicKey SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
	publicKeyObj, err := bls.DeserializePublicKeyCached(publicKey[:])
	if err != nil {
		return err
	}
	defer publicKeyObj.Destroy()

	signatureObj, err := bls.DeserializeSignature(signature)
	if err != nil {
		return err
	}
	defer signatureObj.Destroy()

	err = publicKeyObj.VerifySignature(message, extraData, signatureObj, shouldUseCompositeHasher, cip22)
	return err
}

func (BLS12377) EncodeEpochSnarkDataCIP22(newValSet []SerializedPublicKey, maximumNonSigners, maxValidators uint32, epochIndex uint16, round uint8, blockHash, parentHash bls.EpochEntropy) ([]byte, []byte, error) {
	pubKeys := []*bls.PublicKey{}
	for _, pubKey := range newValSet {
		publicKeyObj, err := bls.DeserializePublicKeyCached(pubKey[:])
		if err != nil {
			return nil, nil, err
		}
		defer publicKeyObj.Destroy()

		pubKeys = append(pubKeys, publicKeyObj)
	}

	return bls.EncodeEpochToBytesCIP22(epochIndex, round, blockHash, parentHash, maximumNonSigners, maxValidators, pubKeys)
}

func (BLS12377) UncompressKey(serialized SerializedPublicKey) ([]byte, error) {
	publicKey, err := bls.DeserializePublicKeyCached(serialized[:])
	if err != nil {
		return nil, err
	}
	uncompressedBytes, err := publicKey.SerializeUncompressed()
	if err != nil {
		return nil, err
	}
	return uncompressedBytes, nil
}

type BN256 struct{}

func (b BN256) ECDSAToBLS(privateKeyECDSA *ecdsa.PrivateKey) ([]byte, error) {
	return crypto.FromECDSA(privateKeyECDSA), nil
}

func (BN256) PrivateToPublic(privateKeyBytes []byte) (SerializedPublicKey, error) {
	pk, err := bn256_dusk_network.PrivateToPublic(privateKeyBytes)
	pubKeyBytesFixed := SerializedPublicKey{}
	copy(pubKeyBytesFixed[:], pk)
	return pubKeyBytesFixed, err
}

func (BN256) VerifyAggregatedSignature(publicKeys []SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
	sigma := &bn256_dusk_network.Signature{}
	err := sigma.Unmarshal(signature)
	if err != nil {
		return err
	}

	var pks []*bn256_dusk_network.PublicKey
	for _, v := range publicKeys {
		var pk2 *bn256_dusk_network.PublicKey
		err = pk2.Decompress(v[:])
		if err != nil {
			return err
		}
		pks = append(pks, pk2)
	}

	apk, err := bn256_dusk_network.AggregateApk(pks)
	if err != nil {
		return err
	}
	err = bn256_dusk_network.Verify(apk, message, sigma)
	if err != nil {
		return err
	}
	return err
}

func (BN256) AggregateSignatures(signatures [][]byte) ([]byte, error) {
	sigma := &bn256_dusk_network.Signature{}
	for _, v := range signatures {
		var sign bn256_dusk_network.Signature
		err := sign.Unmarshal(v)
		if err != nil {
			return nil, err
		}
		sigma.Aggregate(&sign)
	}
	return sigma.Marshal(), nil
}

func (BN256) VerifySignature(publicKey SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
	var sign bn256_dusk_network.Signature
	err := sign.Unmarshal(signature)
	if err != nil {
		return err
	}

	var pk *bn256_dusk_network.PublicKey
	err = pk.Decompress(publicKey[:])
	if err != nil {
		return err
	}

	err = bn256_dusk_network.Verify(bn256_dusk_network.NewApk(pk), message, &sign)
	if err != nil {
		return err
	}
	return nil
}

func (BN256) EncodeEpochSnarkDataCIP22(newValSet []SerializedPublicKey, maximumNonSigners, maxValidators uint32, epochIndex uint16, round uint8, blockHash, parentHash bls.EpochEntropy) ([]byte, []byte, error) {
	type pack1 struct {
		newValSet         []SerializedPublicKey
		maximumNonSigners uint32
		maxValidators     uint32
		epochIndex        uint16
	}

	type pack2 struct {
		round      uint8
		blockHash  bls.EpochEntropy
		parentHash bls.EpochEntropy
	}

	ret1, err := rlp.EncodeToBytes(pack1{newValSet, maximumNonSigners, maxValidators, epochIndex})
	if err != nil {
		return nil, nil, err
	}
	ret2, err := rlp.EncodeToBytes(pack2{round, blockHash, parentHash})
	if err != nil {
		return nil, nil, err
	}
	return ret1, ret2, nil
}

func (BN256) UncompressKey(serialized SerializedPublicKey) ([]byte, error) {
	var pk bn256_dusk_network.PublicKey
	err := pk.Decompress(serialized[:])
	if err != nil {
		return nil, err
	}
	return pk.Marshal(), nil
}
