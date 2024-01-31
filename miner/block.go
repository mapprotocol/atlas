package miner

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/mapprotocol/atlas/consensus/misc"
	"github.com/mapprotocol/atlas/core/vm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/mapprotocol/atlas/consensus"
	"github.com/mapprotocol/atlas/contracts/blockchain_parameters"
	"github.com/mapprotocol/atlas/contracts/random"
	"github.com/mapprotocol/atlas/core"
	"github.com/mapprotocol/atlas/core/chain"
	ethChain "github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/types"
)

// blockState is the collection of modified state that is used to assemble a block
type blockState struct {
	signer types.Signer

	state    *state.StateDB // apply state changes here
	tcount   int            // tx count in cycle
	gasPool  *core.GasPool  // available gas used to pack transactions
	gasLimit uint64

	header         *types.Header
	txs            []*types.Transaction
	receipts       []*types.Receipt
	randomness     *types.Randomness // The types.Randomness of the last block by mined by this worker.
	txFeeRecipient common.Address
}

func getGasLimitByWork(w *worker, parent *types.Block, header *types.Header, vmRunner vm.EVMRunner) uint64 {
	gaslimit := uint64(0)
	if w.chainConfig.IsCalc(header.Number) {
		ceil := blockchain_parameters.GetBlockGasLimitOrDefault(vmRunner, true)
		//fmt.Println("===getGasLimitByWork2", "parent", parent.GasLimit(), "ceil", ceil)
		gaslimit = chain.CalcGasLimit(parent.GasLimit(), ceil)
	} else {
		// fmt.Println("******* not here **********")
		gaslimit = chain.CalcGasLimit(parent.GasLimit(), w.config.GasCeil)
	}
	return gaslimit
}

// prepareBlock intializes a new blockState that is ready to have transaction included to.
func prepareBlock(w *worker) (*blockState, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	timestamp := time.Now().Unix()
	parent := w.chain.CurrentBlock()
	if parent.Time() >= uint64(timestamp) {
		timestamp = int64(parent.Time() + 1)
	}
	num := parent.Number()
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     num.Add(num, common.Big1),
		Extra:      w.extra,
		Time:       uint64(timestamp),
		GasLimit:   chain.CalcGasLimit(parent.GasLimit(), w.config.GasCeil),
	}
	if w.chainConfig.IsLondon(header.Number) {
		header.BaseFee = misc.CalcBaseFee(w.chainConfig, parent.Header())
		if !w.chainConfig.IsLondon(parent.Number()) {
			parentGasLimit := parent.GasLimit() * params.ElasticityMultiplier
			header.GasLimit = chain.CalcGasLimit(parentGasLimit, w.config.GasCeil)
		}
	}

	txFeeRecipient := w.txFeeRecipient
	if w.txFeeRecipient != w.validator {
		txFeeRecipient = w.validator
		log.Warn("TxFeeRecipient and Validator flags set before split etherbase fork is active. Defaulting to the given validator address for the validator.")
	}

	// Only set the validator if our consensus engine is running (avoid spurious block rewards)
	if w.isRunning() {
		if txFeeRecipient == (common.Address{}) {
			return nil, errors.New("refusing to mine without etherbase")
		}
		header.Coinbase = txFeeRecipient
	}
	// Note: The parent seal will not be set when not validating
	if err := w.engine.Prepare(w.chain, header); err != nil {
		log.Error("Failed to prepare header for mining", "err", err)
		return nil, fmt.Errorf("failed to prepare header for mining: %w", err)
	}

	// Initialize the block state itself
	state, err := w.chain.StateAt(parent.Root())
	if err != nil {
		return nil, fmt.Errorf("failed to get the parent state: %w", err)
	}

	vmRunner := w.chain.NewEVMRunner(header, state)
	header.GasLimit = getGasLimitByWork(w, parent, header, vmRunner)

	b := &blockState{
		signer:         types.NewLondonSigner(w.chainConfig.ChainID),
		state:          state,
		tcount:         0,
		gasLimit:       blockchain_parameters.GetBlockGasLimitOrDefault(vmRunner, false),
		header:         header,
		txFeeRecipient: txFeeRecipient,
	}
	if w.chainConfig.IsCalc(header.Number) {
		b.gasLimit = header.GasLimit
	}
	b.gasPool = new(core.GasPool).AddGas(b.gasLimit)

	// Play our part in generating the random beacon.
	if w.isRunning() && random.IsRunning(vmRunner) {
		istanbul, ok := w.engine.(consensus.Istanbul)
		if !ok {
			log.Crit("Istanbul consensus engine must be in use for the randomness beacon")
		}
		lastCommitment, err := random.GetLastCommitment(vmRunner, w.validator)
		if err != nil {
			return nil, fmt.Errorf("failed to get last commitment: %w", err)
		}

		lastRandomness := common.Hash{}
		if (lastCommitment != common.Hash{}) {
			lastRandomnessParentHash := rawdb.ReadRandomCommitmentCache(w.db, lastCommitment)
			if (lastRandomnessParentHash == common.Hash{}) {
				return nil, errors.New("failed to get last randomness cache entry")
			}

			var err error
			lastRandomness, _, err = istanbul.GenerateRandomness(lastRandomnessParentHash)
			if err != nil {
				return nil, fmt.Errorf("failed to generate last randomness: %w", err)
			}
		}

		_, newCommitment, err := istanbul.GenerateRandomness(b.header.ParentHash)
		if err != nil {
			return nil, fmt.Errorf("failed to generate new randomness: %w", err)
		}

		err = random.RevealAndCommit(vmRunner, lastRandomness, newCommitment, w.validator)
		if err != nil {
			return nil, fmt.Errorf("failed to reveal and commit randomness: %w", err)
		}
		// always true (EIP158)
		b.state.IntermediateRoot(true)

		b.randomness = &types.Randomness{Revealed: lastRandomness, Committed: newCommitment}
	} else {
		b.randomness = &types.Randomness{}
	}

	return b, nil
}

