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
	"github.com/ethereum/go-ethereum/rlp"
	blscrypto "github.com/mapprotocol/atlas/helper/bls"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/mapprotocol/atlas/consensus"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/params"
)

//go:generate gencodec -type Genesis -field-override genesisSpecMarshaling -out gen_genesis.go
//go:generate gencodec -type GenesisAccount -field-override genesisAccountMarshaling -out gen_genesis_account.go

var DBGenesisSupplyKey = []byte("genesis-supply-genesis")
var errGenesisNoConfig = errors.New("genesis has no chain configuration")

var (
	faucetAddr    = common.HexToAddress("0xf675187ff5b76d2430b353f6736aa051253118ee")
	faucetBalance = new(big.Int).Mul(big.NewInt(100000000000), big.NewInt(1e18))
	// private key: baf7c2008a568f91a75caf45b6e849b953513c10b5f3d73270d40c62a4ff5002
	tempAddr = common.HexToAddress("0xd13fe09e7a304709b1c4ed6bd3a2d6c272357bbb")

	DefaultGenesisCfg = &Genesis{
		Config:    params.MainnetChainConfig,
		Nonce:     66,
		ExtraData: hexutil.MustDecode(mainnetExtraData),
		GasLimit:  20000000,
	}
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
	Nonce     math.HexOrDecimal64
	Timestamp math.HexOrDecimal64
	ExtraData hexutil.Bytes
	GasLimit  math.HexOrDecimal64
	GasUsed   math.HexOrDecimal64
	Number    math.HexOrDecimal64
	BaseFee   *math.HexOrDecimal256
	Alloc     map[common.UnprefixedAddress]GenesisAccount
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

	// pre compiled
	consensus.InitHeaderStore(statedb, new(big.Int).SetUint64(g.Number))
	consensus.InitTxVerify(statedb, new(big.Int).SetUint64(g.Number))

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
	gs := genesisPOC2Contract()
	l := len(gs)
	log.Info("Writing Main-net poc2 state", "alloc count", l)
	gs1 := genesisRegisterProxyContract()
	for addr, allc := range gs1 {
		// add genesis contract to allc
		gs[addr] = allc
	}
	b1 := new(big.Int).Mul(big.NewInt(6000000000), big.NewInt(1e18))
	b2 := new(big.Int).Mul(big.NewInt(4000000), big.NewInt(1e18))
	balance0 := new(big.Int).Sub(b1, b2)
	preAddr := common.HexToAddress("ec3e016916ba9f10762e33e03e8556409d096fb4")
	gs[preAddr] = GenesisAccount{Balance: balance0}
	DefaultGenesisCfg.Alloc = gs
	return DefaultGenesisCfg
}

// DefaultTestnetGenesisBlock returns the Ropsten network genesis block.
// owner 0x1c0edab88dbb72b119039c4d14b1663525b3ac15
// validator1  0x1c0edab88dbb72b119039c4d14b1663525b3ac15 password ""
// validator2  0x16fdbcac4d4cc24dca47b9b80f58155a551ca2af password ""
// validator3  0x2dc45799000ab08e60b7441c36fcc74060ccbe11 password ""
// validator4  0x6c5938b49bacde73a8db7c3a7da208846898bff5 password ""
// keystore path  atlas/cmd/testnet_genesis
func DefaultTestnetGenesisBlock() *Genesis {
	gs := genesisPOC2Contract()
	l := len(gs)
	log.Info("Writing Test-net poc2 alloc", "alloc count", l)
	gs1 := genesisTestnetRegisterProxyContract()
	for addr, allc := range gs1 {
		// add genesis contract to allc
		gs[addr] = allc
	}
	balance0 := new(big.Int).Mul(big.NewInt(1000000000), big.NewInt(1e18))
	preAddr := common.HexToAddress("0xd9b31120b910c7d239a03062ab1d9403f30fb7d5")
	gs[preAddr] = GenesisAccount{Balance: balance0}

	return &Genesis{
		Config:    params.TestnetConfig,
		Nonce:     66,
		ExtraData: hexutil.MustDecode(testnetExtraData),
		GasLimit:  16777216,
		Alloc:     gs,
	}
}

// DevnetGenesisBlock returns the 'geth --dev' genesis block.
// owner       0x1c0edab88dbb72b119039c4d14b1663525b3ac15 password ""
// validator1  0x16fdbcac4d4cc24dca47b9b80f58155a551ca2af password ""
// keystore path  atlas/cmd/devnet_genesis
func DevnetGenesisBlock() *Genesis {
	gs := genesisPOC2Contract()
	l := len(gs)
	log.Info("Writing Dev-net poc2 alloc", "alloc count", l)
	gs1 := genesisDevnetRegisterProxyContract()
	for addr, allc := range gs1 {
		// add genesis contract to allc
		gs[addr] = allc
	}
	balance0 := new(big.Int).Mul(big.NewInt(1000000000), big.NewInt(1e18))
	preAddr := common.HexToAddress("0x1c0edab88dbb72b119039c4d14b1663525b3ac15")
	gs[preAddr] = GenesisAccount{Balance: balance0}
	return &Genesis{
		Config: params.DevnetConfig,
		//ExtraData: createDevAlloc(pk, faucet),
		ExtraData: hexutil.MustDecode(devnetExtraData),
		GasLimit:  11500000,
		Alloc:     gs,
	}
}

// DefaultGenesisBlock returns the Ethereum main net genesis block.
func UseForGenesisBlock() *Genesis {
	return DefaultGenesisCfg
}

func defaultRelayer() GenesisAlloc {
	balance, _ := new(big.Int).SetString("100000000000000000000000000", 10) //100 million
	relayer := []common.Address{
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

func defaultRelayerMembers() []*RelayerMember {
	relayer := []common.Address{
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

func createDevAlloc(pk blscrypto.SerializedPublicKey, addr common.Address) []byte {
	ads := make([]common.Address, 0)
	apks := make([]blscrypto.SerializedPublicKey, 0)
	ag1pks := make([]blscrypto.SerializedG1PublicKey, 0)
	ads = append(ads, addr)
	apks = append(apks, pk)
	ist := types.IstanbulExtra{
		AddedValidators:             ads,
		AddedValidatorsPublicKeys:   apks,
		AddedValidatorsG1PublicKeys: ag1pks,
		RemovedValidators:           big.NewInt(0),
		Seal:                        []byte(""),
		AggregatedSeal:              types.IstanbulAggregatedSeal{},
		ParentAggregatedSeal:        types.IstanbulAggregatedSeal{},
	}
	payload, _ := rlp.EncodeToBytes(&ist)
	finalExtra := append(bytes.Repeat([]byte{0x00}, types.IstanbulExtraVanity), payload...)
	return finalExtra
}
