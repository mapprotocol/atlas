// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package backend

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/p2p"
)

type MockPeer struct {
	Messages     chan p2p.Msg
	NodeOverride *enode.Node
}

func (p *MockPeer) Send(msgcode uint64, data interface{}) error {
	return nil
}

func (p *MockPeer) Node() *enode.Node {
	if p.NodeOverride != nil {
		return p.NodeOverride
	}
	return enode.MustParse("enode://1dd9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f63270bcc9e1a6f6a439@127.0.0.1:52150")
}

func (p *MockPeer) Version() uint {
	return 0
}

func (p *MockPeer) ReadMsg() (p2p.Msg, error) {
	select {
	case msg := <-p.Messages:
		return msg, nil
	default:
		return p2p.Msg{}, nil
	}
}

func (p *MockPeer) Inbound() bool {
	return false
}

func (p *MockPeer) PurposeIsSet(purpose p2p.PurposeFlag) bool {
	return false
}

func TestIstanbulMessage(t *testing.T) {
	chain, backend := newBlockChain(1, true)
	defer chain.Stop()

	// generate one msg
	data := []byte("data1")
	msg := makeMsg(istanbul.QueryEnodeMsg, data)
	addr := common.BytesToAddress([]byte("address"))

	_, err := backend.HandleMsg(addr, msg, &MockPeer{})
	if err != nil {
		t.Fatalf("handle message failed: %v", err)
	}
}

func TestRecentMessageCaches(t *testing.T) {
	// Define the various voting scenarios to test
	tests := []struct {
		ethMsgCode  uint64
		shouldCache bool
	}{
		{
			ethMsgCode:  istanbul.ConsensusMsg,
			shouldCache: false,
		},
		{
			ethMsgCode:  istanbul.QueryEnodeMsg,
			shouldCache: true,
		},
		{
			ethMsgCode:  istanbul.ValEnodesShareMsg,
			shouldCache: false,
		},
		{
			ethMsgCode:  istanbul.FwdMsg,
			shouldCache: false,
		},
		{
			ethMsgCode:  istanbul.VersionCertificatesMsg,
			shouldCache: true,
		},
		{
			ethMsgCode:  istanbul.EnodeCertificateMsg,
			shouldCache: false,
		},
		{
			ethMsgCode:  istanbul.ValidatorHandshakeMsg,
			shouldCache: false,
		},
	}

	for _, tt := range tests {
		chain, backend := newBlockChain(1, true)

		// generate a msg that is not an Announce
		data := []byte("data1")
		msg := makeMsg(tt.ethMsgCode, data)
		addr := common.BytesToAddress([]byte("address"))

		// 1. this message should not be in cache
		// for peers
		if backend.gossipCache.CheckIfMessageProcessedByPeer(addr, data) {
			t.Fatalf("the cache of messages for this peer should be nil")
		}

		// for self
		if backend.gossipCache.CheckIfMessageProcessedBySelf(data) {
			t.Fatalf("the cache of messages should be nil")
		}

		// 2. this message should be in cache only when ethMsgCode == istanbulQueryEnodeMsg || ethMsgCode == istanbulVersionCertificatesMsg
		_, err := backend.HandleMsg(addr, msg, &MockPeer{})
		if err != nil {
			t.Fatalf("handle message failed: %v", err)
		}

		// Sleep for a bit, since some of the messages are handled in a different thread
		time.Sleep(10 * time.Second)

		// for peers
		if ok := backend.gossipCache.CheckIfMessageProcessedByPeer(addr, data); tt.shouldCache != ok {
			t.Fatalf("the cache of messages for this peer should be nil")
		}
		// for self
		if ok := backend.gossipCache.CheckIfMessageProcessedBySelf(data); tt.shouldCache != ok {
			t.Fatalf("the cache of messages must be nil")
		}

		chain.Stop()
	}
}

func TestReadValidatorHandshakeMessage(t *testing.T) {
	chain, backend := newBlockChain(2, true)
	defer chain.Stop()

	peer := &MockPeer{
		Messages:     make(chan p2p.Msg, 1),
		NodeOverride: backend.p2pserver.Self(),
	}

	// Test an empty message being sent
	emptyMsg := &istanbul.Message{}
	emptyMsgPayload, err := emptyMsg.Payload()
	if err != nil {
		t.Errorf("Error getting payload of empty msg %v", err)
	}
	peer.Messages <- makeMsg(istanbul.ValidatorHandshakeMsg, emptyMsgPayload)
	isValidator, err := backend.readValidatorHandshakeMessage(peer)
	if err != nil {
		t.Errorf("Error from readValidatorHandshakeMessage %v", err)
	}
	if isValidator {
		t.Errorf("Expected isValidator to be false with empty istanbul message")
	}

	var validMsg *istanbul.Message
	// The enodeCertificate is not set synchronously. Wait until it's been set
	for i := 0; i < 10; i++ {
		// Test a legitimate message being sent
		enodeCertMsg := backend.RetrieveEnodeCertificateMsgMap()[backend.SelfNode().ID()]
		if enodeCertMsg != nil {
			validMsg = enodeCertMsg.Msg
		}

		if validMsg != nil {
			break
		}
		time.Sleep(time.Duration(i) * time.Second)
	}
	if validMsg == nil {
		t.Errorf("enodeCertificate is nil")
	}

	validMsgPayload, err := validMsg.Payload()
	if err != nil {
		t.Errorf("Error getting payload of valid msg %v", err)
	}
	peer.Messages <- makeMsg(istanbul.ValidatorHandshakeMsg, validMsgPayload)

	block := backend.currentBlock()
	valSet := backend.getValidators(block.Number().Uint64(), block.Hash())
	// set backend to a different validator
	backend.wallets().Ecdsa.Address = valSet.GetByIndex(1).Address()

	isValidator, err = backend.readValidatorHandshakeMessage(peer)
	if err != nil {
		t.Errorf("Error from readValidatorHandshakeMessage with valid message %v", err)
	}
	if !isValidator {
		t.Errorf("Expected isValidator to be true with valid message")
	}
}

