package ethereum

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mapprotocol/atlas/chains"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

var (
	// EventHash cross-chain transaction event hash
	// LogSwapOut(bytes32,address,address,address,uint256,uint256,uint256)
	EventHash = common.HexToHash("0xcfdd266a10c21b3f2a2da4a807706d3f3825d37ca51d341eef4dce804212a8a3")
)

type TxParams struct {
	From  []byte
	To    []byte
	Value *big.Int
}

type TxProve struct {
	Tx          *TxParams
	Receipt     *types.Receipt
	Prove       light.NodeList
	BlockNumber uint64
	TxIndex     uint
}

type Verify struct {
}

func (v *Verify) Verify(routerContractAddr common.Address, srcChain, dstChain *big.Int, txProveBytes []byte) error {
	txProve, err := v.decode(txProveBytes)
	if err != nil {
		return err
	}

	// debug log
	for i, lg := range txProve.Receipt.Logs {
		ls, _ := json.Marshal(lg)
		log.Printf("receipt log-%d: %s\n", i, ls)
	}

	lg, err := v.queryLog(routerContractAddr, txProve.Receipt.Logs)
	if err != nil {
		return err
	}
	if err := v.verifyTxParams(srcChain, dstChain, txProve.Tx, lg); err != nil {
		return err
	}

	receiptsRoot, err := v.getReceiptsRoot(chains.ChainType(srcChain.Uint64()), txProve.BlockNumber)
	if err != nil {
		return err
	}
	//receiptsRoot := common.HexToHash("0xb350f39d35702cfbc6709470a50255fc2a11248fa91528e5e28fe0fd05c04f4d")
	return v.verifyProof(receiptsRoot, txProve)
}

func (v *Verify) decode(txProveBytes []byte) (*TxProve, error) {
	var txProve TxProve
	if err := rlp.DecodeBytes(txProveBytes, &txProve); err != nil {
		return nil, err
	}
	return &txProve, nil
}

func (v *Verify) queryLog(routerContractAddr common.Address, logs []*types.Log) (*types.Log, error) {
	for _, lg := range logs {
		if bytes.Equal(lg.Address.Bytes(), routerContractAddr.Bytes()) {
			if bytes.Equal(lg.Topics[0].Bytes(), EventHash.Bytes()) {
				return lg, nil
			}
		}
	}
	return nil, fmt.Errorf("not found event log, router contract addr: %v, event hash: %v", routerContractAddr, EventHash)
}

func (v *Verify) verifyTxParams(srcChain, dstChain *big.Int, tx *TxParams, log *types.Log) error {
	if len(log.Topics) < 4 {
		return errors.New("verify tx params failed, the log.Topics`s length cannot be less than 4")
	}

	//from := strings.ToLower(tx.From.String())
	if !bytes.Equal(common.BytesToHash(tx.From).Bytes(), common.HexToHash(log.Topics[2].Hex()).Bytes()) {
		return errors.New("verify tx params failed, invalid from")
	}
	//to := strings.ToLower(tx.To.String())
	if !bytes.Equal(common.BytesToHash(tx.To).Bytes(), log.Topics[3].Bytes()) {
		return errors.New("verify tx params failed, invalid to")
	}

	if len(log.Data) < 128 {
		return errors.New("verify tx params failed, log.Data length cannot be less than 128")
	}

	if !bytes.Equal(common.BigToHash(tx.Value).Bytes(), log.Data[32:64]) {
		return errors.New("verify tx params failed， invalid value")
	}
	if !bytes.Equal(common.BigToHash(srcChain).Bytes(), log.Data[64:96]) {
		return errors.New("verify tx params failed, invalid srcChain")
	}
	if !bytes.Equal(common.BigToHash(dstChain).Bytes(), log.Data[96:128]) {
		return errors.New("verify tx params failed, invalid dstChain")
	}

	return nil
}

func (v *Verify) getReceiptsRoot(chain chains.ChainType, blockNumber uint64) (common.Hash, error) {
	//store, err := chainsdb.GetStoreMgr(chain)
	//if err != nil {
	//	return common.Hash{}, err
	//}
	//header := store.GetHeaderByNumber(blockNumber)
	//if header == nil {
	//	return common.Hash{}, fmt.Errorf("get header by number failed, number: %d", blockNumber)
	//}
	//return header.ReceiptHash, nil
	return common.Hash{}, nil
}

func (v *Verify) verifyProof(receiptsRoot common.Hash, txProve *TxProve) error {
	var buf bytes.Buffer
	rs := types.Receipts{txProve.Receipt}
	rs.EncodeIndex(0, &buf)
	giveReceipt := buf.Bytes()

	var key []byte
	key = rlp.AppendUint64(key[:0], uint64(txProve.TxIndex))

	getReceipt, err := trie.VerifyProof(receiptsRoot, key, txProve.Prove.NodeSet())
	if err != nil {
		return err
	}
	if !bytes.Equal(giveReceipt, getReceipt) {
		return errors.New("receipt mismatch")
	}
	return nil
}
