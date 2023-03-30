package eth2

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	fssz "github.com/prysmaticlabs/fastssz"
	ssz "github.com/prysmaticlabs/fastssz"
	"github.com/prysmaticlabs/go-bitfield"
	"google.golang.org/protobuf/runtime/protoimpl"
	"math/big"
)

const BLSPubkeyLength = 48

type ForkVersion [4]byte
type DomainType [4]byte
type Root []byte
type ValidatorIndex uint64

type ILightClientUpdate interface {
	GetAttestedHeader() *BeaconBlockHeader
	GetNextSyncCommittee() *SyncCommittee
	GetNextSyncCommitteeBranch() [][]byte
	GetFinalizedHeader() *BeaconBlockHeader
	GetFinalityBranch() [][]byte
	GetSyncAggregate() *SyncAggregate
	GetSignatureSlot() uint64
}

type LightClientState struct {
	// Beacon block header that is finalized
	finalizedHeader *BeaconBlockHeader

	// Sync committees corresponding to the header
	currentSyncCommittee *SyncCommittee
	nextSyncCommittee    *SyncCommittee
	chainID              uint64
}

type LightClientVerify struct {
	update ILightClientUpdate
	state  *LightClientState
}

type BeaconBlockHeader struct {
	Slot          uint64
	ProposerIndex ValidatorIndex
	ParentRoot    []byte
	StateRoot     []byte
	BodyRoot      []byte
}

type ExecutionPayload struct {
	ParentHash       common.Hash
	FeeRecipient     common.Address
	StateRoot        common.Hash
	ReceiptsRoot     common.Hash
	LogsBloom        []byte
	PrevRandao       common.Hash
	BlockNumber      *big.Int
	GasLimit         uint64
	GasUsed          uint64
	Timestamp        uint64
	ExtraData        []byte
	BaseFeePerGas    *big.Int
	BlockHash        common.Hash
	TransactionsRoot common.Hash
	WithdrawalsRoot  common.Hash
}

// HashTreeRoot ssz hashes the BeaconBlockHeader object
func (b *BeaconBlockHeader) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(b)
}

// HashTreeRootWith ssz hashes the BeaconBlockHeader object with a hasher
func (b *BeaconBlockHeader) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Slot'
	hh.PutUint64(b.Slot)

	// Field (1) 'ProposerIndex'
	hh.PutUint64(uint64(b.ProposerIndex))

	// Field (2) 'ParentRoot'
	if size := len(b.ParentRoot); size != 32 {
		err = ssz.ErrBytesLengthFn("--.ParentRoot", size, 32)
		return
	}
	hh.PutBytes(b.ParentRoot)

	// Field (3) 'StateRoot'
	if size := len(b.StateRoot); size != 32 {
		err = ssz.ErrBytesLengthFn("--.StateRoot", size, 32)
		return
	}
	hh.PutBytes(b.StateRoot)

	// Field (4) 'BodyRoot'
	if size := len(b.BodyRoot); size != 32 {
		err = ssz.ErrBytesLengthFn("--.BodyRoot", size, 32)
		return
	}
	hh.PutBytes(b.BodyRoot)

	if ssz.EnableVectorizedHTR {
		hh.MerkleizeVectorizedHTR(indx)
	} else {
		hh.Merkleize(indx)
	}
	return
}

type SyncCommittee struct {
	Pubkeys         [][]byte
	AggregatePubkey []byte
}
type SyncAggregate struct {
	SyncCommitteeBits      bitfield.Bitvector512
	SyncCommitteeSignature []byte
}

// HashTreeRoot ssz hashes the ExecutionPayload object
func (e *ExecutionPayload) HashTreeRoot() ([32]byte, error) {
	return fssz.HashWithDefaultHasher(e)
}

