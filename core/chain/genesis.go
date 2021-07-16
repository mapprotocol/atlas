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
	"github.com/mapprotocol/atlas/consensus"
	"github.com/mapprotocol/atlas/core/vm"
	params2 "github.com/mapprotocol/atlas/params"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/types"
)

//go:generate gencodec -type Genesis -field-override genesisSpecMarshaling -out gen_genesis.go
//go:generate gencodec -type GenesisAccount -field-override genesisAccountMarshaling -out gen_genesis_account.go

var errGenesisNoConfig = errors.New("genesis has no chain configuration")

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type Genesis struct {
	Config     *params.ChainConfig `json:"config"`
	Nonce      uint64              `json:"nonce"`
	Timestamp  uint64              `json:"timestamp"`
	ExtraData  []byte              `json:"extraData"`
	GasLimit   uint64              `json:"gasLimit"   gencodec:"required"`
	Difficulty *big.Int            `json:"difficulty" gencodec:"required"`
	Mixhash    common.Hash         `json:"mixHash"`
	Coinbase   common.Address      `json:"coinbase"`
	Alloc      GenesisAlloc        `json:"alloc"      gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Number     uint64      `json:"number"`
	GasUsed    uint64      `json:"gasUsed"`
	ParentHash common.Hash `json:"parentHash"`
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
// error is a *params.ConfigCompatError and the new, unwritten config is returned.
//
// The returned chain configuration is never nil.
func SetupGenesisBlock(db ethdb.Database, genesis *Genesis) (*params.ChainConfig, common.Hash, error) {
	return SetupGenesisBlockWithOverride(db, genesis, nil)
}

func SetupGenesisBlockWithOverride(db ethdb.Database, genesis *Genesis, overrideBerlin *big.Int) (*params.ChainConfig, common.Hash, error) {
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
	if overrideBerlin != nil {
		newcfg.BerlinBlock = overrideBerlin
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

func (g *Genesis) configOrDefault(ghash common.Hash) *params.ChainConfig {
	switch {
	case g != nil:
		return g.Config
	case ghash == params2.MainnetGenesisHash:
		return params2.MainnetChainConfig
	case ghash == params2.TestnetGenesisHash:
		return params2.TestnetConfig
	case ghash == params2.DevnetGenesisHash:
		return params2.DevnetConfig
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
	consensus.OnceInitRegisterState(statedb, new(big.Int).SetUint64(g.Number))
	consensus.InitHeaderStore(statedb, new(big.Int).SetUint64(g.Number))
	register := vm.NewRegisterImpl()
	hh := g.Number
	if hh != 0 {
		hh = hh - 1
	}
	relayer := defaultRelayer2()
	for _, member := range relayer {
		var err error
		err = register.InsertAccount2(hh, member.Coinbase, member.Publickey, params2.ElectionMinLimitForRegister, big.NewInt(100), true)
		if err != nil {
			log.Error("ToBlock InsertSAccount", "error", err)
		} else {
			vm.GenesisAddLockedBalance(statedb, member.Coinbase, params2.ElectionMinLimitForRegister)
		}
	}
	_, err := register.DoElections(statedb, 1, 0)
	if err != nil {
		log.Error("ToBlock DoElections", "error", err)
	}
	err = register.Shift(1, 0)
	if err != nil {
		log.Error("ToBlock Shift", "error", err)
	}
	err = register.Save(statedb, params2.RelayerAddress)
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
		Difficulty: g.Difficulty,
		MixDigest:  g.Mixhash,
		Coinbase:   g.Coinbase,
		Root:       root,
	}
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

// Commit writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func (g *Genesis) Commit(db ethdb.Database) (*types.Block, error) {
	block := g.ToBlock(db)
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
	rawdb.WriteBlock(db, block)
	rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), nil)
	rawdb.WriteCanonicalHash(db, block.Hash(), block.NumberU64())
	rawdb.WriteHeadBlockHash(db, block.Hash())
	rawdb.WriteHeadFastBlockHash(db, block.Hash())
	rawdb.WriteHeadHeaderHash(db, block.Hash())
	rawdb.WriteChainConfig(db, block.Hash(), config)
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
	g := Genesis{Alloc: GenesisAlloc{addr: {Balance: balance}}}
	return g.MustCommit(db)
}

// DefaultGenesisBlock returns the Ethereum main net genesis block.
func DefaultGenesisBlock() *Genesis {
	return &Genesis{
		Config:     params2.MainnetChainConfig,
		Nonce:      66,
		ExtraData:  hexutil.MustDecode("0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa"),
		GasLimit:   50000000,
		Difficulty: big.NewInt(1200000),
		Alloc:      defaultRelayer(),
	}
}

// DefaultTestnetGenesisBlock returns the Ropsten network genesis block.
func DefaultTestnetGenesisBlock() *Genesis {
	return &Genesis{
		Config:     params2.TestnetConfig,
		Nonce:      66,
		ExtraData:  hexutil.MustDecode("0x3535353535353535353535353535353535353535353535353535353535353535"),
		GasLimit:   16777216,
		Difficulty: big.NewInt(1000000),
		Alloc:      defaultRelayer(),
	}
}

// DevnetGenesisBlock returns the 'geth --dev' genesis block.
func DevnetGenesisBlock() *Genesis {
	return &Genesis{
		Config:     params2.DevnetConfig,
		ExtraData:  []byte{1, 2, 3},
		GasLimit:   11500000,
		Difficulty: big.NewInt(1),
		Alloc:      defaultRelayer(),
	}
}

// SingleGenesisBlock returns the 'geth --dev' genesis block.
func SingleGenesisBlock(faucet common.Address) *Genesis {
	// Override the default period to the user requested one
	config := *params2.SingleNetCfg
	//config.Ethash.Period = period
	dc := defaultRelayer()
	dc[faucet] = GenesisAccount{Balance: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))}
	// Assemble and return the genesis with the precompiles and faucet pre-funded
	return &Genesis{
		Config:     &config,
		ExtraData:  append(append(make([]byte, 32), faucet[:]...), make([]byte, crypto.SignatureLength)...),
		GasLimit:   11500000,
		Difficulty: big.NewInt(1),
		Alloc:      dc,
	}
}
func decodePrealloc(data string) GenesisAlloc {
	var p []struct{ Addr, Balance *big.Int }
	if err := rlp.NewStream(strings.NewReader(data), 0).Decode(&p); err != nil {
		panic(err)
	}
	ga := make(GenesisAlloc, len(p))
	for _, account := range p {
		ga[common.BigToAddress(account.Addr)] = GenesisAccount{Balance: account.Balance}
	}
	return ga
}

var relayer []common.Address = []common.Address{
	common.HexToAddress("0x3e3429F72450A39CE227026E8DdeF331E9973E4d"),
	common.HexToAddress("0x1Cfe2A1D7B9CBfce14d06bAFfa338b2465216255"),
	common.HexToAddress("0x1275db492b0d02855a38Bd3Cdf73C92137CD1691"),
	common.HexToAddress("0xF11A544F74a2F4Faa2AF8Aa38F9388A4Cc2F3ACC"),
	common.HexToAddress("0xc30E75016F5a82EE6f0A7989F9DCD5F030c83B3A"),
	common.HexToAddress("0x1e2E48Fa3cC3417474EC264DE53D6305109af1b9"),
	common.HexToAddress("0x7AdC129C637f93C9392c59e9C4d406FDC28aAB43"),
	common.HexToAddress("0xf9621AEa3d6492d43dC96b5472C4680021793109"),
	common.HexToAddress("0x5552FAC84cD38DEdAf8c80a195591CBCED1f4A8D"),
	common.HexToAddress("0xBa9779b7173099354630BD87b5b972441E3605bd"),
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
	key1 := hexutil.MustDecode("0x04600254af4ce74276f54b4f9df193f2cb72ed76b7341cb144f4d6f1408402dc10719eebdcb947ced9ac6fe9a690e004692db6222de7867cbab712246eb23a50b7")
	key2 := hexutil.MustDecode("0x04c042a428a7df304ac7ea81c1555da49310cebb079a905c8256080e8234af804dad4ad9995771f96fba8182b117f62d2f1a6643e27f5f272c293a8301b6a84442")
	key3 := hexutil.MustDecode("0x04dc1da011509b6ea17527550cc480f6eb076a225da2bcc87ec7a24669375f229945d76e4f9dbb4bd26c72392050a18c3922bd7ef38c04e018192b253ef4fc9dcb")
	key4 := hexutil.MustDecode("0x04952af3d04c0b0ba3d16eea8ca0ab6529f5c6e2d08f4aa954ae2296d4ded9f04c8a9e1d52be72e6cebb86b4524645fafac04ac8633c4b33638254b2eb64a89c6a")
	key5 := hexutil.MustDecode("0x04290cdc7fe53df0f93d43264302337751a58bcf67ee56799abea93b0a6205be8b3c8f1c9dac281f4d759475076596d30aa360d0c3b160dc28ea300b7e4925fb32")
	key6 := hexutil.MustDecode("0x04427e32084f7565970d74a3df317b68de59e62f28b86700c8a5e3ae83a781ec163c4c83544bd8f88b8d70c4d71f2827b7b279bfc25481453dd35533cf234b2dfe")
	key7 := hexutil.MustDecode("0x04dd9980aac0edead2de77cc6cde74875c14ac21d95a1cb49d36b810246b50420f1dc7c19f5296d739fcfceb454a18f250fa7802280f5298e5e2b2a591faa15cf9")
	key8 := hexutil.MustDecode("0x04039dd0fb3869e7d2a1eeb95c9a6475771883614b289c604bf6fef2e1e9dd57340d888f59db0129d250394909d4a3b041bd66e6b83f345b38a397fdeb036b3e1c")
	key9 := hexutil.MustDecode("0x042ec25823b375f655117d1a7003f9526e9adc0d6d50150812e0408fbfb3256810c912d7cd7e5441bc5e54ac143fb6274ac496548e1a2aaaf370e8aa8b5b1ced4d")
	key10 := hexutil.MustDecode("0x043e3014c29e42015fe891ca3e97e5fb05961beca9e349b821c6738eadd17d9b784295638e26c1d7ca71beb8703ec8cf944c67f3835bf5119f78192b535ac6a5e0")
	cm := []*RelayerMember{
		&RelayerMember{Coinbase: relayer[0], Publickey: key1},
		&RelayerMember{Coinbase: relayer[1], Publickey: key2},
		&RelayerMember{Coinbase: relayer[2], Publickey: key3},
		&RelayerMember{Coinbase: relayer[3], Publickey: key4},
		&RelayerMember{Coinbase: relayer[4], Publickey: key5},
		&RelayerMember{Coinbase: relayer[5], Publickey: key6},
		&RelayerMember{Coinbase: relayer[6], Publickey: key7},
		&RelayerMember{Coinbase: relayer[7], Publickey: key8},
		&RelayerMember{Coinbase: relayer[8], Publickey: key9},
		&RelayerMember{Coinbase: relayer[9], Publickey: key10},
	}
	return cm
}
