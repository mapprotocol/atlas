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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/consensus/istanbul/uptime"
	"github.com/mapprotocol/atlas/consensus/istanbul/uptime/store"
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/accounts"
	"github.com/mapprotocol/atlas/contracts/election"
	"github.com/mapprotocol/atlas/contracts/epoch_rewards"
	"github.com/mapprotocol/atlas/contracts/gold_token"
	"github.com/mapprotocol/atlas/contracts/validators"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
	"math/big"
	"time"
)

func (sb *Backend) distributeEpochRewards(header *types.Header, state *state.StateDB, EnableRewardBlock, bn256Block *big.Int) error {
	start := time.Now()
	defer sb.rewardDistributionTimer.UpdateSince(start)
	logger := sb.logger.New("func", "Backend.distributeEpochPaymentsAndRewards", "blocknum", header.Number.Uint64())

	vmRunner := sb.chain.NewEVMRunner(header, state)

	communityPartnerAddress, err := epoch_rewards.GetCommunityPartnerAddress(vmRunner)
	if err != nil {
		return err
	}

	validatorVoterReward, communityReward, maintainerReward, err := epoch_rewards.CalculateTargetEpochRewards(vmRunner)
	if err != nil {
		return err
	}

	if communityPartnerAddress == params.ZeroAddress {
		communityReward = big.NewInt(0)
	}

	logger.Info("Calculated target rewards", "validatorReward", validatorVoterReward, "communityReward", communityReward, "maintainerReward", maintainerReward)

	// The validator set that signs off on the last block of the epoch is the one that we need to
	// iterate over.
	signerSet := sb.GetValidators(big.NewInt(header.Number.Int64()-1), header.ParentHash)
	if len(signerSet) == 0 {
		err := errors.New("Unable to fetch validator set to update scores and distribute rewards")
		logger.Error(err.Error())
		return err
	}
	validators_, err := sb.GetAccountsFromSigners(vmRunner, signerSet)
	if err != nil {
		return err
	}
	uptimeRets, ignores, err := sb.updateValidatorScores(header, state, signerSet)
	if err != nil {
		return err
	}

	if header.Number.Cmp(EnableRewardBlock) > 0 {
		scores, err := sb.calculatePaymentScoreDenominator(vmRunner, uptimeRets, ignores)
		if err != nil {
			return err
		}
		// Reward Validators And voters
		totalValidatorRewards, voterRewardData, err := sb.distributeValidatorRewards(vmRunner, signerSet, validators_, validatorVoterReward, scores)
		if err != nil {
			return err
		}
		log.Info("totalValidatorRewards", "maxReward", totalValidatorRewards.String())
		totalVoterRewards, err := sb.distributeVoterRewards(vmRunner, validators_, voterRewardData)
		if err != nil {
			return err
		}
		log.Info("distributeVoterRewards", "totalVoterRewards", totalVoterRewards.String())
		if communityReward.Cmp(new(big.Int)) != 0 {
			if err = gold_token.Mint(vmRunner, communityPartnerAddress, communityReward); err != nil {
				return err
			}
		}
		// mint to mgrMaintainer
		if maintainerReward.Cmp(new(big.Int)) != 0 {
			mmAddress, err := epoch_rewards.GetMgrMaintainerAddress(vmRunner)
			if err != nil {
				return err
			}
			if mmAddress != params.ZeroAddress {
				if err = gold_token.Mint(vmRunner, mmAddress, maintainerReward); err != nil {
					log.Error("reward to maintainer fail", "addr", mmAddress, "maintainerReward", maintainerReward.String())
					return err
				}
				log.Info("reward to maintainer success", "addr", mmAddress, "maintainerReward", maintainerReward.String())
			}
		}
	}
	//----------------------------- deRegister -------------------
	deRegisters, err := sb.deRegisterAllValidatorsInPending(vmRunner)
	if err != nil {
		return err
	}
	log.Info("deRegister AllValidators InPending", "deRegisters", deRegisters)

	//----------------------------- Automatic active -------------------
	var b = false
	if header.Number.Cmp(bn256Block) >= 0 {
		// active the next epoch validators
		vals, err := sb.GetValidatorAccounts(vmRunner)
		if err != nil {
			return err
		}
		b, err = sb.activeAllPending(vmRunner, vals)
		if err != nil {
			return err
		}
	} else {
		b, err = sb.activeAllPending(vmRunner, validators_)
		if err != nil {
			return err
		}
	}

	log.Info("Automatic active pending voter", "success", b)
	//----------------------------------------------------------------------

	return nil
}

