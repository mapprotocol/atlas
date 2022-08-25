package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"time"

	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

const DefaultGasLimit = 4500000

func sendContractTransaction(client *ethclient.Client, from, toAddress common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, input []byte, gasLimitSetting uint64) (common.Hash, error) {
	// Ensure a valid value field and resolve the account nonce
	logger := log.New("func", "sendContractTransaction")
	nonce, err := client.PendingNonceAt(context.Background(), from)
	if err != nil {
		logger.Error("PendingNonceAt", "error", err)
		return common.Hash{}, err
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	//gasPrice = big.NewInt(1000 000 000 000)
	if err != nil {
		logger.Error("SuggestGasPrice", "error", err)
		return common.Hash{}, err
	}
	gasLimit := uint64(DefaultGasLimit) // in units

	//If the contract surely has code (or code is not needed), estimate the transaction

	msg := ethchain.CallMsg{From: from, To: &toAddress, GasPrice: gasPrice, Value: value, Data: input}
	gasLimit, err = client.EstimateGas(context.Background(), msg)
	if err != nil {
		logger.Error("Contract exec failed", "error", err)
		return common.Hash{}, err
	}
	if gasLimit < 1 {
		//gasLimit = 866328
		gasLimit = 2100000
	}
	gasLimit = uint64(DefaultGasLimit)

	if gasLimitSetting != 0 {
		gasLimit = gasLimitSetting // in units
	}

	// Create the transaction, sign it and schedule it for execution
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, input)

	chainID, _ := client.ChainID(context.Background())
	logger.Info("Tx Info", "from", from, "nonce ", nonce, " gasLimit ", gasLimit, " gasPrice ", gasPrice, " chainID ", chainID)
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		logger.Error("SignTx", "error", err)
		return common.Hash{}, err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		logger.Error("SendTransaction", "error", err)
		return common.Hash{}, err
	}
	return signedTx.Hash(), nil
}

func getResult(conn *ethclient.Client, txHash common.Hash, contract bool) {
	logger := log.New("func", "getResult")
	logger.Info("Please waiting ", " txHash ", txHash.String())
	for {
		time.Sleep(time.Millisecond * 500)
		_, isPending, err := conn.TransactionByHash(context.Background(), txHash)
		if err != nil {
			logger.Error("TransactionByHash", "error", err)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if !isPending {
			break
		}
		logger.Info("Please waiting, Transaction is in pending status")
	}

	var (
		err     error
		receipt *types.Receipt
	)
	for {
		time.Sleep(time.Millisecond * 500)
		receipt, err = conn.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			break
		}
		logger.Error("TransactionReceipt", "error", err)
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		logger.Info("Transaction Success", "number", receipt.BlockNumber.Uint64())
	} else if receipt.Status == types.ReceiptStatusFailed {
		isContinueError = false
		logger.Info("Transaction Failed ", "number", receipt.BlockNumber.Uint64())
	}
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
		isContinueError = false
		logger.Info("Transaction Failed ", "Block Number", receipt.BlockNumber.Uint64())
	}
}

func (w writer) handleUnpackMethodSolveType3(m Message) {
	msg := ethchain.CallMsg{From: m.from, To: &m.to, Data: m.input, GasFeeCap: big.NewInt(3000000000000)}
	output, err := w.conn.CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Error("method CallContract error", "error", err)
		isContinueError = false
	}
	err = m.abi.UnpackIntoInterface(&m.ret, m.abiMethod, output)
	if err != nil {
		log.Error("handleUnpackMethodSolveType3", "err", err)
		isContinueError = false
	}
}

func (w writer) handleUnpackMethodSolveType4(m Message) {
	msg := ethchain.CallMsg{From: m.from, To: &m.to, Data: m.input, GasFeeCap: big.NewInt(3000000000000)}
	output, err := w.conn.CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Error("method CallContract error", "error", err)
		isContinueError = false
	}
	m.solveResult(output)
}
