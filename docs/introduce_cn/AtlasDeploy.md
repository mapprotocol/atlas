## Atlas项目启动说明
项目地址 https://github.com/mapprotocol/atlas.git
下载方式：git clone https://github.com/mapprotocol/atlas.git
编译：

- 进入atlas项目执行make命令编译项目
- 编译成功将提示：
  Done building.
  Run "./build/bin/atlas" to launch atlas.

运行项目：进入atlas项目执行一下命令
./build/bin/atlas --datadir /root/atlas_node/data  --rpc --rpcaddr "0.0.0.0" --rpcapi "eth,relayer,header,net,debug,txpool,personal" --mine --miner.threads 5  --rpccorsdomain "*" console

- --datadir /root/atlas_node/data 指定项目数据库路径
- --rpc   开启rpc服务
- --rpcaddr "0.0.0.0" RPC服务器监听地址
- --rpcapi "eth,relayer,header,net,debug,txpool,personal"      指定需要调用的HTTP-RPC API接口，默认只有eth,net,web3
- --mine 开启挖矿服务
- --miner.threads 5 指定挖矿线程数量
- --rpccorsdomain "*" 指定一个可以接收请求来源的以逗号间隔的域名列表
-  console 打开控制台


### 注意事项：

1.relayer参数
以下这几个参数有可能在服务器本地被改动 重新拉代码有可能覆盖

params/common.go:

MaxRedeemHeight uint64 = 200 相对每一届可赎回高度

NewEpochLength uint64  = 200  届高度

ElectionPoint uint64   = 20    距离届高度结束选举点位置