func (sb *Backend) updateValidatorScores(header *types.Header, state *state.StateDB, valSet []istanbul.Validator) ([]*big.Int, []bool, error) {
	epoch := istanbul.GetEpochNumber(header.Number.Uint64(), sb.EpochSize())
	logger := sb.logger.New("func", "Backend.updateValidatorScores", "blocknum", header.Number.Uint64(), "epoch", epoch, "epochsize", sb.EpochSize())
	ignore := make([]bool, len(valSet), len(valSet))
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
		return nil, nil, err
	}

	vmRunner := sb.chain.NewEVMRunner(header, state)

	for i, val := range valSet {
		logger.Trace("Updating validator score", "uptime", uptimes[i], "address", val.Address())
		uptimeRet, isValidator, err := validators.UpdateValidatorScore(vmRunner, val.Address(), uptimes[i])
		if !isValidator {
			ignore[i] = true
		}
		if err != nil {
			sb.logger.Error("Error in updateValidatorScores to validator", "address", val.Address(), "err", err)
			continue
		}
		uptimes[i] = uptimeRet
		logger.Trace("Updating validator score ret", "uptime", uptimes[i], "address", val.Address())
	}
	return uptimes, ignore, nil
}

/*
@param maxReward is epochReward for all validators
*/
func (sb *Backend) distributeValidatorRewards(vmRunner vm.EVMRunner, signerSet []istanbul.Validator, valSets []common.Address, maxReward *big.Int, scoreDenominator *big.Int) (*big.Int, map[common.Address]*big.Int, error) {
	totalValidatorRewards := big.NewInt(0)
	voterRewards := make(map[common.Address]*big.Int, len(signerSet))
	for i, val := range signerSet {
		sb.logger.Debug("Distributing epoch reward for validator", "address", val.Address())
		validatorReward, voterReward, err := validators.DistributeEpochReward(vmRunner, val.Address(), maxReward, scoreDenominator)
		if err != nil {
			sb.logger.Error("Error in distributing rewards to validator", "address", val.Address(), "err", err)
			continue
		}
		voterRewards[valSets[i]] = voterReward
		totalValidatorRewards.Add(totalValidatorRewards, validatorReward)
	}
	return totalValidatorRewards, voterRewards, nil
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

/*
   @notice calculatePaymentScoreDenominator
   @params uptimes  update score for validator return
   @dev     (score + p)/(N*p+s1+s2+s3...)
   @return (N*p+s1+s2+s3...)
*/
func (sb *Backend) calculatePaymentScoreDenominator(vmRunner vm.EVMRunner, uptimes []*big.Int, ignores []bool) (*big.Int, error) {
	PledgeMultiplier, err := validators.GetPledgeMultiplierInReward(vmRunner)
	if err != nil {
		return nil, err
	}
	sum := big.NewInt(0)
	for i, v := range uptimes {
		if ignores[i] {
			continue
		}
		sum.Add(sum, v)
		sum.Add(sum, PledgeMultiplier)
	}
	return sum, nil
}
func (sb *Backend) distributeVoterRewards(vmRunner vm.EVMRunner, validators []common.Address, rewards map[common.Address]*big.Int) (*big.Int, error) {
	lockedGoldAddress, err := contracts.GetRegisteredAddress(vmRunner, params.LockedGoldRegistryId)
	totalReward, err := election.DistributeEpochRewards(vmRunner, validators, rewards)
	if err != nil {
		return nil, err
	}
	gold_token.Mint(vmRunner, lockedGoldAddress, totalReward)
	return totalReward, nil
}

func (sb *Backend) activeAllPending(vmRunner vm.EVMRunner, validators []common.Address) (bool, error) {
	b, err := election.ActiveAllPending(vmRunner, validators)
	if err != nil {
		return false, err
	}
	return b, nil
}
func (sb *Backend) deRegisterAllValidatorsInPending(vmRunner vm.EVMRunner) (*[]common.Address, error) {
	deValidators, err := validators.DeRegisterValidatorsInPending(vmRunner)
	if err != nil {
		return nil, err
	}
	return deValidators, nil
}

func (sb *Backend) GetAccountsFromSigners(vmRunner vm.EVMRunner, signers []istanbul.Validator) ([]common.Address, error) {
	var accountVals []common.Address
	for i := 0; i < len(signers); i++ {
		regVals, err := accounts.GetSignerToAccountMethod(vmRunner, signers[i].Address())
		if err != nil {
			sb.logger.Error("failed to get account from signer", "signer", signers[i], "err", err)
			return accountVals, err
		}
		accountVals = append(accountVals, regVals)
	}
	return accountVals, nil
}
func (sb *Backend) GetAccountsFromSignersAddress(vmRunner vm.EVMRunner, signers []common.Address) ([]common.Address, error) {
	var accountVals []common.Address
	for i := 0; i < len(signers); i++ {
		regVals, err := accounts.GetSignerToAccountMethod(vmRunner, signers[i])
		if err != nil {
			sb.logger.Error("failed to get account from signer", "signer", signers[i], "err", err)
			return accountVals, err
		}
		accountVals = append(accountVals, regVals)
	}
	return accountVals, nil
}

func (sb *Backend) GetValidatorAccounts(vmRunner vm.EVMRunner) ([]common.Address, error) {
	signers, err := election.GetElectedValidators(vmRunner)
	if err != nil {
		return nil, err
	}
	return sb.GetAccountsFromSignersAddress(vmRunner, signers)
}
