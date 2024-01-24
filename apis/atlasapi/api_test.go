package atlasapi

import (
	"fmt"

	"testing"

	"github.com/ethereum/go-ethereum/trie"
	"github.com/mapprotocol/atlas/core/types"
)

func Test01(t *testing.T) {
	EmptyRootHash0 := types.DeriveSha(types.Transactions{}, trie.NewStackTrie(nil))
	fmt.Println(EmptyRootHash0)
}
