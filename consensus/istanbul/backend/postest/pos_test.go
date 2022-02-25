package postest

//import (
//	"fmt"
//	//"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/crypto"
//	"github.com/mapprotocol/atlas/accounts/abi"
//	//"github.com/mapprotocol/atlas/consensus/istanbul"
//	mockEngine "github.com/mapprotocol/atlas/consensus/consensustest"
//	"github.com/mapprotocol/atlas/contracts/abis"
//	"github.com/mapprotocol/atlas/core/chain"
//	"github.com/mapprotocol/atlas/core/rawdb"
//	//"github.com/mapprotocol/atlas/core/state"
//	"github.com/mapprotocol/atlas/core/types"
//	"github.com/mapprotocol/atlas/core/vm"
//	"github.com/mapprotocol/atlas/params"
//	"math/big"
//	"strings"
//	"testing"
//)
//
//var (
//	abiStaking, _ = abi.JSON(strings.NewReader(abis.ElectionsStr))
//	priKey, _     = crypto.HexToECDSA("53ee6ae610b7478404ae2fd07501cfd7688af191e22b553afafa293fbe364980") //私钥
//	signer        = types.NewEIP155Signer(big.NewInt(0))
//)
//
//func TestBackend_AutoActive(t *testing.T) {
//	fmt.Println("=== address ===", crypto.PubkeyToAddress(priKey.PublicKey))
//	var (
//		db      = rawdb.NewMemoryDatabase()
//		genesis = chain.DefaultGenesisBlock().MustCommit(db)
//		//engine  = New(istanbul.DefaultConfig, db)
//
//	)
//	//engine, _ := New(istanbul.DefaultConfig, db).(*Backend)
//	engine := mockEngine.NewFaker()
//
//	blockchain, err := chain.NewBlockChain(db, nil, params.MainnetChainConfig, engine, vm.Config{}, nil, nil)
//	//engine.SetChain(
//	//	blockchain, blockchain.CurrentBlock,
//	//	func(hash common.Hash) (*state.StateDB, error) {
//	//		stateRoot := blockchain.GetHeaderByHash(hash).Root
//	//		return blockchain.StateAt(stateRoot)
//	//	})
//	if err != nil {
//		panic(err)
//	}
//
//	sBlocks := 3
//	//genesis.Coinbase().SetBytes()
//	parent := genesis
//	for i := 0; i < sBlocks; i++ {
//		//_chain, _ := chain.GenerateChain(params.MainnetChainConfig, parent, engine, db, 40, func(i int, gen *chain.BlockGen) {
//		//	header := gen.Header()
//		//	//stateDB := gen.GetStateDB()
//		//	//if gspec.Config.TIP8.FastNumber != nil && gspec.Config.TIP8.FastNumber.Sign() > 0 {
//		//	//	executableTx(header.Number.Uint64()-gspec.Config.TIP8.FastNumber.Uint64()+9600, gen, blockchain, header, stateDB)
//		//	//}
//		//	vote(header.Number.Uint64(), gen, blockchain, header)
//		//})
//		_chain, _ := chain.GenerateChain(params.MainnetChainConfig, parent, engine, db, 40, func(i int, gen *chain.BlockGen) {
//			gen.SetExtra(genesis.Extra())
//		})
//
//		if _, err := blockchain.InsertChain(_chain); err != nil {
//			panic(err)
//		}
//		parent = blockchain.CurrentBlock()
//	}
//
//}
//
//func vote(number uint64, gen *chain.BlockGen, blockchain *chain.BlockChain, header *types.Header) {
//	sendContractTransaction(number, gen, myAddresss, big.NewInt(6000000000000000000), priKey, signer, blockchain, abiStaking)
//}
//
//func Test_ss(t *testing.T) {
//	db := rawdb.NewMemoryDatabase()
//	genesis := chain.DefaultGenesisBlock().MustCommit(db)
//	_chain, _ := chain.GenerateChain(params.MainnetChainConfig, genesis, mockEngine.NewFaker(), db, 1000, func(i int, gen *chain.BlockGen) {
//		gen.SetExtra(genesis.Extra())
//	})
//	blockchain, _ := chain.NewBlockChain(db, nil, params.MainnetChainConfig, mockEngine.NewFaker(), vm.Config{}, nil, nil)
//
//	if i, err := blockchain.InsertChain(_chain); err != nil {
//		fmt.Printf("insert error (block %d): %v\n", _chain[i].NumberU64(), err)
//		return
//	}
//}
