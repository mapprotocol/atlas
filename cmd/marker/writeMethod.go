package main

import (
	"context"
	"crypto/ecdsa"
	ethchain "github.com/ethereum/go-ethereum"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

func sendContractTransaction(client *ethclient.Client, from, toAddress common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, input []byte, gasLimitSeting uint64) common.Hash {
	// Ensure a valid value field and resolve the account nonce
	logger := log.New("func", "sendContractTransaction")
	nonce, err := client.PendingNonceAt(context.Background(), from)
	if err != nil {
		logger.Error("PendingNonceAt", "error", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Error("SuggestGasPrice", "error", err)
	}
	gasLimit := uint64(3100000) // in units

	//If the contract surely has code (or code is not needed), estimate the transaction

	msg := ethchain.CallMsg{From: from, To: &toAddress, GasPrice: gasPrice, Value: value, Data: input}
	gasLimit, err = client.EstimateGas(context.Background(), msg)
	if err != nil {
		logger.Error("Contract exec failed", "error", err)
	}
	if gasLimit < 1 {
		//gasLimit = 866328
		gasLimit = 2100000
	}
	if gasLimitSeting != 0 {
		gasLimit = gasLimitSeting // in units
	}

	// Create the transaction, sign it and schedule it for execution
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, input)

	chainID, _ := client.ChainID(context.Background())
	logger.Info("TxInfo", "TX data nonce ", nonce, " gasLimit ", gasLimit, " gasPrice ", gasPrice, " chainID ", chainID)
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		log.Error("SignTx", "error", err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Error("SendTransaction", "error", err)
	}

	return signedTx.Hash()
}

func getResult(conn *ethclient.Client, txHash common.Hash, contract bool) {
	logger := log.New("func", "getResult")
	logger.Info("Please waiting ", " txHash ", txHash.String())
	for {
		time.Sleep(time.Millisecond * 200)
		_, isPending, err := conn.TransactionByHash(context.Background(), txHash)
		if err != nil {
			logger.Info("TransactionByHash", "error", err)
		}
		if !isPending {
			break
		}
	}

	queryTx(conn, txHash, contract, false)
}

func queryTx(conn *ethclient.Client, txHash common.Hash, contract bool, pending bool) {
	logger := log.New("func", "queryTx")
	if pending {
		_, isPending, err := conn.TransactionByHash(context.Background(), txHash)
		if err != nil {
			logger.Error("TransactionByHash", "error", err)
		}
		if isPending {
			println("In tx_pool no validator  process this, please query later")
			os.Exit(0)
		}
	}

	receipt, err := conn.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		for {
			time.Sleep(time.Millisecond * 200)
			receipt, err = conn.TransactionReceipt(context.Background(), txHash)
			if err == nil {
				break
			}
		}
		logger.Error("TransactionReceipt", "error", err)
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		//block, err := conn.BlockByHash(context.Background(), receipt.BlockHash)
		//if err != nil {
		//	logger.Error("BlockByHash", err)
		//}
		//logger.Info("Transaction Success", " block Number", receipt.BlockNumber.Uint64(), " block txs", len(block.Transactions()), "blockhash", block.Hash().Hex())
		logger.Info("Transaction Success", "block Number", receipt.BlockNumber.Uint64())
	} else if receipt.Status == types.ReceiptStatusFailed {
		logger.Info("Transaction Failed ", "Block Number", receipt.BlockNumber.Uint64())
	}
}

func (w writer) handleUnpackMethodSolveType3(m Message) {
	header, err := w.conn.HeaderByNumber(context.Background(), nil)
	msg := ethchain.CallMsg{From: m.from, To: &m.to, Data: m.input}
	output, err := w.conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "error", err)
	}
	err = m.abi.UnpackIntoInterface(&m.ret, m.abiMethod, output)
	if err != nil {
		log.Error("handleUnpackMethodSolveType3", "err", err)
	}
}

func (w writer) handleUnpackMethodSolveType4(m Message) {
	header, err := w.conn.HeaderByNumber(context.Background(), nil)
	msg := ethchain.CallMsg{From: m.from, To: &m.to, Data: m.input}
	output, err := w.conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "error", err)
	}
	m.solveResult(output)
}
