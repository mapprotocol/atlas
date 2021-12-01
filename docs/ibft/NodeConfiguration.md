---
sort: 3
---

## Become validator

Encode types.IstanbulExtra with params you define and put it into the extraData field of the genesis block.
In the types.IstanbulExtra, we can assign the validators at the `first epoch` by providing addresses and BLS public keys.
After rlp encoded, maximum size extra data should not exceed `32`. See example at https://github.com/mapprotocol/atlas/blob/main/tools/makeValidator_test.go

- types.IstanbulExtra struct

```
type IstanbulExtra struct {
	// AddedValidators are the validators that have been added in the block
	AddedValidators []common.Address
	// AddedValidatorsPublicKeys are the BLS public keys for the validators added in the block
	AddedValidatorsPublicKeys []blscrypto.SerializedPublicKey
	// RemovedValidators is a bitmap having an active bit for each removed validator in the block
	// It use binary of big.Int to record removed validators, and number of big.Int is meaningless. 
	RemovedValidators *big.Int
	// Seal is an ECDSA signature by the proposer, it's created when proposer packs block
	// the seal is used for other validators verify legality of block.  
	Seal []byte
	// AggregatedSeal contains the aggregated BLS signature created via IBFT consensus.
	AggregatedSeal IstanbulAggregatedSeal
	// ParentAggregatedSeal contains and aggregated BLS signature for the previous block.
	ParentAggregatedSeal IstanbulAggregatedSeal
}
```

- encode example

```
package test

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/core/types"
	blscrypto "github.com/mapprotocol/atlas/params/bls"
	"math/big"
	"testing"
)

func TestEncodeData(t *testing.T) {

	// There are at least four addresses
	// In here, you should use private key created by yourself, else the data is not valid
	ads := make([]common.Address, 0)
	sliceOfAddressHexStr := []string{
		"0x1c0eDab88dbb72B119039c4d14b1663525b3aC15",
		"0x16FdBcAC4D4Cc24DCa47B9b80f58155a551ca2aF",
		"0x2dC45799000ab08E60b7441c36fCC74060Ccbe11",
		"0x6C5938B49bACDe73a8Db7C3A7DA208846898BFf5",
	}

	for _, s := range sliceOfAddressHexStr {
		ads = append(ads, common.HexToAddress(s))
	}

    // account publick key, use your public key to instead of them
	apks := make([]blscrypto.SerializedPublicKey, 0)
	sliceOfBlsPubKeyHexStr := []string{
		"0xbe77f945929d5dd3fe99aa825df0f5b1e8ea11786333b4492a8624a4d08dcee0e89df327359e8ec3f2d8ae01e938b7003414aa2d6523ffa02fde42b278cbae311fd39f1fbcad8e3188442ea31dee662389599751f8e73b99215cefc2e0003f81",
		"0x4f38a71fb13ab20f7bbfc2749ab15d775b7729842d967ca4f4115d1fcb3f378c892d073344f84e2abd8995a16eeee8004f4e588c30261e08a5dae70c581f904ea86b574bfe279222cf6b7913bebb0d3bd6c2bbe2e2ea1d338f145c4d95b99201",
		"0x8cf3bfcbfc76e9a99b70cad65ae51f8a8972e3e230445a55c8cf6b96dea7a2d0d970e3545e1316554d5d3b0a53582800ad4de92e3b06b62aa6f7677fdc2885a90b75fd80e2db2775512d8f3d3900aabae5b0525786d65615994b07afe7f69481",
		"0x1bbb8eb14a7f5dddc9de3356ce4247dab8e554fa83cd33e663db148b5d2dd14485f090978c84074154b450329de06b018eac04113ede1eedadf891ee862877af92a648c162be62182db90e8c83f8fd154fc14f13676bcb1fe3503260b6261a01",
	}

	for _, s := range sliceOfBlsPubKeyHexStr {
		pk1 := blscrypto.SerializedPublicKey{}
		pk1.UnmarshalText([]byte(s))
		apks = append(apks, pk1)
	}

	ist := types.IstanbulExtra{
		AddedValidators:           ads,
		AddedValidatorsPublicKeys: apks,
		RemovedValidators:         big.NewInt(0),
		Seal:                      []byte(""),
		AggregatedSeal:            types.IstanbulAggregatedSeal{},
		ParentAggregatedSeal:      types.IstanbulAggregatedSeal{},
	}

	payload, err := rlp.EncodeToBytes(&ist)
	if err != nil {
		fmt.Printf("encode failed: %#v", err.Error())
		return
	}

	finalExtra := append(bytes.Repeat([]byte{0x00}, types.IstanbulExtraVanity), payload...)

	fmt.Printf("finalExtra :%s\n", hexutil.Encode(finalExtra))
}
```

