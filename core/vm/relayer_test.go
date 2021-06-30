package vm

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/mapprotocol/atlas/core/state"
	"math/big"
	"testing"
)

//type precompiledTest struct {
//	Input, Expected string
//	Name            string
//	NoBenchmark     bool // Benchmark primarily the worst-cases
//}

//type dummyContractRef struct {
//	calledForEach bool
//}
//func (dummyContractRef) ReturnGas(*big.Int)          {}
//func (dummyContractRef) Address() common.Address     { return common.Address{} }
//func (dummyContractRef) Value() *big.Int             { return new(big.Int) }
//func (dummyContractRef) SetCode(common.Hash, []byte) {}
//func (d *dummyContractRef) ForEachStorage(callback func(key, value common.Hash) bool) {
//	d.calledForEach = true
//}
//func (d *dummyContractRef) SubBalance(amount *big.Int) {}
//func (d *dummyContractRef) AddBalance(amount *big.Int) {}
//func (d *dummyContractRef) SetBalance(*big.Int)        {}
//func (d *dummyContractRef) SetNonce(uint64)            {}
//func (d *dummyContractRef) Balance() *big.Int          { return new(big.Int) }

//var allPrecompiles = PrecompiledContractsYoloPos

//func testPrecompiled(addr string, test precompiledTest, t *testing.T) {
//	p := allPrecompiles[common.HexToAddress(addr)]
//	in := common.Hex2Bytes(test.Input)
//
//	from := common.BytesToAddress([]byte("truestaking"))
//	statedb, err := state.New(from.Hash(), state.NewDatabase(rawdb.NewMemoryDatabase()))
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	env := NewEVM(BlockContext{}, TxContext{}, statedb,params.TestChainConfig, Config{})
//	contract := NewContract(&dummyContractRef{}, &dummyContractRef{}, new(big.Int), 0)
//	gas := p.RequiredGas(env, in)
//	t.Run(fmt.Sprintf("%s-Gas=%d", test.Name, gas), func(t *testing.T) {
//		if res, _, err := RunPrecompiledContract(env, p, in, gas, contract); err != nil {
//			t.Error(err)
//		} else if common.Bytes2Hex(res) != test.Expected {
//			t.Errorf("Expected %v, got %v", test.Expected, common.Bytes2Hex(res))
//		}
//		// Verify that the precompile did not touch the input buffer
//		exp := common.Hex2Bytes(test.Input)
//		if !bytes.Equal(in, exp) {
//			t.Errorf("Precompiled %v modified input data", addr)
//		}
//	})
//}

func TestRegister(t *testing.T) {
	priKey, _ := crypto.GenerateKey()
	//from := common.BytesToAddress([]byte("truestaking"))//crypto.PubkeyToAddress(priKey.PublicKey)
	pub := crypto.FromECDSAPub(&priKey.PublicKey)
	value := big.NewInt(1000)
	db := rawdb.NewMemoryDatabase()

	statedb, err := state.New(common.Hash{}, state.NewDatabase(db),nil)
	fmt.Println("!!!!!!!!!!!!",statedb)
	fmt.Println("!!!!!!!!!!!!!@@",state.NewDatabase(db))
	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@",db)
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
	db := rawdb.NewMemoryDatabase()
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db),nil)
	statedb.GetOrNewStateObject(StakingAddress)
	evm := NewEVM(BlockContext{}, TxContext{}, statedb , params.TestChainConfig, Config{})
	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, common.Address{'1'})
	impawn.AppendSAAmount(h, common.Address{'1'}, value)

}
func TestWithdraw(t *testing.T) {
	value := big.NewInt(1000)
	var h uint64 = 1000
	db := rawdb.NewMemoryDatabase()
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db),nil)
	statedb.GetOrNewStateObject(StakingAddress)
	evm := NewEVM(BlockContext{}, TxContext{}, statedb , params.TestChainConfig, Config{})

	//log.Info("Staking deposit", "address", from.StringToAbey(), "value", value)
	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, common.Address{'1'})
	impawn.RedeemSAccount(h, common.Address{'1'}, value)
}
func TestGetBalance(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db),nil)
	statedb.GetOrNewStateObject(StakingAddress)
	evm := NewEVM(BlockContext{}, TxContext{}, statedb , params.TestChainConfig, Config{})

	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, common.Address{'1'})
	fmt.Println(impawn.GetBalance(common.Address{'1'}))
}
func TestGetRelayer(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()),nil)
	statedb.GetOrNewStateObject(StakingAddress)
	evm := NewEVM(BlockContext{}, TxContext{}, statedb , params.TestChainConfig, Config{})

	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, common.Address{'1'})
	impawn.GetAllStakingAccount()
	impawn.GetCurrentEpochInfo()
}
func TestGetPeriodHeight(t *testing.T) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()),nil)
	statedb.GetOrNewStateObject(StakingAddress)
	evm := NewEVM(BlockContext{}, TxContext{}, statedb , params.TestChainConfig, Config{})
	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, common.Address{'1'})
	info,h := impawn.GetCurrentEpochInfo()
	isRelayer,_ := impawn.GetStakingAccount(h,common.Address{'1'})

	for _,v := range info{
		if h == v.EpochID {
			fmt.Println("查询编号，开始高度，结束高度: ",v.EpochID,v.BeginHeight,v.EndHeight,isRelayer)
		}
	}

}

