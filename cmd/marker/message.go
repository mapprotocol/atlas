package main

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/marker/config"
	"github.com/mapprotocol/atlas/cmd/marker/mapprotocol"
	"math/big"
)

const (
	SolveType1 = "type1"
	SolveType2 = "type2"
	SolveType3 = "type3"
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
}

func NewMessage(messageType string, ch chan<- struct{}, cfg *config.Config, to common.Address, value *big.Int, abi *abi.ABI, abiMethod string, params ...interface{}) Message {
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
	}
}

//NewMessageRet need to handle return params
func NewMessageRet(messageType string, ch chan<- struct{}, cfg *config.Config, ret interface{}, to common.Address, value *big.Int, abi *abi.ABI, abiMethod string, params ...interface{}) Message {
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
	}
}
