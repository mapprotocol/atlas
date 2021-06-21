package relayer

import (
	"fmt"
	"github.com/abeychain/go-abey/abeydb"
	"github.com/abeychain/go-abey/common"
	"github.com/abeychain/go-abey/core/state"
	"github.com/abeychain/go-abey/core/types"
	"github.com/abeychain/go-abey/core/vm"
	//"github.com/mapprotocol/atlas/core/relayer"
	"github.com/abeychain/go-abey/crypto"
	"github.com/abeychain/go-abey/params"
	"math/big"
	"testing"
)

func TestRegister(t *testing.T) {

	priKey, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(priKey.PublicKey)
	pub := crypto.FromECDSAPub(&priKey.PublicKey)
	value := big.NewInt(1000)

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(abeydb.NewMemDatabase()))
	statedb.GetOrNewStateObject(types.StakingAddress)
	evm := vm.NewEVM(vm.Context{}, statedb, params.TestChainConfig, vm.Config{})

	//log.Info("Staking deposit", "address", from.StringToAbey(), "value", value)
	impawn := vm.NewImpawnImpl()
	impawn.Load(evm.StateDB, types.StakingAddress)

	impawn.InsertSAccount2(1000, 0, from, pub, value, big.NewInt(0), true)
	impawn.Save(evm.StateDB, types.StakingAddress)
}
func TestAppend(t *testing.T) {
	priKey, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(priKey.PublicKey)
	value := big.NewInt(1000)
	var h uint64 = 1000

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(abeydb.NewMemDatabase()))
	statedb.GetOrNewStateObject(types.StakingAddress)
	evm := vm.NewEVM(vm.Context{}, statedb, params.TestChainConfig, vm.Config{})

	//log.Info("Staking deposit", "address", from.StringToAbey(), "value", value)
	impawn := vm.NewImpawnImpl()
	impawn.Load(evm.StateDB, types.StakingAddress)
	impawn.AppendSAAmount(h, from, value)

}
func TestWithdraw(t *testing.T) {
	priKey, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(priKey.PublicKey)
	value := big.NewInt(1000)
	var h uint64 = 1000

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(abeydb.NewMemDatabase()))
	statedb.GetOrNewStateObject(types.StakingAddress)
	evm := vm.NewEVM(vm.Context{}, statedb, params.TestChainConfig, vm.Config{})

	//log.Info("Staking deposit", "address", from.StringToAbey(), "value", value)
	impawn := vm.NewImpawnImpl()
	impawn.Load(evm.StateDB, types.StakingAddress)
	impawn.RedeemSAccount(h, from, value)
}
func TestGetBalance(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(abeydb.NewMemDatabase()))
	statedb.GetOrNewStateObject(types.StakingAddress)
	evm := vm.NewEVM(vm.Context{}, statedb, params.TestChainConfig, vm.Config{})

	priKeyDA, _ := crypto.GenerateKey()
	daAddress := crypto.PubkeyToAddress(priKeyDA.PublicKey)
	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, types.StakingAddress)
	impawn.GetBalance(daAddress)
	fmt.Println(impawn.GetBalance(daAddress))
}
func TestGetRelayer(t *testing.T) {

}
func TestGetPeriodHeight(t *testing.T) {

}
