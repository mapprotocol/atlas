package ethereum

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mapprotocol/atlas/params"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/mapprotocol/atlas/chains/chainsdb"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/types"
)

// EventHash event hash
// todo replace
var EventHash = common.Hash{}

type TxParams struct {
	From  []byte
	To    []byte
	Value *big.Int
}

//type Chain struct {
//	SrcChain *big.Int
//	DstChain *big.Int
//}

type TxProve struct {
	Tx               *TxParams
	Receipt          *types.Receipt
	Prove            light.NodeList
	TransactionIndex uint
}

type Verify struct {
}

func (v *Verify) Verify(srcChain, dstChain *big.Int, txProveBytes []byte) error {
	txProve, err := v.decode(txProveBytes)
	if err != nil {
		return err
	}

	log, number, err := v.getLogAndBlockNumber(txProve.Receipt.Logs)
	if err != nil {
		return err
	}

	if err := v.verifyTxParams(srcChain, dstChain, txProve.Tx, log); err != nil {
		return err
	}
	receiptsRoot, err := v.getReceiptsRoot(rawdb.ChainType(srcChain.Uint64()), number)
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

func (v *Verify) getLogAndBlockNumber(logs []*types.Log) (*types.Log, uint64, error) {
	for _, log := range logs {
		if bytes.Equal(log.Address.Bytes(), params.TxVerifyAddress.Bytes()) {
			if bytes.Equal(log.Topics[0].Bytes(), EventHash.Bytes()) {
				return log, log.BlockNumber, nil
			}
		}
	}
	return nil, 0, errors.New("not found match log")
}

func (v *Verify) verifyTxParams(srcChain, dstChain *big.Int, tx *TxParams, log *types.Log) error {
	if len(log.Topics) < 5 {
		return errors.New("verify tx params failed, the log.topics length must be 5")
	}

	// todo
	//from := strings.ToLower(tx.From.String())
	topics2 := strings.ToLower(log.Topics[1].String())
	if !bytes.Equal(tx.From, common.Hex2Bytes(topics2)) {
		return errors.New("verify tx params failed, From")
	}

	// todo
	//to := strings.ToLower(tx.To.String())
	topics3 := strings.ToLower(log.Topics[2].String())
	if !bytes.Equal(tx.To, common.Hex2Bytes(topics3)) {
		return errors.New("verify tx params failed")
	}
	if !bytes.Equal(common.BigToHash(tx.Value).Bytes(), log.Topics[4].Bytes()) {
		return errors.New("verify tx params failed， Value")
	}

	// todo data split
	//log.Data
	if !bytes.Equal(common.BigToHash(srcChain).Bytes(), log.Topics[0].Bytes()) {
		return errors.New("verify tx params failed, SrcChain")
	}
	if !bytes.Equal(common.BigToHash(dstChain).Bytes(), log.Topics[1].Bytes()) {
		return errors.New("verify tx params failed, DstChain")
	}

	return nil
}

func (v *Verify) getReceiptsRoot(chain rawdb.ChainType, blockNumber uint64) (common.Hash, error) {
	store, err := chainsdb.GetStoreMgr(chain)
	if err != nil {
		return common.Hash{}, err
	}
	header := store.GetHeaderByNumber(blockNumber)
	if header == nil {
		return common.Hash{}, fmt.Errorf("get header by number failed, number: %d", blockNumber)
	}
	return header.ReceiptHash, nil
}

func (v *Verify) verifyProof(receiptsRoot common.Hash, txProve *TxProve) error {
	key, err := rlp.EncodeToBytes(txProve.TransactionIndex)
	if err != nil {
		return err
	}
	getReceipt, err := trie.VerifyProof(receiptsRoot, key, txProve.Prove.NodeSet())
	if err != nil {
		return err
	}
	giveReceipt, err := rlp.EncodeToBytes(txProve.Receipt)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(giveReceipt, getReceipt) {
		return errors.New("receipt mismatch")
	}
	return nil
}
