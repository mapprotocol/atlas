package eth2

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

var _ ILightClientUpdate = (*LightClientUpdateV1)(nil)

type LightClientUpdateV1 struct {
	// The beacon block header that is attested to by the sync committee
	attestedHeader *BeaconBlockHeader
	// Next sync committee corresponding to `attested_header`
	nextSyncCommittee       *SyncCommittee
	nextSyncCommitteeBranch [][]byte
	// The finalized beacon block header attested to by Merkle branch
	finalizedHeader    *BeaconBlockHeader
	finalityBranch     [][]byte
	finalizedExeHeader *types.Header
	exeFinalityBranch  [][]byte
	// Sync committee aggregate signature
	syncAggregate *SyncAggregate
	// Slot at which the aggregate signature was created (untrusted)
	signatureSlot uint64
}

func (l LightClientUpdateV1) GetAttestedHeader() *BeaconBlockHeader {
	return l.attestedHeader
}

func (l LightClientUpdateV1) GetNextSyncCommittee() *SyncCommittee {
	return l.nextSyncCommittee
}

func (l LightClientUpdateV1) GetNextSyncCommitteeBranch() [][]byte {
	return l.nextSyncCommitteeBranch
}

func (l LightClientUpdateV1) GetFinalizedHeader() *BeaconBlockHeader {
	return l.finalizedHeader
}

func (l LightClientUpdateV1) GetFinalityBranch() [][]byte {
	return l.finalityBranch
}

func (l LightClientUpdateV1) GetSyncAggregate() *SyncAggregate {
	return l.syncAggregate
}

func (l LightClientUpdateV1) GetSignatureSlot() uint64 {
	return l.signatureSlot
}

func (verify *ILightNodeLightClientVerify) toLightClientVerify() *LightClientVerify {
	return &LightClientVerify{
		update: verify.Update.toLightClientUpdate(),
		state:  verify.State.toLightClientState(),
	}
}
func (update *ILightNodeLightClientUpdateV1) toLightClientUpdate() *LightClientUpdateV1 {
	return &LightClientUpdateV1{
		attestedHeader:          update.AttestedHeader.toBeaconBlockHeader(),
		nextSyncCommittee:       update.NextSyncCommittee.toSyncCommittee(),
		nextSyncCommitteeBranch: bytes32ArrayToBytesArray(update.NextSyncCommitteeBranch),
		finalizedHeader:         update.FinalizedHeader.toBeaconBlockHeader(),
		finalityBranch:          bytes32ArrayToBytesArray(update.FinalityBranch),
		finalizedExeHeader:      update.FinalizedExeHeader.toBlockHeader(),
		exeFinalityBranch:       bytes32ArrayToBytesArray(update.ExeFinalityBranch),
		syncAggregate:           update.SyncAggregate.toSyncAggregate(),
		signatureSlot:           update.SignatureSlot,
	}
}

func (state *ILightNodeLightClientState) toLightClientState() *LightClientState {
	return &LightClientState{
		finalizedHeader:      state.FinalizedHeader.toBeaconBlockHeader(),
		currentSyncCommittee: state.CurrentSyncCommittee.toSyncCommittee(),
		nextSyncCommittee:    state.NextSyncCommittee.toSyncCommittee(),
		chainID:              state.ChainID,
	}
}

// ILightNodeLightClientState is an auto generated low-level Go binding around an user-defined struct.
type ILightNodeLightClientState struct {
	FinalizedHeader      ILightNodeBeaconBlockHeader
	CurrentSyncCommittee ILightNodeSyncCommittee
	NextSyncCommittee    ILightNodeSyncCommittee
	ChainID              uint64
}

// ILightNodeLightClientUpdate is an auto generated low-level Go binding around an user-defined struct.
type ILightNodeLightClientUpdateV1 struct {
	AttestedHeader          ILightNodeBeaconBlockHeader
	NextSyncCommittee       ILightNodeSyncCommittee
	NextSyncCommitteeBranch [][32]byte
	FinalizedHeader         ILightNodeBeaconBlockHeader
	FinalityBranch          [][32]byte
	FinalizedExeHeader      ILightNodeBlockHeader
	ExeFinalityBranch       [][32]byte
	SyncAggregate           ILightNodeSyncAggregate
	SignatureSlot           uint64
}

// ILightNodeLightClientVerify is an auto generated low-level Go binding around an user-defined struct.
type ILightNodeLightClientVerify struct {
	Update ILightNodeLightClientUpdateV1
	State  ILightNodeLightClientState
}

// ILightNodeBlockHeader is an auto generated low-level Go binding around an user-defined struct.
type ILightNodeBlockHeader struct {
	ParentHash       []byte
	Sha3Uncles       []byte
	Miner            common.Address
	StateRoot        []byte
	TransactionsRoot []byte
	ReceiptsRoot     []byte
	LogsBloom        []byte
	Difficulty       *big.Int
	Number           *big.Int
	GasLimit         *big.Int
	GasUsed          *big.Int
	Timestamp        *big.Int
	ExtraData        []byte
	MixHash          []byte
	Nonce            []byte
	BaseFeePerGas    *big.Int
}

func (header *ILightNodeBlockHeader) toBlockHeader() *types.Header {
	blockHeader := &types.Header{
		ParentHash:  common.BytesToHash(header.ParentHash),
		UncleHash:   common.BytesToHash(header.Sha3Uncles),
		Coinbase:    header.Miner,
		Root:        common.BytesToHash(header.StateRoot),
		TxHash:      common.BytesToHash(header.TransactionsRoot),
		ReceiptHash: common.BytesToHash(header.ReceiptsRoot),
		Bloom:       types.BytesToBloom(header.LogsBloom),
		Difficulty:  header.Difficulty,
		Number:      header.Number,
		GasLimit:    header.GasLimit.Uint64(),
		GasUsed:     header.GasUsed.Uint64(),
		Time:        header.Timestamp.Uint64(),
		Extra:       header.ExtraData,
		MixDigest:   common.BytesToHash(header.MixHash),
		Nonce:       types.BlockNonce{},
		BaseFee:     header.BaseFeePerGas,
	}

	copy(blockHeader.Nonce[:], header.Nonce)

	return blockHeader
}
