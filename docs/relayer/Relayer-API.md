---
sort: 2
---

# Relayer-API

## GetAllRelayers

get all relayers in specified epoch on the basis of block number

### request parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| number    | string | block number |

### response parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| result    | address[] | relayers address |

### example

```shell
# request:
curl -X POST -H 'Content-Type:application/json' --data '{"jsonrpc":"2.0","method":"relayer_getAllRelayers","params":["latest"],"id":83}' http://127.0.0.1:7445 | jq

# response:
{
  "jsonrpc": "2.0",
  "id": 83,
  "result": [
    "0xdf945e6ffd840ed5787d367708307bd1fa3d40f4",
    "0x32cd75ca677e9c37fd989272afa8504cb8f6eb52",
    "0x3e3429f72450a39ce227026e8ddef331e9973e4d",
    "0x81f02fd21657df80783755874a92c996749777bf",
    "0x84d46b3055454646a419d023f73472561b6cf20f",
    "0x480fb8301d0d357956fb8db06988d4e5650c5fc7",
    "0x85273b522f9e17a57cec59f31f24a49a60c54e17",
    "0xf058b45ed9a2b781558c0b9ef8c63c79d615c3bb",
    "0x8ee567be17fb027cbb107ff70fc02dc475ce3f3e",
    "0xb5ac31a4a887e9f773b5fd0aba3fc0fe95c2a750"
  ]
}
```

## GetAccountInfo

query your account is registered or not, is elected for relayer or not in specified epoch on the basis of block number

### request parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| account   | address| account address |
| number    | string | block number |

### response parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| result    | string | account status |

### example

```shell
# request:
curl -X POST -H 'Content-Type:application/json' --data '{"jsonrpc":"2.0","method":"relayer_getAccountInfo","params":["0xdf945e6ffd840ed5787d367708307bd1fa3d40f4","latest"],"id":83}' http://127.0.0.1:7445 | jq

# response:
{
  "jsonrpc": "2.0",
  "id": 83,
  "result": "current epoch: 1, register status: true, relayer status: true"
}
```

## GetSyncNumber

query block number relayer Synchronized in specified epoch on the basis of block number

### request parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| number    | string | block number |
| relayer   | address | relayer address |

### response parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| result    | uint256| block number relayer Synchronized |

### example

```shell
# request:
curl -X POST -H 'Content-Type:application/json' --data '{"jsonrpc":"2.0","method":"relayer_getSyncNumber","params":["0xdf945e6ffd840ed5787d367708307bd1fa3d40f4","latest"],"id":83}' http://127.0.0.1:7445 | jq

#response:
{
  "jsonrpc": "2.0",
  "id": 83,
  "result": 540
}
```

## GetRelayerBalance

query registered balance in your account

### request parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| number    | string | block number |
| relayer   | address | relayer address |

### response parameters

| parameter | type   | comment |
| --------- | ------ | ------- |
| result    | string | balance message |

### example

```shell
# request:
curl -X POST -H 'Content-Type:application/json' --data '{"jsonrpc":"2.0","method":"relayer_getBalance","params":["0xdf945e6ffd840ed5787d367708307bd1fa3d40f4","latest"],"id":83}' http://127.0.0.1:7445 | jq

#response:
{
  "jsonrpc": "2.0",
  "id": 83,
  "result": " registered balance:100020 ETH, unregistering balance:0 ETH, unregistered balance:100000 ETH"
}
```