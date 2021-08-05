package vm

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/params"
	"math/big"
	"strings"
)

var relayerABI abi.ABI

func init() {
	relayerABI, _ = abi.JSON(strings.NewReader(params.RelayerABIJSON))
}

func RunContract(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	var method *abi.Method
	method, err = relayerABI.MethodById(input)
	if err != nil {
		log.Error("No method found")
		return nil, errors.New("execution reverted")
	}
	data := input[4:]
	switch method.Name {
	case "register":
		ret, err = register(evm, contract, data)
	case "append":
		ret, err = append_(evm, contract, data)
	case "getRelayerBalance":
		ret, err = getBalance(evm, contract, data)
	case "getRelayer":
		ret, err = getRelayer(evm, contract, data)
	case "getPeriodHeight":
		ret, err = getPeriodHeight(evm, contract, data)
	case "withdraw":
		ret, err = withdraw(evm, contract, data)
	case "unregister":
		ret, err = unregister(evm, contract, data)
	default:
		log.Warn("RelayerContract call fallback function")
		err = errors.New("execution reverted")
	}
	if err != nil {
		log.Warn("PreCompiledContract error code", "code", err)
	}
	return ret, err
}

func register(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	method, _ := relayerABI.Methods["register"]
	output, err := method.Inputs.Unpack(input)
	args := struct {
		Value *big.Int
	}{
		output[0].(*big.Int),
	}

	from := contract.CallerAddress
	if evm.StateDB.GetUnlockedBalance(from).Cmp(args.Value) < 0 { //if evm.StateDB.GetUnlockedBalance(from).Cmp(args.Value) < 0
		log.Error("register balance insufficient", "address", contract.CallerAddress, "Value", args.Value)
		return nil, errors.New("invalid input for register")
	}
	register := NewRegisterImpl()
	err = register.Load(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("contract load error", "error", err)
		return nil, err
	}
	err = register.InsertAccount2(evm.Context.BlockNumber.Uint64(), from, args.Value)
	if err != nil {
		log.Error("register", "address", contract.CallerAddress, "Value", args.Value, "error", err)
		return nil, err
	}
	err = register.Save(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("register save state error", "error", err)
		return nil, err
	}
	addLockedBalance(evm.StateDB, from, args.Value)
	event := relayerABI.Events["Register"]
	logData, err := event.Inputs.Pack(contract.CallerAddress, args.Value)
	if err != nil {
		log.Error("Pack register log error", "error", err)
		return nil, err
	}
	log.Info("register log: ", logData)
	return nil, nil
}

func append_(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	from := contract.CallerAddress
	method, _ := relayerABI.Methods["append"]
	output, err := method.Inputs.Unpack(input)
	args := struct {
		Value *big.Int
	}{
		output[0].(*big.Int),
	}

	if evm.StateDB.GetUnlockedBalance(from).Cmp(args.Value) < 0 {
		log.Error("register balance insufficient", "address", contract.CallerAddress, "Value", args.Value)
		return nil, errors.New("invalid input for register")
	}

	register := NewRegisterImpl()
	err = register.Load(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("contract load error", "error", err)
		return nil, err
	}

	err = register.AppendAmount(evm.Context.BlockNumber.Uint64(), from, args.Value)
	if err != nil {
		log.Error("register extra", "address", contract.CallerAddress, "Value", args.Value, "error", err)
		return nil, err
	}

	addLockedBalance(evm.StateDB, from, args.Value)
	err = register.Save(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("register save state error", "error", err)
		return nil, err
	}

	event := relayerABI.Events["Append"]
	logData, err := event.Inputs.Pack(from, args.Value)
	if err != nil {
		log.Error("Pack register log error", "error", err)
		return nil, err
	}
	log.Info("append log: ", logData)
	return nil, nil
}

//if locked amount < withdrawn amount, it will cancel lacked amount from registered amount
//if locked amount > withdrawn amount, it will redeem withdrawn amount soon
func withdraw(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	from := contract.CallerAddress
	method, _ := relayerABI.Methods["withdraw"]
	output, err := method.Inputs.Unpack(input)
	args := struct {
		Value *big.Int
	}{
		output[0].(*big.Int),
	}

	register := NewRegisterImpl()
	err = register.Load(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("contract load error", "error", err)
		return nil, err
	}
	if evm.StateDB.GetLockedBalance(from).Cmp(args.Value) < 0 {
		log.Error("register balance insufficient", "address", args.Value, "value", args.Value)
		return nil, ErrInsufficientBalance
	}

	log.Info("register withdraw", "number", evm.Context.BlockNumber.Uint64(), "address", contract.CallerAddress, "Value", args.Value)
	err = register.RedeemAccount(evm.Context.BlockNumber.Uint64(), from, args.Value)
	if err != nil {
		log.Error("register withdraw error", "address", from, "Value", args.Value, "err", err)
	}

	err = register.Save(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("register save state error", "error", err)
		return nil, err
	}
	subLockedBalance(evm.StateDB, from, args.Value)

	event := relayerABI.Events["Withdraw"]
	logData, err := event.Inputs.Pack(from, args.Value)
	if err != nil {
		log.Error("Pack withdraw log error", "error", err)
		return nil, err
	}
	log.Info("withdraw log: ", logData)
	return nil, nil
}

