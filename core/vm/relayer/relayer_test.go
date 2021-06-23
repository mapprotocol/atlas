package relayer

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/mapprotocol/atlas/atlasdb"
	//"github.com/mapprotocol/atlas/params"
	"github.com/mapprotocol/atlas/core/vm/state"
	"math/big"
	"testing"
)

func TestRegister(t *testing.T) {
	priKey, _ := crypto.GenerateKey()
	//from := crypto.PubkeyToAddress(priKey.PublicKey)
	pub := crypto.FromECDSAPub(&priKey.PublicKey)
	value := big.NewInt(1000)

	statedb, _ := state.New(common.Hash{'1'}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)

	evm := vm.NewEVM(vm.BlockContext{},vm.TxContext{}, statedb , params.TestChainConfig, vm.Config{})
	//log.Info("Staking deposit", "address", from.StringToAbey(), "value", value)
	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, common.Address{'1'})

	impawn.InsertSAccount2(1000, 0, common.Address{}, pub, value, big.NewInt(0), true)
	impawn.Save(*statedb, common.Address{'1'})
}
func TestAppend(t *testing.T) {
	value := big.NewInt(1000)
	var h uint64 = 1000

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)
	//evm := vm.NewEVM(vm.BlockContext{},vm.TxContext{}, vm.StateDB() , params.TestChainConfig, vm.Config{})
	//log.Info("Staking deposit", "address", from.StringToAbey(), "value", value)
	impawn := NewImpawnImpl()
	impawn.Load(*statedb, common.Address{'1'})
	impawn.AppendSAAmount(h, common.Address{'1'}, value)

}
func TestWithdraw(t *testing.T) {
	value := big.NewInt(1000)
	var h uint64 = 1000

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)
	//evm := vm.NewEVM(vm.Context{}, statedb, params.TestChainConfig, vm.Config{})

	//log.Info("Staking deposit", "address", from.StringToAbey(), "value", value)
	impawn := NewImpawnImpl()
	impawn.Load(*statedb, common.Address{'1'})
	impawn.RedeemSAccount(h, common.Address{'1'}, value)
}
func TestGetBalance(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)
	//statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	//statedb.GetOrNewStateObject(StakingAddress)
	//evm := vm.NewEVM(vm.Context{}, statedb, params.TestChainConfig, vm.Config{})

	//priKeyDA, _ := crypto.GenerateKey()
	//daAddress := crypto.PubkeyToAddress(priKeyDA.PublicKey)
	impawn := NewImpawnImpl()
	impawn.Load(*statedb, common.Address{'1'})
	//impawn.GetToken(daAddress)
	//impawn.GetBalance(daAddress)
	fmt.Println(impawn.GetBalance(common.Address{'1'}))
}
func TestGetRelayer(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)
	//statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	//statedb.GetOrNewStateObject(StakingAddress)
	//evm := vm.NewEVM(vm.Context{}, statedb, params.TestChainConfig, vm.Config{})

	impawn := NewImpawnImpl()
	impawn.Load(*statedb, common.Address{'1'})
	impawn.GetAllStakingAccount()
	impawn.GetCurrentEpochInfo()
}
func TestGetPeriodHeight(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)
	//statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	//statedb.GetOrNewStateObject(StakingAddress)
	//evm := vm.NewEVM(vm.Context{}, statedb, params.TestChainConfig, vm.Config{})
	impawn := NewImpawnImpl()
	impawn.Load(*statedb, common.Address{'1'})
	//[]*types.EpochIDInfo
	info,h := impawn.GetCurrentEpochInfo()
	isRelayer,_ := impawn.GetStakingAccount(h,common.Address{'1'})

	for _,v := range info{
		if h == v.EpochID {
			fmt.Println("查询编号，开始高度，结束高度: ",v.EpochID,v.BeginHeight,v.EndHeight,isRelayer)
		}
	}

}

