
 ## Atlas multilChain introduce

Multichain databases are used to store data from different chains，Distinguishing different chains with ChainType uint64。The data structure is：
```golang
type HeaderChainStore struct {
	chainDb          ethdb.Database
	currentChainType rawdb.ChainType
	Mu               sync.RWMutex // blockchaindb insertion lock
	rand             *mrand.Rand
}
```
Call through the following function when you want to use multilChainDb.   ps:The only identification of the chain is chainType
```golang
func GetStoreMgr(chainType rawdb.ChainType) (*HeaderChainStore, error) {
	if storeMgr == nil {
		return nil, error01
	}
	storeMgr.currentChainType = chainType
	return storeMgr, nil
}
```
```golang
for Example:
    hc, error := GetStoreMgr(chainType)
    if error !=nil {
        ....
    }
```

## Building the multilChain db source

When the project starts, the multilChain db is initialized within the makeFullNode() function.    ps：【multiChain.NewStoreDb(ctx, &cfg.Eth) 】

## Interface function
Write block header information:

| name |params  | introduce|
|:---:|:---:|:---:|
|  WriteHeader| header | headerInfo (ethereum.Header) |

Get block header information with hash and number .  

| name |params  | introduce| return type|
|:---:|:---:|:---:|:----:|
|  ReadHeader|  Hash  | hashValue | *ethereum.Header |
|  |  number| blockNumber||
Insert what you want to deposit

| name |params  | introduce|
|:---:|:---:|:---:|
|  InsertHeaderChain |  chains | Incoming  []*ethereum.Header| 
|  |  start| time.Time for log|
```golang

for Example:
   	status, error := hc.InsertHeaderChain(chain, time.Now())
    if error !=nil {
        ....
    }
   	if status != wantStatus {
   		t.Errorf("wrong write status from InsertHeaderChain: got %v, want %v", status, wantStatus)
   	}
ps:
status value
NonStatTyState   WriteStatus = iota // not the Canonical will be ignore
CanonStatTyState                    // the Canonical
SideStatTyState                     // the branch
```
Get block number with hash . 

| name |params  | return type|
|:---:|:---:|:---:|
|  GetBlockNumber  |  hash |   *uint64|

Gets the block height of the current type chain:

| name |return type|
|:---:|:---:|
|CurrentHeaderNumber|uint64|

Gets the block lastHash of the current type chain：

| name |return type|
|:---:|:---:|
|CurrentHeaderHash|common.Hash|

Get difficulty with hash:

| name |return type|
|:---:|:---:|
|  GetTdByHash  |     *big.Int|
Get header information by hash:

| name |return type|
|:---:|:---:|
|  GetHeaderByHash  |    *Header|
Get header information for a type of chain through number:

| name |return type|
|:---:|:---:|
|  GetHeaderByNumber     |    *Header|

Get the headhash of a type of specification chain through number:

| name |return type|
|:---:|:---:|
|  ReadCanonicalHash  |common.Hash|







