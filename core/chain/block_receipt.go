package chain

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/types"
)

// AddBlockReceipt checks whether logs were emitted by the core contract calls made as part
// of block processing outside of transactions.  If there are any, it creates a receipt for
// them (the so-called "block receipt") and appends it to receipts
func AddBlockReceipt(receipts types.Receipts, statedb *state.StateDB, blockHash common.Hash) types.Receipts {
	if len(statedb.GetLogs(common.Hash{})) > 0 {
		receipt := types.NewReceipt(nil, false, 0)
		receipt.Logs = statedb.GetLogs(common.Hash{})
		receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
		for i := range receipt.Logs {
			receipt.Logs[i].TxIndex = uint(len(receipts))
			receipt.Logs[i].TxHash = blockHash
			receipt.Logs[i].BlockHash = blockHash
		}
		receipts = append(receipts, receipt)
	}
	return receipts
}
