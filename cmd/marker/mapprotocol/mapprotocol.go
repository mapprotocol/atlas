package mapprotocol

import (
	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/abi"
	"math/big"
)

type EventSig string

const (
	//event ValidatorEpochPaymentDistributed(address indexed validator, uint256 validatorPayment);
	ValidatorEpochPaymentDistributed EventSig = "ValidatorEpochPaymentDistributed(address,uint256)"
)

type ProposalStatus int

func (es EventSig) GetTopic() common.Hash {
	return crypto.Keccak256Hash([]byte(es))
}
func PackInput(abi *abi.ABI, abiMethod string, params ...interface{}) []byte {
	input, err := abi.Pack(abiMethod, params...)
	if err != nil {
		log.Error(abiMethod, " error", err)
	}
	return input
}

// buildQuery constructs a query for the bridgeContract by hashing sig to get the event topic
func BuildQuery(contract common.Address, sig EventSig, startBlock *big.Int, endBlock *big.Int) eth.FilterQuery {
	query := eth.FilterQuery{
		FromBlock: startBlock,
		ToBlock:   endBlock,
		Addresses: []common.Address{contract},
		Topics: [][]common.Hash{
			{sig.GetTopic()},
		},
	}
	return query
}

//func ParseEthLogIntoSwapWithProofArgs(log types.Log, bridgeAddr common.Address, receipts []*types.Receipt) (uint64, uint64, []byte, error) {
//	token := log.Topics[1].Bytes()
//	from := log.Topics[2].Bytes()
//	to := log.Topics[3].Bytes()
//	// every 32 bytes forms a value
//	var orderHash [32]byte
//	copy(orderHash[:], log.Data[:32])
//	amount := log.Data[32:64]
//
//	fromChainID := log.Data[64:96]
//	toChainID := log.Data[96:128]
//	uFromChainID := binary.BigEndian.Uint64(fromChainID[len(fromChainID)-8:])
//	uToChainID := binary.BigEndian.Uint64(toChainID[len(toChainID)-8:])
//
//
//	return uFromChainID, uToChainID, payloads, nil
//}
