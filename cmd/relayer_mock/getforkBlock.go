package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/mapprotocol/atlas/chains/headers/ethereum"
	"github.com/mapprotocol/atlas/consensus"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/vm"
	params2 "github.com/mapprotocol/atlas/params"
	"math/big"
)

func getForkBlock() ([]ethereum.Header, []ethereum.Header) {
	var (
		db001 = rawdb.NewMemoryDatabase()
	)
	genesis := &ethereum.Header{} //cannot unmarshal non-string into Go struct field Header.difficulty of type *hexutil.Big
	err4 := json.Unmarshal([]byte(ethereum.GenesisJSON), genesis)
	if err4 != nil {
		fmt.Println(err4)
	}
	config := params.AllEthashProtocolChanges
	genesis1 := &chain.Genesis{
		Timestamp:  genesis.Time,
		ExtraData:  genesis.Extra,
		GasLimit:   genesis.GasUsed,
		Difficulty: genesis.Difficulty,
		Coinbase:   genesis.Coinbase,
		Number:     genesis.Number.Uint64(),
		Nonce:      genesis.Nonce.Uint64(),
		GasUsed:    genesis.GasUsed,
		ParentHash: genesis.ParentHash,
		Config:     config,
		Alloc:      chain.GenesisAlloc{},
	}
	block, _ := Commit(genesis1, db001)
	//fmt.Println("123::::", block.Header().Hash())
	// chain A: G->A1->A2...A128
	chainA := makeHeaderChain(block.Header(), 0, 100, ethash.NewFaker(), db001, 10)
	// chain A: G->A1->B2...B128
	chainB := makeHeaderChain(chainA[2], 0, 100, ethash.NewFaker(), db001, 10)
	return getChains(chainA), getChains(chainB)
}

func getChains(HeaderData []*types.Header) []ethereum.Header {
	endNum := len(HeaderData)
	Headers := make([]ethereum.Header, len(HeaderData))
	HeaderBytes := make([]bytes.Buffer, len(HeaderData))
	for i := 0; i < endNum; i++ {
		Header := HeaderData[i]
		convertChain(&Headers[i], &HeaderBytes[i], Header)
	}
	return Headers
}
func Commit(g *chain.Genesis, db ethdb.Database) (*types.Block, error) {
	block := ToBlock(g, db)
	if block.Number().Sign() != 0 {
		return nil, fmt.Errorf("can't commit genesis block with number > 0")
	}
	config := g.Config
	if config == nil {
		config = params.AllEthashProtocolChanges
	}
	if err := config.CheckConfigForkOrder(); err != nil {
		return nil, err
	}
	rawdb.WriteTd(db, block.Hash(), block.NumberU64(), g.Difficulty)
	rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), nil)
	rawdb.WriteCanonicalHash(db, block.Hash(), block.NumberU64())
	rawdb.WriteHeadBlockHash(db, block.Hash())
	rawdb.WriteHeadFastBlockHash(db, block.Hash())
	rawdb.WriteHeadHeaderHash(db, block.Hash())
	rawdb.WriteChainConfig(db, block.Hash(), config)
	return block, nil
}
func ToBlock(g *chain.Genesis, db ethdb.Database) *types.Block {
	if db == nil {
		db = rawdb.NewMemoryDatabase()
	}
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db), nil)
	for addr, account := range g.Alloc {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}

	//////////////////////////////////pro compiled////////////////////////////////////
	consensus.InitHeaderStore(statedb, new(big.Int).SetUint64(g.Number))
	register := vm.NewRegisterImpl()

	//	fmt.Println("DoElection")
	_, err := register.DoElections(statedb, 1, 0)
	if err != nil {
		log.Error("ToBlock DoElections", "error", err)
	}
	err = register.Save(statedb, params2.RelayerAddress)
	if err != nil {
		log.Error("ToBlock IMPL Save", "error", err)
	}
	////////////////////////////////////////////////////////////////////////////
	root := statedb.IntermediateRoot(false)
	genesis := &ethereum.Header{} //cannot unmarshal non-string into Go struct field Header.difficulty of type *hexutil.Big
	err4 := json.Unmarshal([]byte(ethereum.GenesisJSON), genesis)
	if err4 != nil {
		fmt.Println(err4)
	}
	head := &types.Header{
		ParentHash:  genesis.ParentHash,
		UncleHash:   genesis.UncleHash,
		Coinbase:    genesis.Coinbase,
		Root:        root,
		TxHash:      genesis.TxHash,
		ReceiptHash: genesis.ReceiptHash,
		Bloom:       types.Bloom(genesis.Bloom),
		Difficulty:  genesis.Difficulty,
		Number:      genesis.Number,
		GasLimit:    genesis.GasLimit,
		GasUsed:     genesis.GasUsed,
		Time:        genesis.Time,
		Extra:       genesis.Extra,
		MixDigest:   genesis.MixDigest,
		Nonce:       types.BlockNonce(genesis.Nonce),
	}
	fmt.Println("toGenesis block hash: ", head.Hash())
	if g.GasLimit == 0 {
		head.GasLimit = params.GenesisGasLimit
	}
	if g.Difficulty == nil {
		head.Difficulty = params.GenesisDifficulty
	}
	statedb.Commit(false)
	statedb.Database().TrieDB().Commit(root, true, nil)
	return types.NewBlock(head, nil, nil, nil, trie.NewStackTrie(nil))
}
