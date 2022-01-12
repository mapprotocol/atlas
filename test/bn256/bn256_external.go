package bn256

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

type EpochEntropy [EPOCHENTROPYBYTES]byte

func ECDSAToBLS() {}

func PrivateToPublic() {}

func VerifyAggregatedSignature(publicKeys []SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
	sigma := &Signature{}
	err := sigma.Unmarshal(signature)
	if err != nil {
		return err
	}

	var pks []*PublicKey
	for _, v := range publicKeys {
		var pk2 *PublicKey
		err = pk2.UnmarshalText(v[:])
		if err != nil {
			return err
		}
		pks = append(pks, pk2)
	}

	apk, err := AggregateApk(pks)
	if err != nil {
		return err
	}
	err = Verify(apk, message, sigma)
	if err != nil {
		return err
	}
	return err
}

func AggregateSignatures(signatures [][]byte) ([]byte, error) {
	sigma := &Signature{}
	for _, v := range signatures {
		var sign Signature
		err := sign.Unmarshal(v)
		if err != nil {
			return nil, err
		}
		sigma.Aggregate(&sign)
	}
	return sigma.Marshal(), nil
}

func VerifySignature(publicKey PublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
	var sign Signature
	err := sign.Unmarshal(signature)
	if err != nil {
		return err
	}

	p, err := publicKey.MarshalText()
	if err != nil {
		return err
	}

	var pk *PublicKey
	err = pk.UnmarshalText(p)
	if err != nil {
		return err
	}

	err = Verify(NewApk(pk), message, &sign)
	if err != nil {
		return err
	}
	return nil
}

func EncodeEpochSnarkDataCIP22(newValSet []PublicKey, maximumNonSigners, maxValidators uint32, epochIndex uint16, round uint8, blockHash, parentHash EpochEntropy) ([]byte, []byte, error) {
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
	var sign Signature
	err := sign.Decompress(serialized[:])
	if err != nil {
		return nil, err
	}
	return sign.Compress(), nil
}

// EpochEntropyFromHash truncates the given hash to the length of epoch SNARK entropy.
func EpochEntropyFromHash(hash common.Hash) EpochEntropy {
	var entropy EpochEntropy
	copy(entropy[:], hash[:EPOCHENTROPYBYTES])
	return entropy
}
