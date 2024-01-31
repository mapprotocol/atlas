package ethereum

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/log"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/consensus/misc"
	"github.com/mapprotocol/atlas/core/types"
)

const (
	allowedFutureBlockTimeSeconds = int64(15)
)

type Validate struct{}

func (v *Validate) ValidateHeaderChain(db types.StateDB, headers []byte, chainType chains.ChainType) (int, error) {
	var chain []*Header
	if err := rlp.DecodeBytes(headers, &chain); err != nil {
		log.Error("rlp decode ethereum headers failed.", "err", err)
		return 0, chains.ErrRLPDecode
	}

	chainLength := len(chain)
	if chainLength == 0 {
		return 0, errors.New("headers cannot be empty")
	}
	if chainLength == 1 {
		if chain[0].Number == nil || chain[0].Difficulty == nil {
			return 0, errors.New("invalid header number or difficulty is nil")
		}
	}
	// Do a sanity check that the provided chain is actually ordered and linked
	for i := 1; i < chainLength; i++ {
		if chain[i].Number == nil || chain[i].Difficulty == nil {
			return 0, errors.New("invalid header number or difficulty is nil")
		}
		if chain[i].Number.Uint64() != chain[i-1].Number.Uint64()+1 {
			hash := chain[i].Hash()
			parentHash := chain[i-1].Hash()
			// Chain broke ancestry, log a message (programming error) and skip insertion
			log.Error("Non contiguous header insert", "number", chain[i].Number, "hash", hash,
				"parent", chain[i].ParentHash, "prevnumber", chain[i-1].Number, "prevhash", parentHash)

			return 0, fmt.Errorf("non contiguous insert: item %d is #%d [%x..], item %d is #%d [%x..] (parent [%x..])", i-1, chain[i-1].Number,
				parentHash.Bytes()[:4], i, chain[i].Number, hash.Bytes()[:4], chain[i].ParentHash[:4])
		}
	}

	hs := NewHeaderStore()
	if err := hs.Load(db); err != nil {
		return 0, err
	}
	////log.Info("validate header stroe", "header", len(hs.HeaderNumber))
	//for _, h := range hs.HeaderNumber {
	//	if _, err := hs.LoadHeader(h.Uint64(), db); err != nil {
	//		return 0, err
	//	}
	//}
	currentNumber := hs.CurrentNumber()
	firstNumber := chain[0].Number

	if firstNumber.Uint64() > currentNumber+1 {
		return 0, fmt.Errorf("non contiguous insert, current number: %d, first number: %d", currentNumber, firstNumber)
	}

	if firstNumber.Uint64() <= currentNumber-MaxHeaderLimit+1 {
		return 0, fmt.Errorf("obsolete block, current number: %d, first number: %d", currentNumber, firstNumber)
	}

	abort, results := v.VerifyHeaders(hs, chain, chainType, db)
	defer close(abort)

	for i := range chain {
		if err := <-results; err != nil {
			return i, err
		}
	}

	return 0, nil
}

func (v *Validate) VerifyHeaders(hs *HeaderStore, headers []*Header, chainType chains.ChainType, db types.StateDB) (chan<- struct{}, <-chan error) {
	// Spawn as many workers as allowed threads
	workers := runtime.GOMAXPROCS(0)
	if len(headers) < workers {
		workers = len(headers)
	}

	// Create a task channel and spawn the verifiers
	var (
		inputs  = make(chan int)
		done    = make(chan int, workers)
		errors  = make([]error, len(headers))
		abort   = make(chan struct{})
		unixNow = time.Now().Unix()
	)
	for i := 0; i < workers; i++ {
		go func() {
			for index := range inputs {
				errors[index] = v.verifyHeaderWorker(hs, headers, index, unixNow, chainType, db)
				done <- index
			}
		}()
	}

	errorsOut := make(chan error, len(headers))
	go func() {
		defer close(inputs)
		var (
			in, out = 0, 0
			checked = make([]bool, len(headers))
			inputs  = inputs
		)
		for {
			select {
			case inputs <- in:
				if in++; in == len(headers) {
					// Reached end of headers. Stop sending to workers.
					inputs = nil
				}
			case index := <-done:
				for checked[index] = true; checked[out]; out++ {
					errorsOut <- errors[out]
					if out == len(headers)-1 {
						return
					}
				}
			case <-abort:
				return
			}
		}
	}()
	return abort, errorsOut
}

func (v *Validate) verifyHeaderWorker(hs *HeaderStore, headers []*Header, index int, unixNow int64, chainType chains.ChainType, db types.StateDB) error {
	var parent *Header
	if index == 0 {
		parent = hs.GetHeader(headers[0].ParentHash, headers[0].Number.Uint64()-1, db)
	} else if headers[index-1].Hash() == headers[index].ParentHash {
		parent = headers[index-1]
	}
	if parent == nil {
		return errUnknownAncestor
	}
	return v.verifyHeader(headers[index], parent, false, unixNow, chainType)
}

