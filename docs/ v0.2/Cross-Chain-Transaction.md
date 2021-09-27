---
sort: 1
---

# Cross-Chain-Transaction

## 1. Running atlas

Going `atlas -h` can get help infos.

## 2. Deploy contract to ethereum and atlas
(1). Deploy [USDT](https://github.com/mapprotocol/contracts/blob/main/contracts/Usdt.sol)

(2). Deploy [MapERC20](https://github.com/mapprotocol/contracts/blob/main/contracts/MapERC20.sol)

The constructor of this contract receives 3 parameters namely name, symbol and tokenBinding.
specify name and symbol assign the USDT contract address obtained in the previous step to tokenBinding

(3). Deploy [MapRouter](https://github.com/mapprotocol/contracts/blob/main/contracts/MapRouter.sol)

The constructor of this contract receives 2 parameters, namely mpcAddress and verify. specify mpcAddress 
assign the [TxVerify](https://mapprotocol.github.io/atlas/tx_verify/Tx-Verify-Contract#contract-address) contract address to verify

(4). Set Authentication

Call the setAuth method of the MapERC20 contract and pass in the address of the MapRouter contract to add 
authentication so that the MapRouter contract can call the MapERC20 contract。

## 3. Send cross chain transactions on the ethereum
Call the swapOut method of the MapRouter contract to initiate a cross chain transaction. The swapOut method receives 
four parameters, namely token, to, amount, toChainID. token: token contract address, here can be MapERC20 contract address, 
to: the address of the recipient of the transaction, amount: transaction amount, toChainId: the target chain of cross-chain transactions,
see [supported chain list](https://mapprotocol.github.io/atlas/light%20client%20data/Header-Store-API#chain-identification-list).
