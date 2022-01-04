package random

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/abis"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
)

var (
	revealAndCommitMethod    = contracts.NewRegisteredContractMethod(params.RandomRegistryId, abis.Random, "revealAndCommit", params.MaxGasForRevealAndCommit)
	commitmentsMethod        = contracts.NewRegisteredContractMethod(params.RandomRegistryId, abis.Random, "commitments", params.MaxGasForCommitments)
	computeCommitmentMethod  = contracts.NewRegisteredContractMethod(params.RandomRegistryId, abis.Random, "computeCommitment", params.MaxGasForComputeCommitment)
	randomMethod             = contracts.NewRegisteredContractMethod(params.RandomRegistryId, abis.Random, "random", params.MaxGasForBlockRandomness)
	getBlockRandomnessMethod = contracts.NewRegisteredContractMethod(params.RandomRegistryId, abis.Random, "getBlockRandomness", params.MaxGasForBlockRandomness)
)

func IsRunning(vmRunner vm.EVMRunner) bool {

	randomAddress, err := contracts.GetRegisteredAddress(vmRunner, params.RandomRegistryId)

	if err == contracts.ErrSmartContractNotDeployed || err == contracts.ErrRegistryContractNotDeployed {
		//log.Debug("Registry address lookup failed", "err", err, "contract", hexutil.Encode(params.RandomRegistryId[:]))
	} else if err != nil {
		//log.Error(err.Error())
	}
	fmt.Println("IsRunning:", err == nil && randomAddress != params.ZeroAddress)
	return err == nil && randomAddress != params.ZeroAddress
}

// GetLastCommitment returns up the last commitment in the smart contract
func GetLastCommitment(vmRunner vm.EVMRunner, validator common.Address) (common.Hash, error) {
	lastCommitment := common.Hash{}
	err := commitmentsMethod.Query(vmRunner, &lastCommitment, validator)
	if err != nil {
		log.Error("Failed to get last commitment", "err", err)
		return lastCommitment, err
	}

	if (lastCommitment == common.Hash{}) {
		log.Debug("Unable to find last randomness commitment in smart contract")
	}
	fmt.Println("lastCommitment", lastCommitment)
	return lastCommitment, nil
}

// ComputeCommitment calulcates the commitment for a given randomness.
func ComputeCommitment(vmRunner vm.EVMRunner, randomness common.Hash) (common.Hash, error) {
	commitment := common.Hash{}
	//TODO(asa): Make an issue to not have to do this via StaticCall
	err := computeCommitmentMethod.Query(vmRunner, &commitment, randomness)
	if err != nil {
		log.Error("Failed to call computeCommitment()", "err", err)
		return common.Hash{}, err
	}
	//0x0e3f47124a429f254d6cb8f72e80b883fb99c622b1b3a914d76a2ede40afe9b2
	//0x003ddee9481b857e0bc5057eb085e91becad4e97da02a767ef9050968f3a7fba
	fmt.Println("ComputeCommitment", randomness)
	fmt.Println("commitment:", commitment)
	return commitment, err
}

// RevealAndCommit performs an internal call to the EVM that reveals a
// proposer's previously committed to randomness, and commits new randomness for
// a future block.
func RevealAndCommit(vmRunner vm.EVMRunner, randomness, newCommitment common.Hash, proposer common.Address) error {
	log.Trace("Revealing and committing randomness", "randomness", randomness.Hex(), "commitment", newCommitment.Hex())
	err := revealAndCommitMethod.Execute(vmRunner, nil, common.Big0, randomness, newCommitment, proposer)
	return err
}

// Random performs an internal call to the EVM to retrieve the current randomness from the official Random contract.
func Random(vmRunner vm.EVMRunner) (common.Hash, error) {
	randomness := common.Hash{}
	err := randomMethod.Query(vmRunner, &randomness)
	fmt.Println("Random:", randomness)
	return randomness, err
}

func BlockRandomness(vmRunner vm.EVMRunner, blockNumber uint64) (common.Hash, error) {
	randomness := common.Hash{}
	err := getBlockRandomnessMethod.Query(vmRunner, &randomness, big.NewInt(int64(blockNumber)))
	fmt.Println("BlockRandomness:", randomness)
	return randomness, err
}
