package writer

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/new_marker/define"
	"github.com/mapprotocol/atlas/cmd/new_marker/mapprotocol"
	"math/big"
)

const (
	SolveSendTranstion1 = "type1" // value ==nil
	SolveSendTranstion2 = "type2" // value !=nil
	SolveQueryResult3   = "type3" // handle Atomic
	SolveQueryResult4   = "type4" // handle Tuple
)

type Message struct {
	from        common.Address
	priKey      *ecdsa.PrivateKey
	value       *big.Int
	messageType string
	input       []byte
	abiMethod   string
	to          common.Address
	abi         *abi.ABI
	DoneCh      chan<- struct{}
	ret         interface{}
	solveResult func([]byte)
	gasLimit    uint64
}

func NewMessage(messageType string, ch chan<- struct{}, cfg *define.Config, to common.Address, value *big.Int, abi *abi.ABI, abiMethod string, params ...interface{}) Message {
	return Message{
		messageType: messageType,
		from:        cfg.From,
		priKey:      cfg.PrivateKey,
		to:          to,
		value:       value,
		abi:         abi,
		abiMethod:   abiMethod,
		input:       mapprotocol.PackInput(abi, abiMethod, params...),
		DoneCh:      ch,
		gasLimit:    uint64(cfg.GasLimit),
	}
}

//NewMessageRet1 need to handle return params
func NewMessageRet1(messageType string, ch chan<- struct{}, cfg *define.Config, ret interface{}, to common.Address, value *big.Int, abi *abi.ABI, abiMethod string, params ...interface{}) Message {
	return Message{
		messageType: messageType,
		from:        cfg.From,
		priKey:      cfg.PrivateKey,
		to:          to,
		value:       value,
		abi:         abi,
		abiMethod:   abiMethod,
		input:       mapprotocol.PackInput(abi, abiMethod, params...),
		DoneCh:      ch,
		ret:         ret,
		gasLimit:    uint64(cfg.GasLimit),
	}
}

//NewMessageRet2 need to handle return params
func NewMessageRet2(messageType string, ch chan<- struct{}, cfg *define.Config, solveResult func([]byte), to common.Address, value *big.Int, abi *abi.ABI, abiMethod string, params ...interface{}) Message {
	return Message{
		messageType: messageType,
		from:        cfg.From,
		priKey:      cfg.PrivateKey,
		to:          to,
		value:       value,
		abi:         abi,
		abiMethod:   abiMethod,
		input:       mapprotocol.PackInput(abi, abiMethod, params...),
		DoneCh:      ch,
		solveResult: solveResult,
		gasLimit:    uint64(cfg.GasLimit),
	}
}
