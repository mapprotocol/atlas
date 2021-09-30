---
sort: 4
---

# Makalu PoC-1

Makalu PoC-1（Atlas v0.2） continues to use the POW consensus algorithm and basic block structure, and uses light client verification to 
verify on-chain transactions on the opposite chain. In order to achieve the purpose of cross-chain transaction data (assets).

Makalu PoC-1 currently supports the realization of Ethereum Ropsten testnet chain assets (USDT) cross-chain to Atlas chain 
assets in a completely decentralized manner. The Makalu PoC-1 architecture is mainly composed of three parts, Including 
on-chain contracts, relayers and main chain networks. Currently, only one-way data verification function to Atlas is supported.

The on-chain contract is used to define the cross-chain entry of messages between the main chain networks and the burn and mint processes of assets.
```Entrance
function swapOut(address token, address to, uint amount, uint toChainID) external
```

Relayer is used for the synchronization of light client data between chains, cross-chain transaction relay and other functions. 
Every third-party main chain network that joins the map cross-chain system can support any number of relayer relay services.

The main chain network is used to carry assets and verify cross-chain transactions.

{% include list.liquid %}
