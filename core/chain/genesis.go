// Copyright 2014 The go-ethereum Authors
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
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	ethparams "github.com/ethereum/go-ethereum/params"

	"github.com/mapprotocol/atlas/consensus"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
)

//go:generate gencodec -type Genesis -field-override genesisSpecMarshaling -out gen_genesis.go
//go:generate gencodec -type GenesisAccount -field-override genesisAccountMarshaling -out gen_genesis_account.go

const mainnetAllocJSON = "{\"0x11901cf7eEae1E2644995FB2E47Ce46bC7F33246\":{\"balance\":\"120000000000000000000000000\"},\"0xC1cDA18694F5B86cFB80c1B4f8Cc046B0d7E6326\":{\"balance\":\"20000000000000000000000000\"},\"0xa5d40D93b01AfBafec84E20018Aff427628F645E\":{\"balance\":\"20000000000000000000000000\"},\"0x8d485780E84E23437f8F6938D96B964645529127\":{\"balance\":\"20000000000000000000000000\"},\"0x5F857c501b73ddFA804234f1f1418D6f75554076\":{\"balance\":\"20000000000000000000000000\"},\"0xaa9064F57F8d7de4b3e08c35561E21Afd6341390\":{\"balance\":\"20000000000000000000000000\"},\"0x7FA26b50b3e9a2eC8AD1850a4c4FBBF94D806E95\":{\"balance\":\"20000000000000000000000000\"},\"0x08960Ce6b58BE32FBc6aC1489d04364B4f7dC216\":{\"balance\":\"20000000000000000000000000\"},\"0x77B68B2e7091D4F242a8Af89F200Af941433C6d8\":{\"balance\":\"20000000000000000000000000\"},\"0x75Bb69C002C43f5a26a2A620518775795Fd45ecf\":{\"balance\":\"20000000000000000000000000\"},\"0x19992AE48914a178Bf138665CffDD8CD79b99513\":{\"balance\":\"20000000000000000000000000\"},\"0xE23a4c6615669526Ab58E9c37088bee4eD2b2dEE\":{\"balance\":\"20000000000000000000000\"},\"0xDe22679dCA843B424FD0BBd70A22D5F5a4B94fe4\":{\"balance\":\"10200014000000000000000000\"},\"0x743D80810fe10c5C3346D2940997cC9647035B13\":{\"balance\":\"20513322000000000000000000\"},\"0x8e1c4355307F1A59E7eD4Ae057c51368b9338C38\":{\"balance\":\"7291740000000000000000000\"},\"0x417fe63186C388812e342c85FF87187Dc584C630\":{\"balance\":\"20000062000000000000000000\"},\"0xF5720c180a6Fa14ECcE82FB1bB060A39E93A263c\":{\"balance\":\"30000061000000000000000000\"},\"0xB80d1e7F9CEbe4b5E1B1Acf037d3a44871105041\":{\"balance\":\"9581366833333333333333335\"},\"0xf8ed78A113cD2a34dF451Ba3D540FFAE66829AA0\":{\"balance\":\"11218686833333333333333333\"},\"0x9033ff75af27222c8f36a148800c7331581933F3\":{\"balance\":\"11218686833333333333333333\"},\"0x8A07541C2eF161F4e3f8de7c7894718dA26626B2\":{\"balance\":\"11218686833333333333333333\"},\"0xB2fe7AFe178335CEc3564d7671EEbD7634C626B0\":{\"balance\":\"11218686833333333333333333\"},\"0xc471776eA02705004C451959129bF09423B56526\":{\"balance\":\"11218686833333333333333333\"},\"0xeF283eca68DE87E051D427b4be152A7403110647\":{\"balance\":\"14375000000000000000000000\"},\"0x7cf091C954ed7E9304452d31fd59999505Ddcb7a\":{\"balance\":\"14375000000000000000000000\"},\"0xa5d2944C32a8D7b284fF0b84c20fDcc46937Cf64\":{\"balance\":\"14375000000000000000000000\"},\"0xFC89C17525f08F2Bc9bA8cb77BcF05055B1F7059\":{\"balance\":\"14375000000000000000000000\"},\"0x3Fa7C646599F3174380BD9a7B6efCde90b5d129d\":{\"balance\":\"14375000000000000000000000\"},\"0x989e1a3B344A43911e02cCC609D469fbc15AB1F1\":{\"balance\":\"14375000000000000000000000\"},\"0xAe1d640648009DbE0Aa4485d3BfBB68C37710924\":{\"balance\":\"20025000000000000000000000\"},\"0x1B6C64779F42BA6B54C853Ab70171aCd81b072F7\":{\"balance\":\"20025000000000000000000000\"},\"000000000000000000000000000000000000ce10\":{\"code\":\"0x60806040526004361061004a5760003560e01c806303386ba3146101e757806342404e0714610280578063bb913f41146102d7578063d29d44ee14610328578063f7e6af8014610379575b6000600160405180807f656970313936372e70726f78792e696d706c656d656e746174696f6e00000000815250601c019050604051809103902060001c0360001b9050600081549050600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161415610136576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260158152602001807f4e6f20496d706c656d656e746174696f6e20736574000000000000000000000081525060200191505060405180910390fd5b61013f816103d0565b6101b1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f496e76616c696420636f6e74726163742061646472657373000000000000000081525060200191505060405180910390fd5b60405136810160405236600082376000803683855af43d604051818101604052816000823e82600081146101e3578282f35b8282fd5b61027e600480360360408110156101fd57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019064010000000081111561023a57600080fd5b82018360208201111561024c57600080fd5b8035906020019184600183028401116401000000008311171561026e57600080fd5b909192939192939050505061041b565b005b34801561028c57600080fd5b506102956105c1565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b3480156102e357600080fd5b50610326600480360360208110156102fa57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061060d565b005b34801561033457600080fd5b506103776004803603602081101561034b57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506107bd565b005b34801561038557600080fd5b5061038e610871565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b60008060007fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a47060001b9050833f915080821415801561041257506000801b8214155b92505050919050565b610423610871565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146104c3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f73656e64657220776173206e6f74206f776e657200000000000000000000000081525060200191505060405180910390fd5b6104cc8361060d565b600060608473ffffffffffffffffffffffffffffffffffffffff168484604051808383808284378083019250505092505050600060405180830381855af49150503d8060008114610539576040519150601f19603f3d011682016040523d82523d6000602084013e61053e565b606091505b508092508193505050816105ba576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601e8152602001807f696e697469616c697a6174696f6e2063616c6c6261636b206661696c6564000081525060200191505060405180910390fd5b5050505050565b600080600160405180807f656970313936372e70726f78792e696d706c656d656e746174696f6e00000000815250601c019050604051809103902060001c0360001b9050805491505090565b610615610871565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146106b5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f73656e64657220776173206e6f74206f776e657200000000000000000000000081525060200191505060405180910390fd5b6000600160405180807f656970313936372e70726f78792e696d706c656d656e746174696f6e00000000815250601c019050604051809103902060001c0360001b9050610701826103d0565b610773576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f496e76616c696420636f6e74726163742061646472657373000000000000000081525060200191505060405180910390fd5b8181558173ffffffffffffffffffffffffffffffffffffffff167fab64f92ab780ecbf4f3866f57cee465ff36c89450dcce20237ca7a8d81fb7d1360405160405180910390a25050565b6107c5610871565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610865576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f73656e64657220776173206e6f74206f776e657200000000000000000000000081525060200191505060405180910390fd5b61086e816108bd565b50565b600080600160405180807f656970313936372e70726f78792e61646d696e000000000000000000000000008152506013019050604051809103902060001c0360001b9050805491505090565b600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161415610960576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260118152602001807f6f776e65722063616e6e6f74206265203000000000000000000000000000000081525060200191505060405180910390fd5b6000600160405180807f656970313936372e70726f78792e61646d696e000000000000000000000000008152506013019050604051809103902060001c0360001b90508181558173ffffffffffffffffffffffffffffffffffffffff167f50146d0e3c60aa1d17a70635b05494f864e86144a2201275021014fbf08bafe260405160405180910390a2505056fea165627a7a723058206808dd43e7d765afca53fe439122bc5eac16d708ce7d463451be5042426f101f0029\",\"storage\":{\"0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103\":\"0xE23a4c6615669526Ab58E9c37088bee4eD2b2dEE\"},\"balance\":\"0\"}}"

