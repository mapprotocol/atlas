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
