---
sort: 1
---

# Makalu PoC-1

## Build your own network and send cross-chain transactions

### 1. Running atlas

Atlas run introduce
Project address: https://github.com/mapprotocol/atlas.git
Download method: git clone https://github.com/mapprotocol/atlas.git
compile：

- Enter the Atlas Project and execute "make" command to compile the project
- If the compilation is successful, the following message will be prompted:：
  Done building.
  Run "./build/bin/atlas" to launch atlas.

Run project：Enter the Atlas Project and execute the following commands
./build/bin/atlas --datadir /root/atlas_node/data  --rpc --rpcaddr "0.0.0.0" --rpcapi "eth,relayer,header,net,debug,txpool,personal" --mine --miner.threads 1  --rpccorsdomain "*" console

- --datadir /root/atlas_node/data   Specify the project database path
- --rpc   Turn on RPC service
- --rpcaddr "0.0.0.0" RPC server listening address
- --rpcapi "eth,relayer,header,net,debug,txpool,personal"      Specify the http-rpc API interface to be called. By default, there are only eth, net and Web3
- --mine Open mining service
- --miner.threads 1 Specifies the number of mining threads
- --rpccorsdomain "*" Specify a comma separated list of domain names from which requests can be received
-  console Open the console

### 2. Deploy contract to ethereum and atlas
(1). Deploy [USDT](https://github.com/mapprotocol/contracts/blob/884cc2b64ff14c01fa60251851537e2f877a14a2/contracts/Usdt.sol)

(2). Deploy [MapERC20](https://github.com/mapprotocol/contracts/blob/884cc2b64ff14c01fa60251851537e2f877a14a2/contracts/MapERC20.sol)

The constructor of this contract receives 3 parameters namely name, symbol and tokenBinding.
specify name and symbol assign the USDT contract address obtained in the previous step to tokenBinding

(3). Deploy [MapRouter](https://github.com/mapprotocol/contracts/blob/884cc2b64ff14c01fa60251851537e2f877a14a2/contracts/MapRouter.sol)

The constructor of this contract receives 2 parameters, mpcAddress and verify. mpcAddress is the address of the relayer 
that listen and verifies cross-chain transactions. This will be described in detail below. assign the 
[TxVerify](https://mapprotocol.github.io/atlas/tx_verify/Tx-Verify-Contract#contract-address) contract address to verify

(4). Set Authentication

Call the setAuth method of the MapERC20 contract and pass in the address of the MapRouter contract to add
authentication so that the MapRouter contract can call the MapERC20 contract。

### 3. Send cross chain transactions on the ethereum
Call the swapOut method of the MapRouter contract to send a cross chain transaction. The swapOut method receives
four parameters, namely token, to, amount, toChainID. 
token: token contract address, here can be MapERC20 contract address,
to: the address of the recipient of the transaction, 
amount: transaction amount, 
toChainId: the target chain of cross-chain transactions,
see [supported chain list](https://mapprotocol.github.io/atlas/light%20client%20data/Header-Store-API#chain-identification-list).

### 4. Register as a relayer and synchronize the block header to atlas
Atlas uses light client verification to verify on-chain transactions on the opposite chain to achieve the purpose of 
cross-chain transaction data (assets). So this requires you to synchronize the block header to atlas to achieve 
cross-chain transaction verification.

For how to register as a relayer and synchronize the block header, please refer to the document linked below.

[How To Become Relayer](https://mapprotocol.github.io/atlas/relayer/QuickStart.html)

[Sync Block Header](https://mapprotocol.github.io/atlas/light%20client%20data/Header-Store-Contract.html)

### 5. Listen cross chain transaction events and verify transactions
In this step, you need to implement a simple program to listen cross-chain transaction events on ethereum. 
When cross-chain transactions are monitored, cross-chain transaction verification data needs to be constructed, 
and then the sawpIn method of the MapRouter contract deployed on atlas is mobilized. 
The sawpIn method will call the pre-compiled contract of atlas to verify whether the cross-chain transaction exists and is legal. 
If the verification is passed, the transfer amount will be sent to the recipient of the cross-chain transaction.

[How to construct cross-chain transaction verification data](https://mapprotocol.github.io/atlas/tx_verify/Tx-Verify-Contract.html#example)

### 6. Query whether the cross-chain transaction has arrived
After the above steps, you have completed a cross-chain transaction. Now let's verify whether the transaction amount has arrived. 
To check the balance of the account, we can call the balanceOf of the MapERC20 contract deployed on atlas.

## Use our existing network to send cross-chain transactions

#### Atlas  address:

| network  | address | chainid | networkid |
| -------- | ------  | ------ | ------ |
| mainnet  | https://rpc-poc-1.maplabs.io | 177 | 177 |

#### Contract address:

| contract| address |
| -------- | ------ |
| Atlas MapRouter | 0xC8EBb59E097148399D67601038a6c6Aaa6416aF9 |
| Ropsten MapRouter | 0x23dd5A89C3ea51601b0674a4fA6eC6B3B14d0B7a|

Now that we have deployed a cross-chain transaction contract on Ropsten, you can directly call.
Let's take a look at how to send a cross-chain transaction in Ropsten.

The address of the MapRouter contract on the Ropsten testnet is 0x23dd5A89C3ea51601b0674a4fA6eC6B3B14d0B7a. 
Using this address, you can use the remix or Go program to call sawpOut to send cross-chain transactions. 
After sending a cross-chain transaction, you do not need to synchronize block headers, monitor cross-chain events, 
verify cross-chain transactions, etc., just wait for the transaction amount to arrive. 
The premise is that cross-chain transactions exist and are legal.

The Atlas main network address is https://rpc-poc-1.maplabs.io, and the address of the MapERC20 contract deployed 
on Atlas is 0xC8EBb59E097148399D67601038a6c6Aaa6416aF9. Using the above address, you can use the remix or Go program to 
call the balanceOf of the MapERC20 contract to check the balance.

### 7.[How to configure the starting point of the synchronization block header](https://mapprotocol.github.io/atlas/chains_config/ChainsConfig.html)