// selectAndApplyTransactions selects and applies transactions to the in flight block state.
func (b *blockState) selectAndApplyTransactions(ctx context.Context, w *worker) error {
	// Fill the block with all available pending transactions.
	pending := w.eth.TxPool().Pending(false)

	// TODO: should this be a fatal error?
	//if err != nil {
	//	log.Error("Failed to fetch pending transactions", "err", err)
	//	return nil
	//}

	// Short circuit if there is no available pending transactions.
	if len(pending) == 0 {
		return nil
	}
	// Split the pending transactions into locals and remotes
	localTxs, remoteTxs := make(map[common.Address]types.Transactions), pending
	for _, account := range w.eth.TxPool().Locals() {
		if txs := remoteTxs[account]; len(txs) > 0 {
			delete(remoteTxs, account)
			localTxs[account] = txs
		}
	}

	//txComparator := createTxCmp(w.chain, b.header, b.state)
	if len(localTxs) > 0 {
		txs := types.NewTransactionsByPriceAndNonce(b.signer, localTxs, b.header.BaseFee)
		if err := b.commitTransactions(ctx, w, txs, b.txFeeRecipient); err != nil {
			return fmt.Errorf("failed to commit local transactions: %w", err)
		}
	}
	if len(remoteTxs) > 0 {
		txs := types.NewTransactionsByPriceAndNonce(b.signer, remoteTxs, b.header.BaseFee)
		if err := b.commitTransactions(ctx, w, txs, b.txFeeRecipient); err != nil {
			return fmt.Errorf("failed to commit remote transactions: %w", err)
		}
	}
	return nil
}

// commitTransactions attempts to commit every transaction in the transactions list until the block is full or there are no more valid transactions.
func (b *blockState) commitTransactions(ctx context.Context, w *worker, txs *types.TransactionsByPriceAndNonce, txFeeRecipient common.Address) error {
	var coalescedLogs []*types.Log

loop:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// pass
		}
		// If we don't have enough gas for any further transactions then we're done
		if b.gasPool.Gas() < params.TxGas {
			log.Trace("Not enough gas for further transactions", "have", b.gasPool, "want", params.TxGas)
			break
		}
		// Retrieve the next transaction and abort if all done
		tx := txs.Peek()
		if tx == nil {
			break
		}
		// Short-circuit if the transaction requires more gas than we have in the pool.
		// If we didn't short-circuit here, we would get core.ErrGasLimitReached below.
		// Short-circuiting here saves us the trouble of checking the GPM and so on when the tx can't be included
		// anyway due to the block not having enough gas left.
		if b.gasPool.Gas() < tx.Gas() {
			log.Trace("Skipping transaction which requires more gas than is left in the block", "hash", tx.Hash(), "gas", b.gasPool.Gas(), "txgas", tx.Gas())
			txs.Pop()
			continue
		}
		// Error may be ignored here. The error has already been checked
		// during transaction acceptance is the transaction pool.
		//
		// We use the eip155 signer regardless of the current hf.
		from, _ := types.Sender(b.signer, tx)
		// Check whether the tx is replay protected. If we're not in the EIP155 hf
		// phase, start ignoring the sender until we do.
		if tx.Protected() && !w.chainConfig.IsEIP155(b.header.Number) {
			log.Trace("Ignoring reply protected transaction", "hash", tx.Hash(), "eip155", w.chainConfig.EIP155Block)

			txs.Pop()
			continue
		}
		// Start executing the transaction
		b.state.Prepare(tx.Hash(), b.tcount)

		logs, err := b.commitTransaction(w, tx, txFeeRecipient)
		switch err {
		case core.ErrGasLimitReached:
			// Pop the current out-of-gas transaction without shifting in the next from the account
			log.Trace("Gas limit exceeded for current block", "sender", from)
			txs.Pop()

		case core.ErrNonceTooLow:
			// New head notification data race between the transaction pool and miner, shift
			log.Trace("Skipping transaction with low nonce", "sender", from, "nonce", tx.Nonce())
			txs.Shift()

		case core.ErrNonceTooHigh:
			// Reorg notification data race between the transaction pool and miner, skip account =
			log.Trace("Skipping account with hight nonce", "sender", from, "nonce", tx.Nonce())
			txs.Pop()

		case errors.New("gasprice is less than gas price minimum"):
			// We are below the GPM, so we can stop (the rest of the transactions will either have
			// even lower gas price or won't be mineable yet due to their nonce)
			log.Trace("Skipping remaining transaction below the gas price minimum")
			break loop

		case nil:
			// Everything ok, collect the logs and shift in the next transaction from the same account
			coalescedLogs = append(coalescedLogs, logs...)
			b.tcount++
			txs.Shift()

		default:
			// Strange error, discard the transaction and get the next in line (note, the
			// nonce-too-high clause will prevent us from executing in vain).
			log.Debug("Transaction failed, account skipped", "hash", tx.Hash(), "err", err)
			txs.Shift()
		}
	}

	if !w.isRunning() && len(coalescedLogs) > 0 {
		// We don't push the pendingLogsEvent while we are mining. The reason is that
		// when we are mining, the worker will regenerate a mining block every 3 seconds.
		// In order to avoid pushing the repeated pendingLog, we disable the pending log pushing.

		// make a copy, the state caches the logs and these logs get "upgraded" from pending to mined
		// logs by filling in the block hash when the block was mined by the local miner. This can
		// cause a race condition if a log was "upgraded" before the PendingLogsEvent is processed.
		cpy := make([]*types.Log, len(coalescedLogs))
		for i, l := range coalescedLogs {
			cpy[i] = new(types.Log)
			*cpy[i] = *l
		}
		w.pendingLogsFeed.Send(cpy)
	}
	return nil
}

