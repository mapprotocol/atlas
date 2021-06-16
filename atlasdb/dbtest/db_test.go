package dbtest

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mapprotocol/atlas/atlasdb"
	"github.com/mapprotocol/atlas/core"
	"github.com/mapprotocol/atlas/params"
	"math"
	"math/big"
	"testing"
)

// light
func Test(t *testing.T) {

	var (
		key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		addr   = crypto.PubkeyToAddress(key.PublicKey)
		gspec  = &core.Genesis{
			Config: params.TestChainConfig,
			Alloc:  core.GenesisAlloc{addr: {Balance: big.NewInt(math.MaxInt64)}},
		}
		m atlasdb.Mark
	)
	db := atlasdb.NewMemoryDatabase()
	for i := 1; i < 11; i++ {
		m = atlasdb.Mark(i)
		gspec.Commit(db, m)
	}
	for i := 1; i < 11; i++ {
		m = atlasdb.Mark(i)
		//atlasdb.ReadTd(db, hash, m, uint64(0))
		//atlasdb.ReadBlock(db, hash, m, uint64(0))
		//atlasdb.ReadReceipts(db, hash, m, uint64(0), nil)
		fmt.Println(atlasdb.ReadCanonicalHash(db, m, uint64(0)))
		fmt.Println(atlasdb.ReadHeadBlockHash(db, m))
		fmt.Println(atlasdb.ReadHeadFastBlockHash(db, m))
		fmt.Println(atlasdb.ReadHeadHeaderHash(db, m))
		//atlasdb.ReadChainConfig(db, hash)
	}

}
