package ethereum

import (
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/snowfork/go-substrate-rpc-client/v2/scale"
	"github.com/snowfork/go-substrate-rpc-client/v2/types"
)

type HeaderID struct {
	Number types.U64
	Hash   types.H256
}

type headerScale struct {
	ParentHash       types.H256
	Timestamp        types.U64
	Number           types.U64
	Author           types.H160
	TransactionsRoot types.H256
	OmmersHash       types.H256
	ExtraData        types.Bytes
	StateRoot        types.H256
	ReceiptsRoot     types.H256
	LogsBloom        types.Bytes256
	GasUsed          types.U256
	GasLimit         types.U256
	Difficulty       types.U256
	Seal             []types.Bytes
}

type Header struct {
	Fields headerScale
	header *etypes.Header
}

func (h *Header) Decode(decoder scale.Decoder) error {
	var fields headerScale
	err := decoder.Decode(&fields)
	if err != nil {
		return err
	}

	h.Fields = fields
	return nil
}

func (h Header) Encode(encoder scale.Encoder) error {
	return encoder.Encode(h.Fields)
}

func (h *Header) ID() HeaderID {
	return HeaderID{
		Number: h.Fields.Number,
		Hash:   types.NewH256(h.header.Hash().Bytes()),
	}
}

type DoubleNodeWithMerkleProof struct {
	DagNodes [2]types.H512
	Proof    [][16]byte
}
