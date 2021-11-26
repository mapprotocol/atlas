package testutil

import (
	"math/big"
	"testing"

	"github.com/mapprotocol/atlas/contracts/blockchain_parameters"
	. "github.com/onsi/gomega"
)

func TestRunnerWorks(t *testing.T) {
	g := NewGomegaWithT(t)

	atlas := NewAtlasMock()

	lw, err := blockchain_parameters.GetLookbackWindow(atlas.Runner)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(lw).To((Equal(uint64(3))))

	atlas.BlockchainParameters.LookbackWindow = big.NewInt(10)

	lw, err = blockchain_parameters.GetLookbackWindow(atlas.Runner)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(lw).To((Equal(uint64(10))))
}