package testutil

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/core/types"
)

// decapitalise makes a camel-case string which starts with a lower case character.
func decapitalise(input string) string {
	if len(input) == 0 {
		return input
	}

	goForm := abi.ToCamelCase(input)
	return strings.ToLower(goForm[:1]) + goForm[1:]
}

type mockStateDB struct {
	types.StateDB
	isContract func(common.Address) bool
}

func (msdb *mockStateDB) GetCodeSize(addr common.Address) int {
	if msdb.isContract(addr) {
		return 100
	}
	return 0
}

func (msdb *mockStateDB) Finalise(bool) {}
