package params

const HeaderStoreABIJSON = `[
    {
        "inputs": [
            {
                "internalType": "uint256",
                "name": "chainID",
                "type": "uint256"
            }
        ],
        "name": "currentNumberAndHash",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "number",
                "type": "uint256"
            },
            {
                "internalType": "bytes",
                "name": "hash",
                "type": "bytes"
            }
        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "uint256",
                "name": "from",
                "type": "uint256"
            },
            {
                "internalType": "uint256",
                "name": "to",
                "type": "uint256"
            },
            {
                "internalType": "bytes",
                "name": "headers",
                "type": "bytes"
            }
        ],
        "name": "save",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`

// TxVerifyABIJSON  tx verify abi json
/*
contract TxVerify {
	function txVerify(address router, address coin, uint256 srcChain, uint256 dstChain, bytes memory txProve) public returns(bool success, string memory message){}
}
*/
const TxVerifyABIJSON = `[
    {
        "inputs": [
            {
                "internalType": "address",
                "name": "router",
                "type": "address"
            },
            {
                "internalType": "address",
                "name": "coin",
                "type": "address"
            },
            {
                "internalType": "uint256",
                "name": "srcChain",
                "type": "uint256"
            },
            {
                "internalType": "uint256",
                "name": "dstChain",
                "type": "uint256"
            },
            {
                "internalType": "bytes",
                "name": "txProve",
                "type": "bytes"
            }
        ],
        "name": "txVerify",
        "outputs": [
            {
                "internalType": "bool",
                "name": "success",
                "type": "bool"
            },
            {
                "internalType": "string",
                "name": "message",
                "type": "string"
            }
        ],
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`
