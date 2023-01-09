package cmd

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/new_marker/connections"
	"github.com/mapprotocol/atlas/cmd/new_marker/define"
	"github.com/mapprotocol/atlas/cmd/new_marker/writer"
	"math/big"
)

type base struct {
	msgCh chan struct{} // wait for msg handles
}

func newBase() *base {
	return &base{
		msgCh: make(chan struct{}),
	}
}

// waitUntilMsgHandled this function will block untill message is handled
func (b *base) waitUntilMsgHandled(counter int) {
	log.Debug("waitUntilMsgHandled", "counter", counter)
	for counter > 0 {
		<-b.msgCh
		counter -= 1
	}
}

func (b *base) newConn(addr string) *ethclient.Client {
	return connections.DialConn(addr)
}

func (b base) handleType1Msg(cfg *define.Config, to common.Address, value *big.Int, abi *abi.ABI, abiMethod string, params ...interface{}) {
	m := writer.NewMessage(writer.SolveSendTranstion1, b.msgCh, cfg, to, value, abi, abiMethod, params...)
	b.handleMessage(cfg.RPCAddr, m)
}

func (b base) handleType2Msg(cfg *define.Config, to common.Address, value *big.Int, abi *abi.ABI, abiMethod string, params ...interface{}) {
	m := writer.NewMessage(writer.SolveSendTranstion2, b.msgCh, cfg, to, value, abi, abiMethod, params...)
	b.handleMessage(cfg.RPCAddr, m)
}

func (b base) handleType3Msg(cfg *define.Config, ret interface{}, to common.Address, value *big.Int, abi *abi.ABI, abiMethod string, params ...interface{}) {
	m := writer.NewMessageRet1(writer.SolveQueryResult3, b.msgCh, cfg, &ret, to, value, abi, abiMethod, params...)
	b.handleMessage(cfg.RPCAddr, m)
}

func (b base) handleType4Msg(cfg *define.Config, solveResult func([]byte), to common.Address, value *big.Int, abi *abi.ABI, abiMethod string, params ...interface{}) {
	m := writer.NewMessageRet2(writer.SolveQueryResult4, b.msgCh, cfg, solveResult, to, value, abi, abiMethod, params...)
	b.handleMessage(cfg.RPCAddr, m)
}

func (b *base) handleMessage(rpcAddr string, msg writer.Message) {
	w := writer.New(b.newConn(rpcAddr))
	go w.ResolveMessage(msg)
	b.waitUntilMsgHandled(1)
}
