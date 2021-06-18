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
	// "encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

const (
	c      = float64(0.5)
	lambda = uint64(50)
)

func BytesToHash(b []byte) common.Hash {
	var a common.Hash
	a.SetBytes(b)
	return a
}
func RlpHash(x interface{}) (h common.Hash) {
	hw := sha3.New256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
func equal_hash(h1, h2 common.Hash) bool {
	return bytes.Equal(h1[:], h2[:])
}

type Node struct {
	value      common.Hash
	difficulty *big.Int
	timeCost   uint64
	ds_diff    *big.Int // difficulty of the block
	de_diff    *big.Int // difficulty of next block
	leafs      uint64
	index      uint64 // position in array
}

func NewNode(v common.Hash, d, ds, de *big.Int, timeCost uint64) *Node {
	return &Node{
		value:      v,
		difficulty: new(big.Int).Set(d),
		ds_diff:    ds,
		de_diff:    de,
		timeCost:   timeCost,
	}
}

func (n *Node) GetHash() common.Hash {
	return n.value
}
func (n *Node) setHash(h common.Hash) {
	n.value = h
}
func (n *Node) getDifficulty() *big.Int {
	return new(big.Int).Set(n.difficulty)
}
func (n *Node) setDifficulty(td *big.Int) {
	n.difficulty = new(big.Int).Set(td)
}
func (n *Node) setIndex(i uint64) {
	n.index = i
}
func (n *Node) getIndex() uint64 {
	return n.index
}
func (n *Node) setDS(ds *big.Int) {
	n.ds_diff = new(big.Int).Set(ds)
}
func (n *Node) getDS() *big.Int {
	return new(big.Int).Set(n.ds_diff)
}
func (n *Node) setDE(de *big.Int) {
	n.de_diff = new(big.Int).Set(de)
}
func (n *Node) getDE() *big.Int {
	return new(big.Int).Set(n.de_diff)
}
func (n *Node) setLeafs(ls uint64) {
	n.leafs = ls
}

func (n *Node) clone() *Node {
	return &Node{
		value:      n.value,
		difficulty: new(big.Int).Set(n.difficulty),
		index:      n.index,
		// ds_diff:	new(big.Int).Set(n.ds_diff),
		// de_diff:	new(big.Int).Set(n.de_diff),
		timeCost: n.timeCost,
		leafs:    n.leafs,
	}
}
func (n *Node) hasChildren(m *Mmr) bool {
	elem_node_number, curr_root_node_number, aggr_node_number := n.index, m.getSize(), uint64(0)
	for {
		if curr_root_node_number > 2 {
			leaf_number := nodeToLeafNumber(curr_root_node_number)
			left_tree_leaf_number := getLeftLeafNumber(leaf_number)
			left_tree_node_number := leafToNodeNumber(left_tree_leaf_number)
			if (aggr_node_number + curr_root_node_number) == (elem_node_number + 1) {
				return true
			}

			if elem_node_number < (aggr_node_number + left_tree_node_number) {
				// branch left
				curr_root_node_number = left_tree_node_number
			} else {
				// branch right
				curr_root_node_number = curr_root_node_number - left_tree_node_number - 1
				aggr_node_number += left_tree_node_number
			}
		} else {
			break
		}
	}
	return false
}
func (n *Node) getChildren(m *Mmr) (*Node, *Node) {
	elem_node_number, curr_root_node_number, aggr_node_number := n.index, m.getSize(), uint64(0)

	for {
		if curr_root_node_number > 2 {
			leaf_number := nodeToLeafNumber(curr_root_node_number)
			left_tree_leaf_number := getLeftLeafNumber(leaf_number)
			left_tree_node_number := leafToNodeNumber(left_tree_leaf_number)

			if (aggr_node_number + curr_root_node_number) == (elem_node_number + 1) {
				leaf_number = nodeToLeafNumber(curr_root_node_number)
				left_tree_leaf_number = getLeftLeafNumber(leaf_number)
				left_tree_node_number = leafToNodeNumber(left_tree_leaf_number)

				left_node_position := aggr_node_number + left_tree_node_number - 1
				right_node_position := aggr_node_number + curr_root_node_number - 2

				left_elem, right_elem := m.getNode(left_node_position), m.getNode(right_node_position)

				return left_elem, right_elem
			}

			if elem_node_number < (aggr_node_number + left_tree_node_number) {
				// branch left
				curr_root_node_number = left_tree_node_number
			} else {
				// branch right
				curr_root_node_number = curr_root_node_number - left_tree_node_number - 1
				aggr_node_number += left_tree_node_number
			}
		} else {
			break
		}
	}

	panic("This node has no children!")
}
func (n *Node) String() string {
	return fmt.Sprintf("{value:%s, index:%v,difficulty:%v}", n.value.Hex(), n.index, n.difficulty)
}
func (r *ProofRes) String() string {
	return fmt.Sprintf("{hash:%s, TD:%v}", r.H.Hex(), r.TD)
}
func (p *ProofElem) String() string {
	if p.Cat == 2 {
		return fmt.Sprintf("[Child,%s]", p.Res.String())
	} else if p.Cat == 1 {
		return fmt.Sprintf("[Node,%s,Right:%v]", p.Res.String(), p.Right)
	} else {
		return fmt.Sprintf("[Root,%s,LeafNum:%v]", p.Res.String(), p.LeafNum)
	}
}
func (p *ProofInfo) String() string {
	elems := make([]string, len(p.Elems))
	for i, v := range p.Elems {
		elems[i] = v.String()
	}
	return fmt.Sprintf("RootHash:%s \n,RootDiff:%v,LeafNum:%v \n,Elems:%s", p.RootHash.Hex(),
		p.RootDifficulty, p.LeafNumber, strings.Join(elems, "\n "))
}

/////////////////////////////////////////////////////////////////////////////////
type ProofRes struct {
	H  common.Hash
	TD *big.Int
}

type VerifyElem struct {
	Res        *ProofRes
	Index      uint64
	LeafNumber uint64
}

type ProofElem struct {
	Cat     uint8 // 0--root,1--node,2 --child
	Res     *ProofRes
	Right   bool
	LeafNum uint64
}

type ProofInfo struct {
	RootHash       common.Hash
	RootDifficulty *big.Int
	LeafNumber     uint64
	Elems          []*ProofElem
	Checked        []uint64
}

func ProofInfoToBytes(info *ProofInfo) ([]byte, error) {
	return rlp.EncodeToBytes(info)
}
func ProofInfoFromBytes(data []byte) (*ProofInfo, error) {
	obj := &ProofInfo{}
	err := rlp.DecodeBytes(data, obj)
	return obj, err
}

type ProofElems []*ProofElem

func (p *ProofElems) pop_back() *ProofElem {
	if len(*p) <= 0 {
		return nil
	}
	index := len(*p) - 1
	last := (*p)[index]
	*p = append((*p)[:index], (*p)[index+1:]...)
	return last
}
func (p *ProofElems) pop_front() *ProofElem {
	if len(*p) <= 0 {
		return nil
	}
	index := 0
	last := (*p)[index]
	*p = append((*p)[:index], (*p)[index+1:]...)
	return last
}
func (p *ProofElems) is_empty() bool {
	return len(*p) == 0
}

type VerifyElems []*VerifyElem

func (v *VerifyElems) pop_back() *VerifyElem {
	if len(*v) <= 0 {
		return nil
	}
	index := len(*v) - 1
	last := (*v)[index]
	*v = append((*v)[:index], (*v)[index+1:]...)
	return last
}
func (v *VerifyElems) is_empty() bool {
	return len(*v) == 0
}

type ProofBlock struct {
	Number     uint64
	AggrWeight float64
}

func (p *ProofBlock) equal(oth *ProofBlock) bool {
	if oth == nil || p == nil {
		return false
	}
	return p.Number == oth.Number
}

type ProofBlocks []*ProofBlock

func (p *ProofBlocks) pop() *ProofBlock {
	if len(*p) <= 0 {
		return nil
	}
	index := len(*p) - 1
	last := (*p)[index]
	*p = append((*p)[:index], (*p)[index+1:]...)
	return last
}
func (a ProofBlocks) Len() int           { return len(a) }
func (a ProofBlocks) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ProofBlocks) Less(i, j int) bool { return a[i].Number < a[j].Number }

