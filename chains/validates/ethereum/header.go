package ethereum

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"time"

	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/chainsdb"
	"github.com/mapprotocol/atlas/chains/headers/ethereum"
	"github.com/mapprotocol/atlas/core/rawdb"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

const (
	allowedFutureBlockTimeSeconds = int64(15)
)

type Validate struct{}

func (v *Validate) GetCurrentHeaderNumber(chain rawdb.ChainType) (uint64, error) {
	if !chains.IsSupportedChain(chain) {
		return 0, errNotSupportChain
	}

	store, err := chainsdb.GetStoreMgr(chain)
	if err != nil {
		return 0, err
	}
	return store.CurrentHeaderNumber(), nil
}

func (v *Validate) GetHashByNumber(chain rawdb.ChainType, number uint64) (common.Hash, error) {
	if !chains.IsSupportedChain(chain) {
		return common.Hash{}, errNotSupportChain
	}

	store, err := chainsdb.GetStoreMgr(chain)
	if err != nil {
		return common.Hash{}, err
	}
	return store.ReadCanonicalHash(number), nil
}

func (v *Validate) ValidateHeaderChain(chainID uint64, chain []*ethereum.Header) (int, error) {
	chainLength := len(chain)
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

	firstNumber := chain[0].Number
	currentNumber, err := v.GetCurrentHeaderNumber(chains.ChainTypeETH)
	if err != nil {
		return 0, err
	}
	if firstNumber.Uint64() > currentNumber+1 {
		return 0, fmt.Errorf("non contiguous insert, current number: %d, first number: %d", currentNumber, firstNumber)
	}

	if firstNumber.Cmp(big.NewInt(1)) == 0 {
		chainsdb.Genesis(chain[0].Genesis(chainID), chains.ChainTypeETH)
	}

	abort, results := v.VerifyHeaders(chain)
	defer close(abort)

	for i := range chain {
		if err := <-results; err != nil {
			return i, err
		}
	}

	return 0, nil
}

func (v *Validate) VerifyHeaders(headers []*ethereum.Header) (chan<- struct{}, <-chan error) {
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
				errors[index] = v.verifyHeaderWorker(headers, index, unixNow)
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

func (v *Validate) verifyHeaderWorker(headers []*ethereum.Header, index int, unixNow int64) error {
	var parent *ethereum.Header
	if index == 0 {
		s, err := chainsdb.GetStoreMgr(chains.ChainTypeETH)
		if err != nil {
			return err
		}
		parent = s.ReadHeader(headers[0].ParentHash, headers[0].Number.Uint64()-1)
	} else if headers[index-1].Hash() == headers[index].ParentHash {
		parent = headers[index-1]
	}
	if parent == nil {
		return errUnknownAncestor
	}
	return v.verifyHeader(headers[index], parent, false, unixNow)
}

func (v *Validate) verifyHeader(header, parent *ethereum.Header, uncle bool, unixNow int64) error {
	// Ensure that the header's extra-data section is of a reasonable size
	if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), params.MaximumExtraDataSize)
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

	// Verify that the gas limit remains within allowed bounds
	diff := int64(parent.GasLimit) - int64(header.GasLimit)
	if diff < 0 {
		diff *= -1
	}
	limit := parent.GasLimit / params.GasLimitBoundDivisor

	if uint64(diff) >= limit || header.GasLimit < params.MinGasLimit {
		return fmt.Errorf("invalid gas limit: have %d, want %d += %d", header.GasLimit, parent.GasLimit, limit)
	}
	// Verify that the block number is parent's +1
	if diff := new(big.Int).Sub(header.Number, parent.Number); diff.Cmp(big.NewInt(1)) != 0 {
		return errInvalidNumber
	}

	return nil
}
