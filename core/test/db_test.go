package test

import (
	//"fmt"
	//"github.com/ethereum/go-ethereum/core"
	//"github.com/ethereum/go-ethereum/core/types"
	//"github.com/ethereum/go-ethereum/crypto"
	//"github.com/ethereum/go-ethereum/params"
	//atlasdbCore "github.com/mapprotocol/atlas/core"
	//"github.com/mapprotocol/atlas/core/rawdb"
	//ethHeader "github.com/mapprotocol/atlas/core/vm/sync"
	//"math"
	//"math/big"
	"testing"
	//"time"
)

func Test_WriteReadHeader(t *testing.T) {
	//var (
	//	key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	//	addr   = crypto.PubkeyToAddress(key.PublicKey)
	//	g      = &core.Genesis{
	//		Config: params.TestChainConfig,
	//		Alloc:  core.GenesisAlloc{addr: {Balance: big.NewInt(math.MaxInt64)}},
	//	}
	//)
	//db0, _ := atlasdbCore.OpenDatabase("C:\\Users\\m1843\\Desktop\\data1", 20, 20)
	//db, _ := atlasdbCore.NewStoreDb(db0, rawdb.ChainType(123))
	//db.SetChainType(rawdb.ChainType(123))
	//block := g.ToBlock(nil)
	//db.WriteHeader(block.Header())
	//fmt.Println(db.ReadHeader(block.Hash(), uint64(0)))
}

func Test_InsertHeaderChain(t *testing.T) {
	//var (
	//	key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	//	addr   = crypto.PubkeyToAddress(key.PublicKey)
	//	g      = &core.Genesis{
	//		Config: params.TestChainConfig,
	//		Alloc:  core.GenesisAlloc{addr: {Balance: big.NewInt(math.MaxInt64)}},
	//	}
	//)
	//m := rawdb.ChainType(0)
	//db0, _ := atlasdbCore.OpenDatabase("C:\\Users\\m1843\\Desktop\\data1", 20, 20)
	//
	//block := g.ToBlock(nil)
	//rawdb.WriteTd(db0, block.Hash(), block.NumberU64(), g.Difficulty, m)
	//rawdb.WriteReceipts(db0, block.Hash(), block.NumberU64(), nil, m)
	//rawdb.WriteCanonicalHash(db0, block.Hash(), block.NumberU64(), m)
	//rawdb.WriteHeadBlockHash(db0, block.Hash(), m)
	//rawdb.WriteHeadFastBlockHash(db0, block.Hash(), m)
	//rawdb.WriteHeadHeaderHash(db0, block.Hash(), m)
	//db, _ := atlasdbCore.NewStoreDb(db0, atlasdbCore.DefultChainType)
	//db.InsertHeaderChain([]*ethHeader.ETHHeader{block.Header()}, time.Now())
	//fmt.Println(db.ReadHeader(block.Hash(), uint64(0)))
}
