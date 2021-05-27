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
	"github.com/ethereum/go-ethereum/common"
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

type ChainInfoProof struct {
	Proof  *ProofInfo
	Header []*types.Header
}

func newChainInforProof() *ChainInfoProof {
	return &ChainInfoProof{
		Proof:  &ProofInfo{},
		Header: []*types.Header{},
	}
}

type ChainProofs struct {
	HeaderProof *ChainHeaderProof
	InfoProof   *ChainInfoProof
}

func NewMapProofs() *ChainProofs {
	return &ChainProofs{
		HeaderProof: newChainHeaderProof(),
		InfoProof:   newChainInforProof(),
	}
}

func (b *ChainProofs) Datas() ([]byte, error) {
	data, err := rlp.EncodeToBytes(b)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (b *ChainProofs) checkMmrRoot() error {
	if b.HeaderProof != nil && b.InfoProof != nil {
		fRoot, sRoot := b.HeaderProof.Proof.RootHash, b.InfoProof.Proof.RootHash
		if !bytes.Equal(fRoot[:], sRoot[:]) {
			fmt.Println("mmr root not match for second proof,first:", hex.EncodeToString(fRoot[:]), "second:", hex.EncodeToString(sRoot[:]))
			return errors.New("mmr root not match for second proof")
		}
		return nil
	}
	return errors.New("invalid params in checkMmrRoot")
}

///////////////////////////////////////////////////////////////////////////////////

type OtherChainAdapter struct {
	Genesis      common.Hash
	ConfirmBlock *types.Header
	ProofHeader  *types.Header
	Leatest      []*types.Header
}

func (o *OtherChainAdapter) Copy() *OtherChainAdapter {
	tmp := &OtherChainAdapter{
		Genesis:      o.Genesis,
		ConfirmBlock: &types.Header{},
		ProofHeader:  &types.Header{},
		Leatest:      o.Leatest,
	}
	if o.ConfirmBlock != nil {
		tmp.ConfirmBlock = types.CopyHeader(o.ConfirmBlock)
	}
	if o.ProofHeader != nil {
		tmp.ProofHeader = types.CopyHeader(o.ProofHeader)
	}
	fmt.Println("***Genesis***", hex.EncodeToString(tmp.Genesis[:]))
	return tmp
}

// header block check
func (o *OtherChainAdapter) originHeaderCheck(head []*types.Header) error {
	// check difficult
	return nil
}

func (o *OtherChainAdapter) GenesisCheck(head *types.Header) error {

	rHash, lHash := head.Hash(), o.Genesis
	if !bytes.Equal(rHash[:], lHash[:]) {
		fmt.Println("genesis not match,local:", hex.EncodeToString(lHash[:]), "remote:", hex.EncodeToString(rHash[:]))
		return errors.New("genesis not match")
	}
	return nil
}
func (o *OtherChainAdapter) checkAndSetHeaders(heads []*types.Header, setcur bool) error {
	if len(heads) == 0 {
		return errors.New("invalid params")
	}

	if err := o.originHeaderCheck(heads); err != nil {
		return err
	}

	if setcur {
		head := heads[0]
		o.setProofHeader(head)
	} else {
		o.setLeatestHeader(heads[1], heads[2:])
	}
	return nil
}
func (o *OtherChainAdapter) setProofHeader(head *types.Header) {
	o.ProofHeader = types.CopyHeader(head)
}
func (o *OtherChainAdapter) setLeatestHeader(confirm *types.Header, leatest []*types.Header) {
	o.ConfirmBlock = types.CopyHeader(confirm)
	tmp := []*types.Header{}
	for _, v := range leatest {
		tmp = append(tmp, types.CopyHeader(v))
	}
	o.Leatest = tmp
}
func (o *OtherChainAdapter) checkMmrRootForFirst(root common.Hash) error {
	if len(o.Leatest) > 0 {
		l := o.Leatest[len(o.Leatest)-1]
		rHash := l.MmrRoot
		if !bytes.Equal(root[:], rHash[:]) {
			fmt.Println("mmr root not match for first proof in header:", hex.EncodeToString(root[:]), "root in proof:", hex.EncodeToString(rHash[:]))
			return errors.New("genesis not match")
		}
		return nil
	}
	return errors.New("not get the first proof")
}

///////////////////////////////////////////////////////////////////////////////////

type MapProofs struct {
	Remote *OtherChainAdapter `json:"remote"     rlp:"nil"`
	Proofs *ChainProofs
}

func (mps *MapProofs) Verify() error {

	if pBlocks, err := VerifyRequiredBlocks(mps.Proofs.HeaderProof.Proof, mps.Proofs.HeaderProof.Right); err != nil {
		return err
	} else {
		if !mps.Proofs.HeaderProof.Proof.VerifyProof(pBlocks) {
			return errors.New("Verify Proof Failed on first msg")
		} else {
			if err := mps.Remote.GenesisCheck(mps.Proofs.HeaderProof.Header[0]); err != nil {
				return err
			}
			if err := mps.Remote.checkAndSetHeaders(mps.Proofs.HeaderProof.Header, false); err != nil {
				return err
			}
			if err := mps.Remote.checkMmrRootForFirst(mps.Proofs.HeaderProof.Proof.RootHash); err != nil {
				return err
			}
			if pBlocks, err := VerifyRequiredBlocks2(mps.Proofs.InfoProof.Proof); err != nil {
				return err
			} else {
				if !mps.Proofs.InfoProof.Proof.VerifyProof2(pBlocks) {
					return errors.New("Verify Proof2 Failed on first msg")
				}
				if err := mps.checkMmrRoot(); err != nil {
					return err
				}
				// check headers
				return mps.Remote.checkAndSetHeaders(mps.Proofs.InfoProof.Header, true)
			}
		}
	}
	return nil
}

func (mps *MapProofs) checkMmrRoot() error {
	return mps.Proofs.checkMmrRoot()
}