type nodesAdapter []*Node

func (p *nodesAdapter) pop() *Node {
	if len(*p) <= 0 {
		return nil
	}
	index := len(*p) - 1
	last := (*p)[index]
	*p = append((*p)[:index], (*p)[index+1:]...)
	return last
}
func (p *nodesAdapter) push(n *Node) {
	*p = append(*p, n)
}

//////////////////////////////////////////////////////////////////////////////////////

type Mmr struct {
	values  []*Node
	curSize uint64 // unused
	leafNum uint64
	mu      *sync.Mutex
}

func NewMMR() *Mmr {
	return &Mmr{
		values:  make([]*Node, 0, 0),
		curSize: 0,
		leafNum: 0,
		mu:      new(sync.Mutex),
	}
}
func (m *Mmr) getNode(pos uint64) *Node {
	if int(pos) > int(len(m.values)-1) {
		return nil
	}
	return m.values[pos]
}
func (m *Mmr) getLeafNumber() uint64 {
	return m.leafNum
}
func (m *Mmr) pop() *Node {
	if m.leafNum <= 0 {
		return nil
	}
	peakNode, curr_tree_number, aggr_node_number := nodesAdapter(make([]*Node, 0, 0)), m.leafNum, uint64(0)
	for !IsPowerOfTwo(curr_tree_number) {
		m.removeLastElem()
		left_tree_number := NextPowerOfTwo(curr_tree_number) / 2
		aggr_node_number += left_tree_number
		right_tree_number := curr_tree_number - left_tree_number

		left_root_node_number := GetNodeFromLeaf(aggr_node_number) - 1
		peakNode.push(m.getNode(left_root_node_number))
		curr_tree_number = right_tree_number
	}
	peakNode.push(m.getRootNode())
	lastPeak := peakNode.pop()
	tmp, remove := m.splitPeak(lastPeak)
	if remove == nil {
		return nil
	}
	m.removeLastLeafNode(remove)
	peakNode = append(peakNode, tmp...)
	for len(peakNode) > 1 {
		right := peakNode.pop()
		left := peakNode.pop()
		parent := merge(left, right)
		m.values = append(m.values, parent)
		parent.index = uint64(len(m.values) - 1)
		peakNode.push(parent)
	}
	return remove
}
func (m *Mmr) Pop2() *Node {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pop()
}
func (m *Mmr) splitPeak(lastPeak *Node) ([]*Node, *Node) {
	peakNodes := nodesAdapter(make([]*Node, 0, 0))
	height := posHeightInTree(lastPeak.index)
	if height == 0 {
		return peakNodes, lastPeak
	}
	peaksize := (uint64(1) << uint64(height+1)) - 1
	offset := lastPeak.index + 1 - peaksize
	peaksize = peaksize - 1

	for peaksize > 0 {
		m.removeLastElem()
		height, pos := leftPeakHeightPos(peaksize)
		if height == 0 {
			peakNodes.push(m.getNode(pos + offset))
			return peakNodes, m.getNode(pos + offset + 1)
		}
		if height > 0 {
			leafPeak := m.getNode(pos + offset)
			peakNodes.push(leafPeak)
			height, pos = getRightPeak(height, pos, peaksize)
			rightPeak := m.getNode(pos + offset)
			peaksize = (uint64(1) << uint64(height+1)) - 1
			offset = rightPeak.index + 1 - peaksize
			peaksize = peaksize - 1
		}
	}
	return peakNodes, nil
}
func (m *Mmr) removeLastLeafNode(leaf *Node) {
	index := leaf.index
	m.values = append(m.values[:index], m.values[index+1:]...)
	m.leafNum = m.leafNum - 1
	return
}

