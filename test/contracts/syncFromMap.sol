//pragma solidity >=0.4.0 < 0.6.0;

import "./Helper.sol"

/**
 * @title RLPEncode
 * @dev A simple RLP encoding library.
 * @author Bakaoh
 */
contract sync {

    struct Header {
        bytes32 parentHash;
        bytes32 extra;
        bytes32 stateRoot;
        bytes32 transactionsRoot;
        bytes32 receiptsRoot;
        uint difficulty;
        uint number;
        uint gasLimit;
        uint gasUsed;
        uint timestamp;
        uint nonce;
    }
    Header[] headers;

    struct ExtraData {
        bytes32 validators;
        bytes32 blsKey;
        bytes32 removeList;
        bytes32 seal;
        bytes32 aggregatedSeal;
        bytes32 parentAggregatedSeal;
    }
    ExtraData[] extarDatas;

    uint epoch;
    bytes[] memory blsKey;
    mapping(uint256 => blsKey) blsKeyStorage
    uint maxLength;
    address _rootAccount;

    constructor(bytes[] memory _epochHeaders,uint _epoch) ERC20(name_,symbol_) {
        _rootAccount = _msgSender();
        epoch = _epoch;
        initEpochData(_epochHeaders);
    }

    function verifyHeader(bytes[] memory _headers){
        Header[] hs;
        ExtraData[] extra;

        hs = new Header[](_headers.length);
        es = new ExtraData[](_headers.length);

        for(uint i=0;i<_headers.length;i++){
            hs[i] = Helper.toBlockHeader(_headers[i]);
            es[i] = Helper.toExtraData(hs[i].extra,hs.[i].number);
            varifyExtra(es);
        }
    }

    function verifyExtra(ExtraData memory e,uint256 number){
        verifySign(e.seal);

        if number%epoch != 0 {
            getBLSPublickKey(number);
        }else{
            getBLSPublickKey(number-2);
        }

        verifyAggregatedSeal(e.aggregatedSeal);

        //verifyParentAggregatedSeal
        if number > 1 {
            if (number-1)%epoch != 0 {
                getBLSPublickKey(number-2);
            }
            verifyAggregatedSeal(e.parentAggregatedSeal);
        }

    }

    function verifySign(){

    }

    function verifyAggregatedSeal(){

    }

    function getBLSPublickKey(queryEpoch uint256) public returns([]bytes){
        return extarDatas[queryEpoch]
    }

    function initEpochData(bytes[] memory _epochHeaders){
        for(uint i=0;i<_headers.length;i++){
            Header storage h = Helper.toBlockHeader(headers[i]);
            ExtraData storage e = Helper.toExtraData(h.extra);
            updateValidatorList(e);
        }
    }

    function updateValidatorList(ExtraData storage extra){

    }

    function reviseBLSPublickKey(reviseEpoch uint256,bytes memoey blsKeys){
        require(_msgSender() == _rootAccount, "reviseBLSPublickKey: onlyRoot");
        extraDatas[reviseEpoch] = blsKeys;
    }

    function setMaxLength(max uint256){
        require(_msgSender() == _rootAccount, "setMaxLength: onlyRoot");
        maxLength = max;
    }
}