---
sort: 1
---

# Proof-of-Stake

Atlas is a proof-of-stake blockchain. In comparison to Proof of Work systems like Bitcoin and Ethereum, this eliminates the negative environmental impact and means that users can make transactions that are cheaper, faster, and whose outcome cannot be changed once complete.

The Atlas Blockchain implements a Istanbul Byzantine Fault Tolerant (IBFT) consensus algorithm in which a well-defined set of validator nodes broadcast signed messages between themselves in a sequence of steps to reach agreement even when up to a third of the total nodes are offline, faulty or malicious. When a quorum of validators have reached agreement, that decision is final.

## Validators
Validators gather transactions received from other nodes and execute any associated smart contracts to form new blocks, then participate in a Byzantine Fault Tolerant (BFT) consensus protocol to advance the state of the network. Since BFT protocols can scale only to a few hundred participants, and can tolerate at most a third of the participants acting maliciously, a proof-of-stake mechanism admits only a limited set of nodes to this role.

## Running a private miner
In a private network setting, you need to encode a types.IstanbulExtra data struct with rlp.
and put it into the extraData field of the genesis block.
```
type IstanbulExtra struct {
	// AddedValidators are the validators that have been added in the block
	AddedValidators []common.Address
	// AddedValidatorsPublicKeys are the BLS public keys for the validators added in the block
	AddedValidatorsPublicKeys []blscrypto.SerializedPublicKey
	// RemovedValidators is a bitmap having an active bit for each removed validator in the block
	RemovedValidators *big.Int
	// Seal is an ECDSA signature by the proposer
	Seal []byte
	// AggregatedSeal contains the aggregated BLS signature created via IBFT consensus.
	AggregatedSeal IstanbulAggregatedSeal
	// ParentAggregatedSeal contains and aggregated BLS signature for the previous block.
	ParentAggregatedSeal IstanbulAggregatedSeal
}
```
in the types.IstanbulExtra data struct,we can assigned the validators at the first epoch by providing 
addresses and BLS public keys.

To start a `atlas` instance for mining, run it with all your usual flags, extended
by:

```shell
$ atlas <usual-flags> --mine --miner.etherbase=0x0000000000000000000000000000000000000000 --unlock=0x0000000000000000000000000000000000000000 --password 
```

Since validitor needs to sign messages with private key.when turn on the mining switch,you also need
to unlock the etherbase account.