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

package validator

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/consensus/istanbul/validator/random"
)

func proposerIndex(valSet istanbul.ValidatorSet, proposer common.Address) uint64 {
	if idx := valSet.GetIndex(proposer); idx >= 0 {
		return uint64(idx)
	}
	return 0
}

// ShuffledRoundRobinProposer selects the next proposer with a round robin strategy according to a shuffled order.
func ShuffledRoundRobinProposer(valSet istanbul.ValidatorSet, proposer common.Address, round uint64) istanbul.Validator {
	if valSet.Size() == 0 {
		return nil
	}
	seed := valSet.GetRandomness()

	shuffle := random.Permutation(seed, valSet.Size())
	reverse := make([]int, len(shuffle))
	for i, n := range shuffle {
		reverse[n] = i
	}
	idx := round
	if proposer != (common.Address{}) {
		idx += uint64(reverse[proposerIndex(valSet, proposer)]) + 1
	}
	return valSet.List()[shuffle[idx%uint64(valSet.Size())]]
}

// RoundRobinProposer selects the next proposer with a round robin strategy according to storage order.
func RoundRobinProposer(valSet istanbul.ValidatorSet, proposer common.Address, round uint64) istanbul.Validator {
	if valSet.Size() == 0 {
		return nil
	}
	idx := round
	if proposer != (common.Address{}) {
		idx += proposerIndex(valSet, proposer) + 1
	}
	return valSet.List()[idx%uint64(valSet.Size())]
}

// StickyProposer selects the next proposer with a sticky strategy, advancing on round change.
func StickyProposer(valSet istanbul.ValidatorSet, proposer common.Address, round uint64) istanbul.Validator {
	if valSet.Size() == 0 {
		return nil
	}
	idx := round
	if proposer != (common.Address{}) {
		idx += proposerIndex(valSet, proposer)
	}
	return valSet.List()[idx%uint64(valSet.Size())]
}

// GetProposerSelector returns the ProposerSelector for the given Policy
func GetProposerSelector(pp istanbul.ProposerPolicy) istanbul.ProposerSelector {
	switch pp {
	case istanbul.Sticky:
		return StickyProposer
	case istanbul.RoundRobin:
		return RoundRobinProposer
	case istanbul.ShuffledRoundRobin:
		return ShuffledRoundRobinProposer
	default:
		// Programming error.
		panic(fmt.Sprintf("unknown proposer selection policy: %v", pp))
	}
}
