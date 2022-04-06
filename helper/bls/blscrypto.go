package bls

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"errors"
	"github.com/celo-org/celo-bls-go/bls"
	"math/big"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"reflect"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	BLSCryptoType     = 1
	BN256Curve        = 1
	BLS12377Curve     = 2
	BLS12381Curve     = 3
	PUBLICKEYBYTES    = 129
	G1PUBLICKEYBYTES  = 129
	SIGNATUREBYTES    = 64
	EPOCHENTROPYBYTES = 16
)

var (
	serializedPublicKeyT = reflect.TypeOf(SerializedPublicKey{})
	serializedSignatureT = reflect.TypeOf(SerializedSignature{})
)

// EpochEntropy is a string of unprediactable bytes included in the epoch SNARK data
// to make prediction of future epoch message values infeasible.
type EpochEntropy [EPOCHENTROPYBYTES]byte

type SerializedPublicKey [PUBLICKEYBYTES]byte

// EpochEntropyFromHash truncates the given hash to the length of epoch SNARK entropy.
func EpochEntropyFromHash(hash common.Hash) EpochEntropy {
	var entropy EpochEntropy
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

type SerializedG1PublicKey [G1PUBLICKEYBYTES]byte

// MarshalText returns the hex representation of pk.
func (pk SerializedG1PublicKey) MarshalText() ([]byte, error) {
	return hexutil.Bytes(pk[:]).MarshalText()
}

// UnmarshalText parses a BLS public key in hex syntax.
func (pk *SerializedG1PublicKey) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("SerializedPublicKey", input, pk[:])
}

// UnmarshalJSON parses a BLS public key in hex syntax.
func (pk *SerializedG1PublicKey) UnmarshalJSON(input []byte) error {
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
	PrivateToG1Public(privateKeyBytes []byte) (SerializedG1PublicKey, error)
	VerifyAggregatedSignature(publicKeys []SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error
	AggregateSignatures(signatures [][]byte) ([]byte, error)
	VerifySignature(publicKey SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error
	EncodeEpochSnarkDataCIP22(newValSet []SerializedPublicKey, maximumNonSigners, maxValidators uint32, epochIndex uint16, round uint8, blockHash, parentHash EpochEntropy) ([]byte, []byte, error)
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

type BN256 struct{}

const (
	MODULUS256 = "21888242871839275222246405745257275088548364400416034343698204186575808495617"
	MODULUSBITS = 254
	MODULUSMASK = 63 // == 2**(254-(256-8)) - 1
)
//func (BN256) ECDSAToBLS(privateKeyECDSA *ecdsa.PrivateKey) ([]byte, error) {
//	return crypto.FromECDSA(privateKeyECDSA), nil
//}
func (BN256) ECDSAToBLS(privateKeyECDSA *ecdsa.PrivateKey) ([]byte, error) {
	for i := 0; i < 256; i++ { // 最多尝试256次
		modulus := big.NewInt(0)
		modulus, ok := modulus.SetString(MODULUS256, 10)
		if !ok {
			return nil, errors.New("can't parse modulus")
		}
		privateKeyECDSABytes := crypto.FromECDSA(privateKeyECDSA)

		keyBytes := []byte("ecdsatobls")
		keyBytes = append(keyBytes, uint8(i))
		keyBytes = append(keyBytes, privateKeyECDSABytes...) // keyBytes = "ecdsatobls" || byte(i) || bytes(k)

		privateKeyBLSBytes := crypto.Keccak256(keyBytes) // privateKeyBlsBytes =  keccak256(keyBytes)
		privateKeyBLSBytes[0] &= MODULUSMASK
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
func (BN256) PrivateToPublic(privateKeyBytes []byte) (SerializedPublicKey, error) {
	pk, err := PrivateToPublic(privateKeyBytes)
	pubKeyBytesFixed := SerializedPublicKey{}
	copy(pubKeyBytesFixed[:], pk)
	return pubKeyBytesFixed, err
}

func (BN256) PrivateToG1Public(privateKeyBytes []byte) (SerializedG1PublicKey, error) {
	pubKeyBytesFixed := SerializedG1PublicKey{}
	return pubKeyBytesFixed, nil
}

func (BN256) VerifyAggregatedSignature(publicKeys []SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
	sigma := Signature{}
	err := sigma.Unmarshal(signature)
	if err != nil {
		return err
	}

	var pks []*PublicKey
	for _, v := range publicKeys {
		pk, err := UnmarshalPk(v[:])
		if err != nil {
			return err
		}
		pks = append(pks, pk)
	}

	apk, err := AggregateApk(pks)
	if err != nil {
		return err
	}

	err = Verify(apk, message, &sigma)
	if err != nil {
		return err
	}
	return err
}

func (BN256) AggregateSignatures(signatures [][]byte) ([]byte, error) {
	var signs Signature
	err := signs.Unmarshal(signatures[0])
	if err != nil {
		return nil, err
	}
	for i := 1; i < len(signatures); i++ {
		var sign Signature
		err := sign.Unmarshal(signatures[i])
		if err != nil {
			return nil, err
		}
		signs.Aggregate(&sign)
	}
	return signs.Marshal(), nil
}

func (BN256) VerifySignature(publicKey SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
	var sign Signature
	err := sign.Unmarshal(signature)
	if err != nil {
		return err
	}
	pk, err := UnmarshalPk(publicKey[:])
	if err != nil {
		return err
	}

	err = Verify(NewApk(pk), message, &sign)
	if err != nil {
		return err
	}
	return nil
}

func (BN256) EncodeEpochSnarkDataCIP22(newValSet []SerializedPublicKey, maximumNonSigners, maxValidators uint32, epochIndex uint16, round uint8, blockHash, parentHash EpochEntropy) ([]byte, []byte, error) {
	type pack1 struct {
		newValSet         []SerializedPublicKey
		maximumNonSigners uint32
		maxValidators     uint32
		epochIndex        uint16
	}

	type pack2 struct {
		round      uint8
		blockHash  EpochEntropy
		parentHash EpochEntropy
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
	pk, err := UnmarshalPk(serialized[:])
	if err != nil {
		return nil, err
	}
	return pk.Marshal(), nil
}
