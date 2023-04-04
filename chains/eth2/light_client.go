package eth2

import (
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/chains/eth2/bls12381"
	ssz "github.com/prysmaticlabs/fastssz"
)

const MinSyncCommitteeParticipants uint64 = 1
const EpochsPerSyncCommitteePeriod uint64 = 256
const SlotsPerEpoch uint64 = 32

const FinalizedRootIndex uint32 = 105
const NextSyncCommitteeIndex uint32 = 55

const BeaconBlockBodyTreeExecutionPayloadIndex uint64 = 25
const ExecutionPayloadProofSize int = 4

const L1BeaconBlockBodyTreeExecutionPayloadIndex uint64 = 25
const L2ExecutionPayloadTreeExecutionBlockIndex uint64 = 28
const L1BeaconBlockBodyProofSize uint64 = 4
const L2ExecutionPayloadProofSize uint64 = 4
const ExecutionProofSize = L1BeaconBlockBodyProofSize + L2ExecutionPayloadProofSize

var DomainSyncCommittee = [4]byte{0x07, 0x00, 0x00, 0x00}

func VerifyLightClientUpdate(input []byte) error {
	verify, err := decodeLightClientVerifyV2(input)
	if err != nil {
		log.Warn("decodeLightClientVerifyV2", "error", err)
		verify, err = decodeLightClientVerifyV1(input)
		if err != nil {
			log.Warn("decodeLightClientVerifyV1", "error", err)
			return err
		}
	}

	switch verify.update.(type) {
	case *LightClientUpdateV1:
		if err := verifyFinalityV1(verify.update.(*LightClientUpdateV1)); err != nil {
			log.Warn("verifyFinalityV1", "error", err)
			return err
		}
	case *LightClientUpdateV2:
		if err := verifyFinalityV2(verify.update.(*LightClientUpdateV2)); err != nil {
			log.Warn("verifyFinalityV2", "error", err)
			return err
		}

	default:
		log.Warn("invalid light client update type")
		return fmt.Errorf("invalid light client update type")
	}

	if err := verifyNextSyncCommittee(verify.state, verify.update); err != nil {
		log.Warn("verifyNextSyncCommittee", "error", err)
		return err
	}

	if err := verifyBlsSignatures(verify.state, verify.update); err != nil {
		log.Warn("verifyBlsSignatures", "error", err)
		return err
	}

	return nil
}

func verifyFinalityV2(update *LightClientUpdateV2) error {
	leaf, err := update.finalizedHeader.HashTreeRoot()
	if err != nil {
		return fmt.Errorf("failed to compute hash tree root of finalized header: %v", err)
	}
	proof := ssz.Proof{
		Index:  int(FinalizedRootIndex),
		Leaf:   leaf[:],
		Hashes: update.finalityBranch,
	}
	ret, err := ssz.VerifyProof(update.attestedHeader.StateRoot, &proof)
	if err != nil {
		return fmt.Errorf("VerifyProof return err: %v", err)
	}

	if !ret {
		return fmt.Errorf("invalid finality proof")
	}

	if len(update.executionBranch) != ExecutionPayloadProofSize {
		return fmt.Errorf("invalid execution payload proof size, exp: %d, got: %d", ExecutionPayloadProofSize, len(update.executionBranch))
	}

	executionPayloadHash, err := update.finalizedExecution.HashTreeRoot()
	if err != nil {
		return fmt.Errorf("compute execution payload merkel root failed: %v", err)
	}

	proof = ssz.Proof{
		Index:  int(BeaconBlockBodyTreeExecutionPayloadIndex),
		Leaf:   executionPayloadHash[:],
		Hashes: update.executionBranch,
	}
	ret, err = ssz.VerifyProof(update.finalizedHeader.BodyRoot, &proof)
	if err != nil {
		return fmt.Errorf("VerifyProof return err: %v", err)
	}

	if !ret {
		return fmt.Errorf("invalid execution payload proof")
	}

	return nil
}

