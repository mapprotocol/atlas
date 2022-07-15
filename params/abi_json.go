package params

// HeaderStoreABIJSON  header store abi json
/*

type BlockHeader struct {
	From    *big.Int
	To      *big.Int
	Headers []byte
}

contract HeaderStore {
    function updateBlockHeader(bytes memory blockHeader) public {}
    function currentNumberAndHash(uint256 chainID) public returns (uint256 number, bytes memory hash) {}
    function setRelayer(address relayer) public {}
    function getRelayer() public returns (address relayer) {}
    function reset(uint256 from, uint256 td, bytes memory header) public {}
}
*/
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
		"inputs": [],
		"name": "getRelayer",
		"outputs": [
			{
				"internalType": "address",
				"name": "relayer",
				"type": "address"
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
				"name": "td",
				"type": "uint256"
			},
			{
				"internalType": "bytes",
				"name": "header",
				"type": "bytes"
			}
		],
		"name": "reset",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "relayer",
				"type": "address"
			}
		],
		"name": "setRelayer",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "bytes",
				"name": "blockHeader",
				"type": "bytes"
			}
		],
		"name": "updateBlockHeader",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

// TxVerifyABIJSON  tx verify abi json
/*

type ReceiptProof struct{
	Router   common.Address
	Coin     common.Address
	SrcChain *big.Int
	DstChain *big.Int
	TxProve  []byte
}

type TxProve struct {
	Receipt     *ethtypes.Receipt
	Prove       light.NodeList
	BlockNumber uint64
	TxIndex     uint
}

contract TxVerify {
    function verifyProofData(bytes memory receiptProof) public returns(bool success, string memory message, bytes memory logs) {}
}
*/
const TxVerifyABIJSON = `[
	{
		"inputs": [
			{
				"internalType": "bytes",
				"name": "receiptProof",
				"type": "bytes"
			}
		],
		"name": "verifyProofData",
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
			},
			{
				"internalType": "bytes",
				"name": "logs",
				"type": "bytes"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

const RelayerABIJSON = `[
  {
    "name": "Register",
    "inputs": [
      {
        "type": "address",
        "name": "from",
        "indexed": true
      },
      {
        "type": "uint256",
        "name": "value",
        "indexed": false
      }
    ],
    "anonymous": false,
    "type": "event"
  },
  {
    "name": "Withdraw",
    "inputs": [
      {
        "type": "address",
        "name": "from",
        "indexed": true
      },
      {
        "type": "uint256",
        "name": "value",
        "indexed": false
      }
    ],
    "anonymous": false,
    "type": "event"
  },
  {
    "name": "Unregister",
    "inputs": [
      {
        "type": "address",
        "name": "from",
        "indexed": true
      },
      {
        "type": "uint256",
        "name": "value",
        "indexed": false
      }
    ],
    "anonymous": false,
    "type": "event"
  },
  {
    "name": "Append",
    "inputs": [
      {
        "type": "address",
        "name": "from",
        "indexed": true
      },
      {
        "type": "uint256",
        "name": "value",
        "indexed": false
      }
    ],
    "anonymous": false,
    "type": "event"
  },
  {
    "name": "register",
    "outputs": [],
    "inputs": [
      {
        "type": "uint256",
        "name": "value"
      }
    ],
    "constant": false,
    "payable": false,
    "type": "function"
  },
  {
    "name": "append",
    "outputs": [],
    "inputs": [
      {
        "type": "uint256",
        "name": "value"
      }
    ],
    "constant": false,
    "payable": false,
    "type": "function"
  },
  {
    "name": "getRelayerBalance",
    "outputs": [
      {
        "type": "uint256",
        "unit": "wei",
        "name": "registered"
      },
      {
        "type": "uint256",
        "unit": "wei",
        "name": "unregistering"
      },
      {
        "type": "uint256",
        "unit": "wei",
        "name": "unregistered"
      }
    ],
    "inputs": [
      {
        "type": "address",
        "name": "owner"
      }
    ],
    "constant": true,
    "payable": false,
    "type": "function"
  },
  {
    "name": "withdraw",
    "outputs": [],
    "inputs": [
      {
        "type": "uint256",
        "unit": "wei",
        "name": "value"
      }
    ],
    "constant": false,
    "payable": false,
    "type": "function"
  },
  {
    "name": "unregister",
    "outputs": [],
    "inputs": [
      {
        "type": "uint256",
        "unit": "wei",
        "name": "value"
      }
    ],
    "constant": false,
    "payable": false,
    "type": "function"
  },
  {
    "name": "getPeriodHeight",
    "outputs": [
      {
        "type": "uint256",
        "name": "start"
      },
      {
        "type": "uint256",
        "name": "end"
      },
      {
        "type": "bool",
        "name": "relayer"
      }
    ],
    "inputs": [
      {
        "type": "address",
        "name": "owner"
      }
    ],
    "constant": true,
    "payable": false,
    "type": "function"
  },
  {
    "name": "getRelayer",
    "inputs": [
      {
        "type": "address",
        "name": "owner"
      }
    ],
    "outputs": [
      {
        "type": "bool",
        "name": "register"
      },
      {
        "type": "bool",
        "name": "relayer"
      },
      {
        "type": "uint256",
        "name": "epoch"
      }
    ],
    "constant": true,
    "payable": false,
    "type": "function"
  }
]`
