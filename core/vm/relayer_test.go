package vm

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/chains/ethereum"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/params"
)

var (
	testABI2, _        = abi.JSON(strings.NewReader(params.RelayerABIJSON))
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
	statedb.GetOrNewStateObject(params.RelayerAddress)
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
	statedb.GetOrNewStateObject(params.RelayerAddress)

	evm := NewEVM(BlockContext{}, TxContext{}, statedb, params.TestChainConfig, Config{})
	register := NewRegisterImpl()
	// join selection
	err = register.InsertAccount2(0, from, value)
	if err != nil {
		t.Fatal(err)
	}
	//append money
	err = register.AppendAmount(h, from, value)
	if err != nil {
		t.Fatal(err)
	}

	//query money
	fmt.Println(register.GetBalance(from, h))
	//query relayer
	fmt.Println()
	//query epoch
	fmt.Println()
	//save data
	err = register.Save(evm.StateDB, params.RelayerAddress)
	if err != nil {
		t.Fatal(err)
	}
	//load data
	err = register.Load(evm.StateDB, params.RelayerAddress)
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
	statedb.GetOrNewStateObject(params.RelayerAddress)
	//evm := NewEVM(BlockContext{}, TxContext{}, statedb, params.TestChainConfig, Config{})
	//register := NewRegisterImpl()
	register := RegisterImpl{
		curEpochID: 565,
		lastReward: 540,
		accounts:   make(map[uint64]Register),
	}
	//save data
	err = register.Save(statedb, params.RelayerAddress)
	if err != nil {
		t.Fatal(err)
	}
	//load data
	err = register.Load(statedb, params.RelayerAddress)
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

	impl.InsertAccount2(1, from[0], amount)
	impl.Shift(2, 0)
	fmt.Println("Current Epoch Info:", impl.getCurrentEpochInfo())

	impl.InsertAccount2(10001, from[1], amount)
	impl.Shift(3, 0)
	fmt.Println("Current Epoch Info:", impl.getCurrentEpochInfo())

	impl.InsertAccount2(20001, from[2], amount)
	impl.Shift(4, 0)
	fmt.Println("Current Epoch Info:", impl.getCurrentEpochInfo())

	impl.InsertAccount2(30001, from[3], amount)
	impl.Shift(5, 0)
	fmt.Println("Current Epoch Info:", impl.getCurrentEpochInfo())

	impl.InsertAccount2(40001, from[4], amount)

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
	fmt.Println(" FirstEpochID:", params.FirstNewEpochID)
	fmt.Println(" epoch 1 ", GetEpochFromID(1))
	fmt.Println(" epoch 2 ", GetEpochFromID(2))
	fmt.Println(" epoch 3 ", GetEpochFromID(3))
	fmt.Println("------------------ready epoch----------------------")
	fmt.Printf("\n")
	//var testacc common.Address
	// register accounts in epoch 1
	for i := uint64(0); i < 5; i++ {
		value := big.NewInt(100)
		priKey, _ := crypto.GenerateKey()
		from := crypto.PubkeyToAddress(priKey.PublicKey)
		//pub := crypto.FromECDSAPub(&priKey.PublicKey)
		if i%2 == 0 {
			amount := new(big.Int).Mul(big.NewInt(200000), big.NewInt(1e18))
			impl.InsertAccount2(0, from, amount)
			//testacc = from
		} else {
			impl.InsertAccount2(0, from, value)
		}
	}
	fmt.Println("account number:", len(impl.accounts[1]), " all account:", impl.accounts[1])
	//used to redeem
	//v1, v2, v3, v4, v5 := impl.GetBalance(testacc)
	//fmt.Println("insert Balance", v1, v2, v3, v4, v5)
	//impl.CancelAccount(100, from, big.NewInt(10))
	//addLockedBalance(statedb, from, new(big.Int).Mul(big.NewInt(200000), big.NewInt(1e18)))
	//v1, v2, v3, v4, v5 = impl.GetBalance(testacc)
	//fmt.Println("cancel Balance", v1, v2, v3, v4, v5)

	//relayers election
	_, err = impl.DoElections(statedb, 1, 0)
	if err != nil {
		fmt.Println("error : ", err)
	}
	fmt.Println("relayer number:", len(impl.getElections2(1)), " all relayer: ", impl.getElections2(1))
	err = impl.Shift(1, 0)

	fmt.Println("------------------epoch 1----------------------")
	fmt.Printf("\n")
	//register account in epoch 1
	for i := uint64(0); i < 4; i++ {
		value := new(big.Int).Mul(big.NewInt(200000), big.NewInt(1e18))
		priKey, _ := crypto.GenerateKey()
		from := crypto.PubkeyToAddress(priKey.PublicKey)
		//	pub := crypto.FromECDSAPub(&priKey.PublicKey)
		if i%2 == 0 {
			impl.InsertAccount2(8880+i, from, value)
		} else {
			impl.InsertAccount2(9990+i, from, value)
		}
	}
	fmt.Println("epoch1 account number:", len(impl.accounts[1]), " all account:", impl.accounts[1])

	//relayer election
	relayer, err := impl.DoElections(statedb, 1, 9900)
	if err != nil {
		fmt.Println("error : ", err)
	}
	fmt.Println("add relayer:", len(relayer), ", qurry relayer:", len(impl.getElections2(1)))
	fmt.Println("epoch2 relayer: ", len(impl.getElections2(1)), " all relayer: ", impl.getElections2(1))
	fmt.Println()
	err = impl.Shift(2, 0)

	fmt.Println("------------------epoch 2----------------------")
	fmt.Printf("\n")
	//register
	for i := uint64(0); i < 2; i++ {
		value := big.NewInt(100)
		priKey, _ := crypto.GenerateKey()
		from := crypto.PubkeyToAddress(priKey.PublicKey)
		//pub := crypto.FromECDSAPub(&priKey.PublicKey)
		impl.InsertAccount2(13333+i, from, value)
	}
	fmt.Println(" epoch2 account:", len(impl.accounts[2]), impl.accounts[2])
	//relayer election
	relayer, err = impl.DoElections(statedb, 2, 19900)
	if err != nil {
		fmt.Println("error : ", err)
	}
	fmt.Println("epoch3 relayer: ", len(impl.getElections2(2)), " all relayer: ", impl.getElections2(2))
	err = impl.Shift(3, 0)
	fmt.Println(" relayer number:", len(relayer), " getElection 1:", len(impl.getElections2(1)), " getElection 2:", len(impl.getElections2(2)))
	//Redeem test
	//fmt.Println()
	//fmt.Println("--------------------redeem test-------------------------------")
	//c2 := impl.GetAllCancelableAsset(testacc)[testacc]
	//fmt.Println("cancelable", c2)
	//err = impl.RedeemAccount(21000, testacc, big.NewInt(10))
	//if err != nil {
	//	t.Fatal(err)
	//}
	//subLockedBalance(statedb, from, big.NewInt(10))
	//v1, v2, v3, v4, v5 = impl.GetBalance(testacc)
	//fmt.Println("after redeem", v1, v2, v3, v4, v5)
}

