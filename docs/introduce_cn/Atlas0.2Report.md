Atlas0.2 跟新内容介绍
==

### 实现了以太坊到atlas的跨连转账

1.在以太坊和atlas上各自部署了[转账合约](https://mapprotocol.github.io/atlas/Makalu%20PoC-1/Atlas-V0.2.html#2-deploy-contract-to-ethereum-and-atlas)

2.Atlas增加了对以太坊上跨连[交易验证](https://mapprotocol.github.io/atlas/tx_verify/)

3.修改了部署合约调用内置合约时不能调用问题

4.修改了atlas节点数据同步时出现的 merkle hash不一致问题

5.另外实现了[compass  demo](https://github.com/zhangwei812/compass)
模拟relayer对atlas
- 1.数据监控
- 2.执行交易验证
- 3.发送跨链交易

