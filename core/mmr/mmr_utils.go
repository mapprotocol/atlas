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
	"encoding/binary"
	"math"
	"math/big"
	"math/bits"
	"sort"

	"github.com/ethereum/go-ethereum/common"
)

func countZore(num uint64) int {
	return bits.UintSize - bits.OnesCount64(num)
}

func leadingZeros(num uint64) int {
	return bits.LeadingZeros64(num)
}
func allOnes(num uint64) bool {
	return num != 0 && countZore(num) == leadingZeros(num)
}
func jumpLeft(pos uint64) uint64 {
	bitLength := 64 - leadingZeros(pos)
	mostSignificantBits := uint64(1) << uint64(bitLength-1)
	return pos - (mostSignificantBits - 1)
}

func posHeightInTree(pos uint64) int {
	pos += 1
	for {
		if !allOnes(pos) {
			pos = jumpLeft(pos)
		} else {
			break
		}
	}
	return 64 - leadingZeros(pos) - 1
}

func parentOffset(height int) uint64 {
	return uint64(2) << uint64(height)
}

func siblingOffset(height int) uint64 {
	return (uint64(2) << uint64(height)) - 1
}

func merge(left, right *Node) *Node {
	parent := &Node{}
	hashes := make([]common.Hash, 0, 0)
	hashes = append(append(hashes, left.GetHash()), right.GetHash())
	parent.setHash(RlpHash(hashes))
	parent.setDifficulty(new(big.Int).Add(left.getDifficulty(), right.getDifficulty()))
	parent.setIndex(right.getIndex() + 1)
	return parent
}
func merge2(left, right common.Hash) common.Hash {
	hashes := make([]common.Hash, 0, 0)
	hashes = append(append(hashes, left), right)
	return RlpHash(hashes)
}

func leftPeakPosByHeight(height int) uint64 {
	return (uint64(1) << uint64(height+1)) - 2
}

func leftPeakHeightPos(MmrSize uint64) (int, uint64) {
	height := 0
	prev_pos := uint64(0)
	pos := leftPeakPosByHeight(height)
	//increase height and get most left pos of tree
	//once pos is out of Mmr_size we consider previous pos is left peak
	for {
		if pos >= MmrSize {
			break
		}
		height += 1
		prev_pos = pos
		pos = leftPeakPosByHeight(height)
	}
	return height - 1, prev_pos
}

func getRightPeak(height int, peakPos, MmrSize uint64) (int, uint64) {
	//jump to right sibling
	peakPos += siblingOffset(height)
	//jump to left child
	for {
		if peakPos <= MmrSize-1 {
			break
		}
		if height == 0 {
			//no right peak exists
			return height, 0
		}
		height -= 1
		peakPos -= parentOffset(height)
	}
	return height, peakPos
}

func getPeaks(MmrSize uint64) []uint64 {
	res := make([]uint64, 0, 0)
	height, pos := leftPeakHeightPos(MmrSize)
	res = append(res, pos)
	for {
		if height <= 0 {
			break
		}
		height, pos = getRightPeak(height, pos, MmrSize)
		if height == 0 && pos == 0 {
			break
		}
		res = append(res, pos)
	}
	return res
}
func posInPeaks(pos uint64, peaks []uint64) bool {
	for _, v := range peaks {
		if v == pos {
			return true
		}
	}
	return false
}
func IsPowerOfTwo(n uint64) bool {
	return n > 0 && ((n & (n - 1)) == 0)
}
func NextPowerOfTwo(n uint64) uint64 {
	if n == 0 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++
	return n
}
func GetNodeFromLeaf(ln uint64) uint64 {
	position, remaining := uint64(0), ln
	for {
		if remaining == 0 {
			break
		}
		leftTreeLeafNumber := remaining
		if !IsPowerOfTwo(remaining) {
			leftTreeLeafNumber = NextPowerOfTwo(remaining) / 2
		}
		position += leftTreeLeafNumber + leftTreeLeafNumber - 1
		remaining = remaining - leftTreeLeafNumber
	}
	return position
}

// calculate logarithm of x for base b:
//
// y = log_2(x)/log_2(b)
//
func logBOfX(b, x float64) float64 {
	return math.Log2(x) / math.Log2(b)
}

