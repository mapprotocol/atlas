// Copyright 2021 MAP Protocol Authors.
// This file is part of MAP Protocol.

// MAP Protocol is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// MAP Protocol is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with MAP Protocol.  If not, see <http://www.gnu.org/licenses/>.

package gasprice_minimum

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/abis"
	"github.com/mapprotocol/atlas/contracts/blockchain_parameters"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
)

var (
	FallbackGasPriceMinimum *big.Int = big.NewInt(0) // gas price minimum to return if unable to fetch from contract
	suggestionMultiplier    *big.Int = big.NewInt(5) // The multiplier that we apply to the minimum when suggesting gas price
)

var (
	getGasPriceMinimumMethod    = contracts.NewRegisteredContractMethod(params.GasPriceMinimumRegistryId, abis.GasPriceMinimum, "getGasPriceMinimum", params.MaxGasForGetGasPriceMinimum)
	updateGasPriceMinimumMethod = contracts.NewRegisteredContractMethod(params.GasPriceMinimumRegistryId, abis.GasPriceMinimum, "updateGasPriceMinimum", params.MaxGasForUpdateGasPriceMinimum)
)

func GetGasPriceSuggestion(vmRunner vm.EVMRunner, currency *common.Address) (*big.Int, error) {
	gasPriceMinimum, err := GetGasPriceMinimum(vmRunner, currency)
	return new(big.Int).Mul(gasPriceMinimum, suggestionMultiplier), err
}

func GetGasPriceMinimum(vmRunner vm.EVMRunner, currency *common.Address) (*big.Int, error) {
	var currencyAddress common.Address
	var err error

	if currency == nil {
		currencyAddress, err = contracts.GetRegisteredAddress(vmRunner, params.GoldTokenRegistryId)

		if err == contracts.ErrSmartContractNotDeployed || err == contracts.ErrRegistryContractNotDeployed {
			return FallbackGasPriceMinimum, nil
		}
		if err != nil {
			return FallbackGasPriceMinimum, err
		}
	} else {
		currencyAddress = *currency
	}

	var gasPriceMinimum *big.Int
	err = getGasPriceMinimumMethod.Query(vmRunner, &gasPriceMinimum, currencyAddress)

	if err == contracts.ErrSmartContractNotDeployed || err == contracts.ErrRegistryContractNotDeployed {
		return FallbackGasPriceMinimum, nil
	}
	if err != nil {
		return FallbackGasPriceMinimum, err
	}

	return gasPriceMinimum, err
}

func UpdateGasPriceMinimum(vmRunner vm.EVMRunner, lastUsedGas uint64) (*big.Int, error) {
	return big.NewInt(0), nil
	//TODO Replace with the following in the future
	var updatedGasPriceMinimum *big.Int

	// If an error occurs, the default block gas limit will be returned and a log statement will be produced by GetBlockGasLimitOrDefault
	gasLimit := blockchain_parameters.GetBlockGasLimitOrDefault(vmRunner)

	err := updateGasPriceMinimumMethod.Execute(vmRunner, &updatedGasPriceMinimum, common.Big0, big.NewInt(int64(lastUsedGas)), big.NewInt(int64(gasLimit)))

	if err != nil {
		return nil, err
	}
	return updatedGasPriceMinimum, err
}
