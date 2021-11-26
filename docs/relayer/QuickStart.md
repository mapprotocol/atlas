---
sort: 3
---

## How To Become Relayer

### Step 1: Insure your account has sufficient balance

In order to join election, the account balance should have 100000 MAP at least. Then using 100000 MAP to register at contract.

### Step 2: Join relayer election 

When balance is more 100000 MAP at contract, the validator will join election at election point that it is block number that counts backwards 200 at a multiple of ten thousand generally.
It is Auto-completion by our system. the validator is 100% selected at first times, only if the list of being selected isn't full.   

Notify! It wants to join election this term, the validator should register before election point.

### Step 3: Wait for election result

Above introduce, the election point is block number that counts backwards 200 at a multiple of ten thousand generally.
After election point, it can query that the validator is selected or not, but the selected validator start to work at Epoch started point, i.e. block number is a multiple of ten thousand.

### Step 4: Synchronize enough block

If the validator isn't registered first, it wants to become relayer continue that needs to complete enough workload.
In order to reappointment, the current relayers need to synchronize 10000 blocks from other blockchain to atlas.

### Step 5: Withdraw your registered asset

If the validator wants to quit, it should unregister balance at contract. The balance will be locked at this epoch and next epoch.
After two epochs, it will add to unlocked balance, can withdraw directly from the contract.    

## About IBFT

### Introduction

IBFT bases on the original PBFT, which using a 3-phase consensus, PRE-PREPARE, PREPARE and COMMIT. The consensus can tolerate at most F faulty nodes  in 3F + 1 validator network.

The validator will select a proposer from themselves, then the proposer will proposal new blocks and broadcast message in PRE-PREPARE phase.
After other validators receive and validate the message, them will income proposal status and broadcast PREPARE message.
If the validator receive at least 2F + 1 PREPARE messages from other validator, it will change status from  PREPARE to COMMIT.
In order to inform other validators accept proposed a block, it will send COMMIT message.
When other validator receive 2F + 1 COMMIT messages, them will entry COMMITTED state and insert the block to the chain. 


### Round change flow

Under three status will send ROUND CHANGE message:

- Round change timer expires.
- Invalid PREPREPARE message.
- Block insertion fails. 
  
the validator send ROUND CHANGE message with proposed round number, the round number is selected from received F + 1 message 
When a validator notices that one of the above conditions applies, it broadcasts a ROUND CHANGE message along with the proposed round number and waits for ROUND CHANGE messages from other validators.
The proposed round number is selected based on following condition:

### Proposer selection

### Validator list voting