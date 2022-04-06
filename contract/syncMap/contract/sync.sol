// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.7.1;

import "../lib/RLPReader.sol";
import "../lib/RLPEncode.sol";
//import "../bn256/BlsSignatureTest.sol";

contract sync {

    struct blockHeader{
        bytes parentHash;
        address coinbase;
        bytes root;
        bytes txHash;
        bytes receipHash;
        bytes bloom;
        uint256 number;
        uint256 gasLimit;
        uint256 gasUsed;
        uint256 time;
        bytes extraData;
        bytes mixDigest;
        bytes nonce;
        uint256 baseFee;
    }

    struct istanbulAggregatedSeal{
        uint256   round;
        bytes     signature;
        uint256   bitmap;
    }

    struct istanbulExtra{
        address[] validators;
        bytes  seal;
        istanbulAggregatedSeal  aggregatedSeal;
        istanbulAggregatedSeal  parentAggregatedSeal;
        uint256  removeList;
        bytes[]  addedPubKey;
    }

    mapping(uint256 => bytes[]) private allkey;
    uint256 nowEpoch;
    uint256 nowNumber;
    address rootAccount;
    uint256 epochLength;
    uint256 keyNum;

    using RLPReader for bytes;
    using RLPReader for uint;
    using RLPReader for RLPReader.RLPItem;
    using RLPReader for RLPReader.Iterator;

    event log(string s,bytes e);

     constructor(bytes memory firstBlock) {
         rootAccount = msg.sender;
         epochLength = 20;
         nowEpoch = 0;
         initFirstBlock(firstBlock);
     }

    //the first block is used for init params,
    //it was operated specially.
    function initFirstBlock(bytes memory firstBlock) private{
        decodeHeader(firstBlock);
        istanbulExtra memory ist = decodeExtraData(bh.extraData);

        keyNum = ist.addedPubKey.length;
        nowNumber = bh.number;
        require(nowNumber == 0);
        allkey[nowEpoch] = new bytes[](keyNum);
        for(uint8 i = 0;i<keyNum;i++){
            allkey[nowEpoch][i] = ist.addedPubKey[i];
        }
    }

    blockHeader bh;
    function decodeHeader(bytes memory rlpBytes)private{
        decodeHeaderPart1(rlpBytes);
        decodeHeaderPart2(rlpBytes);
    }

    function decodeHeaderPart1(bytes memory rlpBytes)private{
        RLPReader.RLPItem[] memory ls = rlpBytes.toRlpItem().toList();
        RLPReader.RLPItem memory item0 = ls[0]; //parentBlockHash
        RLPReader.RLPItem memory item1 = ls[1]; //coinbase
        RLPReader.RLPItem memory item2 = ls[2]; //root
        RLPReader.RLPItem memory item3 = ls[3]; //txHash
        RLPReader.RLPItem memory item4 = ls[4]; //receipHash
        RLPReader.RLPItem memory item6 = ls[6]; //number
        RLPReader.RLPItem memory item10 = ls[10]; //extra

        bh.parentHash = item0.toBytes();
        bh.coinbase = item1.toAddress();
        bh.root = item2.toBytes();
        bh.txHash = item3.toBytes();
        bh.receipHash = item4.toBytes();
        bh.number = item6.toUint();
        bh.extraData = item10.toBytes();
    }

    function decodeHeaderPart2(bytes memory rlpBytes)private{
        RLPReader.RLPItem[] memory ls = rlpBytes.toRlpItem().toList();
        RLPReader.RLPItem memory item5 = ls[5]; //bloom
        RLPReader.RLPItem memory item7 = ls[7]; //gasLimit
        RLPReader.RLPItem memory item8 = ls[8]; //gasUsed
        RLPReader.RLPItem memory item9 = ls[9]; //time
        RLPReader.RLPItem memory item11 = ls[11]; //mixDigest
        RLPReader.RLPItem memory item12 = ls[12]; //nonce
        RLPReader.RLPItem memory item13 = ls[13]; //baseFee

        bh.bloom = item5.toBytes();
        bh.gasLimit = item7.toUint();
        bh.gasUsed = item8.toUint();
        bh.time = item9.toUint();
        bh.mixDigest  = item11.toBytes();
        bh.nonce  = item12.toBytes();
        bh.baseFee = item13.toUint();
    }

    //the input data is the header after rlp encode within seal by proposer and aggregated seal by validators.
    function verifyHeader(bytes memory rlpHeader) public returns(bool){
        bool ret = true;
        decodeHeader(rlpHeader);
        istanbulExtra memory ist = decodeExtraData(bh.extraData);
        deleteAgg(ist);
        bytes memory headerWithoutAgg = encodeHeader();
        bytes32 hash1 = keccak256(abi.encodePacked(headerWithoutAgg));
        deleteSealAndAgg(ist);
        bytes memory headerWithoutSealAndAgg = encodeHeader();
        bytes32 hash2 = keccak256(abi.encodePacked(headerWithoutSealAndAgg));

        //the ecdsa seal signed by proposer
        ret = verifySign(ist.seal,keccak256(abi.encodePacked(hash2)),bh.coinbase);
        if (ret == false) {
            revert("verifyEscaSign fail");
        }


        //the blockHash is the hash of the header without aggregated seal by validators.
        bytes memory blsMsg1 = addsuffix(hash1,uint8(ist.aggregatedSeal.round));
        if (bh.number%epochLength == 0){
            //ret = verifyAggregatedSeal(allkey[nowEpoch],ist.aggregatedSeal.signature,blsMsg1);
            //it need to update validators at first block of new epoch.
            changeValidators(ist.removeList,ist.addedPubKey);
        }else{
            //ret = verifyAggregatedSeal(allkey[nowEpoch],ist.aggregatedSeal.signature,blsMsg1);
        }
        emit log("verify msg of AggregatedSeal",blsMsg1);

        //the parent seal need to pks of last epoch to verify parent seal,if block number is the first block or the second block at new epoch.
        //because, the parent seal of the first block and the second block is signed by validitors of last epoch.
        //and it need to not verify, when the block number is less than 2, the block is no parent seal.
        bytes memory blsMsg2 = addsuffix(hash1,uint8(ist.aggregatedSeal.round));
        if (bh.number > 1) {
            if ((bh.number-1)%epochLength == 0 || (bh.number)%epochLength == 0){
                //ret = verifyAggregatedSeal(allkey[nowEpoch-1],ist.parentAggregatedSeal.signature,blsMsg2);
            }else{
                //ret = verifyAggregatedSeal(allkey[nowEpoch],ist.parentAggregatedSeal.signature,blsMsg2);
            }
        }
        emit log("verify msg of ParentAggregatedSeal",blsMsg2);
        return ret;
    }

    //suffix's rule is hash + round + commitMsg(the value is 2 usually);
    function addsuffix(bytes32 hash,uint8 round)private pure returns(bytes memory){
        bytes memory result = new bytes(34);
        for (uint i=0;i<32;i++){
            result[i] = hash[i];
        }
        result[32] = bytes1(round);
        result[33] = bytes1(uint8(2));
        return result ;
    }

    //istanbulExtra ist;
    bytes extraDataPre;
    function decodeExtraData(bytes memory extraData) private returns(istanbulExtra memory ist){
        bytes memory decodeBytes = splitExtra(extraData);
        RLPReader.RLPItem[] memory ls = decodeBytes.toRlpItem().toList();
        RLPReader.RLPItem memory item0 = ls[0];
        RLPReader.RLPItem memory item1 = ls[1];
        RLPReader.RLPItem memory item2 = ls[2];
        RLPReader.RLPItem memory item3 = ls[3];
        RLPReader.RLPItem memory item4 = ls[4];
        RLPReader.RLPItem memory item5 = ls[5];

        if (item0.len > 20){
            uint num = item0.len/20;
            ist.validators = new address[](num);
            ist.addedPubKey = new bytes[](num);
            for(uint i=0;i<num;i++){
                ist.validators[i] = item0.toList()[i].toAddress();
                ist.addedPubKey[i] = item1.toList()[i].toBytes();
            }
        }

        ist.removeList = item2.toUint();
        ist.seal = item3.toBytes();
        ist.aggregatedSeal.round = item4.toList()[2].toUint();
        ist.aggregatedSeal.signature = item4.toList()[1].toBytes();
        ist.aggregatedSeal.bitmap = item4.toList()[0].toUint();
        ist.parentAggregatedSeal.round = item5.toList()[2].toUint();
        ist.parentAggregatedSeal.signature = item5.toList()[1].toBytes();
        ist.parentAggregatedSeal.bitmap = item5.toList()[0].toUint();
        return ist;
    }

    function splitExtra(bytes memory extra) private returns (bytes memory newExtra){
        //extraData rlpcode is storaged from No.32 byte to latest byte.
        //So, the extraData need to reduce 32 bytes at the beginning.
        newExtra = new bytes(extra.length - 32);
        extraDataPre = new bytes(32);
        uint n = 0;
        for(uint i=32;i<extra.length;i++){
            newExtra[n] = extra[i];
            n = n + 1;
        }
        uint m = 0;
        for(uint i=0;i<32;i++){
            extraDataPre[m] = extra[i];
            m = m + 1;
        }
        return newExtra;
    }

    function encodeAggregatedSeal(uint bitmap,bytes memory signature,uint round) private pure returns (bytes memory output){
        bytes memory output1 = RLPEncode.encodeUint(bitmap);//round
        bytes memory output2 = RLPEncode.encodeBytes(signature);//signature
        bytes memory output3 = RLPEncode.encodeUint(round);//bitmap

        bytes[] memory list = new bytes[](3);
        list[0] = output1;
        list[1] = output2;
        list[2] = output3;
        output = RLPEncode.encodeList(list);
    }

    //the function will select legal validators from old validator and add new validators.
    function changeValidators(uint256 removedVal,bytes[] memory addVal) private{
        (uint[] memory list,uint8 oldVal) = readRemoveList(removedVal);
        allkey[nowEpoch+1] = new bytes[](oldVal+addVal.length);
        uint j=0;
        //if value is 1, the related address will be not validaor at nest epoch.
        for(uint i=0;i<list.length;i++){
            if (list[i] == 0){
                allkey[nowEpoch+1][j] = allkey[nowEpoch][i];
                j = j + 1;
            }
        }
        for(uint i=0;i<addVal.length;i++){
            allkey[nowEpoch+1][j] = addVal[i];
            j = j + 1;
        }
        nowEpoch = nowEpoch + 1;
        //require(j<101,"the number of validators is more than 100")
    }

    //it return binary data and the number of validator in the list.
    function readRemoveList(uint256 r) private view returns(uint[] memory ret,uint8 sum){
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

    function verifySign(bytes memory seal,bytes32 hash,address coinbase) private pure returns (bool){
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

    function deleteAgg(istanbulExtra memory ist) private{
        bytes[] memory list1 = new bytes[](ist.validators.length);
        bytes[] memory list2 = new bytes[](ist.addedPubKey.length);
        for(uint i=0;i<ist.validators.length;i++){
            list1[i] = RLPEncode.encodeAddress(ist.validators[i]);//
            list2[i] = RLPEncode.encodeBytes(ist.addedPubKey[i]);//
        }

        bytes[] memory list = new bytes[](6);
        list[0] = RLPEncode.encodeList(list1);//
        list[1] = RLPEncode.encodeList(list2);//
        list[2] = RLPEncode.encodeUint(ist.removeList);//
        list[3] = RLPEncode.encodeBytes(ist.seal);//
        list[4] = new bytes(4);
        list[4][0] = bytes1(uint8(195));
        list[4][1] = bytes1(uint8(128));
        list[4][2] = bytes1(uint8(128));
        list[4][3] = bytes1(uint8(128));
        list[5] = encodeAggregatedSeal(ist.parentAggregatedSeal.bitmap,ist.parentAggregatedSeal.signature,ist.parentAggregatedSeal.round);
        bytes memory b = RLPEncode.encodeList(list);
        bytes memory output = new bytes(b.length+32);
        for (uint i=0;i<b.length+32;i++){
            if (i<32){
                output[i] = extraDataPre[i];
            }else{
                output[i] = b[i-32];
            }
        }
        bh.extraData = output;
    }

    function deleteSealAndAgg(istanbulExtra memory ist) private{
        bytes[] memory list1 = new bytes[](ist.validators.length);
        bytes[] memory list2 = new bytes[](ist.addedPubKey.length);
        for(uint i=0;i<ist.validators.length;i++){
            list1[i] = RLPEncode.encodeAddress(ist.validators[i]);//
            list2[i] = RLPEncode.encodeBytes(ist.addedPubKey[i]);//
        }

        bytes[] memory list = new bytes[](6);
        list[0] = RLPEncode.encodeList(list1);//
        list[1] = RLPEncode.encodeList(list2);//
        list[2] = RLPEncode.encodeUint(ist.removeList);//
        list[3] = new bytes(1);
        list[3][0] = bytes1(uint8(128));//
        list[4] = new bytes(4);
        list[4][0] = bytes1(uint8(195));
        list[4][1] = bytes1(uint8(128));
        list[4][2] = bytes1(uint8(128));
        list[4][3] = bytes1(uint8(128));
        list[5] = encodeAggregatedSeal(ist.parentAggregatedSeal.bitmap,ist.parentAggregatedSeal.signature,ist.parentAggregatedSeal.round);
        bytes memory b = RLPEncode.encodeList(list);
        bytes memory output = new bytes(b.length+32);
        for (uint i=0;i<b.length+32;i++){
            if (i<32){
                output[i] = extraDataPre[i];
            }else{
                output[i] = b[i-32];
            }
        }
        bh.extraData = output;
    }

    function encodeHeader()private view returns (bytes memory output){
        bytes[] memory list = new bytes[](14);
        list[0] = RLPEncode.encodeBytes(bh.parentHash);//
        list[1] = RLPEncode.encodeAddress(bh.coinbase);//
        list[2] = RLPEncode.encodeBytes(bh.root);//
        list[3] = RLPEncode.encodeBytes(bh.txHash);//
        list[4] = RLPEncode.encodeBytes(bh.receipHash);//
        list[5] = RLPEncode.encodeBytes(bh.bloom);//
        list[6] = RLPEncode.encodeUint(bh.number);//
        list[7] = RLPEncode.encodeUint(bh.gasLimit);//;
        list[8] = RLPEncode.encodeUint(bh.gasUsed);//
        list[9] = RLPEncode.encodeUint(bh.time);//
        list[10] = RLPEncode.encodeBytes(bh.extraData);//
        list[11] = RLPEncode.encodeBytes(bh.mixDigest);//
        list[12] = RLPEncode.encodeBytes(bh.nonce);//
        list[13] = RLPEncode.encodeUint(bh.baseFee);//
        output = RLPEncode.encodeList(list);
    }
}