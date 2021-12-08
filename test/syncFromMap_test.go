package test

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	istanbulCore "github.com/mapprotocol/atlas/consensus/istanbul/core"
	"github.com/mapprotocol/atlas/core/types"
	blscrypto "github.com/mapprotocol/atlas/params/bls"
	"golang.org/x/crypto/sha3"
	"math"
	"testing"
)

func TestFromMap(t *testing.T) {
	hs := getChains(3300, 3305)

	apks := make([]blscrypto.SerializedPublicKey, 0)
	sliceOfBlsPubKeyHexStr := []string{
		"0x41df7be08167a3c7635716418eb42508bee7d97165e6f3482fb55c0a32d2cdc07c8170b97e427c667a87fb8e6f041700b2b1dce0d01a8adadc5816c2c28762ad28730faa9464e65ae7e8031f45fdd7205c499fd92a41ccec5bc97f2dd15da700",
		"0x051fe96e2b46e5708d4081be01ecebadba33a9ec37c9c4219a509b1ff7f1a5f3a3866e4a67050df207cc6546ced94c006f67908ad64656566bb58ebce7ec6bb1a2534c40bf94f6ad205c686ff1ccad1be221c1c82a00cdf989ff98b418810200",
		"0x38030897213e9b7837e600785e3376214948c9bafda2551315fe969206d0be434661c8b4dd6a6298b7f9896efcf3dc002bfd7c2b4d1c7224b0516c76e5ac7fd58a6e72e22b58debcbcaa2b9c72837d6faa6e8e64e02ca222e3ebfd07f25a0580",
		"0xd8b24d419755d8d82b878993d58e7ddd19a19988e00ba55adff574dd9e3df3b45451fe2e56c5793048b0a2c617b11601c451c63e1ce5730f3877a77c026dfdb40349543dfef722dde6f4e06aaf3070ed740d26ae9193d893f5e9d87b67c46080",
	}

	for _, s := range sliceOfBlsPubKeyHexStr {
		pk1 := blscrypto.SerializedPublicKey{}
		_ = pk1.UnmarshalText([]byte(s))
		apks = append(apks, pk1)
	}

	if err := ValidateHeaderExtra(hs, apks); err != nil {
		fmt.Println("verify fail: ", err)
	}
}

func ValidateHeaderExtra(headers []*types.Header, pubKey []blscrypto.SerializedPublicKey) error {
	chainLength := len(headers)

	ps, err := types.ExtractIstanbulExtra(headers[0])
	if err != nil {
		return err
	}
	tmp := ps.AggregatedSeal
	for i := 1; i < chainLength; i++ {
		ex, err := types.ExtractIstanbulExtra(headers[i])
		if err != nil {
			return err
		}

		//verify sign
		addr, err := istanbul.GetSignatureAddress(sigHash(headers[i]).Bytes(), ex.Seal)
		if err != nil {
			return err
		}

		if addr != headers[i].Coinbase {
			return errors.New("verify fail: Coinbase != SignatureAddress")
		}
		//verify AggregatedSeal
		err = verifyAggregatedSeal(headers[i].Hash(), pubKey, ex.AggregatedSeal)
		if err != nil {
			return err
		}
		////verify ParentAggregatedSeal
		err = verifyAggregatedSeal(headers[i-1].Hash(), pubKey, tmp)
		if err != nil {
			return err
		}
		tmp = ex.AggregatedSeal
	}
	return nil
}

func sigHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()

	// Clean seal is required for calculating proposer seal.
	rlp.Encode(hasher, types.IstanbulFilteredHeader(header, false))
	hasher.Sum(hash[:0])
	return hash
}

func verifyAggregatedSeal(headerHash common.Hash, pubKey []blscrypto.SerializedPublicKey, aggregatedSeal types.IstanbulAggregatedSeal) error {
	if len(aggregatedSeal.Signature) != types.IstanbulExtraBlsSignature {
		return errors.New("len error")
	}

	proposalSeal := istanbulCore.PrepareCommittedSeal(headerHash, aggregatedSeal.Round)
	// Find which public keys signed from the provided validator set
	var publicKeys []blscrypto.SerializedPublicKey
	for i := 0; i < len(pubKey); i++ {
		if aggregatedSeal.Bitmap.Bit(i) == 1 {
			publicKeys = append(publicKeys, pubKey[i])
		}
	}
	pknum := int(math.Ceil(float64(2*len(pubKey)) / 3))
	// The length of a valid seal should be greater than the minimum quorum size
	if len(publicKeys) < pknum {
		return errors.New("no enough publicKey")
	}
	err := blscrypto.VerifyAggregatedSignature(publicKeys, proposalSeal, []byte{}, aggregatedSeal.Signature, false, false)
	if err != nil {
		return err
	}

	return nil
}
