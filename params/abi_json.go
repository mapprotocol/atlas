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

const TxVerifyABIJSON = `[
	{
		"inputs": [
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
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`
