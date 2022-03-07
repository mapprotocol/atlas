// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.7.1;

import "../lib/RLPReader.sol";
import "../lib/RLPEncode.sol";
import "../bn256/BlsSignatureTest.sol";

contract sync {

    struct blockHeader{
        //bytes32 parentHash;
        address coinbase;
        //bytes32 root;
        //bytes32 txHash;
        //bytes32 receipHash;
        //bytes bloom;
        uint256 number;
        //uint256 gasLimit;
        //uint256 gasUsed;
        //uint256 time;
        bytes extraData;
        //bytes32 mixDigest;
        //bytes nonce;
        //uint256 baseFee;
    }
    mapping(uint256 => bytes) private allHeader;

    struct istanbulAggregatedSeal{
        //uint256   round;
        bytes     signature;
        uint256   bitmap;
    }

    struct istanbulExtra{
        //address[] validators;
        bytes  seal;
        istanbulAggregatedSeal  aggregatedSeals;
        istanbulAggregatedSeal  parentAggregatedSeals;
        uint256  removeList;
        bytes[]  addedPubKey;
    }

    mapping(uint256 => bytes[]) private allkey;
    uint256 nowEpoch;
    uint256 nowNumber;
    address rootAccount;
    uint256 epochLength;
    uint256 maxSyncNum;
    uint256 keyNum;

    using RLPReader for bytes;
    using RLPReader for uint;
    using RLPReader for RLPReader.RLPItem;
    using RLPReader for RLPReader.Iterator;

    event setParams(string s,uint256 v);
    event setParams(string s,bytes v);
    event log(string s,bool e);

    constructor(bytes memory firstBlock,uint _epochLength) {
        rootAccount = msg.sender;
        epochLength = _epochLength;
        maxSyncNum = 10;
        nowNumber == 0;
        nowEpoch = 0;
        initFirstBlock(firstBlock);
    }

    //the first block is used for init params,
    //it was operated specially. 
    function initFirstBlock(bytes memory firstBlock) private{
        blockHeader memory bh = decodeHeaderPart1(firstBlock);
        bytes memory extra = splitExtra(bh.extraData);
        istanbulExtra memory ist = decodeExtraData(extra);

        keyNum = ist.addedPubKey.length;
        nowNumber = bh.number;
        require(nowNumber == 0);
        allkey[nowEpoch] = new bytes[](keyNum);

        for(uint8 i = 0;i<keyNum;i++){
            allkey[nowEpoch][i] = ist.addedPubKey[i];
        }
        allHeader[nowNumber] = firstBlock;
    }

    //
    function setBLSPublickKeys(bytes[] memory keys,uint256 epoch) public {
        require(msg.sender == rootAccount, "onlyRoot");
        emit setParams("current epoch",epoch);
        allkey[epoch] = new bytes[](keys.length);
        for (uint i=0;i<keys.length;i++){
            emit setParams("setBLSPublickKey",keys[i]);
            allkey[epoch][i] = keys[i];
        }    
    }

    function setMaxSyncNum(uint8 max) public{
        require(msg.sender == rootAccount, "onlyRoot");
        emit setParams("setMaxSyncNum",max);
        maxSyncNum = max;
    }

    function checkNowParams() public view returns(uint,uint,uint,uint){
        //require(msg.sender == rootAccount, "onlyRoot");
        return (maxSyncNum,nowEpoch,nowNumber,keyNum);
    }

    function checkBLSPublickKeys(uint256 epoch) public view returns(bytes[] memory){
        //require(msg.sender == rootAccount, "onlyRoot");
        return allkey[epoch];
    }

    function checkBlockHeader(uint256 number) public view returns(bytes memory){
        //require(msg.sender == rootAccount, "onlyRoot");
        return allHeader[number];
    }

    //function checkExtraData(uint256 number) public view returns(istanbulExtra memory){
    //    require(msg.sender == rootAccount, "onlyRoot");
    //    return allExtra[number];
    //}

    function decodeHeaderPart1(bytes memory rlpBytes)public pure returns(blockHeader memory bh){
        RLPReader.RLPItem[] memory ls = rlpBytes.toRlpItem().toList();
        //RLPReader.RLPItem memory item0 = ls[0]; //parentBlockHash
        RLPReader.RLPItem memory item1 = ls[1]; //coinbase
        //RLPReader.RLPItem memory item2 = ls[2]; //root
        //RLPReader.RLPItem memory item3 = ls[3]; //txHash
        //RLPReader.RLPItem memory item4 = ls[4]; //receipHash
        RLPReader.RLPItem memory item6 = ls[6]; //number
        RLPReader.RLPItem memory item10 = ls[10]; //extra

        //bh.parentHash = bytes32(item0.toBytes());
        bh.coinbase = item1.toAddress();
        //bh.root = bytes32(item2.toBytes());
        //bh.txHash = bytes32(item3.toBytes());
        //bh.receipHash = bytes32(item4.toBytes());
        bh.number = item6.toUint();
        bh.extraData = item10.toBytes(); 
        return bh;
    }

    //function decodeHeaderPart2(bytes memory rlpBytes,blockHeader memory bh)public pure returns(blockHeader memory){
        //RLPReader.RLPItem[] memory ls = rlpBytes.toRlpItem().toList();
        //RLPReader.RLPItem memory item5 = ls[5]; //bloom
        //RLPReader.RLPItem memory item7 = ls[7]; //gasLimit
        //RLPReader.RLPItem memory item8 = ls[8]; //gasUsed
        //RLPReader.RLPItem memory item9 = ls[9]; //time
        //RLPReader.RLPItem memory item11 = ls[11]; //mixDigest
        //RLPReader.RLPItem memory item12 = ls[12]; //nonce
        //RLPReader.RLPItem memory item13 = ls[13]; //baseFee

        //bh.bloom = item5.toBytes();
        //bh.gasLimit = item7.toUint();
        //bh.gasUsed = item8.toUint();
        //bh.time = item9.toUint();
        //bh.mixDigest  = bytes32(item11.toBytes());
        //bh.nonce  = item12.toBytes();
        //bh.baseFee = item13.toUint();
        //return bh;
    //}


    function verifymoreHeaders(bytes[] memory moreRlpHeader,bytes[] memory moreHeaderBytes/*,bytes32[] moreBlockHash*/)public returns(uint,bool){
        require(moreHeaderBytes.length == moreRlpHeader.length);
        require(maxSyncNum > moreHeaderBytes.length);
        for(uint i=0;i<moreRlpHeader.length;i++){
            bool ret = verifyHeader(moreRlpHeader[i],moreHeaderBytes[i]/*,moreBlockHash[i]*/);
            if (ret == false){
                return (i,false);
            }
        }
        return (moreRlpHeader.length,true);
    }

    //the input data is the header after rlp encode within seal by proposer and aggregated seal by validators.
    function verifyHeader(bytes memory rlpHeader,bytes memory escaMsg/*,bytes32 blockHash*/) public returns(bool){
        bool ret = true;

        //it only decode data about validation,so that reduce storage and calculation.
        blockHeader memory bh = decodeHeaderPart1(rlpHeader);
        //bh = decodeHeaderPart2(rlpHeader,bh);
        bytes memory extra = splitExtra(bh.extraData);
        istanbulExtra memory ist = decodeExtraData(extra);

        //the escaMsg is the hash of the header without seal by proposer and aggregated seal by validators.
        //bytes memory escaMsg = cutSealAndAgg(rlpHeader);
        bytes32 HeaderSignHash = keccak256(abi.encodePacked(escaMsg));
        //the esca seal signed by proposer
        ret = verifySign(ist.seal,HeaderSignHash,bh.coinbase);
        if (ret == false) {
            revert("verifyEscaSign fail");
            //return false;
        }
        emit log("verifyEscaSign pass",true);

        
        //the blockHash is the hash of the header without aggregated seal by validators.
        //bytes memory blockHash = cutAgg(rlpHeader);
        if (bh.number%epochLength == 0){
            //ret = verifyAggregatedSeal(allkey[nowEpoch],ist.aggregatedSeals.Signature,blockHash);
            //it need to update validators at first block of new epoch.
            changeValidators(ist.removeList,ist.addedPubKey);
            emit log("changeValidators pass",true);
        }else{
            //ret = verifyAggregatedSeal(allkey[nowEpoch],ist.aggregatedSeals.Signature,blockHash);
        }
        // if (ret == false) {
        //     revert("verifyBlsSign fail");
        //     //return false;
        // }

        //the parent seal need to pks of last epoch to verify parent seal,if block number is the first block or the second block at new epoch.
        //because, the parent seal of the first block and the second block is signed by validitors of last epoch.
        //and it need to not verify, when the block number is less than 2, the block is no parent seal. 
        // if (blockNumber > 1) {
        //     if ((blockNumber-1)%epochLength == 0 || (blockNumber)%epochLength == 0){
        //         ret = verifyAggregatedSeal(allkey[nowEpoch-1],ist.parentAggregatedSeals.Signature,bh.parentHash);
        //     }else{
        //         ret = verifyAggregatedSeal(allkey[nowEpoch],ist.parentAggregatedSeals.Signature,bh.parentHash);
        //     }
        //     if (ret == false) {
        //         revert("verifyBlsSign fail");
        //         //return false;
        //     }
        // }
        
        nowNumber = nowNumber + 1;
        //if(nowNumber+1 != bh.number){
        //    revert("number error");
        //    //return false;
        //}
        allHeader[nowNumber] = rlpHeader;
        emit log("verifyHeader pass",true);
        return ret;
    } 

    function decodeExtraData(bytes memory rlpBytes) public pure returns(istanbulExtra memory ist){
        RLPReader.RLPItem[] memory ls = rlpBytes.toRlpItem().toList();
        RLPReader.RLPItem memory item1 = ls[1];
        RLPReader.RLPItem memory item2 = ls[2];
        RLPReader.RLPItem memory item3 = ls[3];
        RLPReader.RLPItem memory item4 = ls[4];
        RLPReader.RLPItem memory item5 = ls[5];
        
        //Usually, the length of BLS pk is 98 bytes.
        //According to its length, it can calculate the number of pk.
        if (item1.len > 98){
            uint num = (item1.len - 2)/98;
            ist.addedPubKey = new bytes[](num);
            for(uint i=0;i<num;i++){
                ist.addedPubKey[i] = item1.toList()[i].toBytes();
            }
        }

        ist.removeList = item2.toUint();
        ist.seal = item3.toBytes();
        ist.aggregatedSeals.signature = item4.toList()[1].toBytes();
        ist.aggregatedSeals.bitmap = item4.toList()[0].toUint();
        ist.parentAggregatedSeals.signature = item5.toList()[1].toBytes();
        ist.parentAggregatedSeals.bitmap = item5.toList()[0].toUint();
        
        return  ist;                    
    }

    //the function will select legal validators from old validator and add new validators.   
    function changeValidators(uint256 removedVal,bytes[] memory addVal) public view returns(bytes[] memory ret){
        (uint[] memory list,uint8 oldVal) = readRemoveList(removedVal);
        ret = new bytes[](oldVal+addVal.length); 
        uint j=0;
        //if value is 1, the related address will be not validaor at nest epoch. 
        for(uint i=0;i<list.length;i++){
            if (list[i] == 0){
                ret[j] = allkey[nowEpoch][i];
                j = j + 1;
            }
        }
        for(uint i=0;i<addVal.length;i++){
            ret[j] = addVal[i];
            j = j + 1;
        }
        //require(j<101,"the number of validators is more than 100")
        return ret;
    }
    

    //it return binary data and the number of validator in the list. 
    function readRemoveList(uint256 r) public view returns(uint[] memory ret,uint8 sum){
        //the function transfer uint to binary.  
        sum = 0;
        ret = new uint[](keyNum);
        for(uint i=0;r>0;i++){
            if (r%2 == 1){
                r = (r-1)/2;
                ret[i] = 1;
            }else{
                r = r/2;
                ret[i] = 0;
                sum = sum + 1;
            }
        }
        //the current array is inverted.it needs to count down.
        for(uint i=0;i<ret.length/2;i++) {
            uint temp = ret[i];
            ret[i] = ret[ret.length-1-i];
            ret[ret.length-1-i] = temp;
        }
        return (ret,sum);
    }

    // function verifyAggregatedSeal(bytes memory aggregatedSeal,bytes memory seal) private pure returns (bool){
    // }


    function verifySign(bytes memory seal,bytes32 hash,address coinbase) public pure returns (bool){
        //Signature storaged in extraData sub 27 after proposer signed.
        //So signature need to add 27 when verify it.
        (bytes32 r, bytes32 s, uint8 v) = splitSignature(seal);
        v=v+27;
        return coinbase == ecrecover(hash, v, r, s);
    }

    function splitSignature(bytes memory sig) internal pure returns (bytes32 r,bytes32 s,uint8 v){
        require(sig.length == 65, "invalid signature length");
        assembly {
            r := mload(add(sig, 32))
            s := mload(add(sig, 64))
            v := byte(0, mload(add(sig, 96)))
        }
    }

    function splitExtra(bytes memory extra) internal pure returns (bytes memory newExtra){
       //extraData rlpcode is storaged from No.32 byte to latest byte.
       //So, the extraData need to reduce 32 bytes at the beginning.
       newExtra = new bytes(extra.length - 32);
       uint n = 0;
       for(uint i=32;i<extra.length;i++){
           newExtra[n] = extra[i];
           n = n + 1; 
       }
       return newExtra;
    }

    function cutAgg(bytes memory hb,bytes memory agg) public pure returns(bytes memory data){
        require(hb.length < 65535,"the lenght of header rlpcode is too long.");
        require(hb.length > agg.length,"params error.");

        uint datalen = agg.length-2;
        uint index = 0;
        uint target;
        //the escaSeal rlpcode will be replace of nil,the length of seal will become 1 byte. the aggregatedSeal rlpcode as same as the escaSeal.
        //when the length of header rlpcode more than 257,it will add 1 byte.
        uint len = hb.length - datalen + 1; //+3
        if(len>257){
            data = new bytes(len);
            data[index] = bytes1(uint8(249));
            index++;
            data[index] = bytes2(uint16(len-3))[0];
            index++;
            data[index] = bytes2(uint16(len-3))[1];
            index++;
            //emit log("pre 3 byte",len);
        }else{
            data = new bytes(len-1);
            data[index] = bytes1(uint8(248));
            index++;
            data[index] = bytes1(uint8(len-2));
            index++;
            //emit log("pre 2 byte",len);
        }
        //emit log("len",len);
        
        //it use Brute-Force arithmetic for looking for target.
        for (uint i=0;i<hb.length;i++) {
		    for (uint j=0;j<datalen;j++) {
			    if(i+j == hb.length) {
                    //emit log("fail",j);
				    return hb;
			    }
			    if(hb[i+j] != agg[j+2]) {
                    //emit log("not match",i);
				    break;
			    }
			    if(j == datalen-1) {
				    target = i;
			    }
		    }
	    }
        //emit log("target",target);

        uint k;
        if (hb.length>257){
            k=3;
        }else{
            k=2;
        }
        
        for(;k<hb.length;k++) {
            if (k == target){
                data[k] = bytes1(uint8(128));
                index++;
            }
            if (k<target||k>target+datalen){
                data[index] = hb[k];
                index++;
            }
        }
        //emit log("index",index);

        return data;
    }   

    function encodeAgg(bytes memory signature,uint round,uint bitmap) public pure returns (bytes memory output){
        bytes memory output1 = RLPEncode.encodeUint(round);//round
        bytes memory output2 = RLPEncode.encodeBytes(signature);//signature
        bytes memory output3 = RLPEncode.encodeUint(bitmap);//bitmap

        uint index = 0;
        //

        output = new bytes(output1.length+output2.length+output3.length);
        for(uint i=0;i<output1.length;i++){
            output[index] = output1[i];
            index++;
        }
        for(uint i=0;i<output2.length;i++){
            output[index] = output2[i];
            index++;
        }
        for(uint i=0;i<output3.length;i++){
            output[index] = output3[i];
            index++;
        }
    }

}