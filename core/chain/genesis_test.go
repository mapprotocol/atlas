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
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/consensus"
	"golang.org/x/crypto/sha3"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/types"
	params2 "github.com/mapprotocol/atlas/params"

	"io/ioutil"
	"testing"
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
		t.Errorf("wrong ropsten genesis hash, got %v, want %v", block.Hash(), params2.TestnetGenesisHash)
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
		//{
		//	genesis: DevnetGenesisBlock(common.Address{}),
		//	hash:    params2.DevnetGenesisHash,
		//},
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
	////////////////////////////////////////////////////////////////////////////
	root := statedb.IntermediateRoot(false)
	t.Logf("root %s", root)
	if root.String() != makaluPoc2Number150Root {
		t.Fatalf("root != makaluPoc2Number150Root %s %s", "myroot", root)
	}
}

var poc2Genesis = `{
	"config": {
	"chainId": 22776,
	"homesteadBlock": 0,
	"daoForkBlock": 0,
	"daoForkSupport": true,
	"eip150Block": 0,
	"eip150Hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
	"eip155Block": 0,
	"eip158Block": 0,
	"byzantiumBlock": 0,
	"constantinopleBlock": 0,
	"petersburgBlock": 0,
	"istanbulBlock": 0,
	"muirGlacierBlock": 0,
	"berlinBlock": 0,
	"londonBlock": 0,
	"istanbul": {
	"epoch": 20000,
	"policy": 2,
	"lookbackwindow": 12,
	"blockperiod": 5,
	"requesttimeout": 3000
	},
	"FullHeaderChainAvailable": false
	},
	"nonce": "0x42",
	"timestamp": "0x0",
	"extraData": "0x0000000000000000000000000000000000000000000000000000000000000000f901ebf85494f18d71e825c43e5ee5f3bd0384670eef53a3309e94f22b4ae180279dabcf5cd8f4850545ae44521ce994a47444c9daac489777dfeb5f30b03a6f3b4b6337941f39d97a8f697502884fe01cf23dba4eb66e0481f90188b86041df7be08167a3c7635716418eb42508bee7d97165e6f3482fb55c0a32d2cdc07c8170b97e427c667a87fb8e6f041700b2b1dce0d01a8adadc5816c2c28762ad28730faa9464e65ae7e8031f45fdd7205c499fd92a41ccec5bc97f2dd15da700b860051fe96e2b46e5708d4081be01ecebadba33a9ec37c9c4219a509b1ff7f1a5f3a3866e4a67050df207cc6546ced94c006f67908ad64656566bb58ebce7ec6bb1a2534c40bf94f6ad205c686ff1ccad1be221c1c82a00cdf989ff98b418810200b86038030897213e9b7837e600785e3376214948c9bafda2551315fe969206d0be434661c8b4dd6a6298b7f9896efcf3dc002bfd7c2b4d1c7224b0516c76e5ac7fd58a6e72e22b58debcbcaa2b9c72837d6faa6e8e64e02ca222e3ebfd07f25a0580b860d8b24d419755d8d82b878993d58e7ddd19a19988e00ba55adff574dd9e3df3b45451fe2e56c5793048b0a2c617b11601c451c63e1ce5730f3877a77c026dfdb40349543dfef722dde6f4e06aaf3070ed740d26ae9193d893f5e9d87b67c460808080c3808080c3808080",
	"gasLimit": "0x1312d00",
	"mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
	"coinbase": "0x0000000000000000000000000000000000000000",
	"alloc": {
	"000000000000000000000000000000000000ce10": {
	"code": "0x60806040526004361061004a5760003560e01c806303386ba3146101e757806342404e0714610280578063bb913f41146102d7578063d29d44ee14610328578063f7e6af8014610379575b6000600160405180807f656970313936372e70726f78792e696d706c656d656e746174696f6e00000000815250601c019050604051809103902060001c0360001b9050600081549050600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161415610136576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260158152602001807f4e6f20496d706c656d656e746174696f6e20736574000000000000000000000081525060200191505060405180910390fd5b61013f816103d0565b6101b1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f496e76616c696420636f6e74726163742061646472657373000000000000000081525060200191505060405180910390fd5b60405136810160405236600082376000803683855af43d604051818101604052816000823e82600081146101e3578282f35b8282fd5b61027e600480360360408110156101fd57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019064010000000081111561023a57600080fd5b82018360208201111561024c57600080fd5b8035906020019184600183028401116401000000008311171561026e57600080fd5b909192939192939050505061041b565b005b34801561028c57600080fd5b506102956105c1565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b3480156102e357600080fd5b50610326600480360360208110156102fa57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061060d565b005b34801561033457600080fd5b506103776004803603602081101561034b57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506107bd565b005b34801561038557600080fd5b5061038e610871565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b60008060007fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a47060001b9050833f915080821415801561041257506000801b8214155b92505050919050565b610423610871565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146104c3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f73656e64657220776173206e6f74206f776e657200000000000000000000000081525060200191505060405180910390fd5b6104cc8361060d565b600060608473ffffffffffffffffffffffffffffffffffffffff168484604051808383808284378083019250505092505050600060405180830381855af49150503d8060008114610539576040519150601f19603f3d011682016040523d82523d6000602084013e61053e565b606091505b508092508193505050816105ba576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601e8152602001807f696e697469616c697a6174696f6e2063616c6c6261636b206661696c6564000081525060200191505060405180910390fd5b5050505050565b600080600160405180807f656970313936372e70726f78792e696d706c656d656e746174696f6e00000000815250601c019050604051809103902060001c0360001b9050805491505090565b610615610871565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146106b5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f73656e64657220776173206e6f74206f776e657200000000000000000000000081525060200191505060405180910390fd5b6000600160405180807f656970313936372e70726f78792e696d706c656d656e746174696f6e00000000815250601c019050604051809103902060001c0360001b9050610701826103d0565b610773576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f496e76616c696420636f6e74726163742061646472657373000000000000000081525060200191505060405180910390fd5b8181558173ffffffffffffffffffffffffffffffffffffffff167fab64f92ab780ecbf4f3866f57cee465ff36c89450dcce20237ca7a8d81fb7d1360405160405180910390a25050565b6107c5610871565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610865576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f73656e64657220776173206e6f74206f776e657200000000000000000000000081525060200191505060405180910390fd5b61086e816108bd565b50565b600080600160405180807f656970313936372e70726f78792e61646d696e000000000000000000000000008152506013019050604051809103902060001c0360001b9050805491505090565b600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161415610960576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260118152602001807f6f776e65722063616e6e6f74206265203000000000000000000000000000000081525060200191505060405180910390fd5b6000600160405180807f656970313936372e70726f78792e61646d696e000000000000000000000000008152506013019050604051809103902060001c0360001b90508181558173ffffffffffffffffffffffffffffffffffffffff167f50146d0e3c60aa1d17a70635b05494f864e86144a2201275021014fbf08bafe260405160405180910390a2505056fea165627a7a723058202dbb6037e4381b4ad95015ed99441a23345cc2ae52ef27e2e91d34fb0acd277b0029",
	"storage": {
	"0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103": "0x000000000000000000000000456f41406b32c45d59e539e4bba3d7898c3584da"
	},
	"balance": "0x0"
	},
	"c732efcaa62cba951d81bb889bb0f8f6e952d70d": {
	"balance": "0x33b2e3c9fd0803ce8000000"
	}
	},
	"number": "0x0",
	"gasUsed": "0x0",
	"parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
	"baseFeePerGas": null
}`
var poc3Genesis = `{
"config": {
"chainId": 214,
"homesteadBlock": 0,
"daoForkBlock": 0,
"daoForkSupport": true,
"eip150Block": 0,
"eip150Hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
"eip155Block": 0,
"eip158Block": 0,
"byzantiumBlock": 0,
"constantinopleBlock": 0,
"petersburgBlock": 0,
"istanbulBlock": 0,
"muirGlacierBlock": 0,
"berlinBlock": 0,
"londonBlock": 0,
"istanbul": {
"epoch": 30000,
"policy": 2,
"lookbackwindow": 12,
"blockperiod": 5,
"requesttimeout": 3000
},
"FullHeaderChainAvailable": false
},
"nonce": "0x42",
"timestamp": "0x0",
"extraData": "0x0000000000000000000000000000000000000000000000000000000000000000f8eaf854941c0edab88dbb72b119039c4d14b1663525b3ac159416fdbcac4d4cc24dca47b9b80f58155a551ca2af942dc45799000ab08e60b7441c36fcc74060ccbe11946c5938b49bacde73a8db7c3a7da208846898bff5f888a182b9df317d21429c6f0b74c96c21a610483be1d234c2815c50be454a689c35ae01a132071fff6599fcdefb78d8048abf7d32165e4dad0a00d7667ba4e1933a6f1bff00a166b74fbfc9c23963a9a21e12d79422fc288b7598b58f23d4ec04ea2657a05a9901a140cdae9b90b80179ac73341dd83974fa6dd85f921080770241df8b4f3eb2244e018080c3808080c3808080",
"gasLimit": "0x1312d00",
"mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
"coinbase": "0x0000000000000000000000000000000000000000",
"alloc": {
"c732efcaa62cba951d81bb889bb0f8f6e952d70d": {
"balance": "0x33b2e3c9fd0803ce8000000"
}
},
"number": "0x0",
"gasUsed": "0x0",
"parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
"baseFeePerGas": null
}`

