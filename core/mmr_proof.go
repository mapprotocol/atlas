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
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"math/big"
)

type ChainHeaderProof struct {
	Proof  *ProofInfo // the leatest blockchain and an proof of existence
	Header []*types.Header
	Right  *big.Int
}

func newChainHeaderProof() *ChainHeaderProof {
	return &ChainHeaderProof{
		Proof:  &ProofInfo{},
		Header: []*types.Header{},
		Right:  big.NewInt(0),
	}
}
func (b *ChainHeaderProof) Datas() ([]byte, error) {
	data, err := rlp.EncodeToBytes(b)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type ChainInforProof struct {
	Proof  *ProofInfo
	Header []*types.Header
}

func newChainInforProof() *ChainInforProof {
	return &ChainInforProof{
		Proof:  &ProofInfo{},
		Header: []*types.Header{},
	}
}

type MapProofs struct {
	FirstRes  *ChainHeaderProof
	SecondRes *ChainInforProof
}

func NewMapProofs() *MapProofs {
	return &MapProofs{
		FirstRes:  newChainHeaderProof(),
		SecondRes: newChainInforProof(),
	}
}

func (b *MapProofs) Datas() ([]byte, error) {
	data, err := rlp.EncodeToBytes(b)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (b *MapProofs) checkMmrRoot() error {
	if b.FirstRes != nil && b.SecondRes != nil {
		fRoot, sRoot := b.FirstRes.Proof.RootHash, b.SecondRes.Proof.RootHash
		if !bytes.Equal(fRoot[:], sRoot[:]) {
			fmt.Println("mmr root not match for second proof,first:", hex.EncodeToString(fRoot[:]), "second:", hex.EncodeToString(sRoot[:]))
			return errors.New("mmr root not match for second proof")
		}
		return nil
	}
	return errors.New("invalid params in checkMmrRoot")
}