- Copy data to func `DefaultGenesisBlock`, the path is  `atlas/core/chain/genesis.go`

```
// this is our encoded data
const mainnetExtraData = "0x0000000000000000000000000000000000000000000000000000000000000000f901ebf854941c0edab88dbb72b119039c4d14b1663525b3ac159416fdbcac4d4cc24dca47b9b80f58155a551ca2af942dc45799000ab08e60b7441c36fcc74060ccbe11946c5938b49bacde73a8db7c3a7da208846898bff5f90188b860be77f945929d5dd3fe99aa825df0f5b1e8ea11786333b4492a8624a4d08dcee0e89df327359e8ec3f2d8ae01e938b7003414aa2d6523ffa02fde42b278cbae311fd39f1fbcad8e3188442ea31dee662389599751f8e73b99215cefc2e0003f81b8604f38a71fb13ab20f7bbfc2749ab15d775b7729842d967ca4f4115d1fcb3f378c892d073344f84e2abd8995a16eeee8004f4e588c30261e08a5dae70c581f904ea86b574bfe279222cf6b7913bebb0d3bd6c2bbe2e2ea1d338f145c4d95b99201b8608cf3bfcbfc76e9a99b70cad65ae51f8a8972e3e230445a55c8cf6b96dea7a2d0d970e3545e1316554d5d3b0a53582800ad4de92e3b06b62aa6f7677fdc2885a90b75fd80e2db2775512d8f3d3900aabae5b0525786d65615994b07afe7f69481b8601bbb8eb14a7f5dddc9de3356ce4247dab8e554fa83cd33e663db148b5d2dd14485f090978c84074154b450329de06b018eac04113ede1eedadf891ee862877af92a648c162be62182db90e8c83f8fd154fc14f13676bcb1fe3503260b6261a018080c3808080c3808080"
// add our encoded data to config
func DefaultGenesisBlock() *Genesis {
	dr := defaultRelayer()
	for addr, allc := range genesisRegisterProxyContract() {
		dr[addr] = allc
	}

	return &Genesis{
		Config:    params2.MainnetChainConfig,
		Nonce:     66,
		// this is our encoded data, maximum size extra data may be 32 after Genesis.
		ExtraData: hexutil.MustDecode(mainnetExtraData),
		GasLimit:  50000000,
		Alloc:     dr,
	}
}
```

## Start to mine at mainnet

Before start up, notify:
- It needs to become validator.

- It's best not to use `eth,personal` RPC  When node is unlocked.

- If you need to use `eth,personal` RPC, it's best not to let anyone know in case someone transfers money without a password.

After started up, input command `admin.nodeInfo` to get `enode` in atlas console, use command `admin.addPeer` to link four different nodes and wait for mining blocks.
```
admin.addPeer("enode://cb63c953384918826f4a9413ce54e255918027fe78e6ed1f65ce9705e2c434c57b6e8307044601d098489d243a298984afa4c7a8dcc862b38fc604e4050699e9@127.0.0.1:21221")
admin.addPeer("enode://60e990d0b4ff7c8d9c0403feb7637c4d3f21f7a38777b776501bb09be05622a1ed1090da9cb77ba850fb6fcdea5416e84edcf0477cab8d81b2d19e6c1a813888@127.0.0.1:21222")
admin.addPeer("enode://cd63432c612e3185e24b0be116c5e2a0804a94cb65cf0ba2787c42863bc577ef2004b04faa4e12ce1fcc2d5ff040e27241f354b62d7ab59cc6db63e2d2b19c9c@127.0.0.1:21223")
admin.addPeer("enode://40d2dc2a51298f3d9e2eabea17ae20781cb65ceb351681db6ab45e8086ded21aa1fcac08767bab383dec89723565c9f161dbb3cfa7cfd8ef41d049cdf0f44f26@127.0.0.1:21224")
```

