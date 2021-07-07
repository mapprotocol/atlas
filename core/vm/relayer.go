package vm

import (
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/params"
	"math/big"
	"strings"
)

var relayerABI abi.ABI

func init() {
	relayerABI, _ = abi.JSON(strings.NewReader(RelayerABIJSON))
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
	case "getRelayers":
		ret, err = getRelayers(evm, contract, data)
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
		err = errors.New("execution reverted")
	}
	return ret, err
}

func register(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	type reg struct {
		Pubkey []byte
		Fee    *big.Int
		Value  *big.Int
	}
	var args reg
	err = relayerABI.UnpackIntoInterface(&args, "received", input)
	if err != nil {
		log.Error("Unpack register pubkey error", "err", err)
		return nil, errors.New("invalid input for staking")
	}
	from := contract.CallerAddress
	if evm.StateDB.GetUnlockedBalance(from).Cmp(args.Value) < 0 { //if evm.StateDB.GetUnlockedBalance(from).Cmp(args.Value) < 0
		log.Error("Staking balance insufficient", "address", contract.CallerAddress, "value", args.Value)
		return nil, errors.New("invalid input for staking")
	}
	//
	impawn := NewRegisterImpl()
	err = impawn.Load(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("Staking load error", "error", err)
		return nil, err
	}
	//
	effectHeight := uint64(0)
	err = impawn.InsertSAccount2(evm.Context.BlockNumber.Uint64(), effectHeight, from, args.Pubkey, args.Value, args.Fee, true)
	if err != nil {
		log.Error("Staking register", "address", contract.CallerAddress, "value", args.Value, "error", err)
		return nil, err
	}
	//
	err = impawn.Save(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("Staking save state error", "error", err)
		return nil, err
	}
	addLockedBalance(evm.StateDB, from, args.Value)
	//
	event := relayerABI.Events["Register"]
	logData, err := event.Inputs.Pack(args.Pubkey, args.Value, args.Fee)
	if err != nil {
		log.Error("Pack staking log error", "error", err)
		return nil, err
	}
	log.Info("register log: ", logData)
	return nil, nil
}
func append_(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	from := contract.CallerAddress
	amount := big.NewInt(0)

	//method, _ := relayerABI.Methods["append"]
	//err = method.Inputs.Unpack(&amount, input)
	err = relayerABI.UnpackIntoInterface(&amount, "received", input)
	if err != nil {
		log.Error("Unpack append value error", "err", err)
		return nil, errors.New("invalid input for staking")
	}
	if evm.StateDB.GetUnlockedBalance(from).Cmp(amount) < 0 {
		log.Error("Staking balance insufficient", "address", contract.CallerAddress, "value", amount)
		return nil, errors.New("invalid input for staking")
	}
	//
	impawn := NewRegisterImpl()
	err = impawn.Load(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("Staking load error", "error", err)
		return nil, err
	}
	//
	err = impawn.AppendSAAmount(evm.Context.BlockNumber.Uint64(), from, amount)
	if err != nil {
		log.Error("Staking register extra", "address", contract.CallerAddress, "value", amount, "error", err)
		return nil, err
	}
	//
	err = impawn.Save(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("Staking save state error", "error", err)
		return nil, err
	}
	addLockedBalance(evm.StateDB, from, amount)
	//
	event := relayerABI.Events["Append"]
	logData, err := event.Inputs.Pack(amount)
	if err != nil {
		log.Error("Pack staking log error", "error", err)
		return nil, err
	}
	log.Info("append log: ", logData)
	return nil, nil
}
func withdraw(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	from := contract.CallerAddress
	amount := new(big.Int)

	//method, _ := relayerABI.Methods["withdraw"]
	//err = method.Inputs.Unpack(&amount, input)
	err = relayerABI.UnpackIntoInterface(&amount, "received", input)
	if err != nil {
		log.Error("Unpack withdraw input error")
		return nil, errors.New("invalid input for staking")
	}
	if evm.StateDB.GetLockedBalance(from).Cmp(amount) < 0 {
		log.Error("Staking balance insufficient", "address", contract.CallerAddress, "value", amount)
		return nil, errors.New("insufficient balance for staking transfer")
	}

	impawn := NewRegisterImpl()
	err = impawn.Load(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("Staking load error", "error", err)
		return nil, err
	}

	log.Info("Staking withdraw", "number", evm.Context.BlockNumber.Uint64(), "address", contract.CallerAddress, "value", amount)
	err = impawn.RedeemSAccount(evm.Context.BlockNumber.Uint64(), from, amount)
	if err != nil {
		log.Error("Staking withdraw error", "address", from, "value", amount, "err", err)
		return nil, err
	}

	err = impawn.Save(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("Staking save state error", "error", err)
		return nil, err
	}
	subLockedBalance(evm.StateDB, from, amount)

	event := relayerABI.Events["Withdraw"]
	logData, err := event.Inputs.Pack(amount)
	if err != nil {
		log.Error("Pack staking log error", "error", err)
		return nil, err
	}
	log.Info("append log: ", logData)
	return nil, nil
}
func getBalance(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	var registerAddr common.Address
	method, _ := relayerABI.Methods["getBalance"]
	var (
		staked   = big.NewInt(0)
		locked   = big.NewInt(0)
		unlocked = big.NewInt(0)
	)

	//err = method.Inputs.Unpack(&registerAddr, input)
	err = relayerABI.UnpackIntoInterface(&registerAddr, "received", input)
	if err != nil {
		log.Error("Unpack getBalance input error")
		return nil, errors.New("invalid input for staking")
	}

	impawn := NewRegisterImpl()
	err = impawn.Load(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("Staking load error", "error", err)
		return nil, err
	}

	asset := impawn.GetAllCancelableAsset(registerAddr)
	if stake, ok := asset[registerAddr]; ok {
		staked.Add(staked, stake)
	}

	lockedAsset := impawn.GetLockedAsset2(registerAddr, evm.Context.BlockNumber.Uint64())
	if stake, ok := lockedAsset[registerAddr]; ok {
		for _, item := range stake.Value {
			if item.Locked {
				locked.Add(locked, item.Amount)
			} else {
				unlocked.Add(unlocked, item.Amount)
			}
		}
	}

	log.Info("Get staking getBalance", "address", registerAddr, "staked", staked, "locked", locked, "unlocked", unlocked)

	ret, err = method.Outputs.Pack(staked, locked, unlocked)
	return ret, err
}
func getRelayers(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	args := struct {
		relayer big.Int
	}{}
	method, _ := relayerABI.Methods["getRelayers"]
	//err = method.Inputs.Unpack(&args, input)
	err = relayerABI.UnpackIntoInterface(&args, "received", input)
	if err != nil {
		return nil, err
	}
	//RegisterAccount->relayers
	impawn := NewRegisterImpl()
	impawn.Load(evm.StateDB, params.RelayerAddress)
	relayers := impawn.GetAllStakingAccount()
	_, h := impawn.GetCurrentEpochInfo()
	if relayers == nil {
		ret, err = method.Outputs.Pack(h, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		if err != nil {
			return nil, err
		}
	} else {
		ret, err = method.Outputs.Pack(h, relayers[0], relayers[1], relayers[2], relayers[3], relayers[4], relayers[5], relayers[6], relayers[7], relayers[8], relayers[9])
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}
func getPeriodHeight(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	args := struct {
		relayer common.Address
	}{}
	impawn := NewRegisterImpl()
	impawn.Load(evm.StateDB, params.RelayerAddress)

	info, h := impawn.GetCurrentEpochInfo()
	method, _ := relayerABI.Methods["getPeriodHeight"]
	//err = method.Inputs.Unpack(&args, input)
	err = relayerABI.UnpackIntoInterface(&args, "received", input)
	if err != nil {
		return nil, err
	}
	//stakingAccount->relayer
	impawn.GetStakingAccount(h, args.relayer)
	isRelayer, _ := impawn.GetStakingAccount(h, args.relayer)
	if isRelayer == nil {
		ret, err = method.Outputs.Pack(nil, nil, nil, false)
		return ret, err
	}
	//
	for _, v := range info {
		if h == v.EpochID {
			ret, err = method.Outputs.Pack(big.NewInt(int64(v.BeginHeight)), big.NewInt(int64(v.EndHeight)), big.NewInt(int64(v.BeginHeight-v.EndHeight)), true)
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

const RelayerABIJSON = `[
  {
    "name": "Register",
    "inputs": [
      {
        "type": "address",
        "name": "from",
        "indexed": true
      },
      {
        "type": "bytes",
        "name": "pubkey",
        "indexed": false
      },
      {
        "type": "uint256",
        "name": "value",
        "indexed": false
      },
      {
        "type": "uint256",
        "name": "fee",
        "indexed": false
      }
    ],
    "anonymous": false,
    "type": "event"
  },
  {
    "name": "Withdraw",
    "inputs": [
      {
        "type": "address",
        "name": "from",
        "indexed": true
      },
      {
        "type": "uint256",
        "name": "value",
        "indexed": false
      }
    ],
    "anonymous": false,
    "type": "event"
  },
  {
    "name": "Append",
    "inputs": [
      {
        "type": "address",
        "name": "from",
        "indexed": true
      },
      {
        "type": "uint256",
        "name": "value",
        "indexed": false
      }
    ],
    "anonymous": false,
    "type": "event"
  },
  {
    "name": "register",
    "outputs": [],
    "inputs": [
      {
        "type": "bytes",
        "name": "pubkey"
      },
      {
        "type": "uint256",
        "name": "fee"
      },
      {
        "type": "uint256",
        "name": "value"
      }
    ],
    "constant": false,
    "payable": false,
    "type": "function"
  },
  {
    "name": "append",
    "outputs": [],
    "inputs": [
      {
        "type": "address",
        "name": "holder"
      },
      {
        "type": "uint256",
        "name": "value"
      }
    ],
    "constant": false,
    "payable": false,
    "type": "function"
  },
  {
    "name": "getBalance",
    "outputs": [
      {
        "type": "uint256",
        "unit": "wei",
        "name": "staked"
      },
      {
        "type": "uint256",
        "unit": "wei",
        "name": "locked"
      },
      {
        "type": "uint256",
        "unit": "wei",
        "name": "unlocked"
      },
      {
        "type": "uint256",
        "unit": "wei",
        "name": "reward"
      },
      {
        "type": "uint256",
        "unit": "wei",
        "name": "fine"
      }
    ],
    "inputs": [
      {
        "type": "address",
        "name": "holder"
      }
    ],
    "constant": true,
    "payable": false,
    "type": "function"
  },
  {
    "name": "withdraw",
    "outputs": [],
    "inputs": [
      {
        "type": "address",
        "name": "holder"
      },
      {
        "type": "uint256",
        "unit": "wei",
        "name": "value"
      }
    ],
    "constant": false,
    "payable": false,
    "type": "function"
  },
  {
    "name": "getPeriodHeight",
    "outputs": [
      {
        "type": "uint256",
        "name": "start"
      },
      {
        "type": "uint256",
        "name": "end"
      },  
      {
        "type": "uint256",
        "name": "remain"
      },
      {
        "type": "bool",
        "name": "relayer"
      }
    ],
    "inputs": [
      {
        "type": "address",
        "name": "holder"
      }
    ],
    "constant": true,
    "payable": false,
    "type": "function"
  },
  {
    "name": "getRelayers",
    "inputs": [
      {
        "type": "uint256",
        "name": "period"
      }
    ],
    "outputs": [
      {
        "type": "address",
        "name": "relayer0"
      },
      {
        "type": "address",
        "name": "relayer1"
      },
      {
        "type": "address",
        "name": "relayer2"
      },
      {
        "type": "address",
        "name": "relayer3"
      },
      {
        "type": "address",
        "name": "relayer4"
      },
      {
        "type": "address",
        "name": "relayer5"
      },
      {
        "type": "address",
        "name": "relayer6"
      },
      {
        "type": "address",
        "name": "relayer7"
      },
      {
        "type": "address",
        "name": "relayer8"
      },
      {
        "type": "address",
        "name": "relayer9"
      },
      {
        "type": "uint256",
        "name": "total"
      }
    ],
    "constant": true,
    "payable": false,
    "type": "function"
  }
]`
