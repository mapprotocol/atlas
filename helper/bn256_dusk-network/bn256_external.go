package bn256_dusk_network

//import (
//	"crypto/ecdsa"
//	blscrypto "github.com/celo-org/celo-bls-go/bls"
//	"github.com/ethereum/go-ethereum/rlp"
//	"github.com/mapprotocol/atlas/helper/bls"
//)
//
//type SerializedPublicKey [96]byte
//
//type BN256 struct{}
//
//func (BN256) ECDSAToBLS(privateKeyECDSA *ecdsa.PrivateKey) ([]byte, error) {
//	return nil, nil
//}
//
//func (BN256) PrivateToPublic(privateKeyBytes []byte) (bls.SerializedPublicKey, error) {
//	pk,err := PrivateToPublic(privateKeyBytes)
//	pubKeyBytesFixed := bls.SerializedPublicKey{}
//	copy(pubKeyBytesFixed[:], pk)
//	return pubKeyBytesFixed, err
//}
//
//func (BN256) VerifyAggregatedSignature(publicKeys []bls.SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
//	sigma := &Signature{}
//	err := sigma.Unmarshal(signature)
//	if err != nil {
//		return err
//	}
//
//	var pks []*PublicKey
//	for _, v := range publicKeys {
//		var pk2 *PublicKey
//		err = pk2.Decompress(v[:])
//		if err != nil {
//			return err
//		}
//		pks = append(pks, pk2)
//	}
//
//	apk, err := AggregateApk(pks)
//	if err != nil {
//		return err
//	}
//	err = Verify(apk, message, sigma)
//	if err != nil {
//		return err
//	}
//	return err
//}
//
//func (BN256) AggregateSignatures(signatures [][]byte) ([]byte, error) {
//	sigma := &Signature{}
//	for _, v := range signatures {
//		var sign Signature
//		err := sign.Unmarshal(v)
//		if err != nil {
//			return nil, err
//		}
//		sigma.Aggregate(&sign)
//	}
//	return sigma.Marshal(), nil
//}
//
//func (BN256) VerifySignature(publicKey bls.SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
//	var sign Signature
//	err := sign.Unmarshal(signature)
//	if err != nil {
//		return err
//	}
//
//	var pk *PublicKey
//	err = pk.Decompress(publicKey[:])
//	if err != nil {
//		return err
//	}
//
//	err = Verify(NewApk(pk), message, &sign)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (BN256) EncodeEpochSnarkDataCIP22(newValSet []bls.SerializedPublicKey, maximumNonSigners, maxValidators uint32, epochIndex uint16, round uint8, blockHash, parentHash blscrypto.EpochEntropy) ([]byte, []byte, error) {
//	type pack1 struct {
//		newValSet         []bls.SerializedPublicKey
//		maximumNonSigners uint32
//		maxValidators     uint32
//		epochIndex        uint16
//	}
//
//	type pack2 struct {
//		round      uint8
//		blockHash  blscrypto.EpochEntropy
//		parentHash blscrypto.EpochEntropy
//	}
//
//	ret1, err := rlp.EncodeToBytes(pack1{newValSet, maximumNonSigners, maxValidators, epochIndex})
//	if err != nil {
//		return nil, nil, err
//	}
//	ret2, err := rlp.EncodeToBytes(pack2{round, blockHash, parentHash})
//	if err != nil {
//		return nil, nil, err
//	}
//	return ret1, ret2, nil
//}
//
//func (BN256) UncompressKey(serialized bls.SerializedPublicKey) ([]byte, error) {
//	var pk PublicKey
//	err := pk.Decompress(serialized[:])
//	if err != nil {
//		return nil, err
//	}
//	return pk.Marshal(), nil
//}