func (m *Mmr) Push(newElem *Node) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.values) <= 0 {
		m.values, m.leafNum, m.curSize = append(m.values, newElem), 1, 1
		newElem.index = 0
	} else {
		nodes_to_hash, curr_tree_number, aggr_node_number := nodesAdapter(make([]*Node, 0, 0)), m.leafNum, uint64(0)

		for !IsPowerOfTwo(curr_tree_number) {
			m.removeLastElem()
			left_tree_number := NextPowerOfTwo(curr_tree_number) / 2
			aggr_node_number += left_tree_number
			right_tree_number := curr_tree_number - left_tree_number

			left_root_node_number := GetNodeFromLeaf(aggr_node_number) - 1
			nodes_to_hash.push(m.getNode(left_root_node_number))
			curr_tree_number = right_tree_number
		}
		nodes_to_hash.push(m.getRootNode())
		m.values = append(m.values, newElem)
		newElem.index = uint64(len(m.values) - 1)
		nodes_to_hash.push(newElem)

		for len(nodes_to_hash) > 1 {
			right := nodes_to_hash.pop()
			left := nodes_to_hash.pop()
			parent := merge(left, right)
			m.values = append(m.values, parent)
			parent.index = uint64(len(m.values) - 1)
			nodes_to_hash.push(parent)
		}
		m.leafNum += 1
	}
}
func (m *Mmr) removeLastElem() {
	if len(m.values) <= 0 {
		return
	}
	index := len(m.values) - 1
	m.values = append(m.values[:index], m.values[index+1:]...)
	return
}
func (m *Mmr) getRootNode() *Node {
	if len(m.values) <= 0 {
		return nil
	}
	return m.values[len(m.values)-1]
}
func (m *Mmr) getRoot() common.Hash {
	root := m.getRootNode()
	if root == nil {
		return common.Hash{0}
	} else {
		return root.GetHash()
	}
}
func (m *Mmr) GetRoot2() common.Hash {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getRoot()
}
func (m *Mmr) getSize() uint64 {
	return uint64(len(m.values))
}
func (m *Mmr) getRootDifficulty() *big.Int {
	root := m.getRootNode()
	if root == nil {
		return nil
	} else {
		return root.getDifficulty()
	}
}
func (m *Mmr) getChildByAggrWeightDisc(weight *big.Int) uint64 {
	AggrWeight, aggr_node_number, curr_tree_number := big.NewInt(0), uint64(0), m.leafNum
	for {
		if curr_tree_number > 1 {
			left_tree_number := curr_tree_number / 2
			if !IsPowerOfTwo(curr_tree_number) {
				left_tree_number = NextPowerOfTwo(curr_tree_number) / 2
			}
			n := m.getNode(GetNodeFromLeaf(aggr_node_number+left_tree_number) - 1)
			if n == nil {
				panic("wrong pos1")
			}
			left_tree_difficulty := n.getDifficulty()
			if weight.Cmp(new(big.Int).Add(AggrWeight, left_tree_difficulty)) >= 0 {
				// branch right
				aggr_node_number += left_tree_number
				left_root_node_number := GetNodeFromLeaf(aggr_node_number) - 1
				n1 := m.getNode(left_root_node_number)
				if n1 == nil {
					panic("wrong pos2")
				}
				AggrWeight = new(big.Int).Add(AggrWeight, n1.getDifficulty())
				curr_tree_number = curr_tree_number - left_tree_number
			} else {
				// branch left
				curr_tree_number = left_tree_number
			}
		} else {
			break
		}
	}
	return aggr_node_number
}
func (m *Mmr) getChildByAggrWeight(weight float64) uint64 {
	root_weight := m.getRootDifficulty()
	v1, _ := new(big.Float).Mul(new(big.Float).SetInt(root_weight), big.NewFloat(weight)).Int64()
	weight_disc := big.NewInt(v1)
	return m.getChildByAggrWeightDisc(weight_disc)
}

