package eth2

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

type LightClientUpdate struct {
	// The beacon block header that is attested to by the sync committee
	attestedHeader BeaconBlockHeader
	// Next sync committee corresponding to `attested_header`
	nextSyncCommittee       SyncCommittee
	nextSyncCommitteeBranch [][]byte
	// The finalized beacon block header attested to by Merkle branch
	finalizedHeader    BeaconBlockHeader
	finalityBranch     [][]byte
	finalizedExeHeader types.Header
	exeFinalityBranch  [][]byte
	// Sync committee aggregate signature
	syncAggregate SyncAggregate
	// Slot at which the aggregate signature was created (untrusted)
	signatureSlot uint64
}

type LightClientState struct {
	// Beacon block header that is finalized
	finalizedHeader BeaconBlockHeader

	// Sync committees corresponding to the header
	currentSyncCommittee SyncCommittee
	nextSyncCommittee    SyncCommittee
	chainID              uint64
}

type LightClientVerify struct {
	update *LightClientUpdate
	state  *LightClientState
}

type BeaconBlockHeader struct {
	Slot          uint64
	ProposerIndex ValidatorIndex
	ParentRoot    []byte
	StateRoot     []byte
	BodyRoot      []byte
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

// ILightNodeLightClientState is an auto generated low-level Go binding around an user-defined struct.
type ILightNodeLightClientState struct {
	FinalizedHeader      ILightNodeBeaconBlockHeader
	CurrentSyncCommittee ILightNodeSyncCommittee
	NextSyncCommittee    ILightNodeSyncCommittee
	ChainID              uint64
}

// ILightNodeLightClientUpdate is an auto generated low-level Go binding around an user-defined struct.
type ILightNodeLightClientUpdate struct {
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
	Update ILightNodeLightClientUpdate
	State  ILightNodeLightClientState
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

func (verify *ILightNodeLightClientVerify) toLightClientVerify() *LightClientVerify {
	return &LightClientVerify{
		update: verify.Update.toLightClientUpdate(),
		state:  verify.State.toLightClientState(),
	}
}

func (update *ILightNodeLightClientUpdate) toLightClientUpdate() *LightClientUpdate {
	return &LightClientUpdate{
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

func (header *ILightNodeBeaconBlockHeader) toBeaconBlockHeader() BeaconBlockHeader {
	return BeaconBlockHeader{
		Slot:          header.Slot,
		ProposerIndex: ValidatorIndex(header.ProposerIndex),
		ParentRoot:    header.ParentRoot[:],
		StateRoot:     header.StateRoot[:],
		BodyRoot:      header.BodyRoot[:],
	}
}

func (header *ILightNodeBlockHeader) toBlockHeader() types.Header {
	blockHeader := types.Header{
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

func (syncCommittee *ILightNodeSyncCommittee) toSyncCommittee() SyncCommittee {
	var pubkeys [][]byte
	count := len(syncCommittee.Pubkeys) / BLSPubkeyLength
	for i := 0; i < count; i++ {
		pubkeys = append(pubkeys, syncCommittee.Pubkeys[i*BLSPubkeyLength:(i+1)*BLSPubkeyLength])
	}
	return SyncCommittee{
		Pubkeys:         pubkeys,
		AggregatePubkey: syncCommittee.AggregatePubkey,
	}
}
func (syncAggregate *ILightNodeSyncAggregate) toSyncAggregate() SyncAggregate {
	return SyncAggregate{
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