func makeMsg(msgcode uint64, data interface{}) p2p.Msg {
	size, r, _ := rlp.EncodeToReader(data)
	return p2p.Msg{Code: msgcode, Size: uint32(size), Payload: r}
}

/*
a=staking weight，b=working weight
P = a, 1-P = b    【P=0.6 / 0.7】
Vi = a * Si + b * Ai :
Vi = P * vote / TotalVote + (1-P)*(score/(totalscore -N*P)))

validator reward:
reward = Vi * totalReward
voters reward:
validator = reward * commision * score
*/
var (
	P           = float64(0.7) // stakingWeight
	commision   = 0.1
	BaseNum     = 100
	totalReward = toWei(big.NewFloat(33333))
)

func toCoin(val *big.Int) *big.Float {
	BaseBig := big.NewInt(1e18)
	return new(big.Float).Quo(new(big.Float).SetInt(val), new(big.Float).SetInt(BaseBig))
}
func toWei(value *big.Float) *big.Int {
	BaseBig := big.NewInt(1e18)
	base := new(big.Float).SetInt(BaseBig)
	val, _ := new(big.Float).Mul(value, base).Int(big.NewInt(0))
	return val
}
func calc_reward(score, totalscore *big.Int, voteAmount, totalVote *big.Int) (*big.Int, *big.Int) {
	v0 := new(big.Float).Quo(new(big.Float).SetInt(voteAmount), new(big.Float).SetInt(totalVote))
	//fmt.Println("---1 v0", v0, voteAmount, "/", totalVote)
	v0 = v0.Mul(big.NewFloat(P), v0)
	//fmt.Println("---2 v0", v0)

	v1 := new(big.Float).Quo(new(big.Float).SetInt(score), new(big.Float).SetInt(totalscore))
	//fmt.Println("---3 v1", v1, score, "/", totalscore)
	v1 = v1.Mul(big.NewFloat(1-P), v1)
	//fmt.Println("---4 v1", v1)

	v2 := v0.Add(v0, v1)
	//fmt.Println("---5 v2", v2)

	score0 := new(big.Float).Quo(new(big.Float).SetInt(score), big.NewFloat(float64(BaseNum)))
	reward := new(big.Float).Mul(v2, new(big.Float).SetInt(totalReward))

	reward0, _ := reward.Int(big.NewInt(0))
	//fmt.Println("---6 all_reward", toCoin(reward0), reward0)

	validator_reward := reward.Mul(reward, big.NewFloat(commision))
	validator_reward = validator_reward.Mul(validator_reward, score0)
	//fmt.Println("---7 val_reward commision", commision, "score0", score0)

	validator_reward0, _ := validator_reward.Int(big.NewInt(0))
	//fmt.Println("---8 val_reward", toCoin(validator_reward0), validator_reward0)

	return reward0, validator_reward0
}
func Test_newReward(t *testing.T) {
	num := 4
	totalScore := big.NewInt(0)
	totalVote := big.NewInt(0)
	scores := make([]*big.Int, num)
	stakings := make([]*big.Int, num)

	stakings[0], stakings[1] = big.NewInt(30499106), big.NewInt(25483631)
	stakings[2], stakings[3] = big.NewInt(25483631), big.NewInt(25483631)

	for i := 0; i < num; i++ {
		scores[i] = big.NewInt(100)
		//if i%3 == 0 {
		//	scores[i] = big.NewInt(0)
		//}
		totalScore = totalScore.Add(totalScore, scores[i])
		//stakings[i] = big.NewInt(int64(100 * (i + 1)))
		totalVote = totalVote.Add(totalVote, stakings[i])
	}
	all0 := big.NewInt(0)
	for i := 0; i < num; i++ {
		all, val := calc_reward(scores[i], totalScore, stakings[i], totalVote)
		fmt.Println("index ", i, "all", toCoin(all), "voters", toCoin(new(big.Int).Sub(all, val)), "validator", toCoin(val))
		all0 = all0.Add(all0, all)
	}
	fmt.Println("total reward:", all0, toCoin(all0))
}
func Test_01(t *testing.T) {
	//
	priv_hex, err := hex.DecodeString("")
	if err != nil {
		fmt.Println(err)
		return
	}
	priv, err := crypto.ToECDSA(priv_hex)
	if err != nil {
		fmt.Println(err)
		return
	}
	addr := crypto.PubkeyToAddress(priv.PublicKey)
	fmt.Println(addr.String())
}
