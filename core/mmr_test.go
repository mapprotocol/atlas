package core

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
	"hash"
	"math"
	"math/big"
	"testing"
)

func IntToBytes(n int) []byte {
	data := int64(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

/////////////////////////////////////////////////////////////////////
type testHasher struct {
	hasher hash.Hash
}

func newHasher() *testHasher {
	return &testHasher{hasher: sha3.NewLegacyKeccak256()}
}
func (h *testHasher) Reset() {
	h.hasher.Reset()
}

func (h *testHasher) Update(key, val []byte) {
	h.hasher.Write(key)
	h.hasher.Write(val)
}

func (h *testHasher) Hash() common.Hash {
	return common.BytesToHash(h.hasher.Sum(nil))
}

func (h *testHasher) Prove(key []byte, fromLevel uint, proofDb mosdb.KeyValueWriter) error {
	return nil
}

/////////////////////////////////////////////////////////////////////
//func run_Mmr(count int, proof_pos uint64) {
//	m := NewMmr()
//	positions := make([]*Node, 0, 0)
//
//	for i := 0; i < count; i++ {
//		positions = append(positions, m.push(&Node{
//			value:      common.BytesToHash(IntToBytes(i)),
//			difficulty: big.NewInt(0),
//		}))
//	}
//	merkle_root := m.getRoot()
//	// proof
//	pos := positions[proof_pos].index
//	// generate proof for proof_elem
//	proof := m.genProof(pos)
//	// verify proof
//	result := proof.verify(merkle_root, pos, positions[proof_pos].getHash())
//	fmt.Println("result:", result)
//}
//func Test01(t *testing.T) {
//	run_Mmr(10000, 50)
//	fmt.Println("finish")
//}

func Test02(t *testing.T) {
	num := uint64(0)
	a := NextPowerOfTwo(num)
	b := float64(100)
	fmt.Println("b:", math.Log(b), "pos_height:", get_depth(6))
	fmt.Println("aa", a, "isPow:", IsPowerOfTwo(num), "GetNodeFromLeaf:", GetNodeFromLeaf(6))
}
func modify_slice(v []int) []int {
	fmt.Println("len(v):", len(v))
	v = append(v, 100)
	fmt.Println("len(v):", len(v))
	return v
}

func Test03(t *testing.T) {
	val := uint64(0x4029000000000000)
	fmt.Println("val:", val, "fval:", ByteToFloat64(Uint64ToBytes(val)))

	aa := [32]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	fmt.Println("aa", HashToF64(common.BytesToHash(aa[:])))
	fmt.Println("finish")
}

func Test05(t *testing.T) {
	mmr := NewMMR()
	for i := 0; i < 1500; i++ {
		mmr.Push(&Node{
			value:      BytesToHash(IntToBytes(i)),
			difficulty: big.NewInt(1000),
		})
	}
	right_difficulty := big.NewInt(1000)
	fmt.Println("leaf_number:", mmr.getLeafNumber(), "root_difficulty:", mmr.getRootDifficulty())
	proof, blocks, eblocks := mmr.CreateNewProof(right_difficulty)
	fmt.Println("blocks_len:", len(blocks), "blocks:", blocks, "eblocks:", len(eblocks))
	fmt.Println("proof:", proof)
	pBlocks, err := VerifyRequiredBlocks(proof, right_difficulty)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	b := proof.VerifyProof(pBlocks)
	fmt.Println("b:", b)
	fmt.Println("finish")
}
func TestO6(t *testing.T) {
	// for i := 10; i < 10; i++ {
	// 	test_O6(i)
	// }
	test_O6(1000000)
	fmt.Println("finish")
}
func test_O6(count int) {
	mmr := NewMMR()
	for i := 0; i < count; i++ {
		if i == 9999 {
			fmt.Println(i)
		}
		mmr.Push(&Node{
			value:      BytesToHash(IntToBytes(i)),
			difficulty: big.NewInt(1000),
		})
	}

	// fmt.Println(mmr.GetSize(), mmr.GetRootNode())
	// fmt.Println("last:", mmr.getNode(19990))
	// mmr.Pop()
	// r := mmr.Pop()
	// fmt.Println("last:", r)
	// fmt.Println(mmr.GetSize(), mmr.GetRootNode())
	right_difficulty := big.NewInt(1000)
	// fmt.Println("leaf_number:", mmr.getLeafNumber(), "root_difficulty:", mmr.GetRootDifficulty())
	proof, _, _ := mmr.CreateNewProof(right_difficulty)

	tmp := &ChainHeaderProofMsg{
		Proof:  proof,
		Header: nil,
		Right:  big.NewInt(100),
	}

	msg1 := &UlvpMsgRes{
		FirstRes: tmp,
		SecondRes: &ChainInProofMsg{
			Proof:  proof,
			Header: nil,
		},
	}

	if data1, err := rlp.EncodeToBytes(msg1); err != nil {
		fmt.Println("error", err)
	} else {
		fmt.Println("data1 len:", len(data1))
	}

	// Genesis:      types.NewBlock(&types.Header{Number: big.NewInt(int64(100))}, nil, nil, nil, newHasher()),
	// tmp2 := &OtherChainAdapter{
	// 	Genesis:      common.Hash{},
	// 	ConfirmBlock: nil,
	// 	ProofHeader:  nil,
	// 	ProofHeight:  4,
	// 	Leatest:      []*types.Header{},
	// }
	msg2 := &UlvpChainProof{
		Res: msg1,
	}
	if data2, err := rlp.EncodeToBytes(msg2); err != nil {
		fmt.Println("error", err)
	} else {
		msg5 := &UlvpChainProof{}
		if err := rlp.DecodeBytes(data2, msg5); err != nil {
			fmt.Println("msg5", msg5, "error", err)
		}
		fmt.Println("data2 len:", len(data2))
	}

	var data []rlp.RawValue
	data = append(data, []byte{1, 2})
	pReceipt := &ReceiptTrieResps{Proofs: data, Index: 1, ReceiptHash: common.Hash{}}

	msg3 := &SimpleUlvpProof{
		ChainProof:   msg2,
		ReceiptProof: pReceipt,
		End:          big.NewInt(120),
		Header:       &types.Header{},
		Result:       false,
		TxHash:       common.Hash{},
	}

	data3, err := rlp.EncodeToBytes(msg3)
	if err != nil {
		fmt.Println("error", err)
	}
	fmt.Println("data3 len:", len(data3))

	msg4 := &SimpleUlvpProof{}
	if err := rlp.DecodeBytes(data3, msg4); err != nil {
		fmt.Println("msg4", msg4, "error", err)
	}

	// data2,_ := msg1.Datas()
	// res1,_ := ParseUvlpMsgRes(data2)
	// if res1 == nil {
	// 	fmt.Println("error")
	// }

	// fmt.Println("blocks_len:", len(blocks), "blocks:", blocks, "eblocks:", len(eblocks))
	// fmt.Println("proof:", proof)
	pBlocks, err := VerifyRequiredBlocks(proof, right_difficulty)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	b := proof.VerifyProof(pBlocks)
	fmt.Println("b:", b)
	fmt.Println("finish:", count)
}

func TestO7(t *testing.T) {
	count := 20
	mmr := NewMMR()
	for i := 0; i < count; i++ {
		if i == 9999 {
			fmt.Println(i)
		}
		mmr.Push(&Node{
			value:      BytesToHash(IntToBytes(i)),
			difficulty: big.NewInt(1000),
			leafs:      uint64(i + 1),
			ds_diff:    big.NewInt(10),
			de_diff:    big.NewInt(10),
		})
	}
	right_difficulty := big.NewInt(3000)
	// fmt.Println("leaf_number:", mmr.getLeafNumber(), "root_difficulty:", mmr.GetRootDifficulty())
	proof, _, _ := mmr.CreateNewProof(right_difficulty)
	if proofBlock1, err := VerifyRequiredBlocks2(proof); err != nil {
		fmt.Println(err)
	} else {
		b := proof.VerifyProof2(proofBlock1)
		fmt.Println("b:", b)
	}

	proof2 := mmr.GenerateProof(15, 20)
	if proofBlock, err := VerifyRequiredBlocks2(proof2); err != nil {
		fmt.Println(err)
	} else {
		b := proof2.VerifyProof2(proofBlock)
		fmt.Println("b:", b)
	}

	if !bytes.Equal(proof.RootHash[:], proof2.RootHash[:]) {
		fmt.Println("RootHash NOT match,root1:", hex.EncodeToString(proof.RootHash[:]), "root2:", hex.EncodeToString(proof2.RootHash[:]))
	}

	fmt.Println("finish:", count)
}
func Test08(t *testing.T) {

}