// commitTransaction attempts to appply a single transaction. If the transaction fails, it's modifications are reverted.
func (b *blockState) commitTransaction(w *worker, tx *types.Transaction, txFeeRecipient common.Address) ([]*types.Log, error) {
	snap := b.state.Snapshot()
	//vmRunner := w.chain.NewEVMRunner(b.header, b.state)

	receipt, err := chain.ApplyTransaction(
		w.chainConfig,
		w.chain,
		&txFeeRecipient,
		b.gasPool,
		b.state,
		b.header,
		tx,
		&b.header.GasUsed,
		*w.chain.GetVMConfig(),
		//vmRunner
	)
	if err != nil {
		b.state.RevertToSnapshot(snap)
		return nil, err
	}
	b.txs = append(b.txs, tx)
	b.receipts = append(b.receipts, receipt)

	return receipt.Logs, nil
}

// finalizeAndAssemble runs post-transaction state modification and assembles the final block.
func (b *blockState) finalizeAndAssemble(w *worker) (*types.Block, error) {
	// Need to copy the state here otherwise block production stalls. Not sure why.
	b.state = b.state.Copy()

	//block, err := w.engine.FinalizeAndAssemble(w.chain, b.header, b.state, b.txs, b.receipts, b.randomness)
	block, err := w.engine.FinalizeAndAssemble(w.chain, b.header, b.state, b.txs, b.receipts, b.randomness)
	if err != nil {
		return nil, fmt.Errorf("error in FinalizeAndAssemble: %w", err)
	}

	// Set the validator set diff in the new header if we're using Istanbul and it's the last block of the epoch
	if istanbul, ok := w.engine.(consensus.Istanbul); ok {
		if err := istanbul.UpdateValSetDiff(w.chain, block.MutableHeader(), b.state); err != nil {
			return nil, fmt.Errorf("unable to update Validator Set Diff: %w", err)
		}
	}

	b.receipts = ethChain.AddBlockReceipt(b.receipts, b.state, block.Hash())

	return block, nil
}

// createTxCmp creates a Transaction comparator
//func createTxCmp(chain *chain.BlockChain, header *types.Header, state *state.StateDB) func(tx1 *types.Transaction, tx2 *types.Transaction) int {
//	vmRunner := chain.NewEVMRunner(header, state)
//	currencyManager := currency.NewManager(vmRunner)
//
//	return func(tx1 *types.Transaction, tx2 *types.Transaction) int {
//		return currencyManager.CmpValues(tx1.GasPrice(), tx1.FeeCurrency(), tx2.GasPrice(), tx2.FeeCurrency())
//	}
//}

// totalFees computes total consumed fees in ETH. Block transactions and receipts have to have the same order.
func totalFees(block *types.Block, receipts []*types.Receipt) *big.Float {
	feesWei := new(big.Int)
	for i, tx := range block.Transactions() {
		feesWei.Add(feesWei, new(big.Int).Mul(new(big.Int).SetUint64(receipts[i].GasUsed), tx.GasPrice()))
	}
	return new(big.Float).Quo(new(big.Float).SetInt(feesWei), new(big.Float).SetInt(big.NewInt(params.Ether)))
}
