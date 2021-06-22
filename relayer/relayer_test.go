package relayer

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/atlasdb"
	//"github.com/ethereum/go-ethereum/core/state"
	gcommon "github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/relayer/state"
	//"github.com/abeychain/go-abey/core/state"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"testing"
)

func TestRegister(t *testing.T) {

	//StakingAddress:=common.BytesToAddress([]byte("truestaking"))
	priKey, _ := crypto.GenerateKey()
	//from := crypto.PubkeyToAddress(priKey.PublicKey)
	pub := crypto.FromECDSAPub(&priKey.PublicKey)
	value := big.NewInt(1000)

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)
	//log.Info("Staking deposit", "address", from.StringToAbey(), "value", value)
	impawn := NewImpawnImpl()
	impawn.Load(*statedb, gcommon.Address{})

	impawn.InsertSAccount2(1000, 0, gcommon.Address{}, pub, value, big.NewInt(0), true)
	impawn.Save(*statedb, gcommon.Address{})
}
func TestAppend(t *testing.T) {
	//priKey, _ := crypto.GenerateKey()
	//from := crypto.PubkeyToAddress(priKey.PublicKey)
	value := big.NewInt(1000)
	var h uint64 = 1000

	//statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	//statedb.GetOrNewStateObject(StakingAddress)
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)
	//evm := vm.NewEVM(nil,nil,, statedb, params.TestChainConfig, vm.Config{})
	//log.Info("Staking deposit", "address", from.StringToAbey(), "value", value)
	impawn := NewImpawnImpl()
	impawn.Load(*statedb, gcommon.Address{'1'})
	impawn.AppendSAAmount(h, gcommon.Address{'1'}, value)

}
func TestWithdraw(t *testing.T) {
	//priKey, _ := crypto.GenerateKey()
	//from := crypto.PubkeyToAddress(priKey.PublicKey)
	//Delegate := crypto.PubkeyToAddress(priKey.PublicKey)
	value := big.NewInt(1000)
	var h uint64 = 1000

	//statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	//statedb.GetOrNewStateObject(StakingAddress)
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)
	//evm := vm.NewEVM(vm.Context{}, statedb, params.TestChainConfig, vm.Config{})

	//log.Info("Staking deposit", "address", from.StringToAbey(), "value", value)
	impawn := NewImpawnImpl()
	impawn.Load(*statedb, gcommon.Address{'1'})
	impawn.RedeemSAccount(h, gcommon.Address{'1'}, value)
	//impawn.RedeemDAccount(h,Stake,Delegate,value)
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
	impawn.Load(*statedb, gcommon.Address{'1'})
	//impawn.GetToken(daAddress)
	//impawn.GetBalance(daAddress)
	fmt.Println(impawn.GetBalance(gcommon.Address{'1'}))
}
func TestGetRelayer(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)
	//statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	//statedb.GetOrNewStateObject(StakingAddress)
	//evm := vm.NewEVM(vm.Context{}, statedb, params.TestChainConfig, vm.Config{})

	impawn := NewImpawnImpl()
	impawn.Load(*statedb, gcommon.Address{'1'})
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
	impawn.Load(*statedb, gcommon.Address{'1'})
	//[]*types.EpochIDInfo
	info,h := impawn.GetCurrentEpochInfo()
	isRelayer,_ := impawn.GetStakingAccount(h,gcommon.Address{'1'})

	for _,v := range info{
		if h == v.EpochID {
			fmt.Println("查询编号，开始高度，结束高度: ",v.EpochID,v.BeginHeight,v.EndHeight,isRelayer)
		}
	}

}

