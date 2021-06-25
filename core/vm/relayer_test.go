package vm

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/mapprotocol/atlas/atlasdb"
	"github.com/mapprotocol/atlas/core/vm/state"
	"math/big"
	"testing"
)

func TestRegister(t *testing.T) {
	priKey, _ := crypto.GenerateKey()
	from := common.BytesToAddress([]byte("truestaking"))//crypto.PubkeyToAddress(priKey.PublicKey)
	pub := crypto.FromECDSAPub(&priKey.PublicKey)
	value := big.NewInt(1000)

	tr, err := trie.NewSecure(from.Hash(), atlasdb.NewMemoryDatabase())
	if err != nil {
		fmt.Println("tr",tr)
		t.Fatal(err)
	}

	testdb := state.NewDatabase(atlasdb.NewMemoryDatabase())
	te, err := testdb.OpenTrie(from.Hash())
	if err != nil {
		fmt.Println("te",te)
		t.Fatal(err)
	}

	statedb, err := state.New(from.Hash(), state.NewDatabase(atlasdb.NewMemoryDatabase()))
	fmt.Println("!!!!!!!!!!!!",statedb)
	fmt.Println("!!!!!!!!!!!!!@@",state.NewDatabase(atlasdb.NewMemoryDatabase()))
	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@",atlasdb.NewMemoryDatabase())
	if err != nil {
		t.Fatal(err)
	}
	statedb.GetOrNewStateObject(StakingAddress)

	evm := NewEVM(BlockContext{}, TxContext{},statedb , params.TestChainConfig, Config{})
	//log.Info("Staking deposit", "address", from.StringToAbey(), "value", value)
	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, common.Address{'1'})

	impawn.InsertSAccount2(1000, 0, common.Address{}, pub, value, big.NewInt(0), true)
	impawn.Save(evm.StateDB, common.Address{'1'})
}
func TestAppend(t *testing.T) {
	value := big.NewInt(1000)
	var h uint64 = 1000

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)
	evm := NewEVM(BlockContext{}, TxContext{}, statedb , params.TestChainConfig, Config{})
	//evm := vm.NewEVM(vm.BlockContext{},vm.TxContext{}, vm.StateDB() , params.TestChainConfig, vm.Config{})
	//log.Info("Staking deposit", "address", from.StringToAbey(), "value", value)
	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, common.Address{'1'})
	impawn.AppendSAAmount(h, common.Address{'1'}, value)

}
func TestWithdraw(t *testing.T) {
	value := big.NewInt(1000)
	var h uint64 = 1000

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)
	evm := NewEVM(BlockContext{}, TxContext{}, statedb , params.TestChainConfig, Config{})
	//evm := vm.NewEVM(vm.Context{}, statedb, params.TestChainConfig, vm.Config{})

	//log.Info("Staking deposit", "address", from.StringToAbey(), "value", value)
	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, common.Address{'1'})
	impawn.RedeemSAccount(h, common.Address{'1'}, value)
}
func TestGetBalance(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)
	evm := NewEVM(BlockContext{}, TxContext{}, statedb , params.TestChainConfig, Config{})
	//statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	//statedb.GetOrNewStateObject(StakingAddress)
	//evm := vm.NewEVM(vm.Context{}, statedb, params.TestChainConfig, vm.Config{})

	//priKeyDA, _ := crypto.GenerateKey()
	//daAddress := crypto.PubkeyToAddress(priKeyDA.PublicKey)
	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, common.Address{'1'})
	//impawn.GetToken(daAddress)
	//impawn.GetBalance(daAddress)
	fmt.Println(impawn.GetBalance(common.Address{'1'}))
}
func TestGetRelayer(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)
	evm := NewEVM(BlockContext{}, TxContext{}, statedb , params.TestChainConfig, Config{})
	//statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	//statedb.GetOrNewStateObject(StakingAddress)
	//evm := vm.NewEVM(vm.Context{}, statedb, params.TestChainConfig, vm.Config{})

	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, common.Address{'1'})
	impawn.GetAllStakingAccount()
	impawn.GetCurrentEpochInfo()
}
func TestGetPeriodHeight(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	statedb.GetOrNewStateObject(StakingAddress)
	evm := NewEVM(BlockContext{}, TxContext{}, statedb , params.TestChainConfig, Config{})
	//statedb, _ := state.New(common.Hash{}, state.NewDatabase(atlasdb.NewMemoryDatabase()))
	//statedb.GetOrNewStateObject(StakingAddress)
	//evm := vm.NewEVM(vm.Context{}, statedb, params.TestChainConfig, vm.Config{})
	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, common.Address{'1'})
	//[]*types.EpochIDInfo
	info,h := impawn.GetCurrentEpochInfo()
	isRelayer,_ := impawn.GetStakingAccount(h,common.Address{'1'})

	for _,v := range info{
		if h == v.EpochID {
			fmt.Println("查询编号，开始高度，结束高度: ",v.EpochID,v.BeginHeight,v.EndHeight,isRelayer)
		}
	}

}

