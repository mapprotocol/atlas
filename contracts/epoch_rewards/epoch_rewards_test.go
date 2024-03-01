package epoch_rewards

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/testutil"
	"github.com/mapprotocol/atlas/params"
	. "github.com/onsi/gomega"
)

func TestGetCarbonOffsettingPartnerAddress(t *testing.T) {
	fn := GetCommunityPartnerAddress

	testutil.TestFailOnFailingRunner(t, fn)
	testutil.TestFailsWhenContractNotDeployed(t, contracts.ErrSmartContractNotDeployed, fn)

	t.Run("should indicate if reserve is low", func(t *testing.T) {
		g := NewGomegaWithT(t)

		runner := testutil.NewSingleMethodRunner(
			params.EpochRewardsRegistryId,
			"carbonOffsettingPartner",
			func() common.Address { return common.HexToAddress("0x00045") },
		)

		ret, err := GetCommunityPartnerAddress(runner)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ret).To(Equal(common.HexToAddress("0x00045")))
	})
}
func Test_01(t *testing.T) {
	//fn := GetMgrMaintainerAddress
	runner := testutil.NewSingleMethodRunner(
		params.EpochRewardsRegistryId,
		"carbonOffsettingPartner",
		func() common.Address { return common.HexToAddress("0x00045") },
	)

	ret, err := GetMgrMaintainerAddress(runner)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret)
}
