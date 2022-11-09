package hash_test

import (
	"encoding/hex"
	"github.com/mapprotocol/atlas/chains/eth2/hash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHash(t *testing.T) {
	hashOf0 := [32]byte{110, 52, 11, 156, 255, 179, 122, 152, 156, 165, 68, 230, 187, 120, 10, 44, 120, 144, 29, 63, 179, 55, 56, 118, 133, 17, 163, 6, 23, 175, 160, 29}
	h := hash.Hash([]byte{0})
	assert.Equal(t, hashOf0, h)

	hashOf1 := [32]byte{75, 245, 18, 47, 52, 69, 84, 197, 59, 222, 46, 187, 140, 210, 183, 227, 209, 96, 10, 214, 49, 195, 133, 165, 215, 204, 226, 60, 119, 133, 69, 154}
	h = hash.Hash([]byte{1})
	assert.Equal(t, hashOf1, h)
	assert.Equal(t, false, hashOf0 == hashOf1)
}

func BenchmarkHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		hash.Hash([]byte("abc"))
	}
}

func TestHashKeccak256(t *testing.T) {
	hashOf0 := [32]byte{188, 54, 120, 158, 122, 30, 40, 20, 54, 70, 66, 41, 130, 143, 129, 125, 102, 18, 247, 180, 119, 214, 101, 145, 255, 150, 169, 224, 100, 188, 201, 138}
	h := hash.HashKeccak256([]byte{0})
	assert.Equal(t, hashOf0, h)

	hashOf1 := [32]byte{95, 231, 249, 119, 231, 29, 186, 46, 161, 166, 142, 33, 5, 123, 238, 187, 155, 226, 172, 48, 198, 65, 10, 163, 141, 79, 63, 190, 65, 220, 255, 210}
	h = hash.HashKeccak256([]byte{1})
	assert.Equal(t, hashOf1, h)
	assert.Equal(t, false, hashOf0 == hashOf1)

	// Same hashing test from go-ethereum for keccak256
	hashOfabc, err := hex.DecodeString("4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45")
	require.NoError(t, err)
	h = hash.HashKeccak256([]byte("abc"))
	h32 := hash.ToBytes32(hashOfabc)
	assert.Equal(t, h32, h)
}

func BenchmarkHashKeccak256(b *testing.B) {
	for i := 0; i < b.N; i++ {
		hash.HashKeccak256([]byte("abc"))
	}
}
