package params

const HeaderStoreABIJSON = `[
	{
		"inputs": [
			{
				"internalType": "string",
				"name": "chain",
				"type": "string"
			}
		],
		"name": "currentHeaderNumber",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "number",
				"type": "uint256"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "string",
				"name": "from",
				"type": "string"
			},
			{
				"internalType": "string",
				"name": "to",
				"type": "string"
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