var DBGenesisSupplyKey = []byte("genesis-supply-genesis")
var errGenesisNoConfig = errors.New("genesis has no chain configuration")

var (
	faucetAddr    = common.HexToAddress("0xf675187ff5b76d2430b353f6736aa051253118ee")
	faucetBalance = balance
)

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type Genesis struct {
	Config    *params.ChainConfig `json:"config"`
	Nonce     uint64              `json:"nonce"`
	Timestamp uint64              `json:"timestamp"`
	ExtraData []byte              `json:"extraData"`
	GasLimit  uint64              `json:"gasLimit"   gencodec:"required"`
	Mixhash   common.Hash         `json:"mixHash"`
	Coinbase  common.Address      `json:"coinbase"`
	Alloc     GenesisAlloc        `json:"alloc"      gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Number     uint64      `json:"number"`
	GasUsed    uint64      `json:"gasUsed"`
	ParentHash common.Hash `json:"parentHash"`
	BaseFee    *big.Int    `json:"baseFeePerGas"`
}

// GenesisAlloc specifies the initial state that is part of the genesis block.
type GenesisAlloc map[common.Address]GenesisAccount

func (ga *GenesisAlloc) UnmarshalJSON(data []byte) error {
	m := make(map[common.UnprefixedAddress]GenesisAccount)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*ga = make(GenesisAlloc)
	for addr, a := range m {
		(*ga)[common.Address(addr)] = a
	}
	return nil
}

// GenesisAccount is an account in the state of the genesis block.
type GenesisAccount struct {
	Code       []byte                      `json:"code,omitempty"`
	Storage    map[common.Hash]common.Hash `json:"storage,omitempty"`
	Balance    *big.Int                    `json:"balance" gencodec:"required"`
	Nonce      uint64                      `json:"nonce,omitempty"`
	PrivateKey []byte                      `json:"secretKey,omitempty"` // for tests
}

// field type overrides for gencodec
type genesisSpecMarshaling struct {
	Nonce      math.HexOrDecimal64
	Timestamp  math.HexOrDecimal64
	ExtraData  hexutil.Bytes
	GasLimit   math.HexOrDecimal64
	GasUsed    math.HexOrDecimal64
	Number     math.HexOrDecimal64
	Difficulty *math.HexOrDecimal256
	BaseFee    *math.HexOrDecimal256
	Alloc      map[common.UnprefixedAddress]GenesisAccount
}

type genesisAccountMarshaling struct {
	Code       hexutil.Bytes
	Balance    *math.HexOrDecimal256
	Nonce      math.HexOrDecimal64
	Storage    map[storageJSON]storageJSON
	PrivateKey hexutil.Bytes
}

// storageJSON represents a 256 bit byte array, but allows less than 256 bits when
// unmarshaling from hex.
type storageJSON common.Hash

func (h *storageJSON) UnmarshalText(text []byte) error {
	text = bytes.TrimPrefix(text, []byte("0x"))
	if len(text) > 64 {
		return fmt.Errorf("too many hex characters in storage key/value %q", text)
	}
	offset := len(h) - len(text)/2 // pad on the left
	if _, err := hex.Decode(h[offset:], text); err != nil {
		fmt.Println(err)
		return fmt.Errorf("invalid hex storage key/value %q", text)
	}
	return nil
}

func (h storageJSON) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

// GenesisMismatchError is raised when trying to overwrite an existing
// genesis block with an incompatible one.
type GenesisMismatchError struct {
	Stored, New common.Hash
}

func (e *GenesisMismatchError) Error() string {
	return fmt.Sprintf("database contains incompatible genesis (have %x, new %x)", e.Stored, e.New)
}

// SetupGenesisBlock writes or updates the genesis block in db.
// The block that will be used is:
//
//                          genesis == nil       genesis != nil
//                       +------------------------------------------
//     db has no genesis |  main-net default  |  genesis
//     db has genesis    |  from DB           |  genesis (if compatible)
//
// The stored chain configuration will be updated if it is compatible (i.e. does not
// specify a fork block below the local head block). In case of a conflict, the
// error is a *ethparams.ConfigCompatError and the new, unwritten config is returned.
//
// The returned chain configuration is never nil.
func SetupGenesisBlock(db ethdb.Database, genesis *Genesis) (*params.ChainConfig, common.Hash, error) {
	return SetupGenesisBlockWithOverride(db, genesis, nil)
}

func SetupGenesisBlockWithOverride(db ethdb.Database, genesis *Genesis, overrideLondon *big.Int) (*params.ChainConfig, common.Hash, error) {
	if genesis != nil && genesis.Config == nil {
		return params.AllEthashProtocolChanges, common.Hash{}, errGenesisNoConfig
	}
	// Just commit the new block if there is no stored genesis block.
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored == common.Hash{}) {
		if genesis == nil {
			log.Info("Writing default main-net genesis block")
			genesis = DefaultGenesisBlock()
		} else {
			log.Info("Writing custom genesis block")
		}
		block, err := genesis.Commit(db)
		if err != nil {
			return genesis.Config, common.Hash{}, err
		}
		return genesis.Config, block.Hash(), nil
	}
	// We have the genesis block in database(perhaps in ancient database)
	// but the corresponding state is missing.
	header := rawdb.ReadHeader(db, stored, 0)
	if _, err := state.New(header.Root, state.NewDatabaseWithConfig(db, nil), nil); err != nil {
		if genesis == nil {
			genesis = DefaultGenesisBlock()
		}
		// Ensure the stored genesis matches with the given one.
		hash := genesis.ToBlock(nil).Hash()
		if hash != stored {
			return genesis.Config, hash, &GenesisMismatchError{stored, hash}
		}
		block, err := genesis.Commit(db)
		if err != nil {
			return genesis.Config, hash, err
		}
		return genesis.Config, block.Hash(), nil
	}
	// Check whether the genesis block is already written.
	if genesis != nil {
		hash := genesis.ToBlock(nil).Hash()
		if hash != stored {
			return genesis.Config, hash, &GenesisMismatchError{stored, hash}
		}
	}
	// Get the existing chain configuration.
	newcfg := genesis.configOrDefault(stored)
	if overrideLondon != nil {
		newcfg.LondonBlock = overrideLondon
	}
	if err := newcfg.CheckConfigForkOrder(); err != nil {
		return newcfg, common.Hash{}, err
	}
	storedcfg := rawdb.ReadChainConfig(db, stored)
	if storedcfg == nil {
		log.Warn("Found genesis block without chain config")
		rawdb.WriteChainConfig(db, stored, newcfg)
		return newcfg, stored, nil
	}
	// Special case: don't change the existing config of a non-mainnet chain if no new
	// config is supplied. These chains would get AllProtocolChanges (and a compat error)
	// if we just continued here.
	if genesis == nil && stored != params.MainnetGenesisHash {
		return storedcfg, stored, nil
	}
	// Check config compatibility and write the config. Compatibility errors
	// are returned to the caller unless we're already at block zero.
	height := rawdb.ReadHeaderNumber(db, rawdb.ReadHeadHeaderHash(db))
	if height == nil {
		return newcfg, stored, fmt.Errorf("missing block number for head header hash")
	}
	compatErr := storedcfg.CheckCompatible(newcfg, *height)
	if compatErr != nil && *height != 0 && compatErr.RewindTo != 0 {
		return newcfg, stored, compatErr
	}
	rawdb.WriteChainConfig(db, stored, newcfg)
	return newcfg, stored, nil
}

// StoreGenesisSupply computes the total supply of the genesis block and stores
// it in the db.
func (g *Genesis) StoreGenesisSupply(db ethdb.Database) error {
	if db == nil {
		db = rawdb.NewMemoryDatabase()
	}
	genesisSupply := big.NewInt(0)
	for _, account := range g.Alloc {
		genesisSupply.Add(genesisSupply, account.Balance)
	}
	return db.Put(DBGenesisSupplyKey, genesisSupply.Bytes())
}

func (g *Genesis) configOrDefault(ghash common.Hash) *params.ChainConfig {
	switch {
	case g != nil:
		return g.Config
	case ghash == params.MainnetGenesisHash:
		return params.MainnetChainConfig
	case ghash == params.TestnetGenesisHash:
		return params.TestnetConfig
	case ghash == params.DevnetGenesisHash:
		return params.DevnetConfig
	default:
		return params.AllEthashProtocolChanges
	}
}

// ToBlock creates the genesis block and writes state of a genesis specification
// to the given database (or discards it if nil).
func (g *Genesis) ToBlock(db ethdb.Database) *types.Block {
	if db == nil {
		db = rawdb.NewMemoryDatabase()
	}
	statedb, err := state.New(common.Hash{}, state.NewDatabase(db), nil)
	if err != nil {
		panic(err)
	}
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
	consensus.InitTxVerify(statedb, new(big.Int).SetUint64(g.Number))
	register := vm.NewRegisterImpl()
	hh := g.Number
	relayer := defaultRelayer2()
	for _, member := range relayer {
		var err error
		err = register.InsertAccount2(hh, member.Coinbase, params.ElectionMinLimitForRegister)
		if err != nil {
			log.Error("ToBlock InsertAccount", "error", err)
		} else {
			vm.GenesisAddLockedBalance(statedb, member.Coinbase, params.ElectionMinLimitForRegister)
		}
	}
	//	fmt.Println("DoElection")
	_, err = register.DoElections(statedb, 1, 0)
	if err != nil {
		log.Error("ToBlock DoElections", "error", err)
	}
	err = register.Save(statedb, params.RelayerAddress)
	if err != nil {
		log.Error("ToBlock IMPL Save", "error", err)
	}
	////////////////////////////////////////////////////////////////////////////
	root := statedb.IntermediateRoot(false)
	head := &types.Header{
		Number:     new(big.Int).SetUint64(g.Number),
		Nonce:      types.EncodeNonce(g.Nonce),
		Time:       g.Timestamp,
		ParentHash: g.ParentHash,
		Extra:      g.ExtraData,
		GasLimit:   g.GasLimit,
		GasUsed:    g.GasUsed,
		BaseFee:    g.BaseFee,
		MixDigest:  g.Mixhash,
		Coinbase:   g.Coinbase,
		Root:       root,
	}
	if head.GasLimit == 0 {
		head.GasLimit = ethparams.GenesisGasLimit
	}
	if g.Config != nil && g.Config.IsLondon(common.Big0) {
		if g.BaseFee != nil {
			head.BaseFee = g.BaseFee
		} else {
			head.BaseFee = new(big.Int).SetUint64(ethparams.InitialBaseFee)
		}
	}
	statedb.Commit(false)
	statedb.Database().TrieDB().Commit(root, true, nil)

	return types.NewBlock(head, nil, nil, nil)
}

// Commit writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func (g *Genesis) Commit(db ethdb.Database) (*types.Block, error) {
	block := g.ToBlock(db)
	if block.Number().Sign() != 0 {
		return nil, errors.New("can't commit genesis block with number > 0")
	}
	config := g.Config
	if config == nil {
		config = params.AllEthashProtocolChanges
	}
	if err := config.CheckConfigForkOrder(); err != nil {
		return nil, err
	}

	rawdb.WriteTd(db, block.Hash(), block.NumberU64(), block.TotalDifficulty())
	rawdb.WriteBlock(db, block)
	rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), nil)
	rawdb.WriteCanonicalHash(db, block.Hash(), block.NumberU64())
	rawdb.WriteHeadBlockHash(db, block.Hash())
	rawdb.WriteHeadFastBlockHash(db, block.Hash())
	rawdb.WriteHeadHeaderHash(db, block.Hash())
	rawdb.WriteChainConfig(db, block.Hash(), config)
	if err := g.StoreGenesisSupply(db); err != nil {
		log.Error("Unable to store genesisSupply in db", "err", err)
		return nil, err
	}
	return block, nil
}

// MustCommit writes the genesis block and state to db, panicking on error.
// The block is committed as the canonical head block.
func (g *Genesis) MustCommit(db ethdb.Database) *types.Block {
	block, err := g.Commit(db)
	if err != nil {
		panic(err)
	}
	return block
}

// GenesisBlockForTesting creates and writes a block in which addr has the given wei balance.
func GenesisBlockForTesting(db ethdb.Database, addr common.Address, balance *big.Int) *types.Block {
	g := Genesis{
		Alloc:   GenesisAlloc{addr: {Balance: balance}},
		BaseFee: big.NewInt(ethparams.InitialBaseFee),
	}
	return g.MustCommit(db)
}

// DefaultGenesisBlock returns the Ethereum main net genesis block.
func DefaultGenesisBlock() *Genesis {
	dr := defaultRelayer()
	for addr, allc := range genesisRegisterProxyContract() {
		// add genesis contract to allc
		dr[addr] = allc
	}

	return &Genesis{
		Config:    params.MainnetChainConfig,
		Nonce:     66,
		ExtraData: hexutil.MustDecode(mainnetExtraData),
		GasLimit:  50000000,
		Alloc:     dr,
	}
}

// DefaultTestnetGenesisBlock returns the Ropsten network genesis block.
func DefaultTestnetGenesisBlock() *Genesis {
	dr := defaultRelayer()
	for addr, allc := range genesisRegisterProxyContract() {
		// add genesis contract to allc
		dr[addr] = allc
	}
	// faucet
	dr[faucetAddr] = GenesisAccount{Balance: faucetBalance}

	return &Genesis{
		Config:    params.TestnetConfig,
		Nonce:     66,
		ExtraData: hexutil.MustDecode(testnetExtraData),
		GasLimit:  16777216,
		Alloc:     dr,
	}
}

// DevnetGenesisBlock returns the 'geth --dev' genesis block.
func DevnetGenesisBlock(faucet common.Address) *Genesis {
	dc := defaultRelayer()
	defaultBalance, _ := new(big.Int).SetString("100000000000000000000000000", 10)
	dc[faucet] = GenesisAccount{Balance: defaultBalance}
	return &Genesis{
		Config:    params.DevnetConfig,
		ExtraData: []byte{1, 2, 3},
		GasLimit:  11500000,
		Alloc:     dc,
	}
}

// SingleGenesisBlock returns the 'geth --dev' genesis block.
func SingleGenesisBlock(faucet common.Address) *Genesis {
	// Override the default period to the user requested one
	config := *params.SingleNetConfig
	dc := defaultRelayer()
	dc[faucet] = GenesisAccount{Balance: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))}
	// Assemble and return the genesis with the precompiles and faucet pre-funded
	return &Genesis{
		Config:    &config,
		ExtraData: append(append(make([]byte, 32), faucet[:]...), make([]byte, crypto.SignatureLength)...),
		GasLimit:  11500000,
		Alloc:     dc,
	}
}

var relayer = []common.Address{
	common.HexToAddress("0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"),
	common.HexToAddress("0x32CD75ca677e9C37FD989272afA8504CB8F6eB52"),
	common.HexToAddress("0x3e3429F72450A39CE227026E8DdeF331E9973E4d"),
	common.HexToAddress("0x81f02Fd21657DF80783755874a92c996749777Bf"),
	common.HexToAddress("0x84D46B3055454646a419D023f73472561B6cF20F"),
	common.HexToAddress("0x480fB8301D0d357956FB8dB06988d4e5650c5Fc7"),
	common.HexToAddress("0x85273B522f9e17A57CEc59f31f24A49a60C54e17"),
	common.HexToAddress("0xF058b45Ed9A2b781558c0b9ef8C63c79D615c3bB"),
	common.HexToAddress("0x8ee567bE17fB027cBB107Ff70fC02DC475Ce3F3e"),
	common.HexToAddress("0xB5Ac31a4a887e9F773B5Fd0aba3FC0FE95c2a750"),
}
var balance, _ = new(big.Int).SetString("100000000000000000000000000", 10) //100 million
func defaultRelayer() GenesisAlloc {
	dc := make(GenesisAlloc, 10)
	dc[relayer[0]] = GenesisAccount{Balance: balance}
	dc[relayer[1]] = GenesisAccount{Balance: balance}
	dc[relayer[2]] = GenesisAccount{Balance: balance}
	dc[relayer[3]] = GenesisAccount{Balance: balance}
	dc[relayer[4]] = GenesisAccount{Balance: balance}
	dc[relayer[5]] = GenesisAccount{Balance: balance}
	dc[relayer[6]] = GenesisAccount{Balance: balance}
	dc[relayer[7]] = GenesisAccount{Balance: balance}
	dc[relayer[8]] = GenesisAccount{Balance: balance}
	dc[relayer[9]] = GenesisAccount{Balance: balance}
	return dc
}

type RelayerMember struct {
	Coinbase  common.Address `json:"coinbase`
	Publickey []byte
}

