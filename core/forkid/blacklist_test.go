package forkid

import (
	"math/big"
	"testing"

	"github.com/mapprotocol/atlas/params"
)

// Tests that BlacklistBlock does not affect the fork ID, so nodes with and
// without the blacklist fork can keep peering before and after it activates.
func TestBlacklistBlockNotInForkID(t *testing.T) {
	for _, base := range []*params.ChainConfig{params.MainnetChainConfig, params.TestnetConfig, params.DevnetConfig} {
		without := *base
		without.BlacklistBlock = nil
		with := *base
		with.BlacklistBlock = big.NewInt(26654675)

		genesis := params.MainnetGenesisHash
		for _, head := range []uint64{0, 26654674, 26654675, 26654676, 100000000} {
			idWithout := NewID(&without, genesis, head)
			idWith := NewID(&with, genesis, head)
			if idWithout != idWith {
				t.Errorf("chain %v head %d: fork ID mismatch: without %v, with %v", base.ChainID, head, idWithout, idWith)
			}
			headfn := func() uint64 { return head }
			if err := newFilter(&with, genesis, headfn)(idWithout); err != nil {
				t.Errorf("chain %v head %d: upgraded node rejects old node: %v", base.ChainID, head, err)
			}
			if err := newFilter(&without, genesis, headfn)(idWith); err != nil {
				t.Errorf("chain %v head %d: old node rejects upgraded node: %v", base.ChainID, head, err)
			}
		}
	}
}
