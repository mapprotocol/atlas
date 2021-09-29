## 配置流程说明

atlas/chains/chainsdb/config/:

举例: 配置 eth_test_config.json 文件

需要填充配置文件中这些字段

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

### 方法一:

### 1.获取字段

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

### 2.注意这几个字段要转成 10 进制:

```
"difficulty": 1357813117,
"number": 10983100,
"gasLimit": 8000000,
"gasUsed": 113270,
"timestamp": 1630916403,
```

### 3.extraData 获取方式：需要将 extraData 字段转成 base64

tool工具地址:https://github.com/mapprotocol/tools
```

func getBase64(num int64) {
	conn, _ := dialEthConn() // 此方法在tool 工具内 用来创建连接
	h, _ := conn.HeaderByNumber(context.Background(), big.NewInt(num))
	encodedStr := base64.StdEncoding.EncodeToString(h.Extra)
	fmt.Println(encodedStr)
}
```

### 方法二：

tool 工具中 common.go 里 getjson(num) 可直接打印出 json 配置复制到目标文件即可

tool工具地址:https://github.com/mapprotocol/tools

### 测试方法:修改测试路径得到 hash

atlas/chains/chainsdb/ chainsdb_test.go:

```
func TestRead_ethconfig(t \*testing.T) {
	data, err := ioutil.ReadFile(fmt.Sprintf("config/%v_config.json", "eth_test"))
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

### 测试同步:

### 1.运行 tool 工具启动参数为:savemock

修改 common.go 文件中 Ethurl 参数(测试网 url)

账户地址一般用默认 Relayer

Relayer 地址路径:
relayerMock/common.go : keystore1

同步成功 会打印 success 提示

### 2.同步成功查询：

```
POST http://159.138.90.210:7445
Content-Type: application/json

{
	"jsonrpc": "2.0",
	"method": "header_currentNumberAndHash",
	"params": [
		3
	],
	"id": 1
}
```
