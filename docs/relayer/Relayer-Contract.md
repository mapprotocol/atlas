---
sort: 1
---
# Relayer-Contract

## Relayer Introduce

Every node can participate as a relayer to proposal new blocks with account (100 relayer only).
Depending on the performance,the registered accounts will be elected as relayer every epoch (about 10000 block number).
And election time is before 200 block number in the current epoch.
If relayer not completes task (about synchronising 1000 block) in current epoch, it will be not elected as relayer next epoch.

## Interact with contract interface

### register

One node can participate as a relayer to proposal new blocks with `register` function. To
become relayer, you need to `register` with some `eth` coins. You can charge for additional `fee` rate from the reward.

The feeRate calculation:

feeRate = fee / 10000

### parameters

| parameter | type    | comment                                                      |
| --------- | ------- | ------------------------------------------------------------ |
| pubkey    | bytes   | BFT public key of 65 bytes                                   |
| fee       | uint256 | percent of reward charged for relayer, the rate = fee / 10000 |
| value     | uint256(eth)     | `eth` token to register |


### unregister

You can unregister a portion of the registered from account balance. With the `unregister` transaction executed, the unregistering portion is locked in the contract for about 2 epoch.
After the period, you can withdraw the unregistered coins.

| parameter | type    | comment                                        |
| :-------- | ------- | ---------------------------------------------- |
| value     | uint256(eth) | unregister a portion of balance, the unit is `eth` |

### append

After register, you can append extra `eth` token to the contract by `append` function.

| parameter | type | comment                                         |
| --------- | ---- | ----------------------------------------------- |
| value     | uint256(eth)  | amout of coin append by the account |

### withdraw

You can withdraw the unregistered token after a period of 2 epoch.

| parameter | type         | comment                                 |
| --------- | ------------ | --------------------------------------- |
| value     | uint256(eth) | amount of value withdrawed to the owner |

### getBlance

You can query balance status by `getBalance` function. there are 3 states for the
deposit: `registerde`, `unregistering`, `unregistered`

* registered: token which you bond to contract
* unregistering: token which are unregistering but are still locked in the contract util 2 epoch.
* unregistered: you use withdraw token of unregistered state

| parameter | type    | comment         |
| --------- | ------- | --------------- |
| owner     | address | account address |

`getBalance` function outputs tuple of 3 items:

| parameter     | type    | comment                                                   |
| ------------- | ------- | --------------------------------------------------------- |
| registered    | uint256 | amount which is registered in contract                    |
| unregistering | uint256 | amount which is unregistering but still is in lock period |
| unregistered  | uint256 | amount which you can withdraw                             |

### getRelayer

You can query that you are registered or not, relayer or not

| parameter | type    | comment         |
| --------- | ------- | --------------- |
| owner     | address | account address |

`getRelayer` function outputs tuple of 3 items:

| parameter     | type    | comment                 |
| ------------- | ------- | ----------------------- |
| register      | bool    | you are register or not |
| relayer       | bool    | you are relayer or not  | 
| epoch         | uint256 | the current epoch       |  

### getPeriodHeight

You can query that started block number, ended block number and remained block number in the current epoch

| parameter | type    | comment         |
| --------- | ------- | --------------- |
| owner     | address | account address |

`getPeriodHeight` function outputs 4 items:

| parameter     | type    | comment                                   |
| ------------- | ------- | ----------------------------------------- |
| started       | uint256 | started block number in the current epoch |
| ended         | uint256 | ended block number in the current epoch   | 
| remained      | uint256 | remained block number in the current epoch|  
| relayer       | bool    | you are relayer or not                    | 