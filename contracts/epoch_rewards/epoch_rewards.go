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
	communityPartnerMethod            = contracts.NewRegisteredContractMethod(params.EpochRewardsRegistryId, abis.EpochRewards, "communityPartner", params.MaxGasForGetCommunityPartnerSettingPartner)
)

// Returns the per validator epoch reward, the total voter reward, the total community reward, and
// the total carbon offsetting partner award, for the epoch.
func CalculateTargetEpochRewards(vmRunner vm.EVMRunner) (*big.Int, *big.Int, error) {
	var validatorVoterEpochReward *big.Int
	var totalCommunityReward *big.Int
	err := calculateTargetEpochRewardsMethod.Query(vmRunner, &[]interface{}{&validatorVoterEpochReward, &totalCommunityReward})
	if err != nil {
		return nil, nil, err
	}
	return validatorVoterEpochReward, totalCommunityReward, nil
}

// Returns the address of the carbon offsetting partner
func GetCommunityPartnerAddress(vmRunner vm.EVMRunner) (common.Address, error) {
	var communityPartnerPartner common.Address
	err := communityPartnerMethod.Query(vmRunner, &communityPartnerPartner)
	if err != nil {
		return params.ZeroAddress, err
	}
	return communityPartnerPartner, nil
}
