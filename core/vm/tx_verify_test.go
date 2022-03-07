package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/state"
	"log"
	"math/big"
	"testing"

	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/interfaces"
	atlastypes "github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/params"
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

var (
	blockNumber      = big.NewInt(11768672)
	txIndex     uint = 35
	fromAddr         = common.HexToAddress("0x000000000000000000000000a9024d80366a7cc34698c05d6a19fcf7e3f1ad34")
	toAddr           = common.HexToAddress("0x0000000000000000000000002e9b4be739453cddbb3641fb61052ba46873d41f")
	SendValue        = big.NewInt(10)
	srcChain         = big.NewInt(3)
	dstChain         = big.NewInt(212)
	routerAddr       = common.HexToAddress("0x23dd5a89c3ea51601b0674a4fa6ec6b3b14d0b7a")
	coinAddr         = common.HexToAddress("0x23dd5a89c3ea51601b0674a4fa6ec6b3b14d0b7a")
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

func dialConn() *ethclient.Client {
	//conn, err := ethclient.Dial("https://ropsten.infura.io/v3/8cce6b470ad44fb5a3621aa34243647f")
	conn, err := ethclient.Dial("https://ropsten.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161")
	if err != nil {
		log.Fatalf("Failed to connect to the eth: %v", err)
	}
	return conn
}

func dialAtlasConn() *ethclient.Client {
	//conn, err := ethclient.Dial("http://159.138.90.210:7445")
	conn, err := ethclient.Dial("http://127.0.0.1:7445")
	if err != nil {
		log.Fatalf("Failed to connect to the eth: %v", err)
	}
	return conn
}

func getStateDB() *state.StateDB {
	finalDb := rawdb.NewMemoryDatabase()
	finalState, _ := state.New(common.Hash{}, state.NewDatabase(finalDb), nil)
	return finalState
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

	//return GetReceiptsFromJSON()
}

func GetReceiptsFromJSON() []*types.Receipt {
	var rs []*types.Receipt
	if err := json.Unmarshal([]byte(ReceiptsJSON), &rs); err != nil {
		panic(err)
	}
	return rs
}

func getTxProve() []byte {
	// get receipts from ethereum node
	conn := dialConn()
	txsHash := getTransactionsHashByBlockNumber(conn, blockNumber)
	receipts := getReceiptsByTxsHash(conn, txsHash)
	// get receipts from json
	//receipts := GetReceiptsFromJSON()

	tr, err := trie.New(common.Hash{}, trie.NewDatabase(memorydb.New()))
	if err != nil {
		panic(err)
	}

	tr = atlastypes.DeriveTire(receipts, tr)
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
			From:  fromAddr.Bytes(),
			To:    toAddr.Bytes(),
			Value: SendValue,
		},
		Receipt:     receipts[txIndex],
		Prove:       proof.NodeList(),
		BlockNumber: blockNumber.Uint64(),
		TxIndex:     txIndex,
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
		router   = common.HexToAddress("0xd6199276959b95a68c1ee30e8569f5fe060903a6")
	)

	group, err := chains.ChainType2ChainGroup(chains.ChainType(srcChain.Uint64()))
	if err != nil {
		t.Fatal(err)
	}

	//set := flag.NewFlagSet("test", 0)
	//chainsdb.NewStoreDb(cli.NewContext(nil, set, nil), 10, 2)

	v, err := interfaces.VerifyFactory(group)
	if err != nil {
		t.Fatal(err)
	}
	//db := rawdb.NewMemoryDatabase()
	//sdb, _ := state.New(common.Hash{}, state.NewDatabase(db), nil)
	if err := v.Verify(getStateDB(), router, srcChain, dstChain, getTxProve()); err != nil {
		t.Fatal(err)
	}
}

func PackInput(abi abi.ABI, abiMethod string, params ...interface{}) []byte {
	input, err := abi.Pack(abiMethod, params...)
	if err != nil {
		panic(err)
	}
	return input
}

func TestTxVerify(t *testing.T) {
	input := PackInput(abiTxVerify, "txVerify", routerAddr, coinAddr, srcChain, dstChain, getTxProve())
	ret := call(dialAtlasConn(), from, params.TxVerifyAddress, input)
	if !ret[0].(bool) {
		t.Errorf("message: %s", ret[1].(string))
	}
}

func call(client *ethclient.Client, from, toAddress common.Address, input []byte) []interface{} {
	output, err := client.CallContract(context.Background(), ethchain.CallMsg{From: from, To: &toAddress, Data: input}, nil)
	if err != nil {
		panic(err)
	}
	method, _ := abiTxVerify.Methods["txVerify"]
	ret, err := method.Outputs.Unpack(output)
	if err != nil {
		panic(err)
	}
	return ret
}

func TestAddr(t *testing.T) {
	fmt.Println("============================== addr: ", params.TxVerifyAddress)

}

func TestEventHash(t *testing.T) {
	// LogSwapOut(bytes32 hash, address indexed token, address indexed from, address indexed to, uint amount, uint fromChainID, uint toChainID)
	event := "LogSwapOut(bytes32,address,address,address,uint256,uint256,uint256)"
	eventHash := crypto.Keccak256Hash([]byte(event))
	t.Log("event hash: ", eventHash) // 0xcfdd266a10c21b3f2a2da4a807706d3f3825d37ca51d341eef4dce804212a8a3

}
