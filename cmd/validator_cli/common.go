package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/validator_cli/env"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"time"
)

var (
	key   string
	store string
	ip    string
	port  int
)

func checkFee(fee *big.Int) {
	if fee.Sign() < 0 || fee.Cmp(Base) > 0 {
		log.Error("Please set correct fee value")
	}
}

func sendContractTransaction(client *ethclient.Client, from, toAddress common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, input []byte) common.Hash {
	// Ensure a valid value field and resolve the account nonce
	logger := log.New("func", "sendContractTransaction")
	nonce, err := client.PendingNonceAt(context.Background(), from)
	if err != nil {
		logger.Error("PendingNonceAt", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Error("SuggestGasPrice", "err", err)
	}
	gasLimit := uint64(3100000) // in units
	//If the contract surely has code (or code is not needed), estimate the transaction
	msg := ethchain.CallMsg{From: from, To: &toAddress, GasPrice: gasPrice, Value: value, Data: input}
	gasLimit, err = client.EstimateGas(context.Background(), msg)
	if err != nil {
		logger.Error("Contract exec failed", "err", err)
	}
	if gasLimit < 1 {
		//gasLimit = 866328
		gasLimit = 2100000
	}

	// Create the transaction, sign it and schedule it for execution
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, input)

	chainID, _ := client.ChainID(context.Background())
	logger.Info("TxInfo", "TX data nonce ", nonce, " gasLimit ", gasLimit, " gasPrice ", gasPrice, " chainID ", chainID)
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		log.Error("SignTx", "err", err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Error("SendTransaction", "err", err)
	}

	return signedTx.Hash()
}

func loadPrivateKey(keyfile string) common.Address {
	keyjson, err := ioutil.ReadFile(keyfile)
	if err != nil {
		log.Error("loadPrivateKey", fmt.Errorf("failed to read the keyfile at '%s': %v", keyfile, err))
	}
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		log.Error("DecryptKey", fmt.Errorf("error decrypting key: %v", err))
	}
	priKey = key.PrivateKey
	from = crypto.PubkeyToAddress(priKey.PublicKey)
	//fmt.Println("address ", from.Hex(), "key", hex.EncodeToString(crypto.FromECDSA(priKey)))
	return from
}
func loadAccount(path string, password string) env.Account {
	logger := log.New("func", "getResult")
	keyjson, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Crit("loadPrivate ReadFile", "err", fmt.Errorf("failed to read the keyfile at '%s': %v", path, err))
	}
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		logger.Crit("loadPrivate DecryptKey", "err", fmt.Errorf("error decrypting key: %v", err))
	}
	priKey1 := key.PrivateKey
	publicAddr := crypto.PubkeyToAddress(priKey1.PublicKey)
	var addr common.Address
	addr.SetBytes(publicAddr.Bytes())

	return env.Account{
		Address:    addr,
		PrivateKey: priKey1,
	}
}
func getAllFile(path string) (string, error) {
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		log.Error("path ", err)
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fmt.Printf("[%s]\n", path+"\\"+fi.Name())
			getAllFile(path + fi.Name() + "\\")
			return "", errors.New("path error")
		} else {
			fmt.Println(path, "dir has ", fi.Name(), "file")
			return fi.Name(), nil
		}
	}
	return "", err
}

func ethToWei(ctx *cli.Context, zero bool) *big.Int {
	Value = ctx.GlobalUint64(ValueFlag.Name)
	if !zero && Value <= 0 {
		log.Error("value must bigger than 0")
	}
	baseUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	value := new(big.Int).Mul(big.NewInt(int64(Value)), baseUnit)
	return value
}

func weiToEth(value *big.Int) uint64 {
	baseUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	valueT := new(big.Int).Div(value, baseUnit).Uint64()
	return valueT
}

func getResult(conn *ethclient.Client, txHash common.Hash, contract bool) {
	logger := log.New("func", "getResult")
	logger.Info("Please waiting ", " txHash ", txHash.String())
	for {
		time.Sleep(time.Millisecond * 200)
		_, isPending, err := conn.TransactionByHash(context.Background(), txHash)
		if err != nil {
			logger.Info("TransactionByHash", "err", err)
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
			logger.Error("TransactionByHash", "err", err)
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
		logger.Error("TransactionReceipt", "err", err)
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

func packInput(abi *abi.ABI, abiMethod string, params ...interface{}) []byte {
	input, err := abi.Pack(abiMethod, params...)
	if err != nil {
		log.Error(abiMethod, " error", err)
	}
	return input
}

func PrintBalance(conn *ethclient.Client, from common.Address) {
	logger := log.New("func", "PrintBalance")
	balance, err := conn.BalanceAt(context.Background(), from, nil)
	if err != nil {
		logger.Error("BalanceAt", "err", err)
	}
	balance2 := new(big.Float)
	balance2.SetString(balance.String())
	Value := new(big.Float).Quo(balance2, big.NewFloat(math.Pow10(18)))

	fmt.Println("Your wallet balance is ", Value, "'eth ")
}

func loadPrivate(ctx *cli.Context) {
	key = ctx.GlobalString(KeyFlag.Name)
	store = ctx.GlobalString(KeyStoreFlag.Name)
	if key != "" {
		loadPrivateKey(key)
	} else if store != "" {
		loadSigningKey(store)
	} else {
		log.Error("Must specify --key or --keystore")
	}

	if priKey == nil {
		log.Error("load privateKey failed")
	}
}

func dialConn(ctx *cli.Context) (*ethclient.Client, string) {
	logger := log.New("func", "dialConn")
	ip = ctx.GlobalString("rpcaddr") //utils.RPCListenAddrFlag.Name)
	port = ctx.GlobalInt("rpcport")  //utils.RPCPortFlag.Name)
	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	conn, err := ethclient.Dial(url)
	if err != nil {
		logger.Error("Failed to connect to the Atlaschain client: %v", err)
	}
	return conn, url
}

func printBaseInfo(conn *ethclient.Client, url string) *types.Header {
	logger := log.New("func", "printBaseInfo")
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		logger.Error("HeaderByNumber", "err", err)
	}

	if common.IsHexAddress(from.Hex()) {
		fmt.Println("Connect url ", url, " current number ", header.Number.String(), " address ", from.Hex())
	} else {
		fmt.Println("Connect url ", url, " current number ", header.Number.String())
	}

	return header
}

// loadSigningKey loads a private key in Ethereum keystore format.
func loadSigningKey(keyfile string) common.Address {
	keyjson, err := ioutil.ReadFile(keyfile)
	if err != nil {
		log.Error("ReadFile", fmt.Errorf("failed to read the keyfile at '%s': %v", keyfile, err))
	}
	password, _ := prompt.Stdin.PromptPassword("Please enter the password for '" + keyfile + "': ")
	//password := "secret"
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		log.Error("DecryptKey", fmt.Errorf("error decrypting key: %v", err))
	}
	priKey = key.PrivateKey
	from = crypto.PubkeyToAddress(priKey.PublicKey)
	//fmt.Println("address ", from.Hex(), "key", hex.EncodeToString(crypto.FromECDSA(priKey)))
	return from
}
