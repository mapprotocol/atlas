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
package epoch_rewards

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/abis"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
)

var (
	calculateTargetEpochRewardsMethod = contracts.NewRegisteredContractMethod(params.EpochRewardsRegistryId, abis.EpochRewards, "calculateTargetEpochRewards", params.MaxGasForCalculateTargetEpochPaymentAndRewards)
	isReserveLowMethod                = contracts.NewRegisteredContractMethod(params.EpochRewardsRegistryId, abis.EpochRewards, "isReserveLow", params.MaxGasForIsReserveLow)
	carbonOffsettingPartnerMethod     = contracts.NewRegisteredContractMethod(params.EpochRewardsRegistryId, abis.EpochRewards, "carbonOffsettingPartner", params.MaxGasForGetCarbonOffsettingPartner)
	updateTargetVotingYieldMethod     = contracts.NewRegisteredContractMethod(params.EpochRewardsRegistryId, abis.EpochRewards, "updateTargetVotingYield", params.MaxGasForUpdateTargetVotingYield)
)

func UpdateTargetVotingYield(vmRunner vm.EVMRunner) error {
	err := updateTargetVotingYieldMethod.Execute(vmRunner, nil, common.Big0)
	return err
}

// Returns the per validator epoch reward, the total voter reward, the total community reward, and
// the total carbon offsetting partner award, for the epoch.
func CalculateTargetEpochRewards(vmRunner vm.EVMRunner) (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	var validatorEpochReward *big.Int
	var totalVoterRewards *big.Int
	var totalCommunityReward *big.Int
	var totalCarbonOffsettingPartnerReward *big.Int
	err := calculateTargetEpochRewardsMethod.Query(vmRunner, &[]interface{}{&validatorEpochReward, &totalVoterRewards, &totalCommunityReward, &totalCarbonOffsettingPartnerReward})
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return validatorEpochReward, totalVoterRewards, totalCommunityReward, totalCarbonOffsettingPartnerReward, nil
}

// Determines if the reserve is below it's critical threshold
func IsReserveLow(vmRunner vm.EVMRunner) (bool, error) {
	return false, nil
	//TODO Replace with the following in the future
	var isLow bool
	err := isReserveLowMethod.Query(vmRunner, &isLow)
	if err != nil {
		return false, err
	}
	return isLow, nil
}

// Returns the address of the carbon offsetting partner
func GetCarbonOffsettingPartnerAddress(vmRunner vm.EVMRunner) (common.Address, error) {
	var carbonOffsettingPartner common.Address
	err := carbonOffsettingPartnerMethod.Query(vmRunner, &carbonOffsettingPartner)
	if err != nil {
		return params.ZeroAddress, err
	}
	return carbonOffsettingPartner, nil
}
