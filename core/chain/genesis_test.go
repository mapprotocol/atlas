// Copyright 2017 The go-ethereum Authors
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

package chain

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/mapprotocol/atlas/consensus"
	"github.com/mapprotocol/atlas/core/vm"
	params2 "github.com/mapprotocol/atlas/params"
	"io/ioutil"
	"math/big"

	"github.com/mapprotocol/atlas/core/state"
	//"math/big"
	//"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/types"
)

func TestDefaultGenesisBlock(t *testing.T) {
	fmt.Println("address:", common.Address{}.String())
	EmptyRootHash0 := types.DeriveSha(types.Transactions{}, trie.NewStackTrie(nil))
	fmt.Println(EmptyRootHash0)
	block := DefaultGenesisBlock().ToBlock(nil)
	if block.Hash() != params2.MainnetGenesisHash {
		t.Errorf("wrong mainnet genesis hash, got %v, want %v", block.Hash(), params2.MainnetGenesisHash)
	}
	block = DefaultTestnetGenesisBlock().ToBlock(nil)
	if block.Hash() != params2.TestnetGenesisHash {
		t.Errorf("wrong ropsten genesis hash, got %v, want %v", block.Hash(), params.RopstenGenesisHash)
	}
}

//func TestSetupGenesis(t *testing.T) {
//	var (
//		customghash = common.HexToHash("0x89c99d90b79719238d2645c7642f2c9295246e80775b38cfd162b696817fbd50")
//		customg     = Genesis{
//			Config: &params.ChainConfig{HomesteadBlock: big.NewInt(3)},
//			Alloc: GenesisAlloc{
//				{1}: {Balance: big.NewInt(1), Storage: map[common.Hash]common.Hash{{1}: {1}}},
//			},
//		}
//		oldcustomg = customg
//	)
//	oldcustomg.Config = &params.ChainConfig{HomesteadBlock: big.NewInt(2)}
//	tests := []struct {
//		name       string
//		fn         func(ethdb.Database) (*params.ChainConfig, common.Hash, error)
//		wantConfig *params.ChainConfig
//		wantHash   common.Hash
//		wantErr    error
//	}{
//		{
//			name: "genesis without ChainConfig",
//			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
//				return SetupGenesisBlock(db, new(Genesis))
//			},
//			wantErr:    errGenesisNoConfig,
//			wantConfig: params.AllEthashProtocolChanges,
//		},
//		{
//			name: "no block in DB, genesis == nil",
//			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
//				return SetupGenesisBlock(db, nil)
//			},
//			wantHash:   params.MainnetGenesisHash,
//			wantConfig: params.MainnetChainConfig,
//		},
//		{
//			name: "mainnet block in DB, genesis == nil",
//			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
//				DefaultGenesisBlock().MustCommit(db)
//				return SetupGenesisBlock(db, nil)
//			},
//			wantHash:   params.MainnetGenesisHash,
//			wantConfig: params.MainnetChainConfig,
//		},
//		{
//			name: "custom block in DB, genesis == nil",
//			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
//				customg.MustCommit(db)
//				return SetupGenesisBlock(db, nil)
//			},
//			wantHash:   customghash,
//			wantConfig: customg.Config,
//		},
//		{
//			name: "custom block in DB, genesis == ropsten",
//			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
//				customg.MustCommit(db)
//				return SetupGenesisBlock(db, DefaultTestnetGenesisBlock())
//			},
//			wantErr:    &GenesisMismatchError{Stored: customghash, New: params.RopstenGenesisHash},
//			wantHash:   params.RopstenGenesisHash,
//			wantConfig: params2.TestnetConfig,
//		},
//		{
//			name: "compatible config in DB",
//			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
//				oldcustomg.MustCommit(db)
//				return SetupGenesisBlock(db, &customg)
//			},
//			wantHash:   customghash,
//			wantConfig: customg.Config,
//		},
//		{
//			name: "incompatible config in DB",
//			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
//				// Commit the 'old' genesis block with Homestead transition at #2.
//				// Advance to block #4, past the homestead transition block of customg.
//				genesis := oldcustomg.MustCommit(db)
//
//				bc, _ := NewBlockChain(db, nil, oldcustomg.Config, ethash.NewFullFaker(), vm.Config{}, nil, nil)
//				defer bc.Stop()
//
//				blocks, _ := GenerateChain(oldcustomg.Config, genesis, ethash.NewFaker(), db, 4, nil)
//				bc.InsertChain(blocks)
//				bc.CurrentBlock()
//				// This should return a compatibility error.
//				return SetupGenesisBlock(db, &customg)
//			},
//			wantHash:   customghash,
//			wantConfig: customg.Config,
//			wantErr: &params.ConfigCompatError{
//				What:         "Homestead fork block",
//				StoredConfig: big.NewInt(2),
//				NewConfig:    big.NewInt(3),
//				RewindTo:     1,
//			},
//		},
//	}
//
//	for _, test := range tests {
//		db := rawdb.NewMemoryDatabase()
//		config, hash, err := test.fn(db)
//		// Check the return values.
//		if !reflect.DeepEqual(err, test.wantErr) {
//			spew := spew.ConfigState{DisablePointerAddresses: true, DisableCapacities: true}
//			t.Errorf("%s: returned error %#v, want %#v", test.name, spew.NewFormatter(err), spew.NewFormatter(test.wantErr))
//		}
//		if !reflect.DeepEqual(config, test.wantConfig) {
//			t.Errorf("%s:\nreturned %v\nwant     %v", test.name, config, test.wantConfig)
//		}
//		if hash != test.wantHash {
//			t.Errorf("%s: returned hash %s, want %s", test.name, hash.Hex(), test.wantHash.Hex())
//		} else if err == nil {
//			// Check database content.
//			stored := rawdb.ReadBlock(db, test.wantHash, 0)
//			if stored.Hash() != test.wantHash {
//				t.Errorf("%s: block in DB has hash %s, want %s", test.name, stored.Hash(), test.wantHash)
//			}
//		}
//	}
//}

