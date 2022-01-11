pragma solidity ^0.8.7;
//pragma experimental ABIEncoderV2;
import "./RLPReader.sol";

contract sync {

    struct IstanbulAggregatedSeal{
        bytes    Signature;
        bytes   Bitmap;
    }

    struct istanbulExtra{
        bytes  seal;
        IstanbulAggregatedSeal  aggregatedSeals;
        IstanbulAggregatedSeal  parentAggregatedSeals;
        bytes  removeList;
        //bytes[]  addedPubKey;
    }

    struct blsKeys{
        uint256 number;
        bytes blsPublicKeys;
    }
    mapping(uint256 => blsKeys) private allkey;
    uint maxLength;
    address _rootAccount;
    uint epochLength;

    constructor(uint256 _epochLength) {
        _rootAccount = msg.sender;
        epochLength = _epochLength;
    }

    function setEpochKeys(bytes memory publicKeys,uint256 epoch) public {
        require(msg.sender == _rootAccount, "onlyRoot");
        allkey[epoch].number = publicKeys.length;
        allkey[epoch].blsPublicKeys = publicKeys;
    }

    function verifyAggregatedSeal(bytes memory aggregatedSeal,bytes memory seal) private {

    }

    function getBLSPublickKey(uint256 _epoch) public view returns(bytes memory k){
        return allkey[_epoch].blsPublicKeys;
    }

    function setMaxLength(uint256 max) public {
        require(msg.sender == _rootAccount, "onlyRoot");
        require(max > epochLength, "MaxLength more than epochLength.");
        maxLength = max;
    }


    ///////////////////////////////////////////////////////////
    using RLPReader for bytes;
    using RLPReader for uint;
    using RLPReader for RLPReader.RLPItem;
    using RLPReader for RLPReader.Iterator;

    function verifyExtraData(uint256 blockNumber,address coinbase,bytes32 blockHash,bytes memory extraData) public returns(bool){
        bool ret;
        istanbulExtra memory ist = decodeExtraData(extraData);
        ret = verifySign(ist.seal,blockHash,coinbase);
        //if (blockNumber == epochEnd){
            //changeValidators();
        //}
        //verifyAggregatedSeal():
        return ret;
    }


    function decodeExtraData(bytes memory rlpBytes) public returns(istanbulExtra memory ist){
        RLPReader.RLPItem[] memory ls = rlpBytes.toRlpItem().toList();
        RLPReader.RLPItem memory item1 = ls[1];
        RLPReader.RLPItem memory item2 = ls[2];
        RLPReader.RLPItem memory item3 = ls[3];
        RLPReader.RLPItem memory item4 = ls[4];
        RLPReader.RLPItem memory item5 = ls[5];

        //uint num = (item1.len - 2)/98;
        // ist.addedPubKey = new bytes[](num);
        // for(uint i=0;i<num;i++){
        //     ist.addedPubKey[i] = item1.toList()[i].toBytes();
        // }

        ist.removeList = item2.toBytes();
        ist.seal = item3.toBytes();
        ist.aggregatedSeals.Signature = item4.toList()[1].toBytes();
        ist.aggregatedSeals.Bitmap = item4.toList()[2].toBytes();
        ist.parentAggregatedSeals.Signature = item5.toList()[1].toBytes();
        ist.parentAggregatedSeals.Bitmap = item5.toList()[2].toBytes();

        return  ist;
    }

    function changeValidators(uint removeList,bytes memory addedPubKey,uint256 blockNumber) public {

    }

    function verifySign(bytes memory seal,bytes32 hash,address coinbase) public returns (bool){
        (bytes32 r, bytes32 s, uint8 v) = splitSignature(seal);
        v=v+27;
        return coinbase == ecrecover(hash, v, r, s);
    }

    function splitSignature(bytes memory sig) public pure returns (bytes32 r,bytes32 s,uint8 v){
        require(sig.length == 65, "invalid signature length");
        assembly {
            r := mload(add(sig, 32))
            s := mload(add(sig, 64))
            v := byte(0, mload(add(sig, 96)))
        }
    }
}