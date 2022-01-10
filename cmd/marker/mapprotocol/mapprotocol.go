package mapprotocol

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/abi"
)

func PackInput(abi *abi.ABI, abiMethod string, params ...interface{}) []byte {
	input, err := abi.Pack(abiMethod, params...)
	if err != nil {
		log.Error(abiMethod, " error", err)
	}
	return input
}
