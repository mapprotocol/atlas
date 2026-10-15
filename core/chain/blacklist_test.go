package chain

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethparams "github.com/ethereum/go-ethereum/params"

	"github.com/mapprotocol/atlas/core"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/core/vm/vmcontext"
	"github.com/mapprotocol/atlas/params"
)

var blacklistedAddr = common.HexToAddress("0x8d0B87a76C8D38a96dEdA851e737461eF9d03D92")

// Tests that the tx pool rejects transactions sent to a blacklisted account.
func TestTxPoolRejectsBlacklisted(t *testing.T) {
	// NewTxPool requires a full EVM runner, so exercise validateTx on a
	// minimal pool. The blacklist check runs before any state access.
	pool := &TxPool{
		chainconfig:   params.MainnetChainConfig,
		signer:        types.HomesteadSigner{},
		currentMaxGas: 10000000,
		gasPrice:      big.NewInt(1),
	}
	key, _ := crypto.GenerateKey()

	for _, local := range []bool{false, true} {
		tx, _ := types.SignTx(types.NewTransaction(0, blacklistedAddr, big.NewInt(100), 100000, big.NewInt(1), nil), types.HomesteadSigner{}, key)
		if err := pool.validateTx(tx, local); !errors.Is(err, core.ErrBlacklistedAccount) {
			t.Errorf("local=%v: expected %v, got %v", local, core.ErrBlacklistedAccount, err)
		}
	}
}

// Tests that messages from or to a blacklisted account are rejected by the
// state transition only after the Blacklist fork.
func TestStateTransitionBlacklist(t *testing.T) {
	config := *params.TestChainConfig
	config.ChainID = new(big.Int).SetUint64(params.MainNetChainID)
	config.BlacklistBlock = big.NewInt(10)

	key, _ := crypto.GenerateKey()
	sender := crypto.PubkeyToAddress(key.PublicKey)
	other := common.HexToAddress("0x1111111111111111111111111111111111111111")

	apply := func(number int64, from common.Address, to *common.Address) error {
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		statedb.SetBalance(from, big.NewInt(ethparams.Ether))
		blockCtx := vm.BlockContext{
			CanTransfer: CanTransfer,
			Transfer:    vmcontext.TobinTransfer,
			BlockNumber: big.NewInt(number),
			Time:        big.NewInt(0),
			Difficulty:  big.NewInt(0),
			BaseFee:     big.NewInt(0),
			GasLimit:    10000000,
		}
		evm := vm.NewEVM(blockCtx, vm.TxContext{Origin: from, GasPrice: big.NewInt(0)}, statedb, &config, vm.Config{})
		msg := types.NewMessage(from, to, 0, big.NewInt(1), ethparams.TxGas, big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, nil, false)
		_, err := ApplyMessage(evm, msg, new(core.GasPool).AddGas(10000000))
		return err
	}

	for i, tt := range []struct {
		number int64
		from   common.Address
		to     *common.Address
		want   error
	}{
		{9, sender, &blacklistedAddr, nil},                         // before fork
		{9, blacklistedAddr, &other, nil},                          // before fork
		{10, sender, &blacklistedAddr, core.ErrBlacklistedAccount}, // to, at fork
		{10, blacklistedAddr, &other, core.ErrBlacklistedAccount},  // from, at fork
		{11, blacklistedAddr, nil, core.ErrBlacklistedAccount},     // contract creation, after fork
		{11, sender, &other, nil},                                  // unrelated, after fork
	} {
		if err := apply(tt.number, tt.from, tt.to); !errors.Is(err, tt.want) {
			t.Errorf("test %d: have %v, want %v", i, err, tt.want)
		}
	}
}
