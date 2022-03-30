package election

import (
	"testing"

	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/testutil"
)

func TestGetElectedValidators(t *testing.T) {
	testutil.TestFailOnFailingRunner(t, GetElectedValidators)
	testutil.TestFailsWhenContractNotDeployed(t, contracts.ErrSmartContractNotDeployed, GetElectedValidators)
}
