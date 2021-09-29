## Atlas 启动说明

### 1.切换到 atlas 项目下执行 make 命令进行编辑

### 2.执行以下命令

nohup

./build/bin/atlas --datadir /root/atlas_node/data --rpc --rpcaddr "0.0.0.0" --rpcapi "eth,relayer,header,net,debug,txpool,personal" --mine --miner.threads 5 --allow-insecure-unlock --rpccorsdomain "\*" >> /root/atlas_node/log/output.log 2>&1

&

此命令将日志放在 atlas 目录下 output.log 文件

### 3.查看控制台：

./build/bin/atlas attach /root/atlas_node/data/atlas.ipc

### 注意事项：

1.relayer 参数调整 这几个参数有可能在服务器被改动 重新拉代码有可能覆盖

params/common.go:

MaxRedeemHeight uint64 = 200

NewEpochLength uint64 = 200

ElectionPoint uint64 = 20
