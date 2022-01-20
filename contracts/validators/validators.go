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
package validators

import (
	"fmt"
	"github.com/mapprotocol/atlas/helper/bls"
	"math/big"

	blscrypto "github.com/celo-org/celo-bls-go/bls"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/abis"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
)

type ValidatorContractData struct {
	EcdsaPublicKey []byte
	BlsPublicKey   []byte
	Affiliation    common.Address
	Score          *big.Int
	Signer         common.Address
}

var (
	getRegisteredValidatorSignersMethod      = contracts.NewRegisteredContractMethod(params.ValidatorsRegistryId, abis.Validators, "getRegisteredValidatorSigners", params.MaxGasForGetRegisteredValidators)
	getRegisteredValidatorsMethod            = contracts.NewRegisteredContractMethod(params.ValidatorsRegistryId, abis.Validators, "getRegisteredValidators", params.MaxGasForGetRegisteredValidators)
	getValidatorBlsPublicKeyFromSignerMethod = contracts.NewRegisteredContractMethod(params.ValidatorsRegistryId, abis.Validators, "getValidatorBlsPublicKeyFromSigner", params.MaxGasForGetValidator)
	getPledgeMultiplierInRewardMethod        = contracts.NewRegisteredContractMethod(params.ValidatorsRegistryId, abis.Validators, "getPledgeMultiplierInReward", params.MaxGasForGetPledgeMultiplierInReward)
	getValidatorMethod                       = contracts.NewRegisteredContractMethod(params.ValidatorsRegistryId, abis.Validators, "getValidator", params.MaxGasForGetValidator)
	updateValidatorScoreFromSignerMethod     = contracts.NewRegisteredContractMethod(params.ValidatorsRegistryId, abis.Validators, "updateValidatorScoreFromSigner", params.MaxGasForUpdateValidatorScore)
	distributeEpochPaymentsFromSignerMethod  = contracts.NewRegisteredContractMethod(params.ValidatorsRegistryId, abis.Validators, "distributeEpochPaymentsFromSigner", params.MaxGasForDistributeEpochPayment)
)

func RetrieveRegisteredValidatorSigners(vmRunner vm.EVMRunner) ([]common.Address, error) {
	// Get the new epoch's validator signer set
	var regVals []common.Address
	if err := getRegisteredValidatorSignersMethod.Query(vmRunner, &regVals); err != nil {
		return nil, err
	}

	return regVals, nil
}

func RetrieveRegisteredValidators(vmRunner vm.EVMRunner) ([]common.Address, error) {
	// Get the new epoch's validator set
	var regVals []common.Address
	if err := getRegisteredValidatorsMethod.Query(vmRunner, &regVals); err != nil {
		return nil, err
	}

	return regVals, nil
}

func GetValidator(vmRunner vm.EVMRunner, validatorAddress common.Address) (ValidatorContractData, error) {
	var validator ValidatorContractData
	err := getValidatorMethod.Query(vmRunner, &validator, validatorAddress)
	if err != nil {
		return validator, err
	}
	if len(validator.BlsPublicKey) != blscrypto.PUBLICKEYBYTES {
		return validator, fmt.Errorf("length of bls public key incorrect. Expected %d, got %d", blscrypto.PUBLICKEYBYTES, len(validator.BlsPublicKey))
	}
	return validator, nil
}

func GetValidatorData(vmRunner vm.EVMRunner, validatorAddresses []common.Address) ([]istanbul.ValidatorData, error) {
	var validatorData []istanbul.ValidatorData
	for _, addr := range validatorAddresses {
		var blsKey []byte
		err := getValidatorBlsPublicKeyFromSignerMethod.Query(vmRunner, &blsKey, addr)
		if err != nil {
			return nil, err
		}

		if len(blsKey) != blscrypto.PUBLICKEYBYTES {
			return nil, fmt.Errorf("length of bls public key incorrect. Expected %d, got %d", blscrypto.PUBLICKEYBYTES, len(blsKey))
		}
		blsKeyFixedSize := bls.SerializedPublicKey{}
		copy(blsKeyFixedSize[:], blsKey)
		validator := istanbul.ValidatorData{
			Address:      addr,
			BLSPublicKey: blsKeyFixedSize,
		}
		validatorData = append(validatorData, validator)
	}
	return validatorData, nil
}

func UpdateValidatorScore(vmRunner vm.EVMRunner, address common.Address, uptime *big.Int) (*big.Int, bool, error) {
	var uptimeRet *big.Int
	var isValidator bool
	err := updateValidatorScoreFromSignerMethod.Execute(vmRunner, &[]interface{}{&uptimeRet, &isValidator}, common.Big0, address, uptime)
	return uptimeRet, isValidator, err
}

func DistributeEpochReward(vmRunner vm.EVMRunner, address common.Address, maxReward *big.Int, scoreDenominator *big.Int) (*big.Int, error) {
	var epochReward *big.Int
	err := distributeEpochPaymentsFromSignerMethod.Execute(vmRunner, &epochReward, common.Big0, address, maxReward, scoreDenominator)
	return epochReward, err
}

func GetPledgeMultiplierInReward(vmRunner vm.EVMRunner) (*big.Int, error) {
	var pledgeMultiplierInReward *big.Int
	err := getPledgeMultiplierInRewardMethod.Query(vmRunner, &pledgeMultiplierInReward)
	if err != nil {
		return nil, err
	}
	return pledgeMultiplierInReward, nil
}
