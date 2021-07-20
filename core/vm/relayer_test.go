package vm

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/core/state"
	params2 "github.com/mapprotocol/atlas/params"
	"math/big"
	"strings"
	"testing"
)

var (
	testABI2, _        = abi.JSON(strings.NewReader(params2.RelayerABIJSON))
	priKey, _          = crypto.GenerateKey()
	from               = crypto.PubkeyToAddress(priKey.PublicKey)
	pub                = crypto.FromECDSAPub(&priKey.PublicKey)
	value, _           = new(big.Int).SetString("100000000000000000000000", 10)
	h           uint64 = 0
	fee         uint64 = 100
)

func TestContract(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	statedb, err := state.New(common.Hash{}, state.NewDatabase(db), nil)
	if err != nil {
		t.Fatal(err)
	}
	statedb.GetOrNewStateObject(params2.RelayerAddress)
	evm := NewEVM(BlockContext{}, TxContext{}, statedb, params.TestChainConfig, Config{})

	input, err := relayerABI.Pack("register", pub, new(big.Int).SetUint64(fee), value)
	if err != nil {
		t.Fatal(err)
	}
	data := input[4:]
	ret, _ := register(evm, &Contract{}, data)
	method, _ := relayerABI.Methods["register"]
	output, err := method.Inputs.Unpack(ret)
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println(output)
}