func (m *Mmr) Copy() *Mmr {
	tmp := NewMMR()
	m.mu.Lock()
	defer m.mu.Unlock()

	tmp.curSize = m.curSize
	tmp.leafNum = m.leafNum

	for _, n := range m.values {
		tmp.values = append(tmp.values, n.clone())
	}
	return tmp
}

///////////////////////////////////////////////////////////////////////////////////////

func generateProofRecursive(currentNode *Node, blocks []uint64, proofs []*ProofElem,
	max_left_tree_leaf_number uint64, startDepth int, leaf_number_sub_tree uint64, space uint64,
	m *Mmr) []*ProofElem {
	if !currentNode.hasChildren(m) {
		proofs = append(proofs, &ProofElem{
			Cat:     2,
			Right:   false,
			LeafNum: 0,
			Res: &ProofRes{
				H:  currentNode.GetHash(),
				TD: currentNode.getDifficulty(),
			},
		})
		return proofs
	}
	left_node, right_node := currentNode.getChildren(m)
	pos := binarySearch(blocks, max_left_tree_leaf_number)
	left, right := splitAt(blocks, pos)
	next_left_leaf_number_subtree := getLeftLeafNumber(leaf_number_sub_tree)
	if len(left) != 0 {
		depth := getDepth(next_left_leaf_number_subtree)
		diff := uint64(0)
		if depth >= 1 {
			diff = uint64(math.Pow(float64(2), float64(depth-1)))
		}
		proofs = generateProofRecursive(left_node, left, proofs,
			max_left_tree_leaf_number-diff,
			startDepth, next_left_leaf_number_subtree,
			space+1, m)
	} else {
		proofs = append(proofs, &ProofElem{
			Cat:     1,
			Right:   false,
			LeafNum: 0,
			Res: &ProofRes{
				H:  left_node.GetHash(),
				TD: left_node.getDifficulty(),
			},
		})
	}
	if len(right) != 0 {
		depth := getDepth(leaf_number_sub_tree - next_left_leaf_number_subtree)
		diff := uint64(0)
		if depth >= 1 {
			diff = uint64(math.Pow(float64(2), float64(depth-1)))
		}
		proofs = generateProofRecursive(right_node, right, proofs,
			max_left_tree_leaf_number+diff, startDepth,
			leaf_number_sub_tree-next_left_leaf_number_subtree,
			space+1, m)
	} else {
		proofs = append(proofs, &ProofElem{
			Cat:     1,
			Right:   true,
			LeafNum: 0,
			Res: &ProofRes{
				H:  right_node.GetHash(),
				TD: right_node.getDifficulty(),
			},
		})
	}
	return proofs
}

