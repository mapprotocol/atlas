package atlasapi

import (
	"fmt"
	//"github.com/ethereum/go-ethereum/common"
	//"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
	"testing"
	"github.com/mapprotocol/atlas/core/types"
)



func Test01(t *testing.T) {
	EmptyRootHash0 := types.DeriveSha(types.Transactions{},trie.NewStackTrie(nil))
	fmt.Println(EmptyRootHash0)
}


