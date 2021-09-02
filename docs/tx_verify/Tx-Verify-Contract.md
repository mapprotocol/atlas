---
sort: 2
---

# Tx-Verify-Contract

## Contract Address

tx verify contract is deployed at address:

```
0x0000000000747856657269667941646472657373
```

## Contract JSON ABI

```json
[
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "srcChain",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "dstChain",
        "type": "uint256"
      },
      {
        "internalType": "bytes",
        "name": "txProve",
        "type": "bytes"
      }
    ],
    "name": "txVerify",
    "outputs": [
      {
        "internalType": "bool",
        "name": "success",
        "type": "bool"
      },
      {
        "internalType": "string",
        "name": "message",
        "type": "string"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  }
]
```

## Interact with contract interface

### txVerify

judge whether the transaction is true and valid by verifying the transaction receipt

#### input parameters

| parameter| type         | comment |
| -------- | ------------ | ------- |
| Router   | Address      | address of the contract that generated the cross-chain transaction event |
| Coin     | Address      | the address of the token contract |
| SrcChain | *big.Int     | source chain identification |
| DstChain | *big.Int     | destination chain identification|
| TxProve  | [CrossTxProve](https://mapprotocol.github.io/atlas/tx_verify/Tx-Verify) | cross chain transaction prove information |

#### output parameters

| parameter| type         | comment |
| -------- | ------------ | ------- |
| success | bool          | if the verification is successful, is true |
| message | string        | if the verification is successful, is empty |

### example

```
package main

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
)

type BaseParams struct {
	From  common.Address
	To    common.Address
	Value *big.Int
}

type CrossTxProve struct {
	Tx          *BaseParams
	Receipt     *types.Receipt
	Prove       light.NodeList
	BlockNumber uint64
	TxIndex     uint
}

func example() {
	var (
	    coin     = common.Address{}
	    router   = common.Address{}
		srcChain = big.NewInt(1)
		dstChain = big.NewInt(211)
	)

	txProve, err := rlp.EncodeToBytes(CrossTxProve{})
	if err != nil {
		panic(err)
	}

	ABITxVerify, _ := abi.JSON(strings.NewReader(""))
	input, err := ABITxVerify.Pack("txVerify", router, coin, srcChain, dstChain, txProve)
	if err != nil {
		panic(err)
	}
	
	// Send Transaction ...
	
	return
}
```