func (m *Mmr) genProof(right_difficulty *big.Int, blocks []uint64) *ProofInfo {
	blocks = SortAndRemoveRepeatForBlocks(blocks)
	proofs, rootNode, depth := []*ProofElem{}, m.getRootNode(), getDepth(m.getLeafNumber())
	max_leaf_num := uint64(math.Pow(float64(2), float64(depth-1)))
	proofs = generateProofRecursive(rootNode, blocks, proofs, max_leaf_num, depth,
		m.getLeafNumber(), 0, m)

	proofs = append(proofs, &ProofElem{
		Cat:     0,
		Right:   false,
		LeafNum: m.getLeafNumber(),
		Res: &ProofRes{
			H:  rootNode.GetHash(),
			TD: rootNode.getDifficulty(),
		},
	})
	return &ProofInfo{
		RootHash:       m.getRoot(),
		RootDifficulty: m.getRootDifficulty(),
		LeafNumber:     m.getLeafNumber(),
		Elems:          proofs,
	}
}

func (m *Mmr) CreateNewProof(right_difficulty *big.Int) (*ProofInfo, []uint64, []uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	root_hash := m.getRoot()
	r1, _ := new(big.Float).SetInt(right_difficulty).Float64()
	r2, _ := new(big.Float).SetInt(new(big.Int).Add(m.getRootDifficulty(), right_difficulty)).Float64()
	required_queries := uint64(vdCalculateM(float64(lambda), c, r1, r2, m.getLeafNumber()) + 1.0)

	weights, blocks := []float64{}, []uint64{}
	for i := 0; i < int(required_queries); i++ {
		h := RlpHash([]interface{}{root_hash, uint64(i)})
		random := HashToF64(h)
		r3, _ := new(big.Float).SetInt(m.getRootDifficulty()).Float64()
		AggrWeight := cdf(random, vdCalculateDelta(r1, r3))
		weights = append(weights, AggrWeight)
	}
	sort.Float64s(weights)
	for _, v := range weights {
		b := m.getChildByAggrWeight(v)
		blocks = append(blocks, b)
	}
	// Pick up at specific sync point
	// Add extra blocks, which are used for syncing from an already available state
	// 1. block : first block of current 30_000 block interval
	// 2. block : first block of previous 30_000 block interval
	// 3. block : first block of third last 30_000 block interaval
	// 4. block : first block of fourth last 30_000 block interval
	// 5. block : first block of fiftf last 30_000 block interval
	// 6. block : first block of sixth last 30_000 block interval
	// 7. block : first block of seventh last 30_000 block interval
	// 8. block : first block of eighth last 30_000 block interval
	// 9. block : first block of ninth last 30_000 block interval
	// 10. block: first block of tenth last 30_000 block interval
	extra_blocks, current_block := []uint64{}, ((m.getLeafNumber()-1)/30000)*30000
	added := 0
	for {
		if current_block > 30000 && added < 10 {
			// blocks = append(blocks, current_block)
			extra_blocks = append(extra_blocks, current_block)
			current_block -= 30000
			added += 1
		} else {
			break
		}
	}

	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i] < blocks[j]
	})
	info := m.genProof(right_difficulty, blocks)
	info.Checked = blocks
	return info, blocks, extra_blocks
}

///////////////////////////////////////////////////////////////////////////////////////

