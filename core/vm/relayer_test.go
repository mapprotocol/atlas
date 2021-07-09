package vm

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/mapprotocol/atlas/core/state"
	params2 "github.com/mapprotocol/atlas/params"
	"math/big"
	"testing"
)

var (
	priKey, _        = crypto.GenerateKey()
	from             = crypto.PubkeyToAddress(priKey.PublicKey)
	pub              = crypto.FromECDSAPub(&priKey.PublicKey)
	value            = big.NewInt(1000)
	h         uint64 = 0
)

func TestRegister(t *testing.T) {
	fmt.Println(from, "|", priKey, "|", hex.EncodeToString(pub))
	db := rawdb.NewMemoryDatabase()

	statedb, err := state.New(common.Hash{}, state.NewDatabase(db), nil)
	if err != nil {
		t.Fatal(err)
	}
	statedb.GetOrNewStateObject(params2.RelayerAddress)

	evm := NewEVM(BlockContext{}, TxContext{}, statedb, params.TestChainConfig, Config{})
	//log.Info("Staking deposit", "address", from.StringToAbey(), "value", value)
	register := NewRegisterImpl()
	// join selection
	err = register.InsertSAccount2(0, 0, from, pub, value, big.NewInt(0), true)
	if err != nil {
		t.Fatal(err)
	}
	//append money
	err = register.AppendSAAmount(h, from, value)
	if err != nil {
		t.Fatal(err)
	}
	//Redeem money
	err = register.RedeemSAccount(h+1, from, value)
	if err != nil {
		t.Fatal(err)
	}
	//query money
	fmt.Println(register.GetBalance(from))
	//query relayer
	//register.GetAllRegisterAccount()
	//register.GetCurrentEpochInfo()
	//query epoch
	info, h := register.GetCurrentEpochInfo()
	for _, v := range info {
		if h == v.EpochID {
			fmt.Println(v.EpochID, v.BeginHeight, v.EndHeight)
		}
	}
	isRelayer, _ := register.GetRegisterAccount(h, from)
	fmt.Println(isRelayer)
	//save data
	err = register.Save(evm.StateDB, params2.RelayerAddress)
	if err != nil {
		t.Fatal(err)
	}
	//load data
	err = register.Load(evm.StateDB, params2.RelayerAddress)
	if err != nil {
		t.Fatal(err)
	}
}