func defaultRelayer2() []*RelayerMember {
	key1 := hexutil.MustDecode("0x0499ea9aab0498007f662ca5122e39e7353db3f69b9f1aebd96fcd33bd1a098c4cdb41b97c479d7eecd9d5def59ce7e9f0c6534ccca95811b480e39db37f424215")
	key2 := hexutil.MustDecode("0x041ca456b13aeac67364c1a05effe3ea45479d7aa37337ab62600e32b3875ad3deeadbf4f7d2dbfabf86534daa55de19ac31411757cec9bb6674bd556d4fcee0d6")
	key3 := hexutil.MustDecode("0x04600254af4ce74276f54b4f9df193f2cb72ed76b7341cb144f4d6f1408402dc10719eebdcb947ced9ac6fe9a690e004692db6222de7867cbab712246eb23a50b7")
	key4 := hexutil.MustDecode("0x04290ef09419dc28a367a93a4266c646e379ba4dd0bd2fae7f86277d3d4c330179ee2d70b282de4a5d0d8cc1130c36a88b8fe61baa1726dc41f16e192a3d6af8e4")
	key5 := hexutil.MustDecode("0x040aab611fa2df95180df61677900351659f1feb740e650d7496ae3c887553f13fe304a61a3e9a49bb4e68bd846bf8150a866284ea0b404f0b9ecc0cf23fef6cf1")
	key6 := hexutil.MustDecode("0x0403df1c7ca9dfd7af387f29e0d04bb4440d092f1915ee0d446721f36c70dd4e2ee3651647601ec79052aeeb8f85231d191b521278e7fccb2ff4cbc232f55b0f76")
	key7 := hexutil.MustDecode("0x044fffeb7bc8112b82f97f16a42ff13cd6c0d45400654768e200ddae090d4c61e576925faa0f6aa87cef4be1c16012a5462b1599c75ff620984a0bbe6d434c5e54")
	key8 := hexutil.MustDecode("0x042b635e9e33692cfbb52331811ebc937760aeb2e9cafa1e1caeacdc7aad00aad6b6757c4f38da19a0af98499f084e7d17634b3cd09b431d2cfa64df2662ac58bb")
	key9 := hexutil.MustDecode("0x04068f90a637a5de1830203f02fe9b197b0db9ae6d5bf71cc3668b5f5f9e8da09c432949f0e0e9746579ee6c49346b65ccf275924f1ab0909dadc27a479385b284")
	key10 := hexutil.MustDecode("0x04242019bb0969a3de7adcef74012e76dbb6830b244589551488ca36cbe6e1782c68a1d87d29eac3b29d78b5c45b9f04b31d978957a94e65ddd9208fe9c638abe7")
	cm := []*RelayerMember{
		{Coinbase: relayer[0], Publickey: key1},
		{Coinbase: relayer[1], Publickey: key2},
		{Coinbase: relayer[2], Publickey: key3},
		{Coinbase: relayer[3], Publickey: key4},
		{Coinbase: relayer[4], Publickey: key5},
		{Coinbase: relayer[5], Publickey: key6},
		{Coinbase: relayer[6], Publickey: key7},
		{Coinbase: relayer[7], Publickey: key8},
		{Coinbase: relayer[8], Publickey: key9},
		{Coinbase: relayer[9], Publickey: key10},
	}
	return cm
}

// MainnetGenesisBlock returns the Celo main net genesis block.
func MainnetGenesisBlock() *Genesis {
	mainnetAlloc := &GenesisAlloc{}
	mainnetAlloc.UnmarshalJSON([]byte(mainnetAllocJSON))
	return &Genesis{
		Config:    params.MainnetChainConfig,
		Timestamp: 0x5ea06a00,
		ExtraData: hexutil.MustDecode(mainnetExtraData),
		Alloc:     *mainnetAlloc,
	}
}
