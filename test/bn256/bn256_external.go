package bn256

import (
	"crypto/ecdsa"
	blscrypto "github.com/celo-org/celo-bls-go/bls"
	"github.com/mapprotocol/atlas/helper/bls"
)

type BN256 struct{}

func (BN256) ECDSAToBLS(privateKeyECDSA *ecdsa.PrivateKey) ([]byte, error) {
	return nil, nil
}

func (BN256) PrivateToPublic(privateKeyBytes []byte) (bls.SerializedPublicKey, error) {
	return bls.SerializedPublicKey{}, nil
}

func (BN256) VerifyAggregatedSignature(publicKeys []bls.SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
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

func (BN256) AggregateSignatures(signatures [][]byte) ([]byte, error) {
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

func (BN256) VerifySignature(publicKey bls.SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
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

func (BN256) EncodeEpochSnarkDataCIP22(newValSet []bls.SerializedPublicKey, maximumNonSigners, maxValidators uint32, epochIndex uint16, round uint8, blockHash, parentHash blscrypto.EpochEntropy) ([]byte, []byte, error) {
	return nil, nil, nil
}

func (BN256) UncompressKey(serialized bls.SerializedPublicKey) ([]byte, error) {
	var sign Signature
	err := sign.Decompress(serialized[:])
	if err != nil {
		return nil, err
	}
	return sign.Compress(), nil
}
