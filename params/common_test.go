package params

import (
	"fmt"
	"github.com/celo-org/celo-blockchain/common"
	"testing"
)

//0x8b91d837e1684f7353d73b6197230894243cf869282f722841df96b441303f37
func Test_makeRegistryId(t *testing.T) {
	a := makeRegistryId("Election") // common.hash
	//a:= makeRegistryId("Validators") // common.hash
	fmt.Println(common.BytesToHash(a[:]).String())
	fmt.Println(common.ZeroAddress.String())
}
