package eth2

import (
	"fmt"
	"github.com/mapprotocol/atlas/chains/eth2/bls12381"
	ssz "github.com/prysmaticlabs/fastssz"
)

const MinSyncCommitteeParticipants uint64 = 1
const EpochsPerSyncCommitteePeriod uint64 = 256
const SlotsPerEpoch uint64 = 32

const FinalizedRootIndex uint32 = 105
const NextSyncCommitteeIndex uint32 = 55

const L1BeaconBlockBodyTreeExecutionPayloadIndex uint64 = 25
const L2ExecutionPayloadTreeExecutionBlockIndex uint64 = 28
const L1BeaconBlockBodyProofSize uint64 = 4
const L2ExecutionPayloadProofSize uint64 = 4
const ExecutionProofSize = L1BeaconBlockBodyProofSize + L2ExecutionPayloadProofSize

var DomainSyncCommittee = [4]byte{0x07, 0x00, 0x00, 0x00}

func VerifyLightClientUpdate(input []byte) error {
	verify, err := decodeLightClientVerify(input)
	if err != nil {
		return err
	}

	if err := verifyFinality(verify.update); err != nil {
		return err
	}

	if err := verifyNextSyncCommittee(verify.state, verify.update); err != nil {
		return err
	}

	return verifyBlsSignatures(verify.state, verify.update)
}

func verifyFinality(update *LightClientUpdate) error {
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

func verifyNextSyncCommittee(state *LightClientState, update *LightClientUpdate) error {
	// The active header will always be the finalized header because we don't accept updates without the finality update.
	updatePeriod := computeSyncCommitteePeriod(update.finalizedHeader.Slot)
	finalizedPeriod := computeSyncCommitteePeriod(state.finalizedHeader.Slot)

	// Verify that the `next_sync_committee`, if present, actually is the next sync committee saved in the
	// state of the `active_header`
	if updatePeriod != finalizedPeriod {
		leaf, err := SyncCommitteeRoot(&update.nextSyncCommittee)
		if err != nil {
			return fmt.Errorf("failed to compute hash tree root of finalized header: %v", err)
		}
		proof := ssz.Proof{
			Index:  int(NextSyncCommitteeIndex),
			Leaf:   leaf[:],
			Hashes: update.nextSyncCommitteeBranch,
		}
		ret, err := ssz.VerifyProof(update.finalizedHeader.StateRoot, &proof)
		if err != nil {
			return fmt.Errorf("VerifyProof return err: %v", err)
		}

		if !ret {
			return fmt.Errorf("invalid next sync committee proof")
		}
	}

	return nil
}

func verifyBlsSignatures(state *LightClientState, update *LightClientUpdate) error {
	syncCommitteeCount := update.syncAggregate.SyncCommitteeBits.Count()
	if syncCommitteeCount < MinSyncCommitteeParticipants {
		return fmt.Errorf("invalid sync committee participants count, min required %d, got %d", MinSyncCommitteeParticipants, syncCommitteeCount)
	}

	if syncCommitteeCount*3 < update.syncAggregate.SyncCommitteeBits.Len()*2 {
		return fmt.Errorf("not enought sync committe count %d", syncCommitteeCount)
	}

	finalizedPeriod := computeSyncCommitteePeriod(state.finalizedHeader.Slot)
	config, err := newNetworkConfig(state.chainID)
	if err != nil {
		return fmt.Errorf("new network failed: %v", err)
	}

	signaturePeriod := computeSyncCommitteePeriod(update.signatureSlot)
	var syncCommittee SyncCommittee

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
	forkVersion := config.computeForkVersionBySlot(update.signatureSlot)
	if forkVersion == nil {
		return fmt.Errorf("unsupportted fork")
	}

	domain, err := ComputeDomain(DomainSyncCommittee, forkVersion[:], config.GenesisValidatorsRoot[:])
	if err != nil {
		return fmt.Errorf("compute domain failed: %v", err)
	}

	signingRoot, err := ComputeSigningRoot(&update.attestedHeader, domain)
	if err != nil {
		return fmt.Errorf("compute signing root failed: %v", err)
	}

	pubKeys, err := getParticipantPubkeys(syncCommittee.Pubkeys, update.syncAggregate.SyncCommitteeBits)
	if err != nil {
		return fmt.Errorf("get participiant pubkyes failed: %v", err)
	}

	signature, err := bls.SignatureFromBytes(update.syncAggregate.SyncCommitteeSignature)
	if err != nil {
		return fmt.Errorf("ddeserialize signature failed: %v", err)
	}

	if !signature.FastAggregateVerify(pubKeys, signingRoot) {
		return fmt.Errorf("fast aggregate verify failed")
	}

	return nil
}