//if you want withdraw, unregister amount you need in your locked balance
func unregister(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	from := contract.CallerAddress
	method, _ := relayerABI.Methods["unregister"]
	output, err := method.Inputs.Unpack(input)
	args := struct {
		Value *big.Int
	}{
		output[0].(*big.Int),
	}

	register := NewRegisterImpl()
	err = register.Load(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("contract load error", "error", err)
		return nil, err
	}
	err = register.CancelAccount(evm.Context.BlockNumber.Uint64(), from, args.Value)
	log.Info("unregistered", "number", evm.Context.BlockNumber.Uint64(), "address", contract.CallerAddress, "Value", args.Value)
	if err != nil {
		log.Error("unregistered error", "address", from, "Value", args.Value, "err", err)
		return nil, err
	}
	err = register.Save(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("register save state error", "error", err)
		return nil, err
	}

	event := relayerABI.Events["Unregister"]
	logData, err := event.Inputs.Pack(from, args.Value)
	if err != nil {
		log.Error("Pack unregister log error", "error", err)
		return nil, err
	}
	log.Info("unregister log: ", logData)
	return nil, nil
}

//it will return locked asset,register asset,unlocked asset,reward asset and fine
func getBalance(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	method, _ := relayerABI.Methods["getRelayerBalance"]
	output, err := method.Inputs.Unpack(input)
	args := struct {
		Addr common.Address
	}{
		output[0].(common.Address),
	}

	register := NewRegisterImpl()
	err = register.Load(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("contract load error", "error", err)
		return nil, err
	}

	unlocked, unlocking, locked, _, _ := register.GetBalance(args.Addr, evm.Context.BlockNumber.Uint64())
	if unlocked == nil {
		unlocking = big.NewInt(0)
	}
	if unlocking == nil {
		unlocking = big.NewInt(0)
	}
	if locked == nil {
		locked = big.NewInt(0)
	}

	log.Info("Get register getBalance", "address", args.Addr, "locked", locked, "unlocking", unlocking, "unlocked", unlocked)
	ret, err = method.Outputs.Pack(locked, unlocking, unlocked)
	//fmt.Println("log", unlocked, unlocking, locked, ret)
	return ret, err
}

//query your account is registered or not, is relayer or not
func getRelayer(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	method, _ := relayerABI.Methods["getRelayer"]
	output, err := method.Inputs.Unpack(input)
	args := struct {
		register common.Address
	}{
		output[0].(common.Address),
	}
	register := NewRegisterImpl()
	err = register.Load(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("register load error", "error", err)
		return nil, err
	}
	acc := false
	rel := false
	accounts := register.GetAllRegisterAccount()
	for _, v := range accounts {
		if args.register == v.Unit.Address {
			acc = true
			if v.isInRelayer() {
				rel = true
			}
		}
	}

	_, h := register.GetCurrentEpochInfo()
	epoch := new(big.Int).SetUint64(h)
	ret, err = method.Outputs.Pack(acc, rel, epoch)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

//query started time, ended time and remained time in epoch when you are relayer; if you aren't relayer,return null
func getPeriodHeight(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	method, _ := relayerABI.Methods["getPeriodHeight"]
	output, err := method.Inputs.Unpack(input)
	args := struct {
		Addr common.Address
	}{
		output[0].(common.Address),
	}
	register := NewRegisterImpl()
	err = register.Load(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("contract load error", "error", err)
		return nil, err
	}

	info, h := register.GetCurrentEpochInfo()

	rel := false
	for _, v := range register.accounts[h] {
		if v.Unit.Address == args.Addr && v.isInRelayer() {
			rel = true
		}
	}

	if !rel {
		ret, err = method.Outputs.Pack(big.NewInt(0), big.NewInt(0), big.NewInt(0), false)
		return ret, nil
	}

	for _, v := range info {
		if h == v.EpochID {
			ret, err = method.Outputs.Pack(big.NewInt(int64(v.BeginHeight)), big.NewInt(int64(v.EndHeight)), true)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	return ret, nil
}

func addLockedBalance(db StateDB, addr common.Address, amount *big.Int) {
	db.SetLockedBalance(addr, new(big.Int).Add(db.GetLockedBalance(addr), amount))
}

func subLockedBalance(db StateDB, addr common.Address, amount *big.Int) {
	db.SetLockedBalance(addr, new(big.Int).Sub(db.GetLockedBalance(addr), amount))
}

func GenesisAddLockedBalance(db StateDB, addr common.Address, amount *big.Int) {
	db.SetLockedBalance(addr, new(big.Int).Add(db.GetLockedBalance(addr), amount))
}
