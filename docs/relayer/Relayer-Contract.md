---
sort: 1
---

# Relayer-Contract

## Relayer Introduce

Every node can participate as a relayer to proposal new blocks with account (100 relayer only).
Depending on the performance,the registered accounts will be elected as relayer every epoch (about 10000 block number).
And election time is before 200 block number in the current epoch.
If relayer not completes task (about synchronising 1000 block) in current epoch, it will be not elected as relayer next epoch.

## Contract Address

relayer contract is deployed at address:

```
0x00000000000052656c6179657241646472657373
```

## Contract json ABI

```json
[
{
"name": "Register",
"inputs": [
{
"type": "address",
"name": "from",
"indexed": true
},
{
"type": "uint256",
"name": "value",
"indexed": false
}
],
"anonymous": false,
"type": "event"
},
{
"name": "Withdraw",
"inputs": [
{
"type": "address",
"name": "from",
"indexed": true
},
{
"type": "uint256",
"name": "value",
"indexed": false
}
],
"anonymous": false,
"type": "event"
},
{
"name": "Unregister",
"inputs": [
{
"type": "address",
"name": "from",
"indexed": true
},
{
"type": "uint256",
"name": "value",
"indexed": false
}
],
"anonymous": false,
"type": "event"
},
{
"name": "Append",
"inputs": [
{
"type": "address",
"name": "from",
"indexed": true
},
{
"type": "uint256",
"name": "value",
"indexed": false
}
],
"anonymous": false,
"type": "event"
},
{
"name": "register",
"outputs": [],
"inputs": [
{
"type": "uint256",
"name": "value"
}
],
"constant": false,
"payable": false,
"type": "function"
},
{
"name": "append",
"outputs": [],
"inputs": [
{
"type": "uint256",
"name": "value"
}
],
"constant": false,
"payable": false,
"type": "function"
},
{
"name": "getRelayerBalance",
"outputs": [
{
"type": "uint256",
"unit": "wei",
"name": "register"
},
{
"type": "uint256",
"unit": "wei",
"name": "locked"
},
{
"type": "uint256",
"unit": "wei",
"name": "unlocked"
}
],
"inputs": [
{
"type": "address",
"name": "holder"
}
],
"constant": true,
"payable": false,
"type": "function"
},
{
"name": "withdraw",
"outputs": [],
"inputs": [
{
"type": "uint256",
"unit": "wei",
"name": "value"
}
],
"constant": false,
"payable": false,
"type": "function"
},
{
"name": "unregister",
"outputs": [],
"inputs": [
{
"type": "uint256",
"unit": "wei",
"name": "value"
}
],
"constant": false,
"payable": false,
"type": "function"
},
{
"name": "getPeriodHeight",
"outputs": [
{
"type": "uint256",
"name": "start"
},
{
"type": "uint256",
"name": "end"
},
{
"type": "bool",
"name": "relayer"
}
],
"inputs": [
{
"type": "address",
"name": "holder"
}
],
"constant": true,
"payable": false,
"type": "function"
},
{
"name": "getRelayer",
"inputs": [
{
"type": "address",
"name": "holder"
}
],
"outputs": [
{
"type": "bool",
"name": "relayer"
},
{
"type": "bool",
"name": "register"
},
{
"type": "uint256",
"name": "epoch"
}
],
"constant": true,
"payable": false,
"type": "function"
}
]
```

## Interact with contract interface

### register

One node can participate as a relayer to proposal new blocks with `register` function. To
become relayer, you need to `register` with some `eth` coins.

`register` function inputs 1 item:

| parameter | type    | comment                                                      |
| --------- | ------- | ------------------------------------------------------------ |
| value     | uint256(eth)     | `eth` token to register |


### unregister

You can unregister a portion of the registered from account balance. With the `unregister` transaction executed, the unregistering portion is locked in the contract for about 2 epoch.
After the period, you can withdraw the unregistered coins.

`unregister` function inputs 1 item:

| parameter | type    | comment                                        |
| :-------- | ------- | ---------------------------------------------- |
| value     | uint256(eth) | unregister a portion of balance, the unit is `eth` |

### append

After register, you can append extra `eth` token to the contract by `append` function.

`append` function inputs 1 item:

| parameter | type | comment                                         |
| --------- | ---- | ----------------------------------------------- |
| value     | uint256(eth)  | amout of coin append by the account |

### withdraw

You can withdraw the unregistered token after a period of 2 epoch.

`withdraw` function inputs 1 item:

| parameter | type         | comment                                 |
| --------- | ------------ | --------------------------------------- |
| value     | uint256(eth) | amount of value withdrawed to the owner |

### getRelayerBalance

You can query balance status by `getRegisterBalance` function. there are 3 states for the
balance: `registerde`, `unregistering`, `unregistered`

* registered: token which you bond to contract
* unregistering: token which are unregistering but are still locked in the contract util 2 epoch.
* unregistered: you use withdraw token of unregistered state

`getRelayerBalance` function inputs 1 item:

| parameter | type    | comment         |
| --------- | ------- | --------------- |
| owner     | address | account address |

`getRelayerBalance` function outputs 3 items:

| parameter     | type    | comment                                                   |
| ------------- | ------- | --------------------------------------------------------- |
| registered    | uint256 | amount which is registered in contract                    |
| unregistering | uint256 | amount which is unregistering but still is in lock period |
| unregistered  | uint256 | amount which you can withdraw                             |

### getRelayer

You can query that you are registered or not, relayer or not

`getRelayer` function inputs 1 item:

| parameter | type    | comment         |
| --------- | ------- | --------------- |
| owner     | address | account address |

`getRelayer` function outputs 3 items:

| parameter     | type    | comment                 |
| ------------- | ------- | ----------------------- |
| register      | bool    | you are register or not |
| relayer       | bool    | you are relayer or not  | 
| epoch         | uint256 | the current epoch       |  

### getPeriodHeight

You can query that started block number, ended block number and remained block number in the current epoch

`getPeriodHeight` function inputs 1 item:

| parameter | type    | comment         |
| --------- | ------- | --------------- |
| owner     | address | account address |

`getPeriodHeight` function outputs 3 items:

| parameter     | type    | comment                                   |
| ------------- | ------- | ----------------------------------------- |
| started       | uint256 | started block number in the current epoch |
| ended         | uint256 | ended block number in the current epoch   |
| relayer       | bool    | you are relayer or not                    | 