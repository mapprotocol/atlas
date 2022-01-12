package bls12381

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"reflect"
)

const (
	PUBLICKEYBYTES    = 96
	SIGNATUREBYTES    = 48
	EPOCHENTROPYBYTES = 16
)

type SerializedPublicKey [PUBLICKEYBYTES]byte

var (
	serializedPublicKeyT = reflect.TypeOf(SerializedPublicKey{})
	serializedSignatureT = reflect.TypeOf(SerializedSignature{})
)

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

func ECDSAToBLS() {}

func PrivateToPublic() {}

func VerifyAggregatedSignature(publicKeys []SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
	blsm := NewBlsManager()
	sign, err := blsm.DecSignature(signature)
	if err != nil {
		return err
	}

	var pks []PublicKey

	for _, v := range publicKeys {
		blsm2 := NewBlsManager()
		pk, err := blsm2.DecPublicKey(v[:])
		if err != nil {
			return err
		}
		pks = append(pks, pk)
	}

	err = blsm.VerifyAggregatedOne(pks, message, sign)
	if err != nil {
		return err
	}
	return nil
}

func AggregateSignatures(signatures [][]byte) ([]byte, error) {
	blsm := NewBlsManager()
	var signObjs []Signature
	for _, signature := range signatures {
		signatureObj, err := blsm.DecSignature(signature)
		if err != nil {
			return nil, err
		}
		signObjs = append(signObjs, signatureObj)
	}
	agsign, err := blsm.Aggregate(signObjs)
	if err != nil {
		return nil, err
	}
	return agsign.Compress().Bytes(), nil
}

func VerifySignature(publicKey SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
	blsm := NewBlsManager()
	signatureObj, err := blsm.DecSignature(signature)
	if err != nil {
		return err
	}

	blsm2 := NewBlsManager()
	pk, err := blsm2.DecPublicKey(publicKey[:])
	if err != nil {
		return err
	}
	err = pk.Verify(message, signatureObj)
	if err != nil {
		return err
	}
	return nil
}

type EpochEntropy [EPOCHENTROPYBYTES]byte

func EncodeEpochSnarkDataCIP22(newValSet []PublicKey, maximumNonSigners, maxValidators uint32, epochIndex uint16, round uint8, blockHash, parentHash EpochEntropy) ([]byte, []byte, error) {
	//pubKeys := []*PublicKey{}
	//blsm := NewBlsManager()
	//for _, pubKey := range newValSet {
	//	publicKeyObj, err := blsm.DecPublicKey(pubKey.Compress().Bytes())
	//	if err != nil {
	//		return nil, nil, err
	//	}
	//	pubKeys = append(pubKeys, publicKeyObj)
	//}
	//
	//return bls.EncodeEpochToBytesCIP22(epochIndex, round, blockHash, parentHash, maximumNonSigners, maxValidators, pubKeys)
	return nil, nil, nil
}

func SerializedSignatureFromBytes(serializedSignature []byte) (SerializedSignature, error) {
	if len(serializedSignature) != SIGNATUREBYTES {
		return SerializedSignature{}, fmt.Errorf("wrong length for serialized signature: expected %d, got %d", SIGNATUREBYTES, len(serializedSignature))
	}
	signatureBytesFixed := SerializedSignature{}
	copy(signatureBytesFixed[:], serializedSignature)
	return signatureBytesFixed, nil
}

func UncompressKey(serialized SerializedPublicKey) ([]byte, error) {
	blsm := NewBlsManager()
	signature, err := blsm.DecPublicKey(serialized[:])
	if err != nil {
		return nil, err
	}
	return signature.Compress().Bytes(), nil
}

func EpochEntropyFromHash(hash common.Hash) EpochEntropy {
	var entropy EpochEntropy
	copy(entropy[:], hash[:EPOCHENTROPYBYTES])
	return entropy
}
