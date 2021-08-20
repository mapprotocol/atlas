package vm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/mapprotocol/atlas/chains/chainsdb"
	"github.com/mapprotocol/atlas/core/rawdb"
	"io/ioutil"
	"log"
	"math/big"
	"strings"
	"testing"
)

type TxParams struct {
	SrcChain *big.Int
	DstChain *big.Int
	From     common.Address
	To       common.Address
	Value    *big.Int
}

type TXProve struct {
	Tx               *TxParams
	Receipt          *types.Receipt
	Prove            light.NodeList
	TransactionIndex uint
}

var (
	srcChain  = big.NewInt(1001)
	dstChain  = big.NewInt(1000)
	fromAddr  = common.HexToAddress("0xf945e6ffd840ed5787d367708307bd1fa3d40f4")
	toAddr    = common.HexToAddress("0x32cd75ca677e9c37fd989272afa8504cb8f6eb52")
	SendValue = big.NewInt(99)
)

func dialConn() *ethclient.Client {
	url := "http://127.0.0.1:8545"
	conn, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the eth: %v", err)
	}
	return conn
}

func getTransactionsHashByBlockNumber(conn *ethclient.Client, number *big.Int) []common.Hash {
	block, err := conn.BlockByNumber(context.Background(), number)
	if err != nil {
		panic(err)
	}
	if block == nil {
		panic("failed to connect to the eth node, please check the network")
	}

	txs := make([]common.Hash, 0, len(block.Transactions()))
	for _, tx := range block.Transactions() {
		txs = append(txs, tx.Hash())
	}
	return txs
}

func getReceiptsByTxsHash(conn *ethclient.Client, txsHash []common.Hash) []*types.Receipt {
	//rs := make([]*types.Receipt, 0, len(txsHash))
	//for _, h := range txsHash {
	//	r, err := conn.TransactionReceipt(context.Background(), h)
	//	if err != nil {
	//		panic(err)
	//	}
	//	if r == nil {
	//		panic("failed to connect to the eth node, please check the network")
	//	}
	//	rs = append(rs, r)
	//}
	//return rs

	return ReceiptsJSON()
}

func ReceiptsJSON() []*types.Receipt {
	byteValue, err := ioutil.ReadFile("receipts.json")
	if err != nil {
		panic(err)
	}
	var rs []*types.Receipt
	if err := json.Unmarshal(byteValue, &rs); err != nil {
		panic(err)
	}
	return rs
}

func VerifyTxParams(tx *TxParams, topics []common.Hash) error {
	if len(topics) != 5 {
		return errors.New("verify tx failed, the topics length must be 5")
	}
	if !bytes.Equal(common.BigToHash(tx.SrcChain).Bytes(), topics[0].Bytes()) {
		return errors.New("verify tx failed, SrcChain")
	}
	if !bytes.Equal(common.BigToHash(tx.DstChain).Bytes(), topics[1].Bytes()) {
		return errors.New("verify tx failed, DstChain")
	}
	if !bytes.Equal(common.BigToHash(tx.Value).Bytes(), topics[4].Bytes()) {
		return errors.New("verify tx failed， Value")
	}

	from := strings.ToLower(tx.From.String())
	topics2 := strings.ToLower(topics[2].String())
	if !bytes.Equal(common.Hex2Bytes(from), common.Hex2Bytes(topics2)) {
		return errors.New("verify tx failed, From")
	}

	to := strings.ToLower(tx.To.String())
	topics3 := strings.ToLower(topics[2].String())
	if !bytes.Equal(common.Hex2Bytes(to), common.Hex2Bytes(topics3)) {
		return errors.New("verify tx failed, To")
	}

	return nil
}

func getReceiptsRoot(chain rawdb.ChainType, blockNumber uint64) (common.Hash, error) {
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

func _txVerify(input []byte) {
	// RLP decode
	var txProve TXProve
	if err := rlp.DecodeBytes(input, &txProve); err != nil {
		panic(err)
	}

	// todo verify tx params
	//if err := VerifyTxParams(txProve.Tx, txProve.Receipt.Logs[0].Topics); err != nil {
	//	panic(err)
	//}

	//receiptsRoot := getReceiptsRoot(chains.ChainTypeETHTest, txProve.Receipt.Logs[0].BlockNumber)
	receiptsRoot := common.HexToHash("0xb350f39d35702cfbc6709470a50255fc2a11248fa91528e5e28fe0fd05c04f4d")

	key, err := rlp.EncodeToBytes(txProve.TransactionIndex)
	if err != nil {
		panic(err)
	}
	getReceipt, err := trie.VerifyProof(receiptsRoot, key, txProve.Prove.NodeSet())
	if err != nil {
		panic(err)
	}
	giveReceipt, err := rlp.EncodeToBytes(txProve.Receipt)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(giveReceipt, getReceipt) {
		panic("receipt mismatch")
	}
}

// 模拟 relayer 操作
func _relayer() []byte {
	var (
		blockNumber           = big.NewInt(3708807)
		transactionIndex uint = 1
	)

	// 调用以太坊接口获取 receipts
	conn := dialConn()
	txsHash := getTransactionsHashByBlockNumber(conn, blockNumber)
	receipts := getReceiptsByTxsHash(conn, txsHash)

	// 根据 receipts 生成 trie
	tr, err := trie.New(common.Hash{}, trie.NewDatabase(memorydb.New()))
	if err != nil {
		panic(err)
	}
	for i, r := range receipts {
		key, err := rlp.EncodeToBytes(uint(i))
		if err != nil {
			panic(err)
		}
		value, err := rlp.EncodeToBytes(r)
		if err != nil {
			panic(err)
		}

		tr.Update(key, value)
	}

	proof := light.NewNodeSet()
	key, err := rlp.EncodeToBytes(transactionIndex)
	if err != nil {
		panic(err)
	}
	if err = tr.Prove(key, 0, proof); err != nil {
		panic(err)
	}

	txProve := TXProve{
		Tx: &TxParams{
			SrcChain: srcChain,
			DstChain: dstChain,
			From:     fromAddr,
			To:       toAddr,
			Value:    SendValue,
		},
		Receipt:          receipts[transactionIndex],
		Prove:            proof.NodeList(),
		TransactionIndex: transactionIndex,
	}

	input, err := rlp.EncodeToBytes(txProve)
	if err != nil {
		panic(err)
	}
	return input
}

func TestReceiptsRootAndProof(t *testing.T) {
	input := _relayer()
	_txVerify(input)
}
