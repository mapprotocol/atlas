package eth2

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	bls2 "github.com/mapprotocol/atlas/chains/eth2/bls12381"
	"github.com/mapprotocol/atlas/chains/eth2/hash"
	"github.com/mapprotocol/atlas/chains/eth2/ssz"
	"github.com/minio/sha256-simd"
	fssz "github.com/prysmaticlabs/fastssz"
	"github.com/prysmaticlabs/go-bitfield"
)

const ABIJSON = "{\"components\":[{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"slot\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"proposerIndex\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"parentRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"bodyRoot\",\"type\":\"bytes32\"}],\"internalType\":\"structILightNode.BeaconBlockHeader\",\"name\":\"attestedHeader\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes\",\"name\":\"pubkeys\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"aggregatePubkey\",\"type\":\"bytes\"}],\"internalType\":\"structILightNode.SyncCommittee\",\"name\":\"nextSyncCommittee\",\"type\":\"tuple\"},{\"internalType\":\"bytes32[]\",\"name\":\"nextSyncCommitteeBranch\",\"type\":\"bytes32[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"slot\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"proposerIndex\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"parentRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"bodyRoot\",\"type\":\"bytes32\"}],\"internalType\":\"structILightNode.BeaconBlockHeader\",\"name\":\"finalizedHeader\",\"type\":\"tuple\"},{\"internalType\":\"bytes32[]\",\"name\":\"finalityBranch\",\"type\":\"bytes32[]\"},{\"components\":[{\"internalType\":\"bytes\",\"name\":\"parentHash\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"sha3Uncles\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"miner\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"stateRoot\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"transactionsRoot\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"receiptsRoot\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"logsBloom\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"difficulty\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"number\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasUsed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"extraData\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"mixHash\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"nonce\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"baseFeePerGas\",\"type\":\"uint256\"}],\"internalType\":\"structILightNode.BlockHeader\",\"name\":\"finalizedExeHeader\",\"type\":\"tuple\"},{\"internalType\":\"bytes32[]\",\"name\":\"exeFinalityBranch\",\"type\":\"bytes32[]\"},{\"components\":[{\"internalType\":\"bytes\",\"name\":\"syncCommitteeBits\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"syncCommitteeSignature\",\"type\":\"bytes\"}],\"internalType\":\"structILightNode.SyncAggregate\",\"name\":\"syncAggregate\",\"type\":\"tuple\"},{\"internalType\":\"uint64\",\"name\":\"signatureSlot\",\"type\":\"uint64\"}],\"internalType\":\"structILightNode.LightClientUpdate\",\"name\":\"update\",\"type\":\"tuple\"},{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"slot\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"proposerIndex\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"parentRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"bodyRoot\",\"type\":\"bytes32\"}],\"internalType\":\"structILightNode.BeaconBlockHeader\",\"name\":\"finalizedHeader\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes\",\"name\":\"pubkeys\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"aggregatePubkey\",\"type\":\"bytes\"}],\"internalType\":\"structILightNode.SyncCommittee\",\"name\":\"currentSyncCommittee\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes\",\"name\":\"pubkeys\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"aggregatePubkey\",\"type\":\"bytes\"}],\"internalType\":\"structILightNode.SyncCommittee\",\"name\":\"nextSyncCommittee\",\"type\":\"tuple\"},{\"internalType\":\"uint64\",\"name\":\"chainID\",\"type\":\"uint64\"}],\"internalType\":\"structILightNode.LightClientState\",\"name\":\"state\",\"type\":\"tuple\"}],\"indexed\":false,\"internalType\":\"structILightNode.LightClientVerify\",\"name\":\"verify\",\"type\":\"tuple\"}"

