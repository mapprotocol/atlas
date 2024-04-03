package gold_token

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/abis"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
)

var (
	totalSupplyMethod    = contracts.NewRegisteredContractMethod(params.GoldTokenRegistryId, abis.GoldToken, "totalSupply", params.MaxGasForTotalSupply)
	increaseSupplyMethod = contracts.NewRegisteredContractMethod(params.GoldTokenRegistryId, abis.GoldToken, "increaseSupply", params.MaxGasForIncreaseSupply)
	mintMethod           = contracts.NewRegisteredContractMethod(params.GoldTokenRegistryId, abis.GoldToken, "mint", params.MaxGasForMintGas)
	balanceOf            = contracts.NewRegisteredContractMethod(params.GoldTokenRegistryId, abis.GoldToken, "balanceOf", params.MaxGasForIncreaseSupply)
)

func GetTotalSupply(vmRunner vm.EVMRunner) (*big.Int, error) {
	var totalSupply *big.Int
	err := totalSupplyMethod.Query(vmRunner, &totalSupply)
	return totalSupply, err
}

func IncreaseSupply(vmRunner vm.EVMRunner, value *big.Int) error {
	err := increaseSupplyMethod.Execute(vmRunner, nil, common.Big0, value)
	return err
}

func Mint(vmRunner vm.EVMRunner, beneficiary common.Address, value *big.Int) error {
	if value.Cmp(new(big.Int)) <= 0 {
		return nil
	}

	err := mintMethod.Execute(vmRunner, nil, common.Big0, beneficiary, value)
	return err
}

func BalanceOf(vmRunner vm.EVMRunner, owner common.Address) (*big.Int, error) {
	var balance *big.Int
	err := balanceOf.Query(vmRunner, &balance, owner)
	return balance, err
}
