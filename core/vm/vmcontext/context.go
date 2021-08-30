package vmcontext

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/consensus"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/core/vm"
	params2 "github.com/mapprotocol/atlas/params"
	"math/big"
)

// chainContext defines methods required to build a context
// a copy of this exist on core.ChainContext (needed to break dependency)
type chainContext interface {
	// Engine retrieves the blockchain's consensus engine.
	Engine() consensus.Engine

	// GetHeader returns the hash corresponding to the given hash and number.
	GetHeader(common.Hash, uint64) *types.Header

	// GetHeaderByNumber returns the hash corresponding number.
	// in the correct fork.
	GetHeaderByNumber(uint64) *types.Header

	// Config returns the blockchain's chain configuration
	Config() *params2.ChainConfig
}

// New creates a new context for use in the EVM.
func New(from common.Address, gasPrice *big.Int, header *types.Header, chain chainContext, txFeeRecipient *common.Address) vm.BlockContext {
	// If we don't have an explicit txFeeRecipient (i.e. not mining), extract from the header
	// The only call that fills the txFeeRecipient, is the ApplyTransaction from the state processor
	// All the other calls, assume that will be retrieved from the header
	var beneficiary common.Address
	if txFeeRecipient == nil {
		beneficiary = header.Coinbase
	} else {
		beneficiary = *txFeeRecipient
	}

	ctx := vm.BlockContext{
		CanTransfer: CanTransfer,
		//Transfer:    TobinTransfer,
		GetHash: GetHashFn(header, chain),
		//VerifySeal:  VerifySealFn(header, chain),
		//Origin:      from,
		Coinbase:    beneficiary,
		BlockNumber: new(big.Int).Set(header.Number),
		Time:        new(big.Int).SetUint64(header.Time),
		//GasPrice:    new(big.Int).Set(gasPrice),

		//GetRegisteredAddress: GetRegisteredAddress,
	}

	//if chain != nil {
	//	ctx.EpochSize = chain.Engine().EpochSize()
	//	ctx.GetValidators = chain.Engine().GetValidators
	//	ctx.GetHeaderByNumber = chain.GetHeaderByNumber
	//} else {
	//	ctx.GetValidators = func(blockNumber *big.Int, headerHash common.Hash) []istanbul.Validator { return nil }
	//	ctx.GetHeaderByNumber = func(uint64) *types.Header { panic("evm context without blockchain context") }
	//}
	return ctx
}

// GetHashFn returns a GetHashFunc which retrieves header hashes by number
func GetHashFn(ref *types.Header, chain chainContext) func(uint64) common.Hash {
	// Cache will initially contain [refHash.parent],
	// Then fill up with [refHash.p, refHash.pp, refHash.ppp, ...]
	var cache []common.Hash

	return func(n uint64) common.Hash {
		// If there's no hash cache yet, make one
		if len(cache) == 0 {
			cache = append(cache, ref.ParentHash)
		}
		if idx := ref.Number.Uint64() - n - 1; idx < uint64(len(cache)) {
			return cache[idx]
		}
		// No luck in the cache, but we can start iterating from the last element we already know
		lastKnownHash := cache[len(cache)-1]
		lastKnownNumber := ref.Number.Uint64() - uint64(len(cache))

		for {
			header := chain.GetHeader(lastKnownHash, lastKnownNumber)
			if header == nil {
				break
			}
			cache = append(cache, header.ParentHash)
			lastKnownHash = header.ParentHash
			lastKnownNumber = header.Number.Uint64() - 1
			if n == lastKnownNumber {
				return lastKnownHash
			}
		}
		return common.Hash{}
	}
}

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas into account to make the transfer valid.
func CanTransfer(db vm.StateDB, addr common.Address, amount *big.Int) bool {
	return db.GetBalance(addr).Cmp(amount) >= 0
}
