package test

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/mapprotocol/atlas/atlasdb"
	atlasdbCore "github.com/mapprotocol/atlas/core"
	"github.com/mapprotocol/atlas/core/rawdb"
	"math"
	"math/big"
	"testing"
)

func Test(t *testing.T) {

	var (
		key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		addr   = crypto.PubkeyToAddress(key.PublicKey)
		g      = &core.Genesis{
			Config: params.TestChainConfig,
			Alloc:  core.GenesisAlloc{addr: {Balance: big.NewInt(math.MaxInt64)}},
		}
		m rawdb.Mark
	)
	db := atlasdb.NewMemDatabase()
	for i := 1; i < 11; i++ {
		m = rawdb.Mark(i)
		block := g.ToBlock(nil)
		if block.Number().Sign() != 0 {
			t.Fatal("error1")
		}
		config := g.Config
		if config == nil {
			config = params.AllEthashProtocolChanges
		}
		if err := config.CheckConfigForkOrder(); err != nil {
			t.Fatal("error2")
		}
		//fmt.Println(block.Hash())
		rawdb.WriteTd(db, block.Hash(), block.NumberU64(), g.Difficulty, m)
		rawdb.WriteBlock(db, block, m)
		rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), nil, m)
		rawdb.WriteCanonicalHash(db, block.Hash(), block.NumberU64(), m)
		rawdb.WriteHeadBlockHash(db, block.Hash(), m)
		rawdb.WriteHeadFastBlockHash(db, block.Hash(), m)
		rawdb.WriteHeadHeaderHash(db, block.Hash(), m)
		rawdb.WriteChainConfig(db, block.Hash(), config, m)
	}
	for i := 1; i < 11; i++ {
		m = rawdb.Mark(i)
		//atlasdb.ReadTd(db, hash, m, uint64(0))
		//atlasdb.ReadBlock(db, hash, m, uint64(0))
		//atlasdb.ReadReceipts(db, hash, m, uint64(0), nil)
		fmt.Println(rawdb.ReadCanonicalHash(db, uint64(0), m))
		fmt.Println(rawdb.ReadHeadBlockHash(db, m))
		fmt.Println(rawdb.ReadHeadFastBlockHash(db, m))
		fmt.Println(rawdb.ReadHeadHeaderHash(db, m))
		//atlasdb.ReadChainConfig(db, hash)
	}

}
func Test2(t *testing.T) {
	var (
		key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		addr   = crypto.PubkeyToAddress(key.PublicKey)
		g      = &core.Genesis{
			Config: params.TestChainConfig,
			Alloc:  core.GenesisAlloc{addr: {Balance: big.NewInt(math.MaxInt64)}},
		}
	)
	db, _ := atlasdbCore.NewStoreDb(nil, rawdb.Mark(123), "C:\\Users\\m1843\\Desktop\\data1", 20, 20)
	db.SetMark(rawdb.Mark(123))
	block := g.ToBlock(nil)
	db.WriteHeader(block.Header())
	fmt.Println(db.ReadHeader(block.Hash(), uint64(0)))

}
