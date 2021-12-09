package test

import (
	"context"
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
	"math/big"
	"testing"
)

var epoch = int64(1000)
var cache = make(map[int64][]blscrypto.SerializedPublicKey)

func TestFromMap(t *testing.T) {
	hs := getChains(998, 1004)
	if err := ValidateHeaderExtra(hs); err != nil {
		fmt.Println("verify fail: ", err)
	}
}

func ValidateHeaderExtra(headers []*types.Header) error {
	chainLength := len(headers)
	extra, err := types.ExtractIstanbulExtra(headers[0])
	if err != nil {
		return err
	}
	tmp := extra.AggregatedSeal
	for i := 1; i < chainLength; i++ {
		extra, err = types.ExtractIstanbulExtra(headers[i])
		if err != nil {
			return err
		}

		//verify sign
		addr, err := istanbul.GetSignatureAddress(sigHash(headers[i]).Bytes(), extra.Seal)
		if err != nil {
			return err
		}
		if addr != headers[i].Coinbase {
			return errors.New("verify fail: Coinbase != SignatureAddress")
		}

		pubKey, err := getBLSPublickKey(headers[i].Number.Int64())
		if err != nil {
			return err
		}

		//verify AggregatedSeal
		err = verifyAggregatedSeal(headers[i].Hash(), pubKey, extra.AggregatedSeal)
		if err != nil {
			return err
		}
		//verify ParentAggregatedSeal
		if headers[i].Number.Int64() > 1 {
			err = verifyAggregatedSeal(headers[i-1].Hash(), pubKey, tmp)
			if err != nil {
				return err
			}
		}
		tmp = extra.AggregatedSeal
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

func getBLSPublickKey(blockNum int64) ([]blscrypto.SerializedPublicKey, error) {
	num := blockNum / epoch
	//get data quickly
	if cache[num] == nil {
		for num != -1 {
			conn, _ := dialEthConn()

			header, err := conn.HeaderByNumber(context.Background(), big.NewInt(num*epoch))
			if err != nil {
				return nil, err
			}

			extra, err := types.ExtractIstanbulExtra(header)
			if err != nil {
				return nil, err
			}

			if len(extra.AddedValidatorsPublicKeys) != 0 {
				cache[num] = extra.AddedValidatorsPublicKeys
				return cache[num], nil
			} else if cache[num-1] != nil {
				cache[num] = cache[num-1]
				return cache[num], nil
			}

			num = num - 1
		}
	}
	return cache[num], nil
}
