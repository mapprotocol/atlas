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
	"github.com/ethereum/go-ethereum/trie"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/rlp"
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
func (b *ChainHeaderProof) Data() ([]byte, error) {
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

func newChainInfoProof() *ChainInfoProof {
	return &ChainInfoProof{
		Proof:  &ProofInfo{},
		Header: []*types.Header{},
	}
}

///////////////////////////////////////////////////////////////////////////////////

type ChainAdapter struct {
	Genesis      common.Hash
	ConfirmBlock *types.Header
	ProofHeader  *types.Header
	Latest       []*types.Header
}

func (o *ChainAdapter) Copy() *ChainAdapter {
	tmp := &ChainAdapter{
		Genesis:      o.Genesis,
		ConfirmBlock: &types.Header{},
		ProofHeader:  &types.Header{},
		Latest:       o.Latest,
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
func (o *ChainAdapter) originHeaderCheck(head []*types.Header) error {
	// check difficult
	return nil
}

func (o *ChainAdapter) GenesisCheck(head *types.Header) error {

	rHash, lHash := head.Hash(), o.Genesis
	if !bytes.Equal(rHash[:], lHash[:]) {
		fmt.Println("genesis not match,local:", hex.EncodeToString(lHash[:]), "remote:", hex.EncodeToString(rHash[:]))
		return errors.New("genesis not match")
	}
	return nil
}
func (o *ChainAdapter) checkAndSetHeaders(heads []*types.Header, setcur bool) error {
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
func (o *ChainAdapter) setProofHeader(head *types.Header) {
	o.ProofHeader = types.CopyHeader(head)
}
func (o *ChainAdapter) setLeatestHeader(confirm *types.Header, leatest []*types.Header) {
	o.ConfirmBlock = types.CopyHeader(confirm)
	tmp := []*types.Header{}
	for _, v := range leatest {
		tmp = append(tmp, types.CopyHeader(v))
	}
	o.Latest = tmp
}
func (o *ChainAdapter) checkMmrRootForFirst(root common.Hash) error {
	if len(o.Latest) > 0 {
		l := o.Latest[len(o.Latest)-1]
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

type ChainProofs struct {
	Remote      *ChainAdapter `json:"remote"     rlp:"nil"`
	HeaderProof *ChainHeaderProof
	InfoProof   *ChainInfoProof
}

func NewChainProofs() *ChainProofs {
	return &ChainProofs{
		HeaderProof: newChainHeaderProof(),
		InfoProof:   newChainInfoProof(),
	}
}

func (cps *ChainProofs) Data() ([]byte, error) {
	data, err := rlp.EncodeToBytes(cps)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (cps *ChainProofs) checkMmrRoot() error {
	if cps.HeaderProof != nil && cps.InfoProof != nil {
		fRoot, sRoot := cps.HeaderProof.Proof.RootHash, cps.InfoProof.Proof.RootHash
		if !bytes.Equal(fRoot[:], sRoot[:]) {
			fmt.Println("mmr root not match for second proof,first:", hex.EncodeToString(fRoot[:]), "second:", hex.EncodeToString(sRoot[:]))
			return errors.New("mmr root not match for second proof")
		}
		return nil
	}
	return errors.New("invalid params in checkMmrRoot")
}

func (cps *ChainProofs) Verify() error {

	if pBlocks, err := VerifyRequiredBlocks(cps.HeaderProof.Proof, cps.HeaderProof.Right); err != nil {
		return err
	} else {
		if !cps.HeaderProof.Proof.VerifyProof(pBlocks) {
			return errors.New("Verify Proof Failed on first msg")
		} else {
			if err := cps.Remote.GenesisCheck(cps.HeaderProof.Header[0]); err != nil {
				return err
			}
			if err := cps.Remote.checkAndSetHeaders(cps.HeaderProof.Header, false); err != nil {
				return err
			}
			if err := cps.Remote.checkMmrRootForFirst(cps.HeaderProof.Proof.RootHash); err != nil {
				return err
			}
			if pBlocks, err := VerifyRequiredBlocks2(cps.InfoProof.Proof); err != nil {
				return err
			} else {
				if !cps.InfoProof.Proof.VerifyProof2(pBlocks) {
					return errors.New("Verify Proof2 Failed on first msg")
				}
				if err := cps.checkMmrRoot(); err != nil {
					return err
				}
				// check headers
				return cps.Remote.checkAndSetHeaders(cps.InfoProof.Header, true)
			}
		}
	}
	return nil
}

type ReceiptProof struct { // describes all responses, not just a single one
	Proofs      NodeList
	Index       uint64
	ReceiptHash common.Hash
}

func (r *ReceiptProof) Verify() (*types.Receipt, error) {
	keybuf := new(bytes.Buffer)
	keybuf.Reset()
	rlp.Encode(keybuf, r.Index)
	value, err := trie.VerifyProof(r.ReceiptHash, keybuf.Bytes(), r.Proofs.NodeSet())
	if err != nil {
		return nil, err
	}

	var receipt *types.Receipt
	if err := rlp.DecodeBytes(value, &receipt); err != nil {
		return nil, err
	}

	return receipt, err
}

// newBlockData is the network packet for the block propagation message.
type MapProofs struct {
	ChainProof   *ChainProofs
	ReceiptProof *ReceiptProof
	End          *big.Int
	Header       *types.Header
	Result       bool
	TxHash       common.Hash
}

// UlvpTransaction is the network packet for the block propagation message.
type MapTransaction struct {
	SimpUlvpP *MapProofs
	Tx        *types.Transaction
}

func (mr *MapProofs) VerifyMapTransaction(txHash common.Hash) (*types.Receipt, error) {
	if !mr.Result {
		return nil, errors.New("no proof return")
	}
	if err := mr.ChainProof.Verify(); err != nil {
		return nil, err
	}

	if mr.ChainProof.Remote.ProofHeader.Number.Uint64() != mr.Header.Number.Uint64() {
		return nil, errors.New("mmr proof not match receipt proof")
	}
	receipt, err := mr.ReceiptProof.Verify()
	if err != nil {
		return nil, err
	}

	//if !reflect.DeepEqual(receipt.Bloom, mr.ReceiptProof.Receipt.Bloom) {
	//	return nil, errors.New("receipt Bloom proof not match receipt")
	//}
	//
	//if !reflect.DeepEqual(receipt.Logs, mr.ReceiptProof.Receipt.Logs) {
	//	return nil, errors.New("receipt Logs proof not match receipt")
	//}
	//
	//if !reflect.DeepEqual(receipt.CumulativeGasUsed, mr.ReceiptProof.Receipt.CumulativeGasUsed) {
	//	return nil, errors.New("receipt Logs proof not match receipt")
	//}

	if mr.TxHash != txHash {
		return nil, errors.New("txHash checkout failed")
	}
	return receipt, nil
}

func MapVerify(proof []byte, txHash common.Hash) error {
	su := &MapProofs{}
	if err := rlp.DecodeBytes(proof, su); err != nil {
		return err
	}
	_, err := su.VerifyMapTransaction(txHash)
	return err
}
