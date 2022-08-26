package handler

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

const DefaultGasLimit = 4500000

var zeroAddr = common.Address{}

func dial(endpoint string) *ethclient.Client {
	cli, err := ethclient.Dial(endpoint)
	if err != nil {
		log.Crit("dail failed", err.Error())
	}
	return cli
}

func parseABI(abiStr string) *abi.ABI {
	parsed, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		panic(err)
	}
	return &parsed
}

func packInput(abi *abi.ABI, abiMethod string, params ...interface{}) []byte {
	input, err := abi.Pack(abiMethod, params...)
	if err != nil {
		log.Error(abiMethod, " error", err)
	}
	return input

}

func getResult(conn *ethclient.Client, txHash common.Hash) {
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

	queryTx(conn, txHash, false)
}

func queryTx(conn *ethclient.Client, txHash common.Hash, pending bool) {
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
		logger.Info("Transaction Success", "block Number", receipt.BlockNumber.Uint64())
	} else if receipt.Status == types.ReceiptStatusFailed {
		logger.Error("Transaction Failed ", "Block Number", receipt.BlockNumber.Uint64())
	}
}

func CallContract(client *ethclient.Client, to common.Address, input []byte) []byte {
	msg := ethereum.CallMsg{
		From: zeroAddr,
		To:   &to,
		Data: input,
	}
	output, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Crit("method CallContract error", "err", err)
	}
	return output
}

func sendContractTransaction(client *ethclient.Client, from, toAddress common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, input []byte, gasLimitSetting uint64) common.Hash {
	// Ensure a valid value field and resolve the account nonce
	logger := log.New("func", "sendContractTransaction")
	nonce, err := client.PendingNonceAt(context.Background(), from)
	if err != nil {
		logger.Error("PendingNonceAt", "error", err)
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	//gasPrice = big.NewInt(1000 000 000 000)
	if err != nil {
		log.Error("SuggestGasPrice", "error", err)
	}
	gasLimit := uint64(DefaultGasLimit) // in units

	//If the contract surely has code (or code is not needed), estimate the transaction

	msg := ethereum.CallMsg{From: from, To: &toAddress, GasPrice: gasPrice, Value: value, Data: input}
	gasLimit, err = client.EstimateGas(context.Background(), msg)
	if err != nil {
		logger.Error("Contract exec failed", "error", err)
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
	logger.Info("TxInfo", "TX data nonce ", nonce, " gasLimit ", gasLimit, " gasPrice ", gasPrice, " chainID ", chainID)
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		log.Crit("SignTx", "error", err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Crit("SendTransaction", "error", err)
	}
	return signedTx.Hash()
}