const UpdateABIJSON = "{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"slot\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"proposerIndex\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"parentRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"bodyRoot\",\"type\":\"bytes32\"}],\"internalType\":\"struct Types.BeaconBlockHeader\",\"name\":\"finalizedHeader\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes\",\"name\":\"pubkeys\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"aggregatePubkey\",\"type\":\"bytes\"}],\"internalType\":\"struct Types.SyncCommittee\",\"name\":\"nextSyncCommittee\",\"type\":\"tuple\"},{\"internalType\":\"bytes32[]\",\"name\":\"nextSyncCommitteeBranch\",\"type\":\"bytes32[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"slot\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"proposerIndex\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"parentRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"bodyRoot\",\"type\":\"bytes32\"}],\"internalType\":\"struct Types.BeaconBlockHeader\",\"name\":\"finalizedHeader\",\"type\":\"tuple\"},{\"internalType\":\"bytes32[]\",\"name\":\"finalityBranch\",\"type\":\"bytes32[]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"parentHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"feeRecipient\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"receiptsRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"logsBloom\",\"type\":\"bytes\"},{\"internalType\":\"bytes32\",\"name\":\"prevRandao\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasUsed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"extraData\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"baseFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"blockHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"transactionsRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"withdrawalsRoot\",\"type\":\"bytes32\"}],\"internalType\":\"struct Types.Execution\",\"name\":\"finalizedExecution\",\"type\":\"tuple\"},{\"internalType\":\"bytes32[]\",\"name\":\"executionBranch\",\"type\":\"bytes32[]\"},{\"components\":[{\"internalType\":\"bytes\",\"name\":\"syncCommitteeBits\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"syncCommitteeSignature\",\"type\":\"bytes\"}],\"internalType\":\"struct Types.SyncAggregate\",\"name\":\"syncAggregate\",\"type\":\"tuple\"},{\"internalType\":\"uint64\",\"name\":\"signatureSlot\",\"type\":\"uint64\"}],\"internalType\":\"struct Types.LightClientUpdate\",\"type\":\"tuple\"}"
const BeaconHeaderABIJSON = "{\"components\":[{\"internalType\":\"uint64\",\"name\":\"slot\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"proposerIndex\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"parentRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"bodyRoot\",\"type\":\"bytes32\"}],\"internalType\":\"struct Types.BeaconBlockHeader\",\"type\":\"tuple\"}"
const SyncCommitteeABIJSON = "{\"components\":[{\"internalType\":\"bytes\",\"name\":\"pubkeys\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"aggregatePubkey\",\"type\":\"bytes\"}],\"internalType\":\"struct Types.SyncCommittee\",\"type\":\"tuple\"}"
const ChainIdABIJSON = "{\"internalType\":\"uint64\",\"type\":\"uint64\"}"

func decodeLightClientVerify(input []byte) (*LightClientVerify, error) {
	verify, err := decodeLightClientVerifyV2(input)
	if err != nil {
		log.Warn("decodeLightClientVerifyV2", "error", err)
		verify, err = decodeLightClientVerifyV1(input)
		if err != nil {
			log.Warn("decodeLightClientVerifyV1", "error", err)
			return nil, err
		}
	}

	return verify, nil
}

func decodeLightClientVerifyV1(input []byte) (*LightClientVerify, error) {
	var arg abi.Argument
	if err := arg.UnmarshalJSON([]byte(ABIJSON)); err != nil {
		return nil, fmt.Errorf("unmarshal abi json failed: %v", err)
	}

	args := abi.Arguments{arg}
	verify := new(ILightNodeLightClientVerify)
	ret, err := args.Unpack(input)
	if err != nil {
		return nil, fmt.Errorf("unpack input failed: %v", err)
	}

	if err := args.Copy(&verify, ret); err != nil {
		return nil, fmt.Errorf("copy unpacked result failed: %v", err)
	}

	return verify.toLightClientVerify(), nil
}

func decodeLightClientVerifyV2(input []byte) (*LightClientVerify, error) {
	args, err := genAbiArgs()
	if err != nil {
		return nil, fmt.Errorf("gen abi args failed: %v", err)
	}

	ret, err := args.Unpack(input)
	if err != nil {
		return nil, fmt.Errorf("unpack input failed: %v", err)
	}

	update := new(ILightNodeLightClientUpdateV2)
	finalizedBeaconHeader := new(ILightNodeBeaconBlockHeader)
	curSyncCommittee := new(ILightNodeSyncCommittee)
	nextSyncCommittee := new(ILightNodeSyncCommittee)
	chainId := new(uint64)
	if err := args.Copy(&[]interface{}{update, finalizedBeaconHeader, curSyncCommittee, nextSyncCommittee, chainId}, ret); err != nil {
		return nil, fmt.Errorf("copy unpacked result failed: %v", err)
	}

	return ConvertToLightClientVerify(update, finalizedBeaconHeader, curSyncCommittee, nextSyncCommittee, *chainId), nil
}