func (v *Validate) verifyHeader(header, parent *Header, uncle bool, unixNow int64, chainType chains.ChainType) error {
	// Ensure that the header's extra-data section is of a reasonable size
	if uint64(len(header.Extra)) > ethparams.MaximumExtraDataSize {
		return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), ethparams.MaximumExtraDataSize)
	}
	// Verify the header's timestamp
	if !uncle {
		if header.Time > uint64(unixNow+allowedFutureBlockTimeSeconds) {
			return errFutureBlock
		}
	}
	if header.Time <= parent.Time {
		return errOlderBlockTime
	}
	// Verify the block's difficulty based on its timestamp and parent's difficulty
	//expected := v.CalcDifficulty(chain, header.Time, parent)
	//if expected.Cmp(header.Difficulty) != 0 {
	//	return fmt.Errorf("invalid difficulty: have %v, want %v", header.Difficulty, expected)
	//}

	// Verify that the gas limit is <= 2^63-1
	maxGas := uint64(0x7fffffffffffffff)
	if header.GasLimit > maxGas {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, maxGas)
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}

	// Verify the block's gas usage and (if applicable) verify the base fee.
	lb, _ := chains.ChainType2LondonBlock(chainType)
	cfg := &ethparams.ChainConfig{LondonBlock: lb}
	if !cfg.IsLondon(header.Number) {
		// Verify BaseFee not present before EIP-1559 fork.
		if header.BaseFee != nil {
			return fmt.Errorf("invalid baseFee before fork: have %d, expected 'nil'", header.BaseFee)
		}
		if err := misc.VerifyGaslimit(parent.GasLimit, header.GasLimit); err != nil {
			return err
		}
	} else if err := VerifyEip1559Header(cfg, parent, header); err != nil {
		// Verify the header's EIP-1559 attributes.
		return err
	}
	// Verify that the block number is parent's +1
	if diff := new(big.Int).Sub(header.Number, parent.Number); diff.Cmp(big.NewInt(1)) != 0 {
		return errInvalidNumber
	}

	if err := VerifySeal(header); err != nil {
		return err
	}
	return nil
}

func VerifyEip1559Header(config *ethparams.ChainConfig, parent, header *Header) error {
	// Verify that the gas limit remains within allowed bounds
	parentGasLimit := parent.GasLimit
	if !config.IsLondon(parent.Number) {
		parentGasLimit = parent.GasLimit * ethparams.ElasticityMultiplier
	}
	if err := misc.VerifyGaslimit(parentGasLimit, header.GasLimit); err != nil {
		return err
	}
	// Verify the header is not malformed
	if header.BaseFee == nil {
		return fmt.Errorf("header is missing baseFee")
	}
	// Verify the baseFee is correct based on the parent header.
	expectedBaseFee := CalcBaseFee(config, parent)
	if header.BaseFee.Cmp(expectedBaseFee) != 0 {
		return fmt.Errorf("invalid baseFee: have %s, want %s, parentBaseFee %s, parentGasUsed %d",
			expectedBaseFee, header.BaseFee, parent.BaseFee, parent.GasUsed)
	}
	return nil
}

// CalcBaseFee calculates the basefee of the header.
func CalcBaseFee(config *ethparams.ChainConfig, parent *Header) *big.Int {
	// If the current block is the first EIP-1559 block, return the InitialBaseFee.
	if !config.IsLondon(parent.Number) {
		return new(big.Int).SetUint64(ethparams.InitialBaseFee)
	}

	var (
		parentGasTarget          = parent.GasLimit / ethparams.ElasticityMultiplier
		parentGasTargetBig       = new(big.Int).SetUint64(parentGasTarget)
		baseFeeChangeDenominator = new(big.Int).SetUint64(ethparams.BaseFeeChangeDenominator)
	)
	// If the parent gasUsed is the same as the target, the baseFee remains unchanged.
	if parent.GasUsed == parentGasTarget {
		return new(big.Int).Set(parent.BaseFee)
	}
	if parent.GasUsed > parentGasTarget {
		// If the parent block used more gas than its target, the baseFee should increase.
		gasUsedDelta := new(big.Int).SetUint64(parent.GasUsed - parentGasTarget)
		x := new(big.Int).Mul(parent.BaseFee, gasUsedDelta)
		y := x.Div(x, parentGasTargetBig)
		baseFeeDelta := math.BigMax(
			x.Div(y, baseFeeChangeDenominator),
			common.Big1,
		)

		return x.Add(parent.BaseFee, baseFeeDelta)
	} else {
		// Otherwise if the parent block used less gas than its target, the baseFee should decrease.
		gasUsedDelta := new(big.Int).SetUint64(parentGasTarget - parent.GasUsed)
		x := new(big.Int).Mul(parent.BaseFee, gasUsedDelta)
		y := x.Div(x, parentGasTargetBig)
		baseFeeDelta := x.Div(y, baseFeeChangeDenominator)

		return math.BigMax(
			x.Sub(parent.BaseFee, baseFeeDelta),
			common.Big0,
		)
	}
}