// HashTreeRootWith ssz hashes the ExecutionPayload object with a hasher
func (e *ExecutionPayload) HashTreeRootWith(hh *fssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'ParentHash'
	hh.PutBytes(e.ParentHash.Bytes())

	// Field (1) 'FeeRecipient'
	hh.PutBytes(e.FeeRecipient.Bytes())

	// Field (2) 'StateRoot'
	hh.PutBytes(e.StateRoot.Bytes())

	// Field (3) 'ReceiptsRoot'
	hh.PutBytes(e.ReceiptsRoot.Bytes())

	// Field (4) 'LogsBloom'
	if size := len(e.LogsBloom); size != 256 {
		err = fssz.ErrBytesLengthFn("--.LogsBloom", size, 256)
		return
	}
	hh.PutBytes(e.LogsBloom)

	// Field (5) 'PrevRandao'
	hh.PutBytes(e.PrevRandao.Bytes())

	// Field (6) 'BlockNumber'
	hh.PutUint64(e.BlockNumber.Uint64())

	// Field (7) 'GasLimit'
	hh.PutUint64(e.GasLimit)

	// Field (8) 'GasUsed'
	hh.PutUint64(e.GasUsed)

	// Field (9) 'Timestamp'
	hh.PutUint64(e.Timestamp)

	// Field (10) 'ExtraData'
	{
		elemIndx := hh.Index()
		byteLen := uint64(len(e.ExtraData))
		if byteLen > 32 {
			err = fssz.ErrIncorrectListSize
			return
		}
		hh.PutBytes(e.ExtraData)
		if fssz.EnableVectorizedHTR {
			hh.MerkleizeWithMixinVectorizedHTR(elemIndx, byteLen, (32+31)/32)
		} else {
			hh.MerkleizeWithMixin(elemIndx, byteLen, (32+31)/32)
		}
	}

	// Field (11) 'BaseFeePerGas'
	hh.PutBytes(PadTo(ReverseByteOrder(e.BaseFeePerGas.Bytes()), 32))

	// Field (12) 'BlockHash'
	hh.PutBytes(e.BlockHash.Bytes())

	// Field (13) 'Transactions'
	hh.PutBytes(e.TransactionsRoot.Bytes())

	// Field (14) 'WithdrawalsRoot'
	if !bytes.Equal(e.WithdrawalsRoot[:], make([]byte, 32)) {
		hh.PutBytes(e.WithdrawalsRoot.Bytes())
	}

	if fssz.EnableVectorizedHTR {
		hh.MerkleizeVectorizedHTR(indx)
	} else {
		hh.Merkleize(indx)
	}
	return
}

type ForkData struct {
	CurrentVersion        []byte
	GenesisValidatorsRoot []byte
}

func (f *ForkData) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'CurrentVersion'
	if size := len(f.CurrentVersion); size != 4 {
		err = ssz.ErrBytesLengthFn("--.CurrentVersion", size, 4)
		return
	}
	hh.PutBytes(f.CurrentVersion)

	// Field (1) 'GenesisValidatorsRoot'
	if size := len(f.GenesisValidatorsRoot); size != 32 {
		err = ssz.ErrBytesLengthFn("--.GenesisValidatorsRoot", size, 32)
		return
	}
	hh.PutBytes(f.GenesisValidatorsRoot)

	if ssz.EnableVectorizedHTR {
		hh.MerkleizeVectorizedHTR(indx)
	} else {
		hh.Merkleize(indx)
	}
	return
}

// HashTreeRoot ssz hashes the ForkData object
func (f *ForkData) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(f)
}

type SigningData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ObjectRoot []byte `protobuf:"bytes,1,opt,name=object_root,json=objectRoot,proto3" json:"object_root,omitempty" ssz-size:"32"`
	Domain     []byte `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty" ssz-size:"32"`
}

// HashTreeRoot ssz hashes the SigningData object
func (s *SigningData) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(s)
}

