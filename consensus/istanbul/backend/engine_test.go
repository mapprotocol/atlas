// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package backend

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
	. "github.com/onsi/gomega"

	"github.com/mapprotocol/atlas/consensus"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/consensus/istanbul/core"
	bccore "github.com/mapprotocol/atlas/core"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/helper/bls"
)

func stopEngine(engine *Backend) {
	engine.StopValidating()
	engine.StopAnnouncing()
}

func TestPrepare(t *testing.T) {
	g := NewGomegaWithT(t)

	chain, engine := newBlockChain(1, true)
	defer stopEngine(engine)
	defer chain.Stop()
	header := makeHeader(chain.Genesis(), engine.config)
	err := engine.Prepare(chain, header)
	g.Expect(err).ToNot(HaveOccurred())

	header.ParentHash = common.BytesToHash([]byte("1234567890"))
	err = engine.Prepare(chain, header)
	g.Expect(err).To(BeIdenticalTo(consensus.ErrUnknownAncestor))
}

func TestMakeBlockWithSignature(t *testing.T) {
	g := NewGomegaWithT(t)

	numValidators := 1
	genesisCfg, nodeKeys := getGenesisAndKeys(numValidators, true)
	chain, engine, _ := newBlockChainWithKeys(false, common.Address{}, false, genesisCfg, nodeKeys[0])

	defer stopEngine(engine)
	defer chain.Stop()
	genesis := chain.Genesis()

	block, err := makeBlock(nodeKeys, chain, engine, genesis)
	g.Expect(err).ToNot(HaveOccurred())

	block2, err := makeBlock(nodeKeys, chain, engine, block)
	g.Expect(err).ToNot(HaveOccurred())

	_, err = makeBlock(nodeKeys, chain, engine, block2)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestSealCommitted(t *testing.T) {
	chain, engine := newBlockChain(1, true)
	defer stopEngine(engine)
	defer chain.Stop()
	// In normal case, the StateProcessResult should be passed into Commit
	engine.abortCommitHook = func(result *core.StateProcessResult) bool { return result == nil }

	block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
	expectedBlock, _ := engine.signBlock(block)

	go func() {
		if err := engine.Seal(chain, block); err != nil {
			t.Errorf("Failed to seal the block: %v", err)
		}
	}()

	newHeadCh := make(chan bccore.ChainHeadEvent, 10)
	sub := chain.SubscribeChainHeadEvent(newHeadCh)
	defer sub.Unsubscribe()

	select {
	case newHead := <-newHeadCh:
		if newHead.Block.Hash() != expectedBlock.Hash() {
			t.Errorf("Expected result block hash of %v, but got %v", expectedBlock.Hash(), newHead.Block.Hash())
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out when waiting for a new block")
	}
}

func TestVerifyHeader(t *testing.T) {
	g := NewGomegaWithT(t)
	chain, engine := newBlockChain(1, true)
	defer stopEngine(engine)
	defer chain.Stop()

	// errEmptyAggregatedSeal case
	block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
	block, _ = engine.signBlock(block)
	err := engine.VerifyHeader(chain, block.Header(), false)
	g.Expect(err).Should(BeIdenticalTo(errEmptyAggregatedSeal))

	// short extra data
	header := block.Header()
	header.Extra = []byte{}
	err = engine.VerifyHeader(chain, header, false)
	g.Expect(err).Should(BeIdenticalTo(errInvalidExtraDataFormat))

	// incorrect extra format
	header.Extra = []byte("0000000000000000000000000000000012300000000000000000000000000000000000000000000000000000000000000000")
	err = engine.VerifyHeader(chain, header, false)
	g.Expect(err).Should(BeIdenticalTo(errInvalidExtraDataFormat))

	// invalid timestamp
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.Time = chain.Genesis().Time() + engine.config.BlockPeriod - 1
	err = engine.VerifyHeader(chain, header, false)
	g.Expect(err).Should(BeIdenticalTo(errInvalidTimestamp))

	// future block
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.Time = uint64(now().Unix() + 10)
	err = engine.VerifyHeader(chain, header, false)
	g.Expect(err).Should(BeIdenticalTo(consensus.ErrFutureBlock))
}

func TestVerifySeal(t *testing.T) {
	g := NewGomegaWithT(t)
	numValidators := 1
	genesisCfg, nodeKeys := getGenesisAndKeys(numValidators, true)
	chain, engine, _ := newBlockChainWithKeys(false, common.Address{}, false, genesisCfg, nodeKeys[0])
	defer stopEngine(engine)
	defer chain.Stop()

	genesis := chain.Genesis()

	// cannot verify genesis
	err := engine.VerifySeal(genesis.Header())
	g.Expect(err).Should(BeIdenticalTo(errUnknownBlock))

	// should verify
	block, err := makeBlock(nodeKeys, chain, engine, genesis)
	g.Expect(err).ToNot(HaveOccurred())
	header := block.Header()
	err = engine.VerifySeal(header)
	g.Expect(err).ToNot(HaveOccurred())

	// change header content and expect to invalidate signature
	header.Number = big.NewInt(4)
	err = engine.VerifySeal(header)
	g.Expect(err).Should(BeIdenticalTo(errInvalidSignature))

	// delete istanbul extra data and expect invalid extra data format
	header = block.Header()
	header.Extra = nil
	err = engine.VerifySeal(header)
	g.Expect(err).Should(BeIdenticalTo(errInvalidExtraDataFormat))

	// modify seal bitmap and expect to fail the quorum check
	header = block.Header()
	extra, err := types.ExtractIstanbulExtra(header)
	g.Expect(err).ToNot(HaveOccurred())
	extra.AggregatedSeal.Bitmap = big.NewInt(0)
	encoded, err := rlp.EncodeToBytes(extra)
	g.Expect(err).ToNot(HaveOccurred())
	header.Extra = append(header.Extra[:types.IstanbulExtraVanity], encoded...)
	err = engine.VerifySeal(header)
	g.Expect(err).Should(BeIdenticalTo(errInsufficientSeals))

	// verifiy the seal on the unmodified block.
	err = engine.VerifySeal(block.Header())
	g.Expect(err).ToNot(HaveOccurred())
}

func TestVerifyHeaders(t *testing.T) {
	numValidators := 1
	genesisCfg, nodeKeys := getGenesisAndKeys(numValidators, true)
	chain, engine, _ := newBlockChainWithKeys(false, common.Address{}, false, genesisCfg, nodeKeys[0])
	defer stopEngine(engine)
	defer chain.Stop()
	genesis := chain.Genesis()

	// success case
	headers := []*types.Header{}
	blocks := []*types.Block{}
	size := 10

	// generate blocks
	for i := 0; i < size; i++ {
		var b *types.Block
		if i == 0 {
			b, _ = makeBlock(nodeKeys, chain, engine, genesis)
		} else {
			b, _ = makeBlock(nodeKeys, chain, engine, blocks[i-1])
		}

		blocks = append(blocks, b)
		headers = append(headers, blocks[i].Header())
	}

	// mock istanbul now() function
	now = func() time.Time {
		return time.Unix(int64(headers[size-1].Time), 0)
	}

	t.Run("Success case", func(t *testing.T) {
		_, results := engine.VerifyHeaders(chain, headers, nil)

		timeout := time.NewTimer(2 * time.Second)
		index := 0
	OUT1:
		for {
			select {
			case err := <-results:
				if err != nil {
					t.Errorf("error mismatch: have %v, want nil", err)
					break OUT1
				}
				index++
				if index == size {
					break OUT1
				}
			case <-timeout.C:
				break OUT1
			}
		}
	})

	t.Run("Abort case", func(t *testing.T) {
		// abort cases
		abort, results := engine.VerifyHeaders(chain, headers, nil)
		timeout := time.NewTimer(2 * time.Second)

		index := 0
	OUT:
		for {
			select {
			case err := <-results:
				if err != nil {
					t.Errorf("error mismatch: have %v, want nil", err)
					break OUT
				}
				index++
				if index == 1 {
					abort <- struct{}{}
				}
				if index >= size {
					t.Errorf("verifyheaders should be aborted")
					break OUT
				}
			case <-timeout.C:
				break OUT
			}
		}
	})

	t.Run("Error Header cases", func(t *testing.T) {
		// error header cases
		headers[2].Number = big.NewInt(100)
		_, results := engine.VerifyHeaders(chain, headers, nil)
		timeout := time.NewTimer(2 * time.Second)
		index := 0
		errors := 0
		expectedErrors := 8
	OUT3:
		for {
			select {
			case err := <-results:
				if err != nil {
					errors++
				}
				index++
				if index == size {
					if errors != expectedErrors {
						t.Errorf("error mismatch: have %v, want %v", errors, expectedErrors)
					}
					break OUT3
				}
			case <-timeout.C:
				break OUT3
			}
		}
	})
}

func TestVerifyHeaderWithoutFullChain(t *testing.T) {
	chain, engine := newBlockChain(1, false)
	defer stopEngine(engine)
	defer chain.Stop()

	t.Run("should allow future block without full chain available", func(t *testing.T) {
		g := NewGomegaWithT(t)
		block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
		header := block.Header()
		header.Time = uint64(now().Unix() + 3)
		err := engine.VerifyHeader(chain, header, false)
		g.Expect(err).To(BeIdenticalTo(errEmptyAggregatedSeal))
	})

	t.Run("should reject future block without full chain available", func(t *testing.T) {
		g := NewGomegaWithT(t)
		block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
		header := block.Header()
		header.Time = uint64(now().Unix() + 10)
		err := engine.VerifyHeader(chain, header, false)
		g.Expect(err).To(BeIdenticalTo(consensus.ErrFutureBlock))
	})
}

func TestPrepareExtra(t *testing.T) {
	g := NewGomegaWithT(t)

	oldValidators := []istanbul.ValidatorData{
		{Address: common.HexToAddress("0x44add0ec310f115a0e603b2d7db9f067778eaf8a")},
		{Address: common.HexToAddress("0x294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212")},
	}

	newValidators := []istanbul.ValidatorData{
		{Address: common.HexToAddress("0x6beaaed781d2d2ab6350f5c4566a2c6eaac407a6")},
		{Address: common.HexToAddress("0x8be76812f765c24641ec63dc2852b378aba2b440")},
	}

	extra, err := rlp.EncodeToBytes(&types.IstanbulExtra{
		AddedValidators:           []common.Address{},
		AddedValidatorsPublicKeys: []bls.SerializedPublicKey{},
		RemovedValidators:         big.NewInt(0),
		Seal:                      []byte{},
		AggregatedSeal:            types.IstanbulAggregatedSeal{},
		ParentAggregatedSeal:      types.IstanbulAggregatedSeal{},
	})
	g.Expect(err).ToNot(HaveOccurred())

	h := &types.Header{
		Extra: append(make([]byte, types.IstanbulExtraVanity), extra...),
	}

	err = writeValidatorSetDiff(h, oldValidators, newValidators)
	g.Expect(err).ToNot(HaveOccurred())

	// the header must have the updated extra data
	updatedExtra, err := types.ExtractIstanbulExtra(h)
	g.Expect(err).ToNot(HaveOccurred())

	var updatedExtraVals []istanbul.ValidatorData
	for i := range updatedExtra.AddedValidators {
		updatedExtraVals = append(updatedExtraVals, istanbul.ValidatorData{
			Address:      updatedExtra.AddedValidators[i],
			BLSPublicKey: updatedExtra.AddedValidatorsPublicKeys[i],
		})
	}

	g.Expect(updatedExtraVals).To(Equal(newValidators), "validators were not properly updated")

	// the validators which were removed were 2, so the bitmap is 11, meaning it should be 3
	g.Expect(updatedExtra.RemovedValidators.Int64()).To(Equal(int64(3)))
}

func TestWriteSeal(t *testing.T) {
	g := NewGomegaWithT(t)

	vanity := bytes.Repeat([]byte{0x00}, types.IstanbulExtraVanity)
	istExtra := &types.IstanbulExtra{
		AddedValidators: []common.Address{
			common.HexToAddress("0x6beaaed781d2d2ab6350f5c4566a2c6eaac407a6"),
			common.HexToAddress("0x8be76812f765c24641ec63dc2852b378aba2b440"),
		},
		AddedValidatorsPublicKeys: []bls.SerializedPublicKey{},
		RemovedValidators:         big.NewInt(12), // 1100, remove third and fourth validators
		Seal:                      []byte{},
		AggregatedSeal:            types.IstanbulAggregatedSeal{Bitmap: big.NewInt(0), Signature: []byte{}, Round: big.NewInt(0)},
		ParentAggregatedSeal:      types.IstanbulAggregatedSeal{Bitmap: big.NewInt(0), Signature: []byte{}, Round: big.NewInt(0)},
	}
	istExtraRaw, err := rlp.EncodeToBytes(&istExtra)
	g.Expect(err).ToNot(HaveOccurred())

	expectedSeal := hexutil.MustDecode("0x29fe2612266a3965321c23a2e0382cd819e992f293d9a0032439728e41201d2c387cc9de5914a734873d79addb76c59ce73c1085a98b968384811b4ad050dddc56")
	g.Expect(expectedSeal).To(HaveLen(types.IstanbulExtraSeal), "incorrect length for seal")

	expectedIstExtra := istExtra
	expectedIstExtra.Seal = expectedSeal

	h := &types.Header{
		Extra: append(vanity, istExtraRaw...),
	}

	// normal case
	err = writeSeal(h, expectedSeal)
	g.Expect(err).NotTo(HaveOccurred())

	// verify istanbul extra-data
	actualIstExtra, err := types.ExtractIstanbulExtra(h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(actualIstExtra).To(Equal(expectedIstExtra))

	// invalid seal
	unexpectedSeal := append(expectedSeal, make([]byte, 1)...)
	err = writeSeal(h, unexpectedSeal)
	g.Expect(err).To(BeIdenticalTo(errInvalidSignature))
}

func TestWriteAggregatedSeal(t *testing.T) {
	g := NewGomegaWithT(t)

	vanity := bytes.Repeat([]byte{0x00}, types.IstanbulExtraVanity)
	istExtra := &types.IstanbulExtra{
		AddedValidators: []common.Address{
			common.HexToAddress("0x6beaaed781d2d2ab6350f5c4566a2c6eaac407a6"),
			common.HexToAddress("0x8be76812f765c24641ec63dc2852b378aba2b440"),
		},
		AddedValidatorsPublicKeys: []bls.SerializedPublicKey{},
		RemovedValidators:         big.NewInt(12), // 1100, remove third and fourth validators
		Seal:                      []byte{},
		AggregatedSeal:            types.IstanbulAggregatedSeal{},
		ParentAggregatedSeal:      types.IstanbulAggregatedSeal{},
	}
	istExtraRaw, err := rlp.EncodeToBytes(&istExtra)
	g.Expect(err).NotTo(HaveOccurred())

	aggregatedSeal := types.IstanbulAggregatedSeal{
		Round:     big.NewInt(2),
		Bitmap:    big.NewInt(3),
		Signature: append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraBlsSignature-3)...),
	}

	expectedIstExtra := istExtra
	expectedIstExtra.AggregatedSeal = aggregatedSeal
	expectedIstExtra.ParentAggregatedSeal = aggregatedSeal

	h := &types.Header{
		Extra: append(vanity, istExtraRaw...),
	}

	// normal case
	err = writeAggregatedSeal(h, aggregatedSeal, false)
	g.Expect(err).NotTo(HaveOccurred())

	err = writeAggregatedSeal(h, aggregatedSeal, true)
	g.Expect(err).NotTo(HaveOccurred())

	// verify istanbul extra-data
	actualIstExtra, err := types.ExtractIstanbulExtra(h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(actualIstExtra).To(Equal(expectedIstExtra))

	// try to write an invalid length seal to the CommitedSeal or ParentCommit field
	invalidAggregatedSeal := types.IstanbulAggregatedSeal{
		Round:     big.NewInt(3),
		Bitmap:    big.NewInt(3),
		Signature: append(aggregatedSeal.Signature, make([]byte, 1)...),
	}
	err = writeAggregatedSeal(h, invalidAggregatedSeal, false)
	g.Expect(err).To(BeIdenticalTo(errInvalidAggregatedSeal))

	err = writeAggregatedSeal(h, invalidAggregatedSeal, true)
	g.Expect(err).To(BeIdenticalTo(errInvalidAggregatedSeal))
}

func Test_01(t *testing.T) {
	str0 := "0xd883010105846765746888676f312e31392e35856c696e757800000000000000f90297ea94acce1869f2878522301b4b1f33a4d6f60622fe9494d60ca3642c427b0144ede812b85ec32596193bbef90104b88004beb0a42b3bf45194cb5fb29475a55a6f422e15578feea34153ea4487f5d9790387473285d16b2ec1817d577215057a9c7b940fe4094d2e0ab7338b6e05b4b90e802f515ea7bbf94d3d2a8490d262ee8c52910922ebd4c6f3aa0b86cee96f4b1389d40c4c48742fa537b7ba95e32ac7352b7be16adb9fe46c694df5a55065f9b880294066d2b6702ccde3d5929749f32505668daf143f6f41189de0082d2904bc36296c633d6d2dcd240d36773fd76c1b1299ef1e55b149c64755cfc9b8b49c71af221670b6d33c568dc30a6cdbb1d5f7b81d6b28d7524cf7991926780e1b60529a2f209ada12411f598f2a03cff2d84c0a1c955a2ded2c607f1c03038a1b205e05f884b84026ea7c086f13d7f784787a93f86f5db82bddf90ed05a7de3c823b6f81771c11b0bfb1bdaffbce7964e14ba8723b28cbc906e6d84896b5d2b4ad589f7809f7e28b8401cd105c059cad4cf30cb1d0ae6cb332741c7bad079c8fef420f3f6b2213b1f68175f9e466078b4b82e9872a506db90938d6a85a069598f3396c35bc475ab8b8e858040000000b841c21e89d6a089ae0fb72ae9748a7f38a84f4c232184929a042b9923a063972b146736d8f858445200aa8c64f75eae9668a8406d5b5cd925a4e144ca84692d6c4101f849856d14fe3fefb84021dee10b2c0206d65695d374e2431884963ade6e4ca4fb8828ee1e58f50e3a27275fcf8ca84f00139871e8697282c09594db8368addabd2ff0f737f0662fcc3c80f849856db4ffbf5fb840151c72d34ba1ce4e7da7a6d86d91b792d66d47c83f52a58500aa78fc09b87d50028398ac8cee9c7683b8286ab9a5027c9ab2dbe90ad3dba06b1e92b179683fc380"
	Extra0, err := hexutil.Decode(str0)
	if err != nil {
		fmt.Println(err)
		return
	}
	str1 := "0xd883010105846765746888676f312e31392e35856c696e757800000000000000f90297ea94acce1869f2878522301b4b1f33a4d6f60622fe9494d60ca3642c427b0144ede812b85ec32596193bbef90104b88004beb0a42b3bf45194cb5fb29475a55a6f422e15578feea34153ea4487f5d9790387473285d16b2ec1817d577215057a9c7b940fe4094d2e0ab7338b6e05b4b90e802f515ea7bbf94d3d2a8490d262ee8c52910922ebd4c6f3aa0b86cee96f4b1389d40c4c48742fa537b7ba95e32ac7352b7be16adb9fe46c694df5a55065f9b880294066d2b6702ccde3d5929749f32505668daf143f6f41189de0082d2904bc36296c633d6d2dcd240d36773fd76c1b1299ef1e55b149c64755cfc9b8b49c71af221670b6d33c568dc30a6cdbb1d5f7b81d6b28d7524cf7991926780e1b60529a2f209ada12411f598f2a03cff2d84c0a1c955a2ded2c607f1c03038a1b205e05f884b84026ea7c086f13d7f784787a93f86f5db82bddf90ed05a7de3c823b6f81771c11b0bfb1bdaffbce7964e14ba8723b28cbc906e6d84896b5d2b4ad589f7809f7e28b8401cd105c059cad4cf30cb1d0ae6cb332741c7bad079c8fef420f3f6b2213b1f68175f9e466078b4b82e9872a506db90938d6a85a069598f3396c35bc475ab8b8e858040000000b841c21e89d6a089ae0fb72ae9748a7f38a84f4c232184929a042b9923a063972b146736d8f858445200aa8c64f75eae9668a8406d5b5cd925a4e144ca84692d6c4101f849856da5fc3f5fb8402c0af35d2e6d2115270ce14c4c23f300e03c54d3ffc6bc6bfeb76528ad84aabb0c92c9fb85ec2838a476fd987a566bc7cb4a9c929893b8945979ac283a1e551380f849856db4ffbf5fb840151c72d34ba1ce4e7da7a6d86d91b792d66d47c83f52a58500aa78fc09b87d50028398ac8cee9c7683b8286ab9a5027c9ab2dbe90ad3dba06b1e92b179683fc380"
	Extra1, err := hexutil.Decode(str1)
	if err != nil {
		fmt.Println(err)
		return
	}
	var istanbulExtra0, istanbulExtra1 *types.IstanbulExtra
	err = rlp.DecodeBytes(Extra0[types.IstanbulExtraVanity:], &istanbulExtra0)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = rlp.DecodeBytes(Extra1[types.IstanbulExtraVanity:], &istanbulExtra1)
	if err != nil {
		fmt.Println(err)
		return
	}
}
