---
sort: 1
---

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
        "internalType": "string",
        "name": "chain",
        "type": "string"
      }
    ],
    "name": "currentHeaderNumber",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "number",
        "type": "uint256"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "string",
        "name": "from",
        "type": "string"
      },
      {
        "internalType": "string",
        "name": "to",
        "type": "string"
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

### parameters

| parameter| type   | comment |
| -------- | ------ | ------- |
| from     | string | source chain identification |
| to       | string | destination chain identification |
| header   | []byte | block header json serialized data |

### currentHeaderNumber

get the current synchronized part height of the corresponding chain

### parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| chain     | string | chain identification |

