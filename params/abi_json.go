package params

// HeaderStoreABIJSON  header store abi json
/*

type BlockHeader struct {
	From    *big.Int
	To      *big.Int
	Headers []byte
}

contract HeaderStore {
    event UpdateBlockHeader(address indexed account, uint256 indexed blockHeight);
    function updateBlockHeader(bytes memory blockHeader) public {}
    function currentNumberAndHash(uint256 chainID) public returns (uint256 number, bytes memory hash) {}
    function setRelayer(address relayer) public {}
    function getRelayer() public returns (address relayer) {}
    function reset(uint256 from, uint256 td, bytes memory header) public {}
    function verifyProofData(bytes memory receiptProof) public returns(bool success, string memory message, bytes memory logs) {}
}
*/
const HeaderStoreABIJSON = `[
	{
	   "anonymous": false,
	   "inputs": [
		  {
			 "indexed": true,
			 "internalType": "address",
			 "name": "account",
			 "type": "address"
		  },
		  {
			 "indexed": true,
			 "internalType": "uint256",
			 "name": "blockHeight",
			 "type": "uint256"
		  }
	   ],
	   "name": "UpdateBlockHeader",
	   "type": "event"
	},
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
	},
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