// calculate how many independent queries m are required to have the specified security of lambda
// and always check the last specified block difficulty manually in variable difficulty setting
func vdCalculateM(lambda, c, block_difficulty, total_difficulty float64, n uint64) float64 {
	numerator := -lambda - math.Log2(c*float64(n))

	x := 1.0 - (1.0 / (logBOfX(c, block_difficulty/total_difficulty)))
	// x is not allowed to be negative
	if big.NewFloat(x).Sign() == -1 {
		x = 0.0
	}
	denumerator := math.Log2(x)
	return numerator / denumerator
}

// delta in variable difficulty setting is the sum of difficulty checked with probability 1 in the
// end
func vdCalculateDelta(block_difficulty, total_difficulty float64) float64 {
	return block_difficulty / total_difficulty
}

//
//             y(ln(delta))
// f(y) = 1 - e
//
// The cdf takes into account, that the last delta blocks are manually checked
func cdf(y, delta float64) float64 {
	return 1.0 - math.Exp(y*math.Log(delta))
}

func ByteToFloat64(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	return math.Float64frombits(bits)
}
func Uint64ToBytes(i uint64) []byte {
	var buf = make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, i)
	return buf
}

// extract bits from the hash value to generate an f64 value
// this is used for extracting randomness from hash values
// https://news.ycombinator.com/item?id=9207874
// http://www.math.sci.hiroshima-u.ac.jp/~m-mat/MT/ARTICLES/dSFMT.pdf
func HashToF64(x common.Hash) float64 {
	bytes := x[:]
	bytes[0] = byte(63)             // bitmuster to ensure that f64 is between 0 and 1
	bytes[1] = bytes[1] | byte(240) // bitmuster to ensure that f64 is between 0 and 1
	b1 := bytes[:8]
	u1 := binary.BigEndian.Uint64(b1)
	return ByteToFloat64(Uint64ToBytes(u1)) - 1.0
}

//////////////////////////////////////////////////////////////////
func SortAndRemoveRepeatForBlocks(slc []uint64) []uint64 {
	sort.Slice(slc, func(i, j int) bool {
		return slc[i] < slc[j]
	})

	result := []uint64{}
	tempMap := map[uint64]byte{}

	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l {
			result = append(result, e)
		}
	}
	return result
}
func SortAndRemoveRepeatForProofBlocks(slc []*ProofBlock) []*ProofBlock {
	nodes := ProofBlocks(slc)
	sort.Sort(nodes)
	result := []*ProofBlock{}
	var e *ProofBlock

	for _, v := range nodes {
		if e == nil {
			e = v
		}
		if !e.equal(v) {
			result = append(result, e)
			e = v
		}
	}
	if e != nil {
		result = append(result, e)
	}
	return result
}
func reverseForProofBlocks(s []*ProofBlock) []*ProofBlock {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
func binarySearch(arr []uint64, key uint64) int {
	low, height := 0, len(arr)-1
	mid := 0
	for low <= height {
		mid = (height + low) / 2
		if key == arr[mid] {
			return mid
		} else if key > arr[mid] {
			low = mid + 1
			if low > height {
				return low
			}
		} else {
			height = mid - 1
		}
	}
	return mid
}

func splitAt(sli []uint64, pos int) ([]uint64, []uint64) {
	if pos > len(sli) {
		panic("out of slice")
	}
	return sli[:pos], sli[pos:]
}

// Get depth of the Mmr with a specified leaf_number
func getDepth(leaf_number uint64) int {
	depth := 64 - leadingZeros(leaf_number) - 1
	if !IsPowerOfTwo(leaf_number) {
		depth += 1
	}
	return depth
}

// calc leaf number from complete node number
func nodeToLeafNumber(node_number uint64) uint64 {
	return (node_number + 1) / 2
}

func leafToNodeNumber(leaf_number uint64) uint64 {
	return (2 * leaf_number) - 1
}

func getLeftLeafNumber(leaf_number uint64) uint64 {
	if IsPowerOfTwo(leaf_number) {
		return leaf_number / 2
	} else {
		return NextPowerOfTwo(leaf_number) / 2
	}
}
