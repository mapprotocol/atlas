## Configure the starting point of the synchronization block header

path：atlas/chains/chainsdb/config/

forexample: configuration eth_test_config.json file (This file saves the header information of an area of ropsten test network)

Fill in the following information：

```
{
	"parentHash": "",
	"sha3Uncles": "",
	"miner": "",
	"stateRoot": "",
	"transactionsRoot": "",
	"receiptsRoot": "",
	"logsBloom": "",
	"difficulty":,
	"number":,
	"gasLimit":,
	"gasUsed":,
	"timestamp":,
	"extraData": "",
	"mixHash": "",
	"nonce": "",
	"baseFeePerGas":
}
```

### Method 1:

### 1.Get information

```
POST https://ropsten.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161
Content-Type: application/json

{
	"jsonrpc": "2.0",
	"method": "eth_getBlockByNumber",
	"params": [
		"0xa796bc",
		 true
	],
	"id": 1
}
```

### 2.Request data through post. The following fields are in hexadecimal format and need to be converted to hexadecimal:

```
"difficulty": 1357813117,
"number": 10983100,
"gasLimit": 8000000,
"gasUsed": 113270,
"timestamp": 1630916403,
```

### 3.extraData Acquisition method：
The extradata needs to be converted to Base64 format for filling
tool path:https://github.com/mapprotocol/tools
```

func getBase64(num int64) {
	conn, _ := dialEthConn() // This method is used in the tool project  to create an eth connection
	h, _ := conn.HeaderByNumber(context.Background(), big.NewInt(num)) //The built-in API method obtains the corresponding block header
	encodedStr := base64.StdEncoding.EncodeToString(h.Extra)
	fmt.Println(encodedStr)
}
```

### Method 2:

use tool project
path：https://github.com/mapprotocol/tools/relayer_mock：common.go
call getjson(num) method print json config Infomation and  copy it to target file.json
```
func getJson(num int64) {
	conn, _ := dialEthConn()// This method is used in the tool project  to create an eth connection
	h, _ := conn.HeaderByNumber(context.Background(), big.NewInt(num))
	Header := &ethereum.Header{}
	Buffer := &bytes.Buffer{}
	h2, _ := convertChain(Header, Buffer, h)
	bs, err := json.Marshal(h2)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bs))
}
```
### Test:
Test path：
https://github.com/mapprotocol/atlas/tree/main/chains
chainsdb_test.go:

```
func TestRead_ethconfig(t \*testing.T) {
	data, err := ioutil.ReadFile(fmt.Sprintf("config/%v_config.json", "eth_test"))//need  specify config file location
	if err != nil {
		log.Error("read eht store config err", err)
	}
	genesis := &ethereum.Header{}
	err = json.Unmarshal(data, genesis)
	if err != nil {
	log.Error("Unmarshal Err", err.Error())
	}
	fmt.Println(genesis.Hash()) 
}

```
Compare the printed hash value with the test network target block hash to see whether it is consistent
Test network address：https://ropsten.etherscan.io/

### Synchronous data:
Use tool address:https://github.com/mapprotocol/tools/relayer_mock

modify parameters：github.com/mapprotocol/tools/relayer_mock/ common.go
1. EthUrl = "https://ropsten.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161" (Test network url)
2. AtlasRPCListenAddr = "localhost"（atlas host address）
3. AtlasRPCPortFlag   = 8082       （atlas Port number）
4. keystore1          = "./keystore/UTC--2021-07-19T02-04-57.993791200Z--df945e6ffd840ed5787d367708307
5. bd1fa3d40f4"  （Account private key address）The account number needs to meet the following conditions
   - 1.to become relayer（ https://mapprotocol.github.io/atlas/relayer/Relayer-Contract.html#register）
   - 2.The atlas currency of this account must be greater than or equal to 100000

6. Compile to get relay_mock executable

### 1.run tool project tool
run commond：relayer_mock save
If the synchronization is successful, a "success" prompt will be printed

### 2.query：

```
POST http://159.138.90.210:7445 //atlas server address
Content-Type: application/json

{
	"jsonrpc": "2.0",
	"method": "header_currentNumberAndHash",
	"params": [
		3 // Test network chainId
	],
	"id": 1
}
```

