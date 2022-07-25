package ethereum

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/mapprotocol/atlas/core/types"
)

var (
	// EventHash cross-chain transaction event hash
	// mapTransferOut(address indexed token, address indexed from, bytes32 indexed orderId, uint fromChain, uint toChain, bytes to, uint amount, bytes toChainToken);
	// mapTransferOut(address,address,bytes32,uint256,uint256,bytes,uint256,bytes)
	//EventHash = common.HexToHash("0x1d7c4ab437b83807c25950ac63192692227b29e3205a809db6a4c3841836eb02")
)

type TxProve struct {
	Receipt     *ethtypes.Receipt
	Prove       light.NodeList
	BlockNumber uint64
	TxIndex     uint
}

type Verify struct {
}

func (v *Verify) Verify(db types.StateDB, routerContractAddr common.Address, txProveBytes []byte) (logs []byte, err error) {
	txProve, err := v.decode(txProveBytes)
	if err != nil {
		return nil, err
	}

	//lgs, err := v.queryLog(routerContractAddr, txProve.Receipt.Logs)
	//if err != nil {
	//	return nil, err
	//}
	receiptsRoot, err := v.getReceiptsRoot(db, txProve.BlockNumber)
	if err != nil {
		return nil, err
	}

	if err := v.verifyProof(receiptsRoot, txProve); err != nil {
		return nil, err
	}
	return rlp.EncodeToBytes(txProve.Receipt.Logs)
}

func (v *Verify) decode(txProveBytes []byte) (*TxProve, error) {
	var txProve TxProve
	if err := rlp.DecodeBytes(txProveBytes, &txProve); err != nil {
		return nil, err
	}
	return &txProve, nil
}

//func (v *Verify) queryLog(routerContractAddr common.Address, logs []*ethtypes.Log) (*ethtypes.Log, error) {
//	for _, lg := range logs {
//		if bytes.Equal(lg.Address.Bytes(), routerContractAddr.Bytes()) {
//			if bytes.Equal(lg.Topics[0].Bytes(), EventHash.Bytes()) {
//				return lg, nil
//			}
//		}
//	}
//	return nil, fmt.Errorf("not found event log, router contract addr: %v, event hash: %v", routerContractAddr, EventHash)
//}

func (v *Verify) getReceiptsRoot(db types.StateDB, blockNumber uint64) (common.Hash, error) {
	hs := NewHeaderStore()
	if err := hs.Load(db); err != nil {
		return common.Hash{}, err
	}
	header := hs.GetHeaderByNumber(blockNumber)
	if header == nil {
		return common.Hash{}, fmt.Errorf("get header by number failed, number: %d", blockNumber)
	}

	return header.ReceiptHash, nil
}

func (v *Verify) verifyProof(receiptsRoot common.Hash, txProve *TxProve) error {
	var buf bytes.Buffer
	rs := ethtypes.Receipts{txProve.Receipt}
	rs.EncodeIndex(0, &buf)
	giveReceipt := buf.Bytes()

	var key []byte
	key = rlp.AppendUint64(key[:0], uint64(txProve.TxIndex))

	getReceipt, err := trie.VerifyProof(receiptsRoot, key, txProve.Prove.NodeSet())
	if err != nil {
		return err
	}
	if !bytes.Equal(giveReceipt, getReceipt) {
		return errors.New("receipt mismatch")
	}
	return nil
}
