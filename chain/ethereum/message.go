package ethereum

import (
	"bytes"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	etrie "github.com/ethereum/go-ethereum/trie"
	etypes "github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/rlp"
	"github.com/sirupsen/logrus"
	"github.com/snowfork/go-substrate-rpc-client/v2/types"
	"github.com/snowfork/polkadot-ethereum/relayer/chain"
	"github.com/snowfork/polkadot-ethereum/relayer/substrate"
)

func MakeMessageFromEvent(mapping map[common.Address]string, event *etypes.Log, receiptsTrie *etrie.Trie, log *logrus.Entry) (*chain.EthereumOutboundMessage, error) {
	// RLP encode event log's Address, Topics, and Data
	var buf bytes.Buffer
	err := event.EncodeRLP(&buf)
	if err != nil {
		return nil, err
	}

	receiptKey, err := rlp.EncodeToBytes(event.TxIndex)
	if err != nil {
		return nil, err
	}

	proof := substrate.NewProofData()
	err = receiptsTrie.Prove(receiptKey, 0, proof)
	if err != nil {
		return nil, err
	}

	m := substrate.Message{
		Data: buf.Bytes(),
		Proof: substrate.Proof{
			BlockHash: types.NewH256(event.BlockHash.Bytes()),
			TxIndex:   types.NewU32(uint32(event.TxIndex)),
			Data:      proof,
		},
	}

	value := hex.EncodeToString(m.Data)
	log.WithFields(logrus.Fields{
		"payload":    value,
		"blockHash":  m.Proof.BlockHash.Hex(),
		"eventIndex": m.Proof.TxIndex,
	}).Debug("Generated message from Ethereum log")

	var args []interface{}
	args = append(args, m)

	call, ok := mapping[event.Address]
	if !ok {
		return nil, err
	}

	message := chain.EthereumOutboundMessage{
		Call: call,
		Args: args,
	}

	return &message, nil
}