func get_root(nodes []*VerifyElem) (common.Hash, *big.Int) {
	tmp := []*VerifyElem{}
	for _, v := range nodes {
		tmp = append(tmp, v)
	}
	tmp_nodes := VerifyElems(tmp)
	for {
		if len(tmp_nodes) > 1 {
			node2 := tmp_nodes.pop_back()
			node1 := tmp_nodes.pop_back()
			hash := merge2(node1.Res.H, node2.Res.H)
			tmp_nodes = append(tmp_nodes, &VerifyElem{
				Res: &ProofRes{
					H:  hash,
					TD: new(big.Int).Add(node1.Res.TD, node2.Res.TD),
				},
				Index:      math.MaxUint64, // uint64(-1) .. none
				LeafNumber: math.MaxUint64, // uint64(-1) .. none
			})
		} else {
			break
		}
	}
	if len(tmp_nodes) >= 1 {
		return tmp_nodes[0].Res.H, tmp_nodes[0].Res.TD
	}
	return common.Hash{0}, nil
}
func (p *ProofInfo) VerifyProof(blocks []*ProofBlock) bool {
	blocks = SortAndRemoveRepeatForProofBlocks(blocks)
	blocks = reverseForProofBlocks(blocks)
	proof_blocks := ProofBlocks(blocks)

	proofs := ProofElems(p.Elems)
	root_elem := proofs.pop_back()
	if root_elem == nil || root_elem.Cat != 0 {
		return false
	}
	if len(proofs) == 1 {
		if it := proofs.pop_back(); it != nil {
			if it.Cat == 2 {
				return equal_hash(it.Res.H, root_elem.Res.H)
			}
		}
		return false
	}
	nodes := VerifyElems([]*VerifyElem{})
	for {
		if !proofs.is_empty() {
			proof_elem := proofs.pop_front()
			if proof_elem.Cat == 2 {
				proof_block := proof_blocks.pop()
				number := proof_block.Number

				if !nodes.is_empty() {
					//TODO: Verification of previous MMR should happen here
					//weil in einem Mouse block header kein mmr hash vorhanden ist, kann man
					//dies nicht überprüfen, wenn doch irgendwann vorhanden, dann einfach
					//'block_header.mmr == old_root_hash' überprüfen
					_, left_difficulty := get_root(nodes)
					left, middle := new(big.Float).SetInt(left_difficulty), new(big.Float).Mul(new(big.Float).SetInt(root_elem.Res.TD), big.NewFloat(proof_block.AggrWeight))
					right := new(big.Float).Add(new(big.Float).SetInt(left_difficulty), new(big.Float).SetInt(proof_elem.Res.TD))
					if left.Cmp(middle) > 0 || right.Cmp(middle) <= 0 {
						// "aggregated difficulty is not correct, should coincide with: {} <= {} < {}",left, middle, right
						return false
					}
				}
				if number%2 == 0 && number != (root_elem.LeafNum-1) {
					right_node := proofs.pop_front()
					right_node_hash, right_node_diff := right_node.Res.H, new(big.Int).Set(right_node.Res.TD)
					if right_node.Cat == 2 || right_node.Cat == 1 {
						if right_node.Cat == 2 {
							proof_blocks.pop()
						}
					} else {
						// Expected ???
						return false
					}
					hash := merge2(proof_elem.Res.H, right_node_hash)
					nodes = append(nodes, &VerifyElem{
						Res: &ProofRes{
							H:  hash,
							TD: new(big.Int).Add(proof_elem.Res.TD, right_node_diff),
						},
						Index:      number / 2,
						LeafNumber: root_elem.LeafNum / 2,
					})
				} else {
					res0 := nodes.pop_back()
					hash := merge2(res0.Res.H, proof_elem.Res.H)
					nodes = append(nodes, &VerifyElem{
						Res: &ProofRes{
							H:  hash,
							TD: new(big.Int).Add(proof_elem.Res.TD, res0.Res.TD),
						},
						Index:      number / 2,
						LeafNumber: root_elem.LeafNum / 2,
					})
				}
			} else if proof_elem.Cat == 1 {
				if proof_elem.Right {
					left_node := nodes.pop_back()
					hash := merge2(left_node.Res.H, proof_elem.Res.H)
					nodes = append(nodes, &VerifyElem{
						Res: &ProofRes{
							H:  hash,
							TD: new(big.Int).Add(left_node.Res.TD, proof_elem.Res.TD),
						},
						Index:      left_node.Index / 2,
						LeafNumber: left_node.LeafNumber / 2,
					})
				} else {
					nodes = append(nodes, &VerifyElem{
						Res:        proof_elem.Res,
						Index:      math.MaxUint64, // UINT64(-1)
						LeafNumber: math.MaxUint64, // UINT64(-1)
					})
				}
			} else if proof_elem.Cat == 0 {
				// do nothing
			} else {
				panic("invalid Cat...")
			}
			for {
				if len(nodes) > 1 {
					node2 := nodes.pop_back()
					node1 := nodes.pop_back()
					if node2.Index == math.MaxUint64 || (node2.Index%2 != 1 && !proofs.is_empty()) {
						nodes = append(nodes, node1)
						nodes = append(nodes, node2)
						break
					}
					hash := merge2(node1.Res.H, node2.Res.H)
					nodes = append(nodes, &VerifyElem{
						Res: &ProofRes{
							H:  hash,
							TD: new(big.Int).Add(node1.Res.TD, node2.Res.TD),
						},
						Index:      node2.Index / 2,
						LeafNumber: node2.LeafNumber / 2,
					})
				} else {
					break
				}
			}
		} else {
			break
		}
	}

	res0 := nodes.pop_back()
	if res0 != nil {
		return equal_hash(root_elem.Res.H, res0.Res.H) && root_elem.Res.TD.Cmp(res0.Res.TD) == 0
	}
	return false
}
func (p *ProofInfo) VerifyProof2(blocks []*ProofBlock) bool {
	blocks = SortAndRemoveRepeatForProofBlocks(blocks)
	blocks = reverseForProofBlocks(blocks)
	proof_blocks := ProofBlocks(blocks)

	proofs := ProofElems(p.Elems)
	root_elem := proofs.pop_back()
	if root_elem == nil || root_elem.Cat != 0 {
		return false
	}
	if len(proofs) == 1 {
		if it := proofs.pop_back(); it != nil {
			if it.Cat == 2 {
				return equal_hash(it.Res.H, root_elem.Res.H)
			}
		}
		return false
	}
	nodes := VerifyElems([]*VerifyElem{})

	for {
		if !proofs.is_empty() {
			proof_elem := proofs.pop_front()
			if proof_elem.Cat == 2 {
				proof_block := proof_blocks.pop()
				number := proof_block.Number

				if number%2 == 0 && number != (root_elem.LeafNum-1) {
					right_node := proofs.pop_front()
					right_node_hash, right_node_diff := right_node.Res.H, new(big.Int).Set(right_node.Res.TD)
					if right_node.Cat == 2 || right_node.Cat == 1 {
						if right_node.Cat == 2 {
							proof_blocks.pop()
						}
					} else {
						// Expected ???
						return false
					}
					hash := merge2(proof_elem.Res.H, right_node_hash)
					nodes = append(nodes, &VerifyElem{
						Res: &ProofRes{
							H:  hash,
							TD: new(big.Int).Add(proof_elem.Res.TD, right_node_diff),
						},
						Index:      number / 2,
						LeafNumber: root_elem.LeafNum / 2,
					})
				} else {
					res0 := nodes.pop_back()
					hash := merge2(res0.Res.H, proof_elem.Res.H)
					nodes = append(nodes, &VerifyElem{
						Res: &ProofRes{
							H:  hash,
							TD: new(big.Int).Add(proof_elem.Res.TD, res0.Res.TD),
						},
						Index:      number / 2,
						LeafNumber: root_elem.LeafNum / 2,
					})
				}
			} else if proof_elem.Cat == 1 {
				if proof_elem.Right {
					left_node := nodes.pop_back()
					hash := merge2(left_node.Res.H, proof_elem.Res.H)
					nodes = append(nodes, &VerifyElem{
						Res: &ProofRes{
							H:  hash,
							TD: new(big.Int).Add(left_node.Res.TD, proof_elem.Res.TD),
						},
						Index:      left_node.Index / 2,
						LeafNumber: left_node.LeafNumber / 2,
					})
				} else {
					nodes = append(nodes, &VerifyElem{
						Res:        proof_elem.Res,
						Index:      math.MaxUint64, // UINT64(-1)
						LeafNumber: math.MaxUint64, // UINT64(-1)
					})
				}
			} else if proof_elem.Cat == 0 {
				// do nothing
			} else {
				panic("invalid Cat...")
			}
			for {
				if len(nodes) > 1 {
					node2 := nodes.pop_back()
					node1 := nodes.pop_back()
					if node2.Index == math.MaxUint64 || (node2.Index%2 != 1 && !proofs.is_empty()) {
						nodes = append(nodes, node1)
						nodes = append(nodes, node2)
						break
					}
					hash := merge2(node1.Res.H, node2.Res.H)
					nodes = append(nodes, &VerifyElem{
						Res: &ProofRes{
							H:  hash,
							TD: new(big.Int).Add(node1.Res.TD, node2.Res.TD),
						},
						Index:      node2.Index / 2,
						LeafNumber: node2.LeafNumber / 2,
					})
				} else {
					break
				}
			}
		} else {
			break
		}
	}

	res0 := nodes.pop_back()
	if res0 != nil {
		return equal_hash(root_elem.Res.H, res0.Res.H) && root_elem.Res.TD.Cmp(res0.Res.TD) == 0
	}
	return false
}
func VerifyRequiredBlocks(info *ProofInfo, right_difficulty *big.Int) ([]*ProofBlock, error) {
	blocks := info.Checked
	root_hash := info.RootHash
	root_difficulty := info.RootDifficulty
	root_leaf_number := info.LeafNumber
	r1, _ := new(big.Float).SetInt(right_difficulty).Float64()
	r2, _ := new(big.Float).SetInt(new(big.Int).Add(root_difficulty, right_difficulty)).Float64()
	required_queries := uint64(vdCalculateM(float64(lambda), c, r1, r2, root_leaf_number) + 1.0)
	extra_blocks, current_block := []uint64{}, ((root_leaf_number-1)/30000)*30000
	added := 0
	for {
		if current_block > 30000 && added < 10 {
			extra_blocks = append(extra_blocks, current_block)
			current_block -= 30000
			added += 1
		} else {
			break
		}
	}

	// required queries can contain the same block number multiple times
	// TODO: maybe multiple blocks can be pruned away?
	if required_queries != uint64(len(blocks)) {
		return nil, errors.New(fmt.Sprintf("false number of blocks provided: required: %v, got: %v", required_queries, len(blocks)))
	}
	weights := []float64{}
	for i := 0; i < int(required_queries); i++ {
		h := RlpHash([]interface{}{root_hash, uint64(i)})
		random := HashToF64(h)
		r3, _ := new(big.Float).SetInt(root_difficulty).Float64()
		AggrWeight := cdf(random, vdCalculateDelta(r1, r3))
		weights = append(weights, AggrWeight)
	}
	sort.Float64s(weights)
	proof_blocks, weight_pos := []*ProofBlock{}, 0

	for _, v := range blocks {
		AggrWeight := weights[weight_pos]
		// if len(extra_blocks) > 0 {
		// 	index := len(extra_blocks) - 1
		// 	curr_extra_block := extra_blocks[index]
		// 	if v == curr_extra_block {
		// 		extra_blocks = append(extra_blocks[:index], extra_blocks[index+1:]...)
		// 		AggrWeight = 0 // 0--none
		// 	} else {
		// 		weight_pos++
		// 	}
		// } else {
		// 	weight_pos++
		// }
		AggrWeight = weights[weight_pos]
		weight_pos++
		proof_blocks = append(proof_blocks, &ProofBlock{
			Number:     v,
			AggrWeight: AggrWeight,
		})
	}
	return proof_blocks, nil
}
func VerifyRequiredBlocks2(info *ProofInfo) ([]*ProofBlock, error) {
	blocks := info.Checked
	proof_blocks := []*ProofBlock{}
	for _, v := range blocks {
		proof_blocks = append(proof_blocks, &ProofBlock{
			Number: v,
		})
	}
	return proof_blocks, nil
}

///////////////////////////////////////////////////////////////////////////////////////

func (m *Mmr) GenerateProof(proofHeight, EndHeight uint64) *ProofInfo {
	// sort.Slice(blocks, func(i, j int) bool {
	// 	return blocks[i] < blocks[j]
	// })

	mmrClone := m.Copy()
	lRemove := int64(EndHeight - mmrClone.leafNum)
	if lRemove > 0 {
		for i := 0; i < int(lRemove); i++ {
			mmrClone.pop()
		}
	}

	info := mmrClone.genProof(big.NewInt(0), []uint64{proofHeight})
	info.Checked = []uint64{proofHeight}
	return info
}
func (m *Mmr) GenerateProof2(proofHeight, EndHeight uint64) *ProofInfo {

	info := m.genProof(big.NewInt(0), []uint64{proofHeight})
	info.Checked = []uint64{proofHeight}
	return info
}
