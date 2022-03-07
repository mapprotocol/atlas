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
	blscrypto "github.com/mapprotocol/atlas/helper/bls"
	"golang.org/x/crypto/sha3"
	"log"
	"math"
	"math/big"
	"testing"
)

var epoch = int64(200)
var cache = make(map[int64][]blscrypto.SerializedPublicKey)

func TestFromMap(t *testing.T) {
	hs := getChains(598, 604)
	if err := ValidateHeaderExtra(hs); err != nil {
		fmt.Println("verify fail: ", err)
	}
}

func ValidateHeaderExtra(headers []*types.Header) error {
	// init cache for storing epoch
	err := getChangeEpoch(headers[0].Number.Int64())
	if err != nil {
		log.Println("failed to init epoch cache")
		return err
	}

	//first header not verify but get AggregatedSeal
	//because there has no ParentAggregatedSeal
	chainLength := len(headers)

	//verify header from second header to last header
	for i := 1; i < chainLength; i++ {
		extra, err := types.ExtractIstanbulExtra(headers[i])
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

		var pubKey []blscrypto.SerializedPublicKey
		if headers[i].Number.Int64()%epoch != 0 {
			pubKey, err = getBLSPublickKey(headers[i].Number.Int64())
			if err != nil {
				return err
			}
			//log.Println("point",headers[i].Number,headers[i].Number.Int64()%epoch,len(pubKey))
		} else {
			pubKey, err = getBLSPublickKey(headers[i].Number.Int64() - 2)
			if err != nil {
				return err
			}
		}

		//verify AggregatedSeal
		err = verifyAggregatedSeal(headers[i].Hash(), pubKey, extra.AggregatedSeal)
		if err != nil {
			log.Println("failed to verify AggregatedSeal, block num", headers[i].Number)
			return err
		}

		//verify ParentAggregatedSeal except for block 1,
		//because block 1 has no ParentAggregatedSeal.
		if headers[i].Number.Int64() > 1 {
			if headers[i-1].Number.Int64()%epoch == 0 {
				pubKey, err = getBLSPublickKey(headers[i].Number.Int64() - 2)
				if err != nil {
					return err
				}
			}

			err = verifyAggregatedSeal(headers[i-1].Hash(), pubKey, extra.ParentAggregatedSeal)
			if err != nil {
				log.Println("failed to verify ParentAggregatedSeal in epoch point, block num", headers[i].Number)
				return err
			}
		}
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
	//// The length of a valid seal should be greater than the minimum quorum size
	if len(publicKeys) < pknum {
		log.Println("now", len(publicKeys), ",need", pknum, ",all", len(pubKey))
		return errors.New("no enough publicKey")
	}
	err := blscrypto.CryptoType().VerifyAggregatedSignature(publicKeys, proposalSeal, []byte{}, aggregatedSeal.Signature, false, false)
	if err != nil {
		return err
	}

	return nil
}

func TestBLSPublickKey(t *testing.T) {
	_, _ = getBLSPublickKey(801)
	fmt.Println("result ", len(cache))
	for i, v := range cache {
		fmt.Println(i, len(v))
	}
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
				//log.Println("new validator in",num)
				cache[num] = updateValidatorList(extra, num)
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

func TestChangeEpoch(t *testing.T) {
	_ = getChangeEpoch(801)
	fmt.Println("result ", len(cache))
	for i := 0; i < len(cache); i++ {
		fmt.Println(i, len(cache[int64(i)]))
	}
}

func getChangeEpoch(blockNum int64) error {
	num := blockNum / epoch

	for i := int64(0); i <= num; i++ {
		conn, _ := dialEthConn()

		header, err := conn.HeaderByNumber(context.Background(), big.NewInt(i*epoch))
		if err != nil {
			log.Println("failed to get header,num", i*epoch)
			return err
		}
		//log.Println("get block",i*epoch)

		extra, err := types.ExtractIstanbulExtra(header)
		if err != nil {
			return err
		}
		//log.Println(i,"extra",extra.RemovedValidators.BitLen(),extra.AggregatedSeal.Bitmap.BitLen())

		if len(extra.AddedValidatorsPublicKeys) != 0 || extra.RemovedValidators.Int64() != 0 {
			//log.Println("new validator in",i)
			cache[i] = updateValidatorList(extra, i)
			//log.Println("1.new validator in",i,", add ",len(extra.AddedValidatorsPublicKeys))
			//log.Println("2.new validator in",i,", update ",len(cache[i]))
			//log.Println("3.add validator in",i-1,", in fact",extra.AggregatedSeal.Bitmap.BitLen())
		} else if cache[i-1] != nil {
			cache[i] = cache[i-1]
			//log.Println(i,"epoch,no change, member ",len(cache[i]))
		}
		//log.Println("get pk",len(cache[i]),"in epoch",i)
	}

	return nil
}

func updateValidatorList(extra *types.IstanbulExtra, num int64) []blscrypto.SerializedPublicKey {
	if num == 0 {
		//log.Println(len(extra.AddedValidatorsPublicKeys),"pk in block 0")
		return extra.AddedValidatorsPublicKeys
	}

	var valData = make(map[blscrypto.SerializedPublicKey]bool)
	var tempList []blscrypto.SerializedPublicKey
	var oldVal []blscrypto.SerializedPublicKey
	addVal := extra.AddedValidatorsPublicKeys
	list := extra.RemovedValidators
	//log.Println("addVal num",len(addVal),"in epoch",num)

	//
	ok := num
	for ok > -1 {
		if cache[ok] != nil {
			oldVal = cache[ok]
			break
		}
		ok--
	}
	//log.Println("before updated, old num",len(oldVal),",in epoch",ok)

	//
	for _, v := range extra.AddedValidatorsPublicKeys {
		valData[v] = true
	}

	//
	for i, v := range oldVal {
		_, ok := valData[v]
		if list.Bit(i) == 0 && !ok {
			tempList = append(tempList, v)
		}
	}
	//log.Println("after updated, old num",len(tempList),",in epoch",ok)

	tempList = append(tempList, addVal...)
	//log.Println("return num",len(tempList),"in epoch",num)

	return tempList
}