// TestGenesisHashes checks the congruity of default genesis data to corresponding hardcoded genesis hash values.
func TestGenesisHashes(t *testing.T) {
	cases := []struct {
		genesis *Genesis
		hash    common.Hash
	}{
		{
			genesis: DefaultGenesisBlock(),
			hash:    params2.MainnetGenesisHash,
		},
		{
			genesis: DefaultTestnetGenesisBlock(),
			hash:    params2.TestnetGenesisHash,
		},
		{
			genesis: DevnetGenesisBlock(common.Address{}),
			hash:    params2.DevnetGenesisHash,
		},
	}
	for i, c := range cases {
		b := c.genesis.MustCommit(rawdb.NewMemoryDatabase())
		if got := b.Hash(); got != c.hash {
			t.Errorf("case: %d, want: %s, got: %s", i, c.hash.Hex(), got.Hex())
		}
	}
}
func generateAddr() common.Address {
	priv, _ := crypto.GenerateKey()
	privHex := hex.EncodeToString(crypto.FromECDSA(priv))
	fmt.Println(privHex)
	addr := crypto.PubkeyToAddress(priv.PublicKey)
	fmt.Println(addr.String())
	fmt.Println("finish")
	return addr
}

func Test01(t *testing.T) {
	generateAddr()
}

func genesisReadContract(t *testing.T) GenesisAlloc {
	mainnetAlloc := &GenesisAlloc{}
	genesisPreContractPath := "D:\\work\\zhangwei812\\atlas\\poc2.json"
	if common.FileExist(genesisPreContractPath) {
		t.Logf("loaded the genesisPreContractPath%s%s", "buildpath", genesisPreContractPath)
		jsonData, err := ioutil.ReadFile(genesisPreContractPath)
		if err != nil {
			t.Error("loaded the genesisPreContractPath jsonData err ", "err", err)
			return *mainnetAlloc
		}
		mainnetAlloc.UnmarshalJSON(jsonData)
	}
	return *mainnetAlloc
}

func TestReadPoc2Contracts(t *testing.T) {
	makaluPoc2Number150Root := "0x7a230bf7e6bbe4bfdfb19a5b7f8ed77cce884baf67b425cc118d5a6d14d5c13a"
	db := rawdb.NewMemoryDatabase()
	statedb, err := state.New(common.Hash{}, state.NewDatabase(db), nil)
	if err != nil {
		panic(err)
	}
	g := genesisReadContract(t)
	for addr, account := range g {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}

	//////////////////////////////////pro compiled////////////////////////////////////
	Number := uint64(0)
	consensus.InitHeaderStore(statedb, new(big.Int).SetUint64(Number))
	consensus.InitTxVerify(statedb, new(big.Int).SetUint64(Number))
	register := vm.NewRegisterImpl()
	_, err = register.DoElections(statedb, 1, 0)
	if err != nil {
		t.Error("ToBlock DoElections", "error", err)
	}
	err = register.Save(statedb, params2.RelayerAddress)
	if err != nil {
		t.Error("ToBlock IMPL Save", "error", err)
	}
	////////////////////////////////////////////////////////////////////////////
	root := statedb.IntermediateRoot(false)
	t.Logf("root %s", root)
	if root.String() != makaluPoc2Number150Root {
		t.Fatalf("root != makaluPoc2Number150Root%s%s", "myroot", root)
	}

}