func TestRegister(t *testing.T) {
	//fmt.Println(from, "|", priKey, "|", hex.EncodeToString(pub))
	db := rawdb.NewMemoryDatabase()

	statedb, err := state.New(common.Hash{}, state.NewDatabase(db), nil)
	if err != nil {
		t.Fatal(err)
	}
	statedb.GetOrNewStateObject(params2.RelayerAddress)

	evm := NewEVM(BlockContext{}, TxContext{}, statedb, params.TestChainConfig, Config{})
	//log.Info("Staking deposit", "address", from.StringToAbey(), "Value", Value)
	register := NewRegisterImpl()
	// join selection
	err = register.InsertAccount2(0, from, pub, value, big.NewInt(0), true)
	if err != nil {
		t.Fatal(err)
	}
	//append money
	err = register.AppendAmount(h, from, value)
	if err != nil {
		t.Fatal(err)
	}

	//query money
	fmt.Println(register.GetBalance(from))
	//query relayer
	fmt.Println()
	//query epoch
	fmt.Println()
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

func TestPack(t *testing.T) {
	input, err := relayerABI.Pack("register", pub, new(big.Int).SetUint64(fee), value)
	if err != nil {
		t.Fatal(err)
	}
	data := input[4:]
	fmt.Println("input", data)

	method, _ := relayerABI.Methods["register"]
	output, err := method.Inputs.Unpack(data)
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("output", output)
	fmt.Println("output", output[0].([]byte), output[1].(*big.Int), output[2].(*big.Int))
}

func TestSaveAndLoad(t *testing.T) {
	db := rawdb.NewMemoryDatabase()

	statedb, err := state.New(common.Hash{}, state.NewDatabase(db), nil)
	if err != nil {
		t.Fatal(err)
	}
	statedb.GetOrNewStateObject(params2.RelayerAddress)
	evm := NewEVM(BlockContext{}, TxContext{}, statedb, params.TestChainConfig, Config{})
	//register := NewRegisterImpl()
	register := RegisterImpl{
		curEpochID: 565,
		lastReward: 540,
		accounts:   make(map[uint64]Register),
	}
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

	var temp interface{}
	data, _ := rlp.EncodeToBytes(register)
	rlp.DecodeBytes(data, temp)
	fmt.Println("decode:", temp)
}

func TestInsertAccount(t *testing.T) {
	impl := NewRegisterImpl()
	from := make([]common.Address, 0)
	pub := make([][]byte, 0)
	for i := 0; i < 5; i++ {
		priKey, _ := crypto.GenerateKey()
		addr := crypto.PubkeyToAddress(priKey.PublicKey)
		pk := crypto.FromECDSAPub(&priKey.PublicKey)
		from = append(from, addr)
		pub = append(pub, pk)
	}
	amount := new(big.Int).Mul(big.NewInt(200000), big.NewInt(1e18))

	impl.InsertAccount2(1, from[0], pub[0], amount, big.NewInt(50), true)
	impl.Shift(2, 0)
	fmt.Println("Current Epoch Info:", impl.getCurrentEpochInfo())

	impl.InsertAccount2(10001, from[1], pub[1], amount, big.NewInt(50), true)
	impl.Shift(3, 0)
	fmt.Println("Current Epoch Info:", impl.getCurrentEpochInfo())

	impl.InsertAccount2(20001, from[2], pub[2], amount, big.NewInt(50), true)
	impl.Shift(4, 0)
	fmt.Println("Current Epoch Info:", impl.getCurrentEpochInfo())

	impl.InsertAccount2(30001, from[3], pub[3], amount, big.NewInt(50), true)
	impl.Shift(5, 0)
	fmt.Println("Current Epoch Info:", impl.getCurrentEpochInfo())

	impl.InsertAccount2(40001, from[4], pub[4], amount, big.NewInt(50), true)

	fmt.Println("1. all account:", impl.accounts[1])
	fmt.Println("2. all account:", impl.accounts[2])
	fmt.Println("3. all account:", impl.accounts[3])
	fmt.Println("4. all account:", impl.accounts[4])
	fmt.Println("5. all account:", impl.accounts[5])
}

func TestRegisterDoElections(t *testing.T) {
	impl := NewRegisterImpl()
	db := rawdb.NewMemoryDatabase()
	statedb, err := state.New(common.Hash{}, state.NewDatabase(db), nil)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(" FirstEpochID:", params2.FirstNewEpochID)
	fmt.Println(" epoch 1 ", GetEpochFromID(1))
	fmt.Println(" epoch 2 ", GetEpochFromID(2))
	fmt.Println(" epoch 3 ", GetEpochFromID(3))
	fmt.Println("------------------ready epoch----------------------")
	fmt.Printf("\n")
	var testacc common.Address
	// register accounts in epoch 1
	for i := uint64(0); i < 5; i++ {
		value := big.NewInt(100)
		priKey, _ := crypto.GenerateKey()
		from := crypto.PubkeyToAddress(priKey.PublicKey)
		pub := crypto.FromECDSAPub(&priKey.PublicKey)
		if i%2 == 0 {
			amount := new(big.Int).Mul(big.NewInt(200000), big.NewInt(1e18))
			impl.InsertAccount2(0, from, pub, amount, big.NewInt(50), true)
			//used to redeem
			testacc = from
			addLockedBalance(statedb, from, amount)
			impl.CancelAccount(100, from, big.NewInt(10))
		} else {
			impl.InsertAccount2(0, from, pub, value, big.NewInt(50), true)
		}
	}
	fmt.Println("account number:", len(impl.accounts[1]), " all account:", impl.accounts[1])

	//relayers election
	_, err = impl.DoElections(statedb, 1, 0)
	if err != nil {
		fmt.Println("error : ", err)
	}
	fmt.Println("relayer number:", len(impl.getElections2(1)))
	err = impl.Shift(1, 0)

	fmt.Println("------------------epoch 1----------------------")
	fmt.Printf("\n")
	//register account in epoch 1
	for i := uint64(0); i < 4; i++ {
		value := new(big.Int).Mul(big.NewInt(200000), big.NewInt(1e18))
		priKey, _ := crypto.GenerateKey()
		from := crypto.PubkeyToAddress(priKey.PublicKey)
		pub := crypto.FromECDSAPub(&priKey.PublicKey)
		if i%2 == 0 {
			impl.InsertAccount2(8880+i, from, pub, value, big.NewInt(50), true)
		} else {
			impl.InsertAccount2(9990+i, from, pub, value, big.NewInt(50), true)
		}
	}
	fmt.Println("epoch1 account number:", len(impl.accounts[1]), " all account:", impl.accounts[1])

	//relayer election
	relayer, err := impl.DoElections(statedb, 1, 9900)
	if err != nil {
		fmt.Println("error : ", err)
	}
	fmt.Println("add relayer:", len(relayer), ", qurry relayer:", len(impl.getElections2(1)))
	fmt.Println("epoch2 relayer: ", len(impl.getElections2(2)))
	fmt.Println()
	err = impl.Shift(2, 0)

	fmt.Println("------------------epoch 2----------------------")
	fmt.Printf("\n")
	//register
	for i := uint64(0); i < 2; i++ {
		value := big.NewInt(100)
		priKey, _ := crypto.GenerateKey()
		from := crypto.PubkeyToAddress(priKey.PublicKey)
		pub := crypto.FromECDSAPub(&priKey.PublicKey)
		impl.InsertAccount2(13333+i, from, pub, value, big.NewInt(50), true)
	}
	fmt.Println(" epoch2 account:", len(impl.accounts[2]), impl.accounts[2])
	//relayer election
	relayer, err = impl.DoElections(statedb, 2, 19900)
	if err != nil {
		fmt.Println("error : ", err)
	}
	err = impl.Shift(3, 0)
	fmt.Println(" relayer number:", len(relayer), " getElection 1:", len(impl.getElections2(1)), " getElection 2:", len(impl.getElections2(2)))

	//Redeem test
	fmt.Println()
	fmt.Println("--------------------redeem test-------------------------------")
	c2 := impl.GetAllCancelableAsset(testacc)[testacc]
	fmt.Println("cancelable", c2)
	err = impl.RedeemAccount(21000, testacc, big.NewInt(1))
	if err != nil {
		t.Fatal(err)
	}
	_, _, balance, _, _ := impl.GetBalance(testacc)
	fmt.Println("after redeem", balance)
}

func TestGetBalance(t *testing.T) {
	value := big.NewInt(100)
	priKey, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(priKey.PublicKey)
	pub := crypto.FromECDSAPub(&priKey.PublicKey)

	db := rawdb.NewMemoryDatabase()
	statedb, err := state.New(common.Hash{}, state.NewDatabase(db), nil)
	if err != nil {
		t.Fatal(err)
	}
	statedb.GetOrNewStateObject(params2.RelayerAddress)
	//evm := NewEVM(BlockContext{}, TxContext{}, statedb, params.TestChainConfig, Config{})

	register := NewRegisterImpl()
	register.InsertAccount2(0, from, pub, value, big.NewInt(50), true)
	v1, v2, v3, v4, v5 := register.GetBalance(from)
	fmt.Println("getBalance0", v1, v2, v3, v4, v5)
	//addLockedBalance(statedb, from, value)
	register.CancelAccount(1, from, value)
	v1, v2, v3, v4, v5 = register.GetBalance(from)
	fmt.Println("getBalance1", v1, v2, v3, v4, v5)
	////save data
	//err = register.Save(evm.StateDB, params2.RelayerAddress)
	//if err != nil {
	//	t.Fatal(err)
	//}
	////load data
	//err = register.Load(evm.StateDB, params2.RelayerAddress)
	//if err != nil {
	//	t.Fatal(err)
	//}
}
