package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"math/big"
	"os"
	"time"

	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mapprotocol/atlas/accounts/abi"
	blscrypto "github.com/mapprotocol/atlas/helper/bls"
)

type Config struct {
	from          common.Address
	PublicKey     []byte
	PrivateKey    *ecdsa.PrivateKey
	BlsPub        blscrypto.SerializedPublicKey
	BLSProof      []byte
	Value         uint64
	Commission    int64
	lesser        common.Address
	greater       common.Address
	voteNum       *big.Int
	TopNum        *big.Int
	Idx           *big.Int
	targetAddress common.Address
	ip            string
	port          int
	conn          *ethclient.Client
}

func sendContractTransaction(client *ethclient.Client, from, toAddress common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, input []byte) common.Hash {
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
func loadAccount(path string, password string) Account {
	logger := log.New("func", "getResult")
	keyjson, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Crit("loadPrivate ReadFile", "error", fmt.Errorf("failed to read the keyfile at '%s': %v", path, err))
	}
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		logger.Crit("loadPrivate DecryptKey", "error", fmt.Errorf("error decrypting key: %v", err))
	}
	priKey1 := key.PrivateKey
	publicAddr := crypto.PubkeyToAddress(priKey1.PublicKey)
	var addr common.Address
	addr.SetBytes(publicAddr.Bytes())

	return Account{
		Address:    addr,
		PrivateKey: priKey1,
	}
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

func packInput(abi *abi.ABI, abiMethod string, params ...interface{}) []byte {
	input, err := abi.Pack(abiMethod, params...)
	if err != nil {
		log.Error(abiMethod, " error", err)
	}
	return input
}

func dialConn(ctx *cli.Context) (*ethclient.Client, string) {
	logger := log.New("func", "dialConn")
	ip := ctx.GlobalString("rpcaddr") //utils.RPCListenAddrFlag.Name)
	port := ctx.GlobalInt("rpcport")  //utils.RPCPortFlag.Name)
	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	conn, err := ethclient.Dial(url)
	if err != nil {
		logger.Error("Failed to connect to the Atlaschain client: %v", err)
	}
	return conn, url
}

func AssemblyConfig(ctx *cli.Context) *Config {
	var config *Config
	//------------------ pre set --------------------------
	path := ""
	password := "111111"
	config.voteNum = big.NewInt(int64(100))
	config.lesser = params.ZeroAddress
	config.greater = params.ZeroAddress
	config.targetAddress = params.ZeroAddress
	config.Commission = 80
	//-----------------------------------------------------

	if ctx.IsSet(KeyStoreFlag.Name) {
		path = ctx.GlobalString(KeyStoreFlag.Name)
	}
	if ctx.IsSet(PasswordFlag.Name) {
		password = ctx.GlobalString(PasswordFlag.Name)
	}

	if ctx.IsSet(CommissionFlag.Name) {
		config.Commission = ctx.GlobalInt64(CommissionFlag.Name)
	}
	if ctx.IsSet(lesserFlag.Name) {
		config.lesser = common.HexToAddress(ctx.GlobalString(lesserFlag.Name))
	}
	if ctx.IsSet(greaterFlag.Name) {
		config.greater = common.HexToAddress(ctx.GlobalString(greaterFlag.Name))
	}
	if ctx.IsSet(voteNumFlag.Name) {
		config.voteNum = big.NewInt(ctx.Int64(voteNumFlag.Name))
	}
	if ctx.IsSet(TargetAddressFlag.Name) {
		config.targetAddress = common.HexToAddress(ctx.GlobalString(TargetAddressFlag.Name))
	}
	if ctx.IsSet(ValueFlag.Name) {
		config.Value = ctx.GlobalUint64(ValueFlag.Name)
	}
	if ctx.IsSet(TopNumFlag.Name) {
		config.TopNum = big.NewInt(ctx.GlobalInt64(TopNumFlag.Name))
	}

	account := loadAccount(path, password)
	blsPub, err := account.BLSPublicKey()
	if err != nil {
		return nil
	}
	config.PublicKey = account.PublicKey()
	config.from = account.Address
	config.PrivateKey = account.PrivateKey
	config.BlsPub = blsPub
	config.BLSProof = account.MustBLSProofOfPossession()
	conn, _ := dialConn(ctx)
	config.conn = conn
	return config
}
