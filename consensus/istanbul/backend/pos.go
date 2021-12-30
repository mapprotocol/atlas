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

package backend

import (
	"errors"
	"github.com/mapprotocol/atlas/core/chain"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/consensus/istanbul/uptime"
	"github.com/mapprotocol/atlas/consensus/istanbul/uptime/store"
	"github.com/mapprotocol/atlas/contracts/epoch_rewards"
	"github.com/mapprotocol/atlas/contracts/gold_token"
	"github.com/mapprotocol/atlas/contracts/validators"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
)

func (sb *Backend) distributeEpochRewards(header *types.Header, state *state.StateDB) error {
	start := time.Now()
	defer sb.rewardDistributionTimer.UpdateSince(start)
	logger := sb.logger.New("func", "Backend.distributeEpochPaymentsAndRewards", "blocknum", header.Number.Uint64())

	vmRunner := sb.chain.NewEVMRunner(header, state)

	communityPartnerAddress, err := epoch_rewards.GetCommunityPartnerAddress(vmRunner)
	if err != nil {
		return err
	}

	validatorVoterReward, communityReward, err := epoch_rewards.CalculateTargetEpochRewards(vmRunner)
	if err != nil {
		return err
	}

	if communityPartnerAddress == params.ZeroAddress {
		communityReward = big.NewInt(0)
	}

	logger.Debug("Calculated target rewards", "validatorReward", validatorVoterReward, "communityReward", communityReward)

	// The validator set that signs off on the last block of the epoch is the one that we need to
	// iterate over.
	valSet := sb.GetValidators(big.NewInt(header.Number.Int64()-1), header.ParentHash)
	if len(valSet) == 0 {
		err := errors.New("Unable to fetch validator set to update scores and distribute rewards")
		logger.Error(err.Error())
		return err
	}

	_, err = sb.updateValidatorScores(header, state, valSet)
	if err != nil {
		return err
	}

	// Reward Validators And voters
	totalValidatorRewards, err := sb.distributeValidatorRewards(vmRunner, valSet, validatorVoterReward)
	if err != nil {
		return err
	}
	log.Info("totalValidatorRewards", "maxReward", totalValidatorRewards.String())

	if communityReward.Cmp(new(big.Int)) != 0 {
		if err = gold_token.Mint(vmRunner, communityPartnerAddress, communityReward); err != nil {
			return err
		}
	}

	return nil
}

func (sb *Backend) updateValidatorScores(header *types.Header, state *state.StateDB, valSet []istanbul.Validator) ([]*big.Int, error) {
	epoch := istanbul.GetEpochNumber(header.Number.Uint64(), sb.EpochSize())
	logger := sb.logger.New("func", "Backend.updateValidatorScores", "blocknum", header.Number.Uint64(), "epoch", epoch, "epochsize", sb.EpochSize())

	// header (&state) == lastBlockOfEpoch
	// sb.LookbackWindow(header, state) => value at the end of epoch
	// It doesn't matter which was the value at the beginning but how it ends.
	// Notice that exposed metrics compute based on current block (not last of epoch) so if lookback window changed during the epoch, metric uptime score might differ
	lookbackWindow := sb.LookbackWindow(header, state)

	logger = logger.New("window", lookbackWindow)
	logger.Trace("Updating validator scores")

	monitor := uptime.NewMonitor(store.New(sb.db), sb.EpochSize(), lookbackWindow)
	uptimes, err := monitor.ComputeValidatorsUptime(epoch, len(valSet))
	if err != nil {
		return nil, err
	}

	vmRunner := sb.chain.NewEVMRunner(header, state)
	for i, val := range valSet {
		logger.Trace("Updating validator score", "uptime", uptimes[i], "address", val.Address())
		err := validators.UpdateValidatorScore(vmRunner, val.Address(), uptimes[i])
		if err != nil {
			return nil, err
		}
	}
	return uptimes, nil
}

func (sb *Backend) distributeValidatorRewards(vmRunner vm.EVMRunner, valSet []istanbul.Validator, maxReward *big.Int) (*big.Int, error) {
	totalValidatorRewards := big.NewInt(0)
	for _, val := range valSet {
		sb.logger.Debug("Distributing epoch reward for validator", "address", val.Address())
		validatorReward, err := validators.DistributeEpochReward(vmRunner, val.Address(), maxReward)
		if err != nil {
			sb.logger.Error("Error in distributing rewards to validator", "address", val.Address(), "err", err)
			continue
		}
		totalValidatorRewards.Add(totalValidatorRewards, validatorReward)
	}
	return totalValidatorRewards, nil
}

func (sb *Backend) setInitialGoldTokenTotalSupplyIfUnset(vmRunner vm.EVMRunner) error {
	totalSupply, err := gold_token.GetTotalSupply(vmRunner)
	if err != nil {
		return err
	}
	// totalSupply not yet initialized.
	if totalSupply.Cmp(common.Big0) == 0 {
		data, err := sb.db.Get(chain.DBGenesisSupplyKey)
		if err != nil {
			log.Error("Unable to fetch genesisSupply from db", "err", err)
			return err
		}
		genesisSupply := new(big.Int)
		genesisSupply.SetBytes(data)

		err = gold_token.IncreaseSupply(vmRunner, genesisSupply)
		if err != nil {
			return err
		}
	}
	return nil
}
