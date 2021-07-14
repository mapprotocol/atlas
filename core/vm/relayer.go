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

//
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
	case "getBalance":
		ret, err = getBalance(evm, contract, data)
	case "getRelayer":
		ret, err = getRelayer(evm, contract, data)
	case "getPeriodHeight":
		ret, err = getPeriodHeight(evm, contract, data)
	case "withdraw":
		ret, err = withdraw(evm, contract, data)
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
		Pubkey []byte
		Fee    *big.Int
		Value  *big.Int
	}{
		output[0].([]byte),
		output[1].(*big.Int),
		output[2].(*big.Int),
	}

	from := contract.CallerAddress
	if evm.StateDB.GetUnlockedBalance(from).Cmp(args.Value) < 0 { //if evm.StateDB.GetUnlockedBalance(from).Cmp(args.Value) < 0
		log.Error("register balance insufficient", "address", contract.CallerAddress, "Value", args.Value)
		return nil, errors.New("invalid input for register")
	}
	register := NewRegisterImpl()
	//
	effectHeight := uint64(0)
	err = register.InsertAccount2(evm.Context.BlockNumber.Uint64(), effectHeight, from, args.Pubkey, args.Value, args.Fee, true)
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
	logData, err := event.Inputs.Pack(contract.CallerAddress, args.Pubkey, args.Value, args.Fee)
	if err != nil {
		log.Error("Pack register log error", "error", err)
		return nil, err
	}
	log.Info("register log: ", logData)
	return nil, nil
}
func append_(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	//from := contract.CallerAddress
	method, _ := relayerABI.Methods["append"]
	output, err := method.Inputs.Unpack(input)
	args := struct {
		Addr  common.Address
		Value *big.Int
	}{
		output[0].(common.Address),
		output[1].(*big.Int),
	}

	if evm.StateDB.GetUnlockedBalance(args.Addr).Cmp(args.Value) < 0 {
		log.Error("register balance insufficient", "address", contract.CallerAddress, "Value", args.Value)
		return nil, errors.New("invalid input for register")
	}
	//
	register := NewRegisterImpl()
	//
	err = register.AppendAmount(evm.Context.BlockNumber.Uint64(), args.Addr, args.Value)
	if err != nil {
		log.Error("register extra", "address", contract.CallerAddress, "Value", args.Value, "error", err)
		return nil, err
	}
	//
	err = register.Save(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("register save state error", "error", err)
		return nil, err
	}
	addLockedBalance(evm.StateDB, args.Addr, args.Value)
	//
	event := relayerABI.Events["Append"]
	logData, err := event.Inputs.Pack(args.Value)
	if err != nil {
		log.Error("Pack register log error", "error", err)
		return nil, err
	}
	log.Info("append log: ", logData)
	return nil, nil
}
func withdraw(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	method, _ := relayerABI.Methods["withdraw"]
	output, err := method.Inputs.Unpack(input)
	args := struct {
		Addr  common.Address
		Value *big.Int
	}{
		output[0].(common.Address),
		output[1].(*big.Int),
	}

	if evm.StateDB.GetLockedBalance(args.Addr).Cmp(args.Value) < 0 {
		log.Error("register balance insufficient", "address", contract.CallerAddress, "Value", args.Value)
		return nil, errors.New("insufficient balance for register transfer")
	}

	register := NewRegisterImpl()

	log.Info("register withdraw", "number", evm.Context.BlockNumber.Uint64(), "address", contract.CallerAddress, "Value", args.Value)
	err = register.RedeemAccount(evm.Context.BlockNumber.Uint64(), args.Addr, args.Value)
	if err != nil {
		log.Error("register withdraw error", "address", args.Addr, "Value", args.Value, "err", err)
		return nil, err
	}

	err = register.Save(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("register save state error", "error", err)
		return nil, err
	}
	subLockedBalance(evm.StateDB, args.Addr, args.Value)

	event := relayerABI.Events["Withdraw"]
	logData, err := event.Inputs.Pack(args.Value)
	if err != nil {
		log.Error("Pack register log error", "error", err)
		return nil, err
	}
	log.Info("append log: ", logData)
	return nil, nil
}
func getBalance(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	var (
		staked   = big.NewInt(0)
		locked   = big.NewInt(0)
		unlocked = big.NewInt(0)
	)
	method, _ := relayerABI.Methods["getBalance"]
	output, err := method.Inputs.Unpack(input)
	args := struct {
		Addr common.Address
	}{
		output[0].(common.Address),
	}

	register := NewRegisterImpl()
	err = register.Load(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("register load error", "error", err)
		return nil, err
	}

	asset := register.GetAllCancelableAsset(args.Addr)
	if stake, ok := asset[args.Addr]; ok {
		staked.Add(staked, stake)
	}

	lockedAsset := register.GetLockedAsset2(args.Addr, evm.Context.BlockNumber.Uint64())
	if stake, ok := lockedAsset[args.Addr]; ok {
		for _, item := range stake.Value {
			if item.Locked {
				locked.Add(locked, item.Amount)
			} else {
				unlocked.Add(unlocked, item.Amount)
			}
		}
	}

	log.Info("Get register getBalance", "address", args.Addr, "register", staked, "locked", locked, "unlocked", unlocked)
	ret, err = method.Outputs.Pack(staked, locked, unlocked, 0, 0)
	return ret, err
}
func getRelayer(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {

	method, _ := relayerABI.Methods["getRelayer"]
	output, err := method.Inputs.Unpack(input)
	args := struct {
		register common.Address
	}{
		output[0].(common.Address),
	}
	//RegisterAccount->relayers
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
			relayers := GetCurrentRelayer(evm.StateDB)
			for _, relayer := range relayers {
				if args.register == relayer.Coinbase {
					rel = true
				}
			}
		}
	}
	_, h := register.GetCurrentEpochInfo()
	epoch := new(big.Int).SetUint64(h)
	ret, err = method.Outputs.Pack(rel, acc, epoch)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
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
		log.Error("register load error", "error", err)
		return nil, err
	}

	info, h := register.GetCurrentEpochInfo()

	//stakingAccount->relayer
	register.GetRegisterAccount(h, args.Addr)
	isRelayer, _ := register.GetRegisterAccount(h, args.Addr)
	if isRelayer == nil {
		ret, err = method.Outputs.Pack(nil, nil, nil, false)
		return ret, err
	}
	//
	for _, v := range info {
		if h == v.EpochID {
			ret, err = method.Outputs.Pack(big.NewInt(int64(v.BeginHeight)), big.NewInt(int64(v.EndHeight)), big.NewInt(int64(v.EndHeight-evm.Context.BlockNumber.Uint64())), true)
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
