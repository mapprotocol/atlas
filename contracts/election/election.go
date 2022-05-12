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
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/abis"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
	"math/big"
	"sort"
	"time"
)

var (
	electValidatorSignersMethod              = contracts.NewRegisteredContractMethod(params.ElectionRegistryId, abis.Elections, "electValidatorSigners", params.MaxGasForElectValidators)
	getElectableValidatorsMethod             = contracts.NewRegisteredContractMethod(params.ElectionRegistryId, abis.Elections, "getElectableValidators", params.MaxGasForGetElectableValidators)
	electNValidatorSignersMethod             = contracts.NewRegisteredContractMethod(params.ElectionRegistryId, abis.Elections, "electNValidatorSigners", params.MaxGasForElectNValidatorSigners)
	getTotalVotesForEligibleValidatorsMethod = contracts.NewRegisteredContractMethod(params.ElectionRegistryId, abis.Elections, "getTotalVotesForEligibleValidators", params.MaxGasForGetEligibleValidatorsVoteTotals)
	distributeEpochVotersRewardsMethod       = contracts.NewRegisteredContractMethod(params.ElectionRegistryId, abis.Elections, "distributeEpochVotersRewards", params.MaxGasForDistributeVoterEpochRewards)

	activeAllPendingMethod             = contracts.NewRegisteredContractMethod(params.ElectionRegistryId, abis.Elections, "activeAllPending", params.MaxGasForActiveAllPending)
	getPendingVotersForValidatorMethod = contracts.NewRegisteredContractMethod(params.ElectionRegistryId, abis.Elections, "getPendingVotersForValidator", params.MaxGasForActiveAllPending)
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

type voteTotal struct {
	Validator common.Address
	Value     *big.Int
}

func getTotalVotesForEligibleValidators(vmRunner vm.EVMRunner) ([]voteTotal, error) {
	var validators []common.Address
	var values []*big.Int
	err := getTotalVotesForEligibleValidatorsMethod.Query(vmRunner, &[]interface{}{&validators, &values})
	if err != nil {
		return nil, err
	}

	voteTotals := make([]voteTotal, len(validators))
	for i, validator := range validators {
		log.Trace("Got Validator vote total", "Validator", validator, "value", values[i])
		voteTotals[i].Validator = validator
		voteTotals[i].Value = values[i]
	}
	return voteTotals, err
}

func DistributeEpochRewards(vmRunner vm.EVMRunner, validators []common.Address, rewards map[common.Address]*big.Int) (*big.Int, error) {
	totalRewards := big.NewInt(0)
	voteTotals, err := getTotalVotesForEligibleValidators(vmRunner)
	if err != nil {
		return totalRewards, err
	}

	for _, validator := range validators {
		reward := rewards[validator]
		if rewards[validator] == nil {
			reward = big.NewInt(0)
		}
		for _, voteTotal := range voteTotals {
			if voteTotal.Validator == validator {
				if rewards[validator] != nil {
					voteTotal.Value.Add(voteTotal.Value, rewards[validator])
				}
				break
			}
		}

		// Sorting in descending order is necessary to match the order on-chain.
		// TODO: We could make this more efficient by only moving the newly vote member.
		sort.SliceStable(voteTotals, func(j, k int) bool {
			return voteTotals[j].Value.Cmp(voteTotals[k].Value) > 0
		})

		lesser := params.ZeroAddress
		greater := params.ZeroAddress
		for j, voteTotal := range voteTotals {
			if voteTotal.Validator == validator {
				if j > 0 {
					greater = voteTotals[j-1].Validator
				}
				if j+1 < len(voteTotals) {
					lesser = voteTotals[j+1].Validator
				}
				break
			}
		}
		err := distributeEpochVotersRewardsMethod.Execute(vmRunner, nil, common.Big0, validator, reward, lesser, greater)
		if err != nil {
			return totalRewards, err
		}
		totalRewards.Add(totalRewards, reward)
	}
	return totalRewards, nil
}
func ActiveAllPending(vmRunner vm.EVMRunner, validators []common.Address) (bool, error) {
	vmRunner.StopGasMetering()
	defer vmRunner.StartGasMetering()
	start := time.Now()
	//debug voters
	for _, ele := range validators {
		var voters []common.Address
		getPendingVotersForValidatorMethod.Query(vmRunner, &voters, ele)
		log.Info("voters", "validator", ele, "voters", voters)
	}
	log.Info("ActiveAllPending", "start", time.Now().Sub(start))
	// Automatic activation
	var success bool
	err := activeAllPendingMethod.Execute(vmRunner, &success, common.Big0, validators)
	if err != nil {
		return false, err
	}
	log.Info("ActiveAllPending", "end", time.Now().Sub(start))
	return success, nil
}
