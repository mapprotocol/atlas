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