func genAbiArgs() (abi.Arguments, error) {
	var updateArg, beaconHeaderArg, syncCommitteeArg, chainIdArg abi.Argument
	if err := updateArg.UnmarshalJSON([]byte(UpdateABIJSON)); err != nil {
		return nil, fmt.Errorf("unmarshal update abi json failed: %v", err)
	}

	if err := beaconHeaderArg.UnmarshalJSON([]byte(BeaconHeaderABIJSON)); err != nil {
		return nil, fmt.Errorf("unmarshal beacon header abi json failed: %v", err)
	}

	if err := syncCommitteeArg.UnmarshalJSON([]byte(SyncCommitteeABIJSON)); err != nil {
		return nil, fmt.Errorf("unmarshal sync committee abi json failed: %v", err)
	}

	if err := chainIdArg.UnmarshalJSON([]byte(ChainIdABIJSON)); err != nil {
		return nil, fmt.Errorf("unmarshal chain id abi json failed: %v", err)
	}

	return abi.Arguments{updateArg, beaconHeaderArg, syncCommitteeArg, syncCommitteeArg, chainIdArg}, nil
}

func computeEpochAtSlot(slot uint64) uint64 {
	return slot / SlotsPerEpoch
}

func computeSyncCommitteePeriod(slot uint64) uint64 {
	return computeEpochAtSlot(slot) / EpochsPerSyncCommitteePeriod
}

func getParticipantPubkeys(public_keys [][]byte, sync_committee_bits bitfield.Bitvector512) ([]bls2.PublicKey, error) {
	var pubkeys []bls2.PublicKey
	for i := uint64(0); i < sync_committee_bits.Len(); i++ {
		if sync_committee_bits.BitAt(i) {

			pubKey, err := bls2.PublicKeyFromBytes(public_keys[i])
			if err != nil {
				return nil, fmt.Errorf("deserialze sync committe public key failed: %v", err)
			}
			pubkeys = append(pubkeys, pubKey)
		}
	}
	return pubkeys, nil
}

func merkelRootFromBranch(leaf common.Hash, branch [][]byte, depth uint64, index uint64) (common.Hash, error) {
	if uint64(len(branch)) != depth {
		return common.Hash{}, fmt.Errorf("expected proof length %d, but got %d", depth, len(branch))
	}

	node := leaf[:]
	tmp := make([]byte, 64)
	for i, h := range branch {
		if getPosAtLevel(int(index), i) {
			copy(tmp[:32], h[:])
			copy(tmp[32:], node[:])
			node = hashFn(tmp)
		} else {
			copy(tmp[:32], node[:])
			copy(tmp[32:], h[:])
			node = hashFn(tmp)
		}
	}

	return common.BytesToHash(node), nil
}

// Returns the position (i.e. false for left, true for right)
// of an index at a given level.
// Level 0 is the actual index's level, Level 1 is the position
// of the parent, etc.
func getPosAtLevel(index int, level int) bool {
	return (index & (1 << level)) > 0
}

