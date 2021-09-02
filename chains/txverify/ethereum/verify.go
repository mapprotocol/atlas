package ethereum

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/mapprotocol/atlas/chains/chainsdb"
	"github.com/mapprotocol/atlas/core/rawdb"
)

var (
	// EventHash cross-chain transaction event hash
	EventHash = common.HexToHash("0x155e433be3576195943c515e1096620bc754e11b3a4b60fda7c4628caf373635")
)

type TxParams struct {
	From  []byte
	To    []byte
	Value *big.Int
}

type TxProve struct {
	Tx           *TxParams
	Receipt      *types.Receipt
	Prove        light.NodeList
	BlockNumber  uint64
	TxIndex      uint
	CoinAddr     common.Address // the address of the token contract
	ContractAddr common.Address // address of the contract that generated the cross-chain transaction event
}

type Verify struct {
}

func (v *Verify) Verify(srcChain, dstChain *big.Int, txProveBytes []byte) error {
	txProve, err := v.decode(txProveBytes)
	if err != nil {
		return err
	}

	// debug log
	for i, lg := range txProve.Receipt.Logs {
		ls, _ := json.Marshal(lg)
		log.Printf("receipt log-%d: %s\n", i, ls)
	}

	lg, err := v.queryLog(txProve.ContractAddr, txProve.Receipt.Logs)
	if err != nil {
		return err
	}
	if err := v.verifyTxParams(srcChain, dstChain, txProve.Tx, lg); err != nil {
		return err
	}

	receiptsRoot, err := v.getReceiptsRoot(rawdb.ChainType(srcChain.Uint64()), txProve.BlockNumber)
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

func (v *Verify) queryLog(contractAddr common.Address, logs []*types.Log) (*types.Log, error) {
	for _, lg := range logs {
		if bytes.Equal(lg.Address.Bytes(), contractAddr.Bytes()) {
			if bytes.Equal(lg.Topics[0].Bytes(), EventHash.Bytes()) {
				return lg, nil
			}
		}
	}
	return nil, errors.New("not found match log")
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
	key, err := rlp.EncodeToBytes(txProve.TxIndex)
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
