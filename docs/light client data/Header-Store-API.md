---
sort: 2
---

# Header-Store-API

## CurrentHeaderNumber

get the synchronized part height of the corresponding chain m:1

### request parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| chain     | string | chain identification |

### response parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| result    | number | current block height |

### example

```shell

# request:

curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"header_currentHeaderNumber","params":["ETH"],"id":1}' http://127.0.0.1:7445

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
| chain     | string | chain identification |
| number    | number | block number |

### response parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| result    | number | block hash |

### example

```shell

# request:

# curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"header_getHashByNumber","params":["ETH", "1"],"id":1}' http://127.0.0.1:7445

# response:

{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "0x88e96d4537bea4d9c05d12549907b32561d3bf31f45aae734cdc119f13406cb6"
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

curl -X POST -H "Content-Type: application/json" --data '{{"jsonrpc":"2.0","method":"header_getRelayerReward","params":["1", "0xDf945e6FFd840Ed5787d367708307BD1Fa3d40f4"],"id":1}' http://127.0.0.1:7445

#response:

{
  "jsonrpc": "2.0",
  "id": 1,
  "result": 15000000000
}

```
