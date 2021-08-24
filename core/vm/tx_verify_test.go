package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/txverify"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/params"
	"log"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/mapprotocol/atlas/chains/txverify/ethereum"
)

var ReceiptsJSON = `[
  {
    "blockHash": "0x2c8e76744c2febc0d8a281b8506054b7b22ef8117275c07757130d3a5d7a2277",
    "blockNumber": "0x389787",
    "contractAddress": null,
    "cumulativeGasUsed": "0x8323",
    "effectiveGasPrice": "0x174876e800",
    "from": "0x68479fe806493cbbf44333eaee9a91bb82b54daa",
    "gasUsed": "0x8323",
    "logs": [],
    "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "root": "0x3644d2dd9f66425f1d184a8a4e9a4e8de78edc45a741a8441be3a556c88db315",
    "to": "0xabbb6bebfa05aa13e908eaa492bd7a8343760477",
    "transactionHash": "0x89a278dd99b3300cf9c1260d2e04d0544624702942daf98b8d0eb8fd1ca734b1",
    "transactionIndex": "0x0",
    "type": "0x0"
  },
  {
    "blockHash": "0x2c8e76744c2febc0d8a281b8506054b7b22ef8117275c07757130d3a5d7a2277",
    "blockNumber": "0x389787",
    "contractAddress": null,
    "cumulativeGasUsed": "0xd52b",
    "effectiveGasPrice": "0x9502f9000",
    "from": "0x7ed1e469fcb3ee19c0366d829e291451be638e59",
    "gasUsed": "0x5208",
    "logs": [],
    "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "root": "0x7080a444f04a57ce59a3c353402d1be445637f8e24add48f53685d8937f9748c",
    "to": "0x3dfe24eb84f968ac83db65516b64e9687c2e5536",
    "transactionHash": "0xd1817ccba00fe19d9cbd28ea1df5660c3ee1f9da8e6ebfff48495cd66b831452",
    "transactionIndex": "0x1",
    "type": "0x0"
  },
  {
    "blockHash": "0x2c8e76744c2febc0d8a281b8506054b7b22ef8117275c07757130d3a5d7a2277",
    "blockNumber": "0x389787",
    "contractAddress": "0x541ece32a7d8500d3cefc1a4fb5b684fef148cd5",
    "cumulativeGasUsed": "0x55e43",
    "effectiveGasPrice": "0x6fc23ac00",
    "from": "0x0536806df512d6cdde913cf95c9886f65b1d3462",
    "gasUsed": "0x48918",
    "logs": [],
    "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "root": "0x870a5f48945c3d9d777ebacd3a5ce287fabf7277e91639bf6d72702a8133de33",
    "to": null,
    "transactionHash": "0x2f150bcb1b529959011c83125468819e569853867bdffaaa32deaf0a80db76db",
    "transactionIndex": "0x2",
    "type": "0x0"
  }
]`

var (
	fromAddr  = common.HexToAddress("0xf945e6ffd840ed5787d367708307bd1fa3d40f4")
	toAddr    = common.HexToAddress("0x32cd75ca677e9c37fd989272afa8504cb8f6eb52")
	SendValue = big.NewInt(99)
)

type TxParams struct {
	From  []byte
	To    []byte
	Value *big.Int
}

type TxProve struct {
	Tx               *TxParams
	Receipt          *types.Receipt
	Prove            light.NodeList
	TransactionIndex uint
}

func dialConn() *ethclient.Client {
	conn, err := ethclient.Dial("http://192.168.10.215:8545")
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

	return GetReceiptsFromJSON()
}

func GetReceiptsFromJSON() []*types.Receipt {
	var rs []*types.Receipt
	if err := json.Unmarshal([]byte(ReceiptsJSON), &rs); err != nil {
		panic(err)
	}
	return rs
}

func getTxProve() []byte {
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

	txProve := TxProve{
		Tx: &TxParams{
			From:  fromAddr.Bytes(),
			To:    toAddr.Bytes(),
			Value: SendValue,
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
	var (
		srcChain = big.NewInt(1)
		dstChain = big.NewInt(211)
	)

	if err := new(ethereum.Verify).Verify(srcChain, dstChain, getTxProve()); err != nil {
		t.Fatal(err)
	}

	group, err := chains.ChainType2ChainGroup(rawdb.ChainType(srcChain.Uint64()))
	if err != nil {
		t.Fatal(err)
	}

	v, err := txverify.Factory(group)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Verify(srcChain, dstChain, getTxProve()); err != nil {
		t.Fatal(err)
	}
}

func TestAddr(t *testing.T) {
	fmt.Println("============================== addr: ", params.TxVerifyAddress)
}
