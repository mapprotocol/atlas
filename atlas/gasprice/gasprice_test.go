// Copyright 2020 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package gasprice

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/mapprotocol/atlas/consensus/consensustest"
	"github.com/mapprotocol/atlas/core"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
	"github.com/mapprotocol/compass/pkg/ethclient"
)

const testHead = 32

type testBackend struct {
	chain   *chain.BlockChain
	pending bool // pending block available
}

func (b *testBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error) {
	if number > testHead {
		return nil, nil
	}
	if number == rpc.LatestBlockNumber {
		number = testHead
	}
	if number == rpc.PendingBlockNumber {
		if b.pending {
			number = testHead + 1
		} else {
			return nil, nil
		}
	}
	return b.chain.GetHeaderByNumber(uint64(number)), nil
}

func (b *testBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error) {
	if number > testHead {
		return nil, nil
	}
	if number == rpc.LatestBlockNumber {
		number = testHead
	}
	if number == rpc.PendingBlockNumber {
		if b.pending {
			number = testHead + 1
		} else {
			return nil, nil
		}
	}
	return b.chain.GetBlockByNumber(uint64(number)), nil
}

func (b *testBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	return b.chain.GetReceiptsByHash(hash), nil
}

func (b *testBackend) PendingBlockAndReceipts() (*types.Block, types.Receipts) {
	if b.pending {
		block := b.chain.GetBlockByNumber(testHead + 1)
		return block, b.chain.GetReceiptsByHash(block.Hash())
	}
	return nil, nil
}

func (b *testBackend) ChainConfig() *params.ChainConfig {
	return b.chain.Config()
}

func (b *testBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return nil
}

func newTestBackend(t *testing.T, londonBlock *big.Int, pending bool) *testBackend {
	var (
		key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		addr   = crypto.PubkeyToAddress(key.PublicKey)
		config = *params.TestChainConfig // needs copy because it is modified below
		gspec  = &chain.Genesis{
			Config: &config,
			Alloc:  chain.GenesisAlloc{addr: {Balance: big.NewInt(math.MaxInt64)}},
		}
		signer = types.LatestSigner(gspec.Config)
	)
	config.LondonBlock = londonBlock
	engine := consensustest.NewFaker()
	db := rawdb.NewMemoryDatabase()
	genesis, _ := gspec.Commit(db)

	// Generate testing blocks
	blocks, _ := chain.GenerateChain(gspec.Config, genesis, engine, db, testHead+1, func(i int, b *chain.BlockGen) {
		b.SetCoinbase(common.Address{1})

		var txdata types.TxData
		if londonBlock != nil && b.Number().Cmp(londonBlock) >= 0 {
			txdata = &types.DynamicFeeTx{
				ChainID:   gspec.Config.ChainID,
				Nonce:     b.TxNonce(addr),
				To:        &common.Address{},
				Gas:       30000,
				GasFeeCap: big.NewInt(100 * params.GWei),
				GasTipCap: big.NewInt(int64(i+1) * params.GWei),
				Data:      []byte{},
			}
		} else {
			txdata = &types.LegacyTx{
				Nonce:    b.TxNonce(addr),
				To:       &common.Address{},
				Gas:      21000,
				GasPrice: big.NewInt(int64(i+1) * params.GWei),
				Value:    big.NewInt(100),
				Data:     []byte{},
			}
		}
		b.AddTx(types.MustSignNewTx(key, signer, txdata))
	})
	// Construct testing chain
	diskdb := rawdb.NewMemoryDatabase()
	gspec.Commit(diskdb)
	chain, err := chain.NewBlockChain(diskdb, &chain.CacheConfig{TrieCleanNoPrefetch: true}, &config, engine, vm.Config{}, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create local chain, %v", err)
	}
	chain.InsertChain(blocks)
	return &testBackend{chain: chain, pending: pending}
}

func (b *testBackend) CurrentHeader() *types.Header {
	return b.chain.CurrentHeader()
}

func (b *testBackend) GetBlockByNumber(number uint64) *types.Block {
	return b.chain.GetBlockByNumber(number)
}

