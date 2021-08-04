---
sort: 1
---

# Header-Store-Contract

## Contract Address

header store contract is deployed at address:

```
0x000068656164657273746F726541646472657373
```

## Contract json ABI

```json
[
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "chainID",
        "type": "uint256"
      }
    ],
    "name": "currentNumberAndHash",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "number",
        "type": "uint256"
      },
      {
        "internalType": "bytes",
        "name": "hash",
        "type": "bytes"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "from",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "to",
        "type": "uint256"
      },
      {
        "internalType": "bytes",
        "name": "headers",
        "type": "bytes"
      }
    ],
    "name": "save",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  }
]
```

## Interact with contract interface

### save

validate and save the block header synchronized by the relayer

#### parameters

| parameter| type   | comment |
| -------- | ------ | ------- |
| from     | number | source chain identification |
| to       | number | destination chain identification |
| headers  | []byte | block header json serialized data |

### currentHeaderNumber

get the current synchronized part height of the corresponding chain

#### parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| chainID   | number | chain identification |

