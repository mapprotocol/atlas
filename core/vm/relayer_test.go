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
	//relayerABI,_       = abi.JSON(strings.NewReader(RelayerABIJSON))
	priKey, _        = crypto.GenerateKey()
	from             = crypto.PubkeyToAddress(priKey.PublicKey)
	pub              = crypto.FromECDSAPub(&priKey.PublicKey)
	value            = big.NewInt(1000)
	h         uint64 = 0
	fee       uint64 = 1000000
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

func TestPack(t *testing.T) {
	//getBalance
	//input, err := relayerABI.Pack("getBalance", common.Address{'1'})
	//if err != nil {
	//	fmt.Println("----------pack----------", err)
	//}
	//err = relayerABI.UnpackIntoInterface(&input, "getBalance", input)
	//if err != nil {
	//	fmt.Println("----------unpack--------------", err)
	//}

	//register
	type reg struct {
		Pubkey []byte
		Fee    *big.Int
		Value  *big.Int
	}
	var args reg
	priKey, _ := crypto.GenerateKey()
	pk := crypto.FromECDSAPub(&priKey.PublicKey)
	input2, err := relayerABI.Pack("register", pk, new(big.Int).SetUint64(2), big.NewInt(100))
	if err != nil {
		fmt.Println("----------pack----------", err)
	}
	fmt.Println("input", input2)
	err = relayerABI.UnpackIntoInterface(&args, "register", input2)
	if err != nil {
		fmt.Println("----------unpack1--------------")
		fmt.Println(err)
	}
	output, err := relayerABI.Unpack("register", input2)
	if err != nil {
		fmt.Println("----------unpack2--------------", err)
		fmt.Println(err)
	}

	fmt.Println("ouput1", args)
	fmt.Println("ouput2", output)
}

func TestPack2(t *testing.T) {
	input, err := relayerABI.Pack("register", pub, new(big.Int).SetUint64(fee), value)
	if err != nil {
		t.Fatal(err)
	}

	args := struct {
		Pubkey []byte
		Fee    *big.Int
		Value  *big.Int
	}{}
	err = params2.UnpackRegister(&args, input)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("unpack ==> ", args)
	//method, _ := relayerABI.Methods["register"]
	//fmt.Println("input",input)
	//output,err := method.Inputs.Unpack(input)
	//
	//if err != nil {
	//	fmt.Println( "err", err)
	//}
	//fmt.Println("output",output)
}