// HashTreeRootWith ssz hashes the SigningData object with a hasher
func (s *SigningData) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'ObjectRoot'
	if size := len(s.ObjectRoot); size != 32 {
		err = ssz.ErrBytesLengthFn("--.ObjectRoot", size, 32)
		return
	}
	hh.PutBytes(s.ObjectRoot)

	// Field (1) 'Domain'
	if size := len(s.Domain); size != 32 {
		err = ssz.ErrBytesLengthFn("--.Domain", size, 32)
		return
	}
	hh.PutBytes(s.Domain)

	if ssz.EnableVectorizedHTR {
		hh.MerkleizeVectorizedHTR(indx)
	} else {
		hh.Merkleize(indx)
	}
	return
}

// ILightNodeBeaconBlockHeader is an auto generated low-level Go binding around an user-defined struct.
type ILightNodeBeaconBlockHeader struct {
	Slot          uint64
	ProposerIndex uint64
	ParentRoot    [32]byte
	StateRoot     [32]byte
	BodyRoot      [32]byte
}

// ILightNodeSyncAggregate is an auto generated low-level Go binding around an user-defined struct.
type ILightNodeSyncAggregate struct {
	SyncCommitteeBits      []byte
	SyncCommitteeSignature []byte
}

// ILightNodeSyncCommittee is an auto generated low-level Go binding around an user-defined struct.
type ILightNodeSyncCommittee struct {
	Pubkeys         []byte
	AggregatePubkey []byte
}

func (header *ILightNodeBeaconBlockHeader) toBeaconBlockHeader() *BeaconBlockHeader {
	return &BeaconBlockHeader{
		Slot:          header.Slot,
		ProposerIndex: ValidatorIndex(header.ProposerIndex),
		ParentRoot:    header.ParentRoot[:],
		StateRoot:     header.StateRoot[:],
		BodyRoot:      header.BodyRoot[:],
	}
}

func (execution *ILightNodeExecution) toExecutionPayload() *ExecutionPayload {
	return &ExecutionPayload{
		ParentHash:       common.BytesToHash(execution.ParentHash[:]),
		FeeRecipient:     common.BytesToAddress(execution.FeeRecipient[:]),
		StateRoot:        common.BytesToHash(execution.StateRoot[:]),
		ReceiptsRoot:     common.BytesToHash(execution.ReceiptsRoot[:]),
		LogsBloom:        execution.LogsBloom,
		PrevRandao:       common.BytesToHash(execution.PrevRandao[:]),
		BlockNumber:      execution.BlockNumber,
		GasLimit:         execution.GasLimit.Uint64(),
		GasUsed:          execution.GasUsed.Uint64(),
		Timestamp:        execution.Timestamp.Uint64(),
		ExtraData:        execution.ExtraData,
		BaseFeePerGas:    execution.BaseFeePerGas,
		BlockHash:        common.BytesToHash(execution.BlockHash[:]),
		TransactionsRoot: common.BytesToHash(execution.TransactionsRoot[:]),
		WithdrawalsRoot:  common.BytesToHash(execution.WithdrawalsRoot[:]),
	}
}

func (syncCommittee *ILightNodeSyncCommittee) toSyncCommittee() *SyncCommittee {
	var pubkeys [][]byte
	count := len(syncCommittee.Pubkeys) / BLSPubkeyLength
	for i := 0; i < count; i++ {
		pubkeys = append(pubkeys, syncCommittee.Pubkeys[i*BLSPubkeyLength:(i+1)*BLSPubkeyLength])
	}
	return &SyncCommittee{
		Pubkeys:         pubkeys,
		AggregatePubkey: syncCommittee.AggregatePubkey,
	}
}
func (syncAggregate *ILightNodeSyncAggregate) toSyncAggregate() *SyncAggregate {
	return &SyncAggregate{
		SyncCommitteeBits:      syncAggregate.SyncCommitteeBits,
		SyncCommitteeSignature: syncAggregate.SyncCommitteeSignature,
	}
}

func bytes32ArrayToBytesArray(input [][32]byte) [][]byte {
	var output [][]byte
	for _, item := range input {
		itemCopy := item
		output = append(output, itemCopy[:])
	}

	return output
}
