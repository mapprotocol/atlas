package ethereum

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
)

//var (
//	testdb      = rawdb.NewMemoryDatabase()
//	testKey, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
//	testAddress = crypto.PubkeyToAddress(testKey.PublicKey)
//	genesis     = core.GenesisBlockForTesting(testdb, testAddress, big.NewInt(1000000000000000))
//)
//
//type ExtBlock struct {
//	Header       *types.Header
//	Transactions []*types.Transaction
//	Uncles       []*types.Header
//}
//
//// makeChain creates a chain of n blocks starting at and including parent.
//// the returned hash chain is ordered head->parent. In addition, every 3rd block
//// contains a transaction and every 5th an uncle to allow testing correct block
//// reassembly.
//func makeChain(n int, seed byte, parent *types.Block) ([]common.Hash, map[common.Hash]*ExtBlock) {
//	blocks, _ := core.GenerateChain(params.TestChainConfig, parent, ethash.NewFaker(), testdb, n, func(i int, block *core.BlockGen) {
//		block.SetCoinbase(common.Address{seed})
//
//		// If the block number is multiple of 3, send a bonus transaction to the miner
//		if parent == genesis && i%3 == 0 {
//			signer := types.MakeSigner(params.TestChainConfig, block.Number())
//			tx, err := types.SignTx(types.NewTransaction(block.TxNonce(testAddress), common.Address{seed}, big.NewInt(1000), params.TxGas, block.BaseFee(), nil), signer, testKey)
//			if err != nil {
//				panic(err)
//			}
//			block.AddTx(tx)
//		}
//		// If the block number is a multiple of 5, add a bonus uncle to the block
//		if i%5 == 0 {
//			block.AddUncle(&types.Header{ParentHash: block.PrevBlock(i - 1).Hash(), Number: big.NewInt(int64(i - 1))})
//		}
//	})
//	hashes := make([]common.Hash, n+1)
//	hashes[len(hashes)-1] = parent.Hash()
//	blockm := make(map[common.Hash]*ExtBlock, n+1)
//	blockm[parent.Hash()] = &ExtBlock{
//		Header:       parent.Header(),
//		Transactions: parent.Transactions(),
//		Uncles:       parent.Uncles(),
//	}
//	for i, b := range blocks {
//		hashes[len(hashes)-i-2] = b.Hash()
//		blockm[b.Hash()] = &ExtBlock{
//			Header:       b.Header(),
//			Transactions: b.Transactions(),
//			Uncles:       b.Uncles(),
//		}
//	}
//	return hashes, blockm
//}

// makeHeaderChain creates a deterministic chain of headers rooted at parent.
func makeHeaderChain(parent *types.Header, n int, engine consensus.Engine, db ethdb.Database, seed int) []*types.Header {
	blocks := makeBlockChain(types.NewBlockWithHeader(parent), n, engine, db, seed)
	headers := make([]*types.Header, len(blocks))
	for i, block := range blocks {
		headers[i] = block.Header()
	}
	return headers
}

// makeBlockChain creates a deterministic chain of blocks rooted at parent.
func makeBlockChain(parent *types.Block, n int, engine consensus.Engine, db ethdb.Database, seed int) []*types.Block {
	blocks, _ := core.GenerateChain(params.TestChainConfig, parent, engine, db, n, func(i int, b *core.BlockGen) {
		b.SetCoinbase(common.Address{0: byte(seed), 19: byte(i)})
	})
	return blocks
}