func TestComparePOC2_3_Genesis(t *testing.T) {
	bytePoc2, err := json.Marshal(poc2Genesis)
	if err != nil {
		t.Fatalf("read poc2 Genesis err%s", err)
	}
	poc3G := DefaultGenesisBlock()
	bytePoc3, err := json.Marshal(poc3G)
	if err != nil {
		t.Fatalf("read poc2 Genesis err%s", err)
	}
	if !bytes.Equal(bytePoc2, bytePoc3) {
		t.Fatalf("poc2 genesis != poc3  genesis err")
	}
	//poc2 poc3  chainId、 epoch 、extraData are inconsistent
	//poc2 One more 000000000000000000000000000000000000ce10 contracts
}
func TestPrintPreContractAddr(t *testing.T) {
	t.Log(params2.HeaderStoreAddress)
}

func Test06(t *testing.T) {
	type txLogs struct{
		PostStateOrStatus []byte
		CumulativeGasUsed uint
		Bloom []byte
	}
	p1,_ := hex.DecodeString("2")
	b1,_ := hex.DecodeString("1")
	tx1 := txLogs{
		PostStateOrStatus: p1,
		CumulativeGasUsed: 1,
		Bloom: b1,
	}
	r1,_ := rlp.EncodeToBytes(tx1)
	fmt.Println(hex.EncodeToString(r1))
	fmt.Println(rlpHash(tx1).String())
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	if e := rlp.Encode(hw, x); e != nil {
		panic(e.Error())
	}
	hw.Sum(h[:0])
	return h
}
