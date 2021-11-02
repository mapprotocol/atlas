package params

import (
	"fmt"
	"github.com/celo-org/celo-blockchain/common"
	"testing"
)

//0x8b91d837e1684f7353d73b6197230894243cf869282f722841df96b441303f37
func Test_makeRegistryId(t *testing.T) {
	a := makeRegistryId("Election")                // common.hash
	fmt.Println(common.BytesToHash(a[:]).String()) // [:] 的作用就是将数组转成切片 /
	fmt.Println(common.ZeroAddress.String())       // [:] 的作用就是将数组转成切片
}
