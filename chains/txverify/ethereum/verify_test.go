package ethereum

import (
	"context"
	"encoding/json"
	"flag"

	"log"
	"math/big"
	"testing"

	//sm "github.com/cch123/supermonkey"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"gopkg.in/urfave/cli.v1"

	"github.com/mapprotocol/atlas/chains/chainsdb"
)

var ReceiptsJSON = `[
  {
    "blockHash": "0xe02bf0c849a67807d9465398c829938c560af09617eecaff598ba820ae0c1729",
    "blockNumber": "0x111",
    "contractAddress": null,
    "cumulativeGasUsed": "0xbfdf",
    "from": "0x1aec262a9429eb9167ac4033aaf8b4239c2743fe",
    "gasUsed": "0xbfdf",
    "logs": [
      {
        "address": "0xd6199276959b95a68c1ee30e8569f5fe060903a6",
        "topics": [
          "0x155e433be3576195943c515e1096620bc754e11b3a4b60fda7c4628caf373635",
          "0x000000000000000000000000000068656164657273746f726541646472657373",
          "0x0000000000000000000000001aec262a9429eb9167ac4033aaf8b4239c2743fe",
          "0x000000000000000000000000970e05ffbb2c4a3b80082e82b24f48a29a9c7651"
        ],
        "data": "0x0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000024c000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000d3",
        "blockNumber": "0x111",
        "transactionHash": "0x58e102c383f926992093192bdfb6c6d1197013fd0470475dca6b4c3749484755",
        "transactionIndex": "0x0",
        "blockHash": "0xe02bf0c849a67807d9465398c829938c560af09617eecaff598ba820ae0c1729",
        "logIndex": "0x0",
        "removed": false
      }
    ],
    "logsBloom": "0x00000000000000000000000000000000000000000040000800000000000000000000000000000000000000000000000400000000008000000000000000000000000000000000000000000000000000000000000000000000000000000200200000000000000000021000000000000000000000000080000000000000000004000000000000040000000000000000000000002000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008000000000001000000000000000",
    "status": "0x1",
    "to": "0xd6199276959b95a68c1ee30e8569f5fe060903a6",
    "transactionHash": "0x58e102c383f926992093192bdfb6c6d1197013fd0470475dca6b4c3749484755",
    "transactionIndex": "0x0",
    "type": "0x0"
  }
]`

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
	rs := make([]*types.Receipt, 0, len(txsHash))
	for _, h := range txsHash {
		r, err := conn.TransactionReceipt(context.Background(), h)
		if err != nil {
			panic(err)
		}
		if r == nil {
			panic("failed to connect to the eth node, please check the network")
		}
		rs = append(rs, r)
	}
	return rs
}

func GetReceiptsFromJSON(receiptsJSON string) []*types.Receipt {
	var rs []*types.Receipt
	if err := json.Unmarshal([]byte(receiptsJSON), &rs); err != nil {
		panic(err)
	}
	return rs
}

func getTxProve(blockNumber uint64, txIndex uint, receiptsJSON string, txParams *TxParams) []byte {

	// get receipts from eth node
	//conn := dialConn()
	//txsHash := getTransactionsHashByBlockNumber(conn, blockNumber)
	//receipts := getReceiptsByTxsHash(conn, txsHash)

	// get receipts from json
	receipts := GetReceiptsFromJSON(receiptsJSON)

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
	key, err := rlp.EncodeToBytes(txIndex)
	if err != nil {
		panic(err)
	}
	if err = tr.Prove(key, 0, proof); err != nil {
		panic(err)
	}

	txProve := TxProve{
		Tx: &TxParams{
			From:  txParams.From,
			To:    txParams.To,
			Value: txParams.Value,
		},
		Receipt:     receipts[txIndex],
		Prove:       proof.NodeList(),
		BlockNumber: blockNumber,
		TxIndex:     txIndex,
	}

	input, err := rlp.EncodeToBytes(txProve)
	if err != nil {
		panic(err)
	}
	return input
}

func TestVerify_Verify(t *testing.T) {
	type args struct {
		router       common.Address
		srcChain     *big.Int
		dstChain     *big.Int
		blockNumber  uint64
		txIndex      uint
		receiptsJSON string
		txParams     *TxParams
	}
	tests := []struct {
		name            string
		args            args
		wantErr         bool
		wantReceiptHash common.Hash
	}{
		{
			name: "",
			args: args{
				router:       common.HexToAddress("0xd6199276959b95a68c1ee30e8569f5fe060903a6"),
				srcChain:     big.NewInt(10),
				dstChain:     big.NewInt(211),
				blockNumber:  273,
				txIndex:      0,
				receiptsJSON: ReceiptsJSON,
				txParams: &TxParams{
					From:  common.HexToAddress("0x1aec262a9429eb9167ac4033aaf8b4239c2743fe").Bytes(),
					To:    common.HexToAddress("0x970e05ffbb2c4a3b80082e82b24f48a29a9c7651").Bytes(),
					Value: big.NewInt(588),
				},
			},
			wantErr:         false,
			wantReceiptHash: common.HexToHash("0x27022c6416c6a79e82c97f1d25f90b8543ea15fc5adfe11ec941d5ab0dec6d28"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//patch := sm.PatchByFullSymbolName("github.com/mapprotocol/atlas/chains/txverify/ethereum.(*Verify).getReceiptsRoot", func(chain rawdb.ChainType, blockNumber uint64) (common.Hash, error) {
			//	return tt.wantReceiptHash, nil
			//})
			//defer patch.Unpatch()

			set := flag.NewFlagSet("test", 0)
			chainsdb.NewStoreDb(cli.NewContext(nil, set, nil), 10, 2)
			txProve := getTxProve(tt.args.blockNumber, tt.args.txIndex, tt.args.receiptsJSON, tt.args.txParams)
			if err := new(Verify).Verify(tt.args.router, tt.args.srcChain, tt.args.dstChain, txProve); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
