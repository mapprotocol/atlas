## 同步区块头起始点配置说明

路径：atlas/chains/chainsdb/config/

举例: 配置 eth_test_config.json 文件(此文件保存了ropsten测试网某一区块头信息)

填充配置文件中以下字段

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

### 2.通过post方式请求数据,以下字段是16进制格式需要将其转成10进制:

```
"difficulty": 1357813117,
"number": 10983100,
"gasLimit": 8000000,
"gasUsed": 113270,
"timestamp": 1630916403,
```

### 3.extraData 获取方式：
需要将 extraData 字段转成 base64进行填充
tool工具地址:https://github.com/mapprotocol/tools
```

func getBase64(num int64) {
	conn, _ := dialEthConn() // 此方法在tool工具内 用来创建Eth连接
	h, _ := conn.HeaderByNumber(context.Background(), big.NewInt(num)) //内置api方法获取对应块头
	encodedStr := base64.StdEncoding.EncodeToString(h.Extra)//通过base64模块转换
	fmt.Println(encodedStr)
}
```

### 方法二：

使用tool 工具
路径：https://github.com/mapprotocol/tools/relayer_mock：common.go
调用 getjson(num)可直接打印出 json配置信息复制到配置文件即可
```
func getJson(num int64) {
	conn, _ := dialEthConn()// 此方法在tool工具内 用来创建Eth连接
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
### 测试:
测试路径：
https://github.com/mapprotocol/atlas/tree/main/chains
chainsdb_test.go:

```
func TestRead_ethconfig(t \*testing.T) {
	data, err := ioutil.ReadFile(fmt.Sprintf("config/%v_config.json", "eth_test"))//指定配置文件位置
	if err != nil {
		log.Error("read eht store config err", err)
	}
	genesis := &ethereum.Header{}
	err = json.Unmarshal(data, genesis)
	if err != nil {
	log.Error("Unmarshal Err", err.Error())
	}
	fmt.Println(genesis.Hash()) //打印出hash
}

```
将打印出的hash值与测试网目标块hash进行对比看是否一致
测试网地址：https://ropsten.etherscan.io/

### 同步数据:
使用工具地址:https://github.com/mapprotocol/tools/relayer_mock

修改参数：github.com/mapprotocol/tools/relayer_mock/ common.go
1. EthUrl = "https://ropsten.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161" (测试网 url)
2. AtlasRPCListenAddr = "localhost"（atlas host地址）
3. AtlasRPCPortFlag   = 8082       （atlas 端口号）
4. keystore1          = "./keystore/UTC--2021-07-19T02-04-57.993791200Z--df945e6ffd840ed5787d367708307
5. bd1fa3d40f4"  （账号私钥地址）账号需要满足一下条件
    - 1.成为relayer（ https://mapprotocol.github.io/atlas/relayer/Relayer-Contract.html#register）
    - 2.此账号atlas币需要大于等于100000

6. 编译得到relayer_mock可执行程序

### 1.运行tool工具
运行命令：relayer_mock save
同步成功 会打印 success 提示

### 2.查询：

```
POST http://159.138.90.210:7445 //atlas服务器地址
Content-Type: application/json

{
	"jsonrpc": "2.0",
	"method": "header_currentNumberAndHash",
	"params": [
		3 // 测试网 chainId
	],
	"id": 1
}
```

