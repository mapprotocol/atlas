package eth2

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var _ ILightClientUpdate = (*LightClientUpdateV2)(nil)

type LightClientUpdateV2 struct {
	// The beacon block header that is attested to by the sync committee
	attestedHeader *BeaconBlockHeader
	// Next sync committee corresponding to `attested_header`
	nextSyncCommittee       *SyncCommittee
	nextSyncCommitteeBranch [][]byte
	// The finalized beacon block header attested to by Merkle branch
	finalizedHeader    *BeaconBlockHeader
	finalityBranch     [][]byte
	finalizedExecution *ExecutionPayload
	executionBranch    [][]byte
	// Sync committee aggregate signature
	syncAggregate *SyncAggregate
	// Slot at which the aggregate signature was created (untrusted)
	signatureSlot uint64
}

func (l LightClientUpdateV2) GetAttestedHeader() *BeaconBlockHeader {
	return l.attestedHeader
}

func (l LightClientUpdateV2) GetNextSyncCommittee() *SyncCommittee {
	return l.nextSyncCommittee
}

func (l LightClientUpdateV2) GetNextSyncCommitteeBranch() [][]byte {
	return l.nextSyncCommitteeBranch
}

func (l LightClientUpdateV2) GetFinalizedHeader() *BeaconBlockHeader {
	return l.finalizedHeader
}

func (l LightClientUpdateV2) GetFinalityBranch() [][]byte {
	return l.finalityBranch
}

func (l LightClientUpdateV2) GetSyncAggregate() *SyncAggregate {
	return l.syncAggregate
}

func (l LightClientUpdateV2) GetSignatureSlot() uint64 {
	return l.signatureSlot
}

// ILightNodeLightClientUpdateV2 is an auto generated low-level Go binding around an user-defined struct.
type ILightNodeLightClientUpdateV2 struct {
	AttestedHeader          ILightNodeBeaconBlockHeader
	NextSyncCommittee       ILightNodeSyncCommittee
	NextSyncCommitteeBranch [][32]byte
	FinalizedHeader         ILightNodeBeaconBlockHeader
	FinalityBranch          [][32]byte
	FinalizedExecution      ILightNodeExecution
	ExecutionBranch         [][32]byte
	SyncAggregate           ILightNodeSyncAggregate
	SignatureSlot           uint64
}

// ILightNodeExecution is an auto generated low-level Go binding around an user-defined struct.
type ILightNodeExecution struct {
	ParentHash       [32]byte
	FeeRecipient     common.Address
	StateRoot        [32]byte
	ReceiptsRoot     [32]byte
	LogsBloom        []byte
	PrevRandao       [32]byte
	BlockNumber      *big.Int
	GasLimit         *big.Int
	GasUsed          *big.Int
	Timestamp        *big.Int
	ExtraData        []byte
	BaseFeePerGas    *big.Int
	BlockHash        [32]byte
	TransactionsRoot [32]byte
	WithdrawalsRoot  [32]byte
}

func ConvertToLightClientVerify(update *ILightNodeLightClientUpdateV2,
	finalizedBeaconHeader *ILightNodeBeaconBlockHeader,
	curSyncCommittee *ILightNodeSyncCommittee,
	nextSyncCommittee *ILightNodeSyncCommittee,
	chainId uint64) *LightClientVerify {
	return &LightClientVerify{
		update: update.toLightClientUpdateV2(),
		state:  ConvertToLightClientState(finalizedBeaconHeader, curSyncCommittee, nextSyncCommittee, chainId),
	}
}

func (update *ILightNodeLightClientUpdateV2) toLightClientUpdateV2() *LightClientUpdateV2 {
	return &LightClientUpdateV2{
		attestedHeader:          update.AttestedHeader.toBeaconBlockHeader(),
		nextSyncCommittee:       update.NextSyncCommittee.toSyncCommittee(),
		nextSyncCommitteeBranch: bytes32ArrayToBytesArray(update.NextSyncCommitteeBranch),
		finalizedHeader:         update.FinalizedHeader.toBeaconBlockHeader(),
		finalityBranch:          bytes32ArrayToBytesArray(update.FinalityBranch),
		finalizedExecution:      update.FinalizedExecution.toExecutionPayload(),
		executionBranch:         bytes32ArrayToBytesArray(update.ExecutionBranch),
		syncAggregate:           update.SyncAggregate.toSyncAggregate(),
		signatureSlot:           update.SignatureSlot,
	}
}
func ConvertToLightClientState(
	finalizedBeaconHeader *ILightNodeBeaconBlockHeader,
	curSyncCommittee *ILightNodeSyncCommittee,
	nextSyncCommittee *ILightNodeSyncCommittee,
	chainId uint64) *LightClientState {
	return &LightClientState{
		finalizedHeader:      finalizedBeaconHeader.toBeaconBlockHeader(),
		currentSyncCommittee: curSyncCommittee.toSyncCommittee(),
		nextSyncCommittee:    nextSyncCommittee.toSyncCommittee(),
		chainID:              chainId,
	}
}