func TestSuggestTipCap(t *testing.T) {
	config := Config{
		Blocks:     3,
		Percentile: 60,
		Default:    big.NewInt(params.GWei),
	}
	var cases = []struct {
		fork   *big.Int // London fork number
		expect *big.Int // Expected gasprice suggestion
	}{
		{nil, big.NewInt(params.GWei * int64(30))},
		{big.NewInt(0), big.NewInt(params.GWei * int64(30))},  // Fork point in genesis
		{big.NewInt(1), big.NewInt(params.GWei * int64(30))},  // Fork point in first block
		{big.NewInt(32), big.NewInt(params.GWei * int64(30))}, // Fork point in last block
		{big.NewInt(33), big.NewInt(params.GWei * int64(30))}, // Fork point in the future
	}
	for _, c := range cases {
		backend := newTestBackend(t, c.fork, false)
		oracle := NewOracle(backend, config)

		// The gas price sampled is: 32G, 31G, 30G, 29G, 28G, 27G
		got, err := oracle.SuggestTipCap(context.Background())
		if err != nil {
			t.Fatalf("Failed to retrieve recommended gas price: %v", err)
		}
		if got.Cmp(c.expect) != 0 {
			t.Fatalf("Gas price mismatch, want %d, got %d", c.expect, got)
		}
	}
}

func getBlockValues(ctx context.Context, signer types.Signer, blockNum uint64,
	limit int, ignoreUnder *big.Int, result chan results, quit chan struct{}) {
	url := "https://poc3-rpc.maplabs.io"
	c, e := ethclient.Dial(url)
	if e != nil {
		panic(e)
	}

	block, err := c.MAPBlockByNumber(ctx, big.NewInt(int64((blockNum))))
	if block == nil {
		select {
		case result <- results{nil, err}:
		case <-quit:
		}
		return
	}
	// Sort the transaction by effective tip in ascending sort.
	txs := make([]*types.Transaction, len(block.Transactions()))
	copy(txs, block.Transactions())

	sorter := newSorter(txs, block.BaseFee())
	sort.Sort(sorter)

	var prices []*big.Int
	for _, tx := range sorter.txs {
		fmt.Println("===", tx.GasFeeCap(), tx.GasTipCap(), block.BaseFee())
		tip, _ := tx.EffectiveGasTip(block.BaseFee())
		if ignoreUnder != nil && tip.Cmp(ignoreUnder) == -1 {
			continue
		}
		sender, err := types.Sender(signer, tx)
		if err == nil && sender != block.Coinbase() {
			prices = append(prices, tip)
			if len(prices) >= limit {
				break
			}
		}
	}
	fmt.Println("===prices", prices, "len", len(prices), block.Number(), len(txs))
	select {
	case result <- results{prices, nil}:
	case <-quit:
	}
}

func Test03(t *testing.T) {

	checkBlocks := 20
	num := big.NewInt(2865114)
	lastPrice := big.NewInt(0)
	var (
		sent, exp int
		number    = num.Uint64()
		result    = make(chan results, checkBlocks)
		quit      = make(chan struct{})
		results   []*big.Int
	)
	for sent < checkBlocks && number > 0 {
		go getBlockValues(context.Background(), types.MakeSigner(params.MainnetChainConfig,
			big.NewInt(int64(number))), number, sampleNumber, DefaultIgnorePrice, result, quit)
		sent++
		exp++
		number--
	}
	for exp > 0 {
		res := <-result
		if res.err != nil {
			close(quit)
			t.Error(res.err)
		}
		exp--
		// Nothing returned. There are two special cases here:
		// - The block is empty
		// - All the transactions included are sent by the miner itself.
		// In these cases, use the latest calculated price for sampling.
		if len(res.values) == 0 {
			res.values = []*big.Int{lastPrice}
		}
		// Besides, in order to collect enough data for sampling, if nothing
		// meaningful returned, try to query more blocks. But the maximum
		// is 2*checkBlocks.
		if len(res.values) == 1 && len(results)+1+exp < checkBlocks*2 && number > 0 {
			go getBlockValues(context.Background(), types.MakeSigner(params.MainnetChainConfig, big.NewInt(int64(number))),
				number, sampleNumber, DefaultIgnorePrice, result, quit)
			sent++
			exp++
			number--
		}
		results = append(results, res.values...)
	}
	price := lastPrice
	if len(results) > 0 {
		sort.Sort(bigIntArray(results))
		pos := (len(results) - 1) * 60 / 100
		price = results[pos]
		fmt.Println(pos)
		fmt.Println(results)
	}
	if price.Cmp(DefaultMaxPrice) > 0 {
		price = new(big.Int).Set(DefaultMaxPrice)
	}
	fmt.Println(price)
}
