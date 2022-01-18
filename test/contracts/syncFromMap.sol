// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.7.1;

import "./RLPReader.sol";

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
    mapping(uint256 => blockHeader) private allHeader;

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
    mapping(uint256 => istanbulExtra) private allExtra;

    mapping(uint256 => bytes[]) private allkey;
    uint256 nowEpoch;
    uint256 nowNumber;
    address rootAccount;
    uint256 epochLength;
    uint maxSyncNum;

    using RLPReader for bytes;
    using RLPReader for uint;
    using RLPReader for RLPReader.RLPItem;
    using RLPReader for RLPReader.Iterator;

    event setParams(string s,uint256 v);
    event setParams(string s,bytes v);

    constructor(uint256 _epochLength) {
        rootAccount = msg.sender;
        epochLength = _epochLength;
        maxSyncNum = 10;
        nowEpoch = 0;
    }

    function setBLSPublickKeys(bytes[] memory keys,uint256 epoch) public {
        require(msg.sender == rootAccount, "onlyRoot");
        emit setParams("current epoch",epoch);
        allkey[epoch] = new bytes[](keys.length);
        for (uint i=0;i<keys.length;i++){
            emit setParams("setBLSPublickKey",keys[i]);
            allkey[epoch][i] = keys[i];
        }
    }

    // function verifyAggregatedSeal(bytes memory aggregatedSeal,bytes memory seal) private {
    // }

    function checkBLSPublickKeys(uint256 epoch) public view returns(bytes[] memory){
        require(msg.sender == rootAccount, "onlyRoot");
        return allkey[epoch];
    }

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

    function setMaxSyncNum(uint max) public{
        require(msg.sender == rootAccount, "onlyRoot");
        emit setParams("setMaxSyncNum",max);
        maxSyncNum = max;
    }

    function checkNowParams() public view returns(uint,uint,uint){
        //require(msg.sender == rootAccount, "onlyRoot");
        return (maxSyncNum,nowEpoch,nowNumber);
    }

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

    function verifyHeader(bytes memory rlpBytes,bytes memory HeaderBytes/*,bytes32 blockHash*/) public returns(bool){
        bool ret = true;
        blockHeader memory bh = decodeHeaderPart1(rlpBytes);
        //bh = decodeHeaderPart2(rlpBytes,bh);
        bytes memory extra = splitExtra(bh.extraData);
        istanbulExtra memory ist = decodeExtraData(extra);
        bytes32 HeaderSignHash = keccak256(abi.encodePacked(HeaderBytes));
        ret = verifySign(ist.seal,HeaderSignHash,bh.coinbase);
        if (ret == false) {
            return false;
        }

        if (bh.number == 0){
            require(ist.addedPubKey.length != 0);
            changeValidators(ist.removeList,ist.addedPubKey);
            return true;
        }

        if (bh.number%epochLength == 0){
            //ret = verifyAggregatedSeal(allkey[nowEpoch],ist.aggregatedSeals.Signature,blockHash);
            changeValidators(ist.removeList,ist.addedPubKey);
        }else{
            //ret = verifyAggregatedSeal(allkey[nowEpoch],ist.aggregatedSeals.Signature,blockHash);
        }
        // if (ret == false) {
        //     return false;
        // }

        //verify parentSeal
        // if (blockNumber > 1) {
        //     if ((blockNumber-1)%epochLength == 0){
        //         ret = verifyAggregatedSeal(allkey[nowEpoch-1],ist.parentAggregatedSeals.Signature,ParentBlockHash);
        //     }
        //     if (ret == false) {
        //         return false;
        //     }
        // }


        allExtra[nowNumber] = ist;
        //allHeader[nowNumber] = bh;
        return ret;
    }

    function decodeExtraData(bytes memory rlpBytes) public pure returns(istanbulExtra memory ist){
        RLPReader.RLPItem[] memory ls = rlpBytes.toRlpItem().toList();
        RLPReader.RLPItem memory item1 = ls[1];
        RLPReader.RLPItem memory item2 = ls[2];
        RLPReader.RLPItem memory item3 = ls[3];
        RLPReader.RLPItem memory item4 = ls[4];
        RLPReader.RLPItem memory item5 = ls[5];

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

    function changeValidators(uint256 removedVal,bytes[] memory addVal) public {
        uint oldValNum = allkey[nowEpoch].length;
        uint j = 0;
        // for(uint i=0;removedVal>0;i++){
        //     if (removedVal%2 == 1){
        //         removedVal = (removedVal-1)/2;
        //     }else{
        //         removedVal = removedVal/2;
        //         allkey[nowEpoch+1][j] = allkey[nowEpoch][oldValNum-i];
        //     }
        //     j = j + 1;
        // }


        // for(uint i=0;i<addVal.length;i++){
        //     allkey[nowEpoch+1][j] = allkey[nowEpoch+1][i];
        // }
        nowEpoch = nowEpoch + 1;
    }

    function verifySign(bytes memory seal,bytes32 hash,address coinbase) public pure returns (bool){
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
       newExtra = new bytes(extra.length - 32);
       uint n = 0;
       for(uint i=32;i<extra.length;i++){
           newExtra[n] = extra[i];
           n = n + 1;
       }
       return newExtra;
    }

}