- example
  To start a `atlas` instance for mining, run it with all your usual flags, extended by:

```shell
$ atlas <usual-flags> --datadir ./data1 --ipcpath data1 --port 20201 --unlock 0x6c5938b49bacde73a8db7c3a7da208846898bff5 --mine --miner.validator 0x6c5938b49bacde73a8db7c3a7da208846898bff5 console
```

Repeat four times to start up four different nodes, `miner.etherbase` params is just address we defined, `unlock` params is the same as `miner.etherbase`,
it's better to has different `datadir` params every node, `ipcpath` params should be the same as `datadir`.

### Operating a private network

Maintaining your own private network is more involved as a lot of configurations taken for
granted in the official networks need to be manually set up.

#### Defining the private genesis state
First, you'll need to create the genesis state of your networks, which all nodes need to be
aware of and agree upon. This consists of a small JSON file (e.g. call it `genesis.json`):

```json
{
  "config": {
    "chainId": <arbitrary positive integer>,
    "homesteadBlock": 0,
    "daoForkBlock":0,
	"daoForkSupport":true,
    "eip150Block": 0,
    "eip150Hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "muirGlacierBlock":0,
    "istanbulBlock": 0,
    "berlinBlock": 0
  },
  "alloc": {},
  "coinbase": "0x0000000000000000000000000000000000000000",
  "difficulty": "0x20000",
  "extraData": "",
  "gasLimit": "0x2fefd8",
  "nonce": "0x0000000000000042",
  "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "timestamp": "0x00"
}
```

If you'd like to pre-fund some accounts for easier testing, create
the accounts and populate the `alloc` field with their addresses.

```json
"alloc": {
  "0x0000000000000000000000000000000000000001": {
    "balance": "111111111"
  },
  "0x0000000000000000000000000000000000000002": {
    "balance": "222222222"
  }
}
```

With the genesis state defined in the above JSON file, you'll need to initialize **every**
`atlas` node with it prior to starting it up to ensure all blockchain parameters are correctly
set:

```shell
$ atlas init path/to/genesis.json
```


#### Starting up your member nodes

With the bootnode operational and externally reachable (you can try
`telnet <ip> <port>` to ensure it's indeed reachable), start every subsequent `atlas`
node pointed to the bootnode for peer discovery via the `--bootnodes` flag. It will
probably also be desirable to keep the data directory of your private network separated, so
do also specify a custom `--datadir` flag.

```shell
$ atlas --datadir=path/to/custom/data/folder --bootnodes=<bootnode-enode-url-from-above>
```

Repeat four times,the nodes connect to each other for IBFT consensus.

*Note: Since your network will be completely cut off from the main and test networks, you'll
also need to configure a miner to process transactions and create new blocks for you.*

#### Running a private miner

First, it needs to become a validator.

Mining on the public Atlas network is a complex task as it's only feasible using GPUs,
requiring an OpenCL or CUDA enabled `ethminer` instance. For information on such a setup, please consult the EtherMining subreddit and the ethminer repository.

In a private network setting, however a single CPU miner instance is more than enough for
practical purposes as it can produce a stable stream of blocks at the correct intervals
without needing heavy resources (consider running on a single thread, no need for multiple
ones either). To start a `atlas` instance for mining, run it with all your usual flags, extended
by:

```shell
$ atlas <usual-flags> --miner.validator=0x0000000000000000000000000000000000000000
```

Which will start mining blocks and transactions on a single CPU thread, crediting all
proceedings to the account specified by `--miner.validator`. You can further tune the mining by changing the default gas price transactions converge to (`--miner.gasprice`).
