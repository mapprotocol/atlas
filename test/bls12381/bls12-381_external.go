package bls12381

import (
	"crypto/ecdsa"
	blscrypto "github.com/celo-org/celo-bls-go/bls"
	"github.com/mapprotocol/atlas/helper/bls"
)

type BLS12381 struct{}

func (BLS12381) ECDSAToBLS(privateKeyECDSA *ecdsa.PrivateKey) ([]byte, error) {
	return nil, nil
}

func (BLS12381) PrivateToPublic(privateKeyBytes []byte) (bls.SerializedPublicKey, error) {
	return bls.SerializedPublicKey{}, nil
}

func (BLS12381) VerifyAggregatedSignature(publicKeys []bls.SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
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

func (BLS12381) AggregateSignatures(signatures [][]byte) ([]byte, error) {
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

func (BLS12381) VerifySignature(publicKey bls.SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
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

func (BLS12381) EncodeEpochSnarkDataCIP22(newValSet []bls.SerializedPublicKey, maximumNonSigners, maxValidators uint32, epochIndex uint16, round uint8, blockHash, parentHash blscrypto.EpochEntropy) ([]byte, []byte, error) {
	return nil, nil, nil
}

func (BLS12381) UncompressKey(serialized bls.SerializedPublicKey) ([]byte, error) {
	blsm := NewBlsManager()
	signature, err := blsm.DecPublicKey(serialized[:])
	if err != nil {
		return nil, err
	}
	return signature.Compress().Bytes(), nil
}