func hashFn(data []byte) []byte {
	res := sha256.Sum256(data)
	return res[:]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ForkVersionByteLength length of fork version byte array.
const ForkVersionByteLength = 4

// DomainByteLength length of domain byte array.
const DomainByteLength = 4

// ComputeDomain returns the domain version for BLS private key to sign and verify with a zeroed 4-byte
// array as the fork version.
//
// def compute_domain(domain_type: DomainType, fork_version: Version=None, genesis_validators_root: Root=None) -> Domain:
//    """
//    Return the domain for the ``domain_type`` and ``fork_version``.
//    """
//    if fork_version is None:
//        fork_version = GENESIS_FORK_VERSION
//    if genesis_validators_root is None:
//        genesis_validators_root = Root()  # all bytes zero by default
//    fork_data_root = compute_fork_data_root(fork_version, genesis_validators_root)
//    return Domain(domain_type + fork_data_root[:28])
func ComputeDomain(domainType [DomainByteLength]byte, forkVersion, genesisValidatorsRoot []byte) ([]byte, error) {
	if forkVersion == nil || genesisValidatorsRoot == nil {
		return nil, fmt.Errorf("invalid input args")
	}

	forkBytes := [ForkVersionByteLength]byte{}
	copy(forkBytes[:], forkVersion)

	forkDataRoot, err := computeForkDataRoot(forkBytes[:], genesisValidatorsRoot)
	if err != nil {
		return nil, err
	}

	return domain(domainType, forkDataRoot[:]), nil
}

// This returns the bls domain given by the domain type and fork data root.
func domain(domainType [DomainByteLength]byte, forkDataRoot []byte) []byte {
	var b []byte
	b = append(b, domainType[:4]...)
	b = append(b, forkDataRoot[:28]...)
	return b
}

// this returns the 32byte fork data root for the ``current_version`` and ``genesis_validators_root``.
// This is used primarily in signature domains to avoid collisions across forks/chains.
//
// Spec pseudocode definition:
//	def compute_fork_data_root(current_version: Version, genesis_validators_root: Root) -> Root:
//    """
//    Return the 32-byte fork data root for the ``current_version`` and ``genesis_validators_root``.
//    This is used primarily in signature domains to avoid collisions across forks/chains.
//    """
//    return hash_tree_root(ForkData(
//        current_version=current_version,
//        genesis_validators_root=genesis_validators_root,
//    ))
func computeForkDataRoot(version, root []byte) ([32]byte, error) {
	r, err := (&ForkData{
		CurrentVersion:        version,
		GenesisValidatorsRoot: root,
	}).HashTreeRoot()
	if err != nil {
		return [32]byte{}, err
	}
	return r, nil
}

// SyncCommitteeRoot computes the HashTreeRoot Merkleization of a committee root.
// a SyncCommitteeRoot struct according to the eth2
// Simple Serialize specification.
func SyncCommitteeRoot(committee *SyncCommittee) ([32]byte, error) {
	hasher := hash.CustomSHA256Hasher()
	var fieldRoots [][32]byte
	if committee == nil {
		return [32]byte{}, nil
	}

	// Field 1:  Vector[BLSPubkey, SYNC_COMMITTEE_SIZE]
	pubKeyRoots := make([][32]byte, 0)
	for _, pubkey := range committee.Pubkeys {
		r, err := merkleizePubkey(hasher, pubkey)
		if err != nil {
			return [32]byte{}, err
		}
		pubKeyRoots = append(pubKeyRoots, r)
	}
	pubkeyRoot, err := ssz.BitwiseMerkleize(hasher, pubKeyRoots, uint64(len(pubKeyRoots)), uint64(len(pubKeyRoots)))
	if err != nil {
		return [32]byte{}, err
	}

	// Field 2: BLSPubkey
	aggregateKeyRoot, err := merkleizePubkey(hasher, committee.AggregatePubkey)
	if err != nil {
		return [32]byte{}, err
	}
	fieldRoots = [][32]byte{pubkeyRoot, aggregateKeyRoot}

	return ssz.BitwiseMerkleize(hasher, fieldRoots, uint64(len(fieldRoots)), uint64(len(fieldRoots)))
}

func merkleizePubkey(hasher ssz.HashFn, pubkey []byte) ([32]byte, error) {
	chunks, err := ssz.PackByChunk([][]byte{pubkey})
	if err != nil {
		return [32]byte{}, err
	}
	var pubKeyRoot [32]byte
	//if features.Get().EnableVectorizedHTR {
	//	outputChunk := make([][32]byte, 1)
	//	htr.VectorizedSha256(chunks, outputChunk)
	//	pubKeyRoot = outputChunk[0]
	//} else {
	pubKeyRoot, err = ssz.BitwiseMerkleize(hasher, chunks, uint64(len(chunks)), uint64(len(chunks)))
	if err != nil {
		return [32]byte{}, err
	}
	//}
	return pubKeyRoot, nil
}

// ComputeSigningRoot computes the root of the object by calculating the hash tree root of the signing data with the given domain.
//
// Spec pseudocode definition:
//	def compute_signing_root(ssz_object: SSZObject, domain: Domain) -> Root:
//    """
//    Return the signing root for the corresponding signing data.
//    """
//    return hash_tree_root(SigningData(
//        object_root=hash_tree_root(ssz_object),
//        domain=domain,
//    ))
func ComputeSigningRoot(object fssz.HashRoot, domain []byte) ([32]byte, error) {
	return signingData(object.HashTreeRoot, domain)
}

// Computes the signing data by utilising the provided root function and then
// returning the signing data of the container object.
func signingData(rootFunc func() ([32]byte, error), domain []byte) ([32]byte, error) {
	objRoot, err := rootFunc()
	if err != nil {
		return [32]byte{}, err
	}
	container := &SigningData{
		ObjectRoot: objRoot[:],
		Domain:     domain,
	}
	return container.HashTreeRoot()
}

// PadTo pads a byte slice to the given size. If the byte slice is larger than the given size, the
// original slice is returned.
func PadTo(b []byte, size int) []byte {
	if len(b) >= size {
		return b
	}
	return append(b, make([]byte, size-len(b))...)
}

// ReverseByteOrder Switch the endianness of a byte slice by reversing its order.
// this function does not modify the actual input bytes.
func ReverseByteOrder(input []byte) []byte {
	b := make([]byte, len(input))
	copy(b, input)
	for i := 0; i < len(b)/2; i++ {
		b[i], b[len(b)-i-1] = b[len(b)-i-1], b[i]
	}
	return b
}