func verifyFinalityV1(update *LightClientUpdateV1) error {
	leaf, err := update.finalizedHeader.HashTreeRoot()
	if err != nil {
		return fmt.Errorf("failed to compute hash tree root of finalized header: %v", err)
	}
	proof := ssz.Proof{
		Index:  int(FinalizedRootIndex),
		Leaf:   leaf[:],
		Hashes: update.finalityBranch,
	}
	ret, err := ssz.VerifyProof(update.attestedHeader.StateRoot, &proof)
	if err != nil {
		return fmt.Errorf("VerifyProof return err: %v", err)
	}

	if !ret {
		return fmt.Errorf("invalid finality proof")
	}

	l1Proof := update.exeFinalityBranch[0:L1BeaconBlockBodyProofSize]
	l2Proof := update.exeFinalityBranch[L1BeaconBlockBodyProofSize:ExecutionProofSize]

	executionPayloadHash, err := merkelRootFromBranch(
		update.finalizedExeHeader.Hash(),
		l2Proof,
		L2ExecutionPayloadProofSize,
		L2ExecutionPayloadTreeExecutionBlockIndex,
	)
	if err != nil {
		return fmt.Errorf("compute execution payload merkel root failed: %v", err)
	}

	proof = ssz.Proof{
		Index:  int(L1BeaconBlockBodyTreeExecutionPayloadIndex),
		Leaf:   executionPayloadHash[:],
		Hashes: l1Proof,
	}
	ret, err = ssz.VerifyProof(update.finalizedHeader.BodyRoot, &proof)
	if err != nil {
		return fmt.Errorf("VerifyProof return err: %v", err)
	}

	if !ret {
		return fmt.Errorf("invalid execution payload proof")
	}

	return nil
}

func verifyNextSyncCommittee(state *LightClientState, update ILightClientUpdate) error {
	// The active header will always be the finalized header because we don't accept updates without the finality update.
	updatePeriod := computeSyncCommitteePeriod(update.GetFinalizedHeader().Slot)
	finalizedPeriod := computeSyncCommitteePeriod(state.finalizedHeader.Slot)

	// Verify that the `next_sync_committee`, if present, actually is the next sync committee saved in the
	// state of the `active_header`
	if updatePeriod != finalizedPeriod {
		leaf, err := SyncCommitteeRoot(update.GetNextSyncCommittee())
		if err != nil {
			return fmt.Errorf("failed to compute hash tree root of finalized header: %v", err)
		}
		proof := ssz.Proof{
			Index:  int(NextSyncCommitteeIndex),
			Leaf:   leaf[:],
			Hashes: update.GetNextSyncCommitteeBranch(),
		}
		ret, err := ssz.VerifyProof(update.GetAttestedHeader().StateRoot, &proof)
		if err != nil {
			return fmt.Errorf("VerifyProof return err: %v", err)
		}

		if !ret {
			return fmt.Errorf("invalid next sync committee proof")
		}
	}

	return nil
}

func verifyBlsSignatures(state *LightClientState, update ILightClientUpdate) error {
	syncCommitteeCount := update.GetSyncAggregate().SyncCommitteeBits.Count()
	if syncCommitteeCount < MinSyncCommitteeParticipants {
		return fmt.Errorf("invalid sync committee participants count, min required %d, got %d", MinSyncCommitteeParticipants, syncCommitteeCount)
	}

	if syncCommitteeCount*3 < update.GetSyncAggregate().SyncCommitteeBits.Len()*2 {
		return fmt.Errorf("not enought sync committe count %d", syncCommitteeCount)
	}

	finalizedPeriod := computeSyncCommitteePeriod(state.finalizedHeader.Slot)
	config, err := newNetworkConfig(state.chainID)
	if err != nil {
		return fmt.Errorf("new network failed: %v", err)
	}

	signaturePeriod := computeSyncCommitteePeriod(update.GetSignatureSlot())
	var syncCommittee *SyncCommittee

	// Verify signature period does not skip a sync committee period
	if signaturePeriod == finalizedPeriod {
		syncCommittee = state.currentSyncCommittee
	} else if signaturePeriod == finalizedPeriod+1 {
		syncCommittee = state.nextSyncCommittee
	} else {
		return fmt.Errorf("signature period should be %d or %d, but got %d",
			finalizedPeriod, finalizedPeriod+1, signaturePeriod)
	}

	// Verify sync committee aggregate signature
	forkVersionSlot := update.GetSignatureSlot() - 1
	forkVersion := config.computeForkVersionBySlot(forkVersionSlot)
	if forkVersion == nil {
		return fmt.Errorf("unsupportted fork")
	}

	domain, err := ComputeDomain(DomainSyncCommittee, forkVersion[:], config.GenesisValidatorsRoot[:])
	if err != nil {
		return fmt.Errorf("compute domain failed: %v", err)
	}

	signingRoot, err := ComputeSigningRoot(update.GetAttestedHeader(), domain)
	if err != nil {
		return fmt.Errorf("compute signing root failed: %v", err)
	}

	pubKeys, err := getParticipantPubkeys(syncCommittee.Pubkeys, update.GetSyncAggregate().SyncCommitteeBits)
	if err != nil {
		return fmt.Errorf("get participiant pubkyes failed: %v", err)
	}

	signature, err := bls.SignatureFromBytes(update.GetSyncAggregate().SyncCommitteeSignature)
	if err != nil {
		return fmt.Errorf("deserialize signature failed: %v", err)
	}

	if !signature.FastAggregateVerify(pubKeys, signingRoot) {
		return fmt.Errorf("fast aggregate verify failed")
	}

	return nil
}
