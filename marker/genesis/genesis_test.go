package genesis

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mapprotocol/atlas/helper/decimal/fixed"
	"github.com/mapprotocol/atlas/params"
	"math/big"
	"testing"
)

//0x8b91d837e1684f7353d73b6197230894243cf869282f722841df96b441303f37
func Test_makeRegistryId(t *testing.T) {
	makeRegistryId := func(contractName string) [32]byte {
		hash := crypto.Keccak256([]byte(contractName))
		var id [32]byte
		copy(id[:], hash)
		return id
	}
	a := makeRegistryId("Election") // common.hash
	//a:= makeRegistryId("Validators") // common.hash
	fmt.Println(common.BytesToHash(a[:]).String())
	//fmt.Println(common.ZeroAddress.String())
	//fmt.Println(big.NewInt(0).Exp(big.NewInt(2),big.NewInt(4),nil))
}
func Test_fixed(t *testing.T) {
	perNumberReward := new(big.Int).Div(big.NewInt(300000000), big.NewInt(6000000))
	fmt.Println(perNumberReward.String())
	perEpochReward := new(big.Int).Mul(perNumberReward, big.NewInt(int64(params.Epoch))) //MAP
	fmt.Println(perEpochReward.String())
	perEpochReward = new(big.Int).Mul(perEpochReward, big.NewInt(1e18))
	fmt.Println(perEpochReward.String())

	fixed := fixed.MustNew
	fmt.Println(fixed("1").BigInt())                    // 10000 00000 00000 00000 00000  24
	fmt.Println(common.HexToAddress("0xd012").String()) // ValidatorsProxy
}
