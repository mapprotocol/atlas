// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.7.1;

import "./BlsSignatureVerification.sol";


contract BlsSignatureTest is BlsSignatureVerification {
    bool public verified;

    function verifySignature(
        bytes calldata _publicKey,  // an E2 point
        bytes calldata _message,
        bytes calldata _signature   // an E1 point
    ) external {
        E2Point memory pub = decodeE2Point(_publicKey);
        E1Point memory sig = decodeE1Point(_signature);
        verified = verify(pub, _message, sig);
    }

    function verifySignaturePoint(
        bytes calldata _publicKey,  // an E2 point
        bytes calldata _message,    // an E1 point
        bytes calldata _signature   // an E1 point
    ) external {
        E2Point memory pub = decodeE2Point(_publicKey);
        E1Point memory sig = decodeE1Point(_signature);
        verified = verifyForPoint(pub, decodeE1Point(_message), sig);
    }

    function verifyMultisignature(
        bytes calldata _aggregatedPublicKey,  // an E2 point
        bytes calldata _partPublicKey,        // an E2 point
        bytes calldata _message,
        bytes calldata _partSignature,        // an E1 point
        uint _signersBitmask
    ) external {
        E2Point memory aPub = decodeE2Point(_aggregatedPublicKey);
        E2Point memory pPub = decodeE2Point(_partPublicKey);
        E1Point memory pSig = decodeE1Point(_partSignature);
        verified = verifyMultisig(aPub, pPub, _message, pSig, _signersBitmask);
    }

    function verifyAggregatedHash(
        bytes calldata _p,
        uint index
    ) external view returns (bytes memory) {
        E2Point memory pub = decodeE2Point(_p);
        bytes memory message = abi.encodePacked(pub.x, pub.y, index);
        E1Point memory h = hashToCurveE1(message);
        return abi.encodePacked(h.x, h.y);
    }

    function addOnCurveE1(
        bytes calldata _p1,
        bytes calldata _p2
    ) external view returns (bytes memory) {
        E1Point memory res = addCurveE1(decodeE1Point(_p1), decodeE1Point(_p2));
        return abi.encode(res.x, res.y);
    }

    function decodeE2Point(bytes memory _pubKey) private pure returns (E2Point memory pubKey) {
        uint256[] memory output = new uint256[](4);
        for (uint256 i = 32; i <= output.length * 32; i += 32) {
            assembly { mstore(add(output, i), mload(add(_pubKey, i))) }
        }

        pubKey.x[0] = output[0];
        pubKey.x[1] = output[1];
        pubKey.y[0] = output[2];
        pubKey.y[1] = output[3];
    }

    function decodeE1Point(bytes memory _sig) private pure returns (E1Point memory signature) {
        uint256[] memory output = new uint256[](2);
        for (uint256 i = 32; i <= output.length * 32; i += 32) {
            assembly { mstore(add(output, i), mload(add(_sig, i))) }
        }

        signature.x = output[0];
        signature.y = output[1];
    }

    function aggregatedPublicKey(
        bytes calldata _publicKey1,  // an E2 point
        bytes calldata _publicKey2  // an E2 point
    ) external view{
        E1Point memory pub1 = decodeE1Point(_publicKey1);
        E1Point memory pub2 = decodeE1Point(_publicKey2);
        addCurveE1(pub1, pub2);
    }

}