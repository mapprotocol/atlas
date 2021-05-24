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

package core

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"math/big"
)

type ChainHeaderProofMsg struct {
	Proof  *ProofInfo // the leatest blockchain and an proof of existence
	Header []*types.Header
	Right  *big.Int
}

func newChainHeaderProofMsg() *ChainHeaderProofMsg {
	return &ChainHeaderProofMsg{
		Proof:  &ProofInfo{},
		Header: []*types.Header{},
		Right:  big.NewInt(0),
	}
}
func (b *ChainHeaderProofMsg) Datas() ([]byte, error) {
	data, err := rlp.EncodeToBytes(b)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type ChainInProofMsg struct {
	Proof  *ProofInfo
	Header []*types.Header
}

func newChainInProofMsg() *ChainInProofMsg {
	return &ChainInProofMsg{
		Proof:  &ProofInfo{},
		Header: []*types.Header{},
	}
}
