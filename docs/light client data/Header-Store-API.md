---
sort: 2
---

# Header-Store-API

## chain identification list

| chain ID | comment           |
| ---------| ----------------- | 
| 1000     | MAP chain         |
| 1001     | ethereum chain (main chain)    |
| 1002     | ethereum chain (ropsten chain start Number 800 )    |
| 1003     | ethereum chain (Private chain  (--dev))|

## CurrentHeaderNumber

get the synchronized part height of the corresponding chain

### request parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| chainID   | number | [chain identification](#chain identification list)  |

### response parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| result    | number | current block height |

### example

```shell

# request:

curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"header_currentHeaderNumber","params":[1001],"id":1}' http://127.0.0.1:7445

# response:

{
  "jsonrpc": "2.0",
  "id": 1,
  "result": 100
}

```

## GetHashByNumber

get the block hash of the corresponding chain by number

### request parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| chainID   | number | [chain identification](#chain identification list) |
| number    | number | block number |

### response parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| result    | number | block hash |

### example

```shell

# request:

curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"header_getHashByNumber","params":[1001, 1],"id":1}' http://127.0.0.1:7445

# response:

{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "0x88e96d4537bea4d9c05d12549907b32561d3bf31f45aae734cdc119f13406cb6"
}

```

## CurrentNumberAndHash

equivalent to the set of CurrentHeaderNumber and GetHashByNumber

### request parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| chainID   | number | [chain identification](#chain identification list) |

### response parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| number    | number | current block height |
| hash      | string | current block hash |

### example

```shell

# request:

curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"header_currentNumberAndHash","params":[1001],"id":1}' http://127.0.0.1:7445

# response:

{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "number": 0,
    "hash": "0x0000000000000000000000000000000000000000000000000000000000000000"
  }
}

```

## GetRelayerReward

get relayer rewards for the specified epoch

### request parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| epochID   | number | epoch id |
| relayer   | string | relayer address |

### response parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| result    | number | reward of relayer |

### example

```shell

# request:
curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"header_getRelayerReward","params":["0x1", "0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"],"id":1}' http://127.0.0.1:7445

# response:

{
  "jsonrpc": "2.0",
  "id": 1,
  "result": 15000000000
}

```