func TestGetBalance(t *testing.T) {
	value := big.NewInt(100)
	priKey, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(priKey.PublicKey)
	//pub := crypto.FromECDSAPub(&priKey.PublicKey)

	db := rawdb.NewMemoryDatabase()
	statedb, err := state.New(common.Hash{}, state.NewDatabase(db), nil)
	if err != nil {
		t.Fatal(err)
	}
	statedb.GetOrNewStateObject(params.RelayerAddress)
	//evm := NewEVM(BlockContext{}, TxContext{}, statedb, params.TestChainConfig, Config{})

	register := NewRegisterImpl()
	register.InsertAccount2(0, from, value)
	v1, v2, v3, v4 := register.GetBalance(from, h)
	fmt.Println("getBalance0", v1, v2, v3, v4)
}

func TestStateDB(t *testing.T) {
	db := rawdb.NewMemoryDatabase()

	statedb, err := state.New(common.Hash{}, state.NewDatabase(db), nil)
	if err != nil {
		t.Fatal(err)
	}
	statedb.GetOrNewStateObject(params.RelayerAddress)
	reg := NewRegisterImpl()
	reg.curEpochID = 11223
	if err := reg.Save(statedb, params.RelayerAddress); err != nil {
		log.Crit("store failed, ", "err", err)
	}
	reg.curEpochID = 2
	if err := reg.Load(statedb, params.RelayerAddress); err != nil {
		log.Crit("store failed, ", "err", err)
	}

	fmt.Println("------------------------------")
	statedb.GetOrNewStateObject(params.RelayerAddress)
	hs := ethereum.NewHeaderSync()
	//hs.epoch2reward[11233] = big.NewInt(11233)
	if err := hs.Store(statedb, params.RelayerAddress); err != nil {
		log.Crit("store failed, ", "err", err)
	}
	//hs.epoch2reward[2] = big.NewInt(2)
	if err := hs.Load(statedb, params.RelayerAddress); err != nil {
		log.Crit("store failed, ", "err", err)
	}
	fmt.Println(reg)
	fmt.Println(hs)
	fmt.Println(rlp.EncodeToBytes(hs))
}
