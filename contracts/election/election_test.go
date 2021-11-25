package election

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/testutil"
)

func TestGetElectedValidators(t *testing.T) {
	testutil.TestFailOnFailingRunner(t, GetElectedValidators)
	testutil.TestFailsWhenContractNotDeployed(t, contracts.ErrSmartContractNotDeployed, GetElectedValidators)

}

// func TestElectNValidatorSigners(t *testing.T) {

// }

func TestGetTotalVotesForEligibleValidatorGroups(t *testing.T) {
	testutil.TestFailOnFailingRunner(t, getTotalVotesForEligibleValidatorGroups)
	testutil.TestFailsWhenContractNotDeployed(t, contracts.ErrSmartContractNotDeployed, getTotalVotesForEligibleValidatorGroups)
}

func TestGetGroupEpochRewards(t *testing.T) {
	testutil.TestFailOnFailingRunner(t, getGroupEpochRewards, common.HexToAddress("0x05"), big.NewInt(10), []*big.Int{})
	testutil.TestFailsWhenContractNotDeployed(t, contracts.ErrSmartContractNotDeployed, getGroupEpochRewards, common.HexToAddress("0x05"), big.NewInt(10), []*big.Int{})
}

// func TestDistributeEpochRewards(t *testing.T) {

// }
