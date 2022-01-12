package bls12381

func ECDSAToBLS() {}

func PrivateToPublic() {}

func VerifyAggregatedSignature(publicKeys []PublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
	blsm := NewBlsManager()
	sign, err := blsm.DecSignature(signature)
	if err != nil {
		return err
	}
	err = blsm.VerifyAggregatedOne(publicKeys, message, sign)
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

func VerifySignature(publicKey PublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
	blsm := NewBlsManager()
	signatureObj, err := blsm.DecSignature(signature)
	if err != nil {
		return err
	}
	err = publicKey.Verify(message, signatureObj)
	if err != nil {
		return err
	}
	return nil
}

type EpochEntropy [16]byte

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

func SerializedSignatureFromBytes(serializedSignature []byte) (Signature, error) {
	blsm := NewBlsManager()
	signature, err := blsm.DecSignature(serializedSignature)
	return signature, err
}

func UncompressKey(serialized PublicKey) ([]byte, error) {
	return serialized.Compress().Bytes(), nil
}
