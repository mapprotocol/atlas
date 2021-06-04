pragma solidity ^0.6.12;


contract BlockData{
    // block hash and header
    mapping(string => header) public blockHeader;
    
    // block heith and hash
    mapping(uint256 => hashHave) public heightBlockHash;
    
    struct header{
    	string ParentHash;  
    	string UncleHash;  
    	address Coinbase;
    	string Root;
    	string TxHash;     
    	string ReceiptHash;
    	uint256 Difficulty;
    	uint256 Number;
    	uint256 Time;
    	uint256 Nonce;
    }
    
    struct hashHave{
        string hash;
        bool isHave;
    }
    
    
    modifier isBlockHave(uint256 height){
        require(!heightBlockHash[height].isHave,"block header is have");
        _;
    }
    
    function saveBlock(string memory ParentHash,string memory UncleHash,address Coinbase,string memory Root,
        string memory TxHash,string memory ReceiptHash,uint256  Difficulty,uint256  Number,
        uint256 Time, uint256 Nonce) public isBlockHave(Number){
            
            heightBlockHash[Number] = hashHave({hash:TxHash,isHave:true});
            blockHeader[TxHash] = header(
                    {   ParentHash: ParentHash,
                        UncleHash :UncleHash,
                        Coinbase :Coinbase,
                        Root :Root,
                        TxHash :TxHash,
                        ReceiptHash : ReceiptHash,
                        Difficulty :Difficulty,
                        Number : Number,
                        Time : Time,
                        Nonce : Nonce
                    }
                );
    }
}