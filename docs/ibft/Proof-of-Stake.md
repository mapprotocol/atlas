---
sort: 1
---

## Proof-of-Stake

Atlas is a proof-of-stake blockchain. In comparison to Proof of Work systems like Bitcoin and Ethereum, this eliminates the negative environmental impact and means that users can make transactions that are cheaper, faster, and whose outcome cannot be changed once complete.

The Atlas Blockchain implements a Istanbul Byzantine Fault Tolerant (IBFT) consensus algorithm in which a well-defined set of validator nodes broadcast signed messages between themselves in a sequence of steps to reach agreement even when up to a third of the total nodes are offline, faulty or malicious. When a quorum of validators have reached agreement, that decision is final.

## Validators
Validators gather transactions received from other nodes and execute any associated smart contracts to form new blocks, then participate in a Byzantine Fault Tolerant (BFT) consensus protocol to advance the state of the network. Since BFT protocols can scale only to a few hundred participants, and can tolerate at most a third of the participants acting maliciously, a proof-of-stake mechanism admits only a limited set of nodes to this role.

## Staking Requirements

Atlas uses a proof-of-stake consensus mechanism, which requires Validators to have locked MAP to participate in block production. 
The current requirement is `10,000` MAP to register a Validator, and `10,000` MAP per member validator to register a Validator Group.

## About election

The Election contract is called from the IBFT block finalization code to select the validators for the following epoch.
The contract maintains a sorted list of the Locked Gold voting (either pending or activated) for each Validator Group.
The D’Hondt method, a closed party list form of proportional representation, is applied to iteratively select validators from the Validator Groups with the greatest associated vote balances.

The list of groups is first filtered to remove those that have not achieved a certain fraction of the votes of the total voting Locked Gold.

Then, in the first iteration, the algorithm assigns the first seat to the group that has at least one member and with the most votes.
Thereafter, it assigns the seat to the group that would ‘pay’, if its next validator were elected, the highest vote averaged over its candidates that have been selected so far plus the one under consideration.

There is a minimum target and a maximum cap on the number of active validators that may be selected.
If the minimum target is not reached, the election aborts and no change is made to the validator set this epoch.

## Validator number and Reward

The participators make these decisions by locking MAP and voting for Validator Groups, intermediaries that sit between voters and Validators.
Every Validator Group has an ordered list of up to `5` candidate Validators.
Validator elections are held `every epoch` (approximately once per day).
The protocol elects a maximum of `100` Validators. At each epoch, every elected Validator must be re-elected to continue.
Validators are selected in proportion to votes received for each Validator Group.

If you hold MAP, or are a beneficiary of a ReleaseGold contract that allows voting, you can vote for Validator Groups. A single account can split their LockedGold balance to have outstanding votes between `3 groups` and `10 groups`.
MAP that you lock and use to vote for a group that elects one or more Validators receives epoch rewards every epoch (approximately every day) once the community passes a governance proposal enabling rewards.
The initial level of rewards is anticipated to be around `6%` per annum equivalent (but is subject to change).
