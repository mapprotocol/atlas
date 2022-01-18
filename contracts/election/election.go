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
package election

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/abis"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
	"math/big"
)

var (
	electValidatorSignersMethod  = contracts.NewRegisteredContractMethod(params.ElectionRegistryId, abis.Elections, "electValidatorSigners", params.MaxGasForElectValidators)
	getElectableValidatorsMethod = contracts.NewRegisteredContractMethod(params.ElectionRegistryId, abis.Elections, "getElectableValidators", params.MaxGasForGetElectableValidators)
	electNValidatorSignersMethod = contracts.NewRegisteredContractMethod(params.ElectionRegistryId, abis.Elections, "electNValidatorSigners", params.MaxGasForElectNValidatorSigners)
)

func GetElectedValidators(vmRunner vm.EVMRunner) ([]common.Address, error) {
	// Get the new epoch's validator set
	var newValSet []common.Address
	err := electValidatorSignersMethod.Query(vmRunner, &newValSet)
	if err != nil {
		return nil, err
	}
	return newValSet, nil
}

func ElectNValidatorSigners(vmRunner vm.EVMRunner, additionalAboveMaxElectable int64) ([]common.Address, error) {
	// Get the electable min and max
	var minElectableValidators *big.Int
	var maxElectableValidators *big.Int
	err := getElectableValidatorsMethod.Query(vmRunner, &[]interface{}{&minElectableValidators, &maxElectableValidators})
	if err != nil {
		return nil, err
	}
	// Run the validator election for up to maxElectable + getTotalVotesForEligibleValidatorGroup
	var electedValidators []common.Address
	err = electNValidatorSignersMethod.Query(vmRunner, &electedValidators, minElectableValidators, maxElectableValidators.Add(maxElectableValidators, big.NewInt(additionalAboveMaxElectable)))
	if err != nil {
		return nil, err
	}
	return electedValidators, nil
}
