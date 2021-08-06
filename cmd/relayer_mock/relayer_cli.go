package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mapprotocol/atlas/accounts/keystore"
	"github.com/mapprotocol/atlas/cmd/ethclient"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"log"
	"math/big"
	"time"
)

var (
	key   string
	store string
	ip    string
	port  int
)

var (
	//baseUnit   = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	Base = new(big.Int).SetUint64(10000)
)
var (
	contractQueryFailedErr = errors.New("Contract query failed result ")
)

const (
	registerValue          int64 = 100000
	datadirPrivateKey            = "key"
	datadirDefaultKeyStore       = "keystore"
	RegisterAmount               = 100000
	RewardInterval               = 14
)

func getConn11(ctx *cli.Context) *ethclient.Client {
	conn, _ := dialConn()
	return conn
}

func register(ctx *cli.Context, conn *ethclient.Client) (common.Address, *ecdsa.PrivateKey) {
	from, priKey := loadPrivate(ctx)
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	if common.IsHexAddress(from.Hex()) {
		fmt.Println(" current number ", header.Number.String(), " address ", from.Hex())
	} else {
		fmt.Println(" current number ", header.Number.String())
	}

	fmt.Println("Your wallet balance is ", getBalance(conn, from), "'eth ")
	value := ethToWei(registerValue)
	fee := ctx.GlobalUint64(FeeFlag.Name)
	checkFee(new(big.Int).SetUint64(fee))
	pubkey, pk, _ := getPubKey(priKey)
	fmt.Println("Fee", fee, " Pubkey ", pubkey, " value ", value)
	input := packInput("register", pk, new(big.Int).SetUint64(fee), value)
	sendContractTransaction(conn, from, RelayerAddress, nil, priKey, input)
	return from, priKey
}

func checkFee(fee *big.Int) {
	if fee.Sign() < 0 || fee.Cmp(Base) > 0 {
		log.Fatal("Please set correct fee value")
	}
}

func getPubKey(priKey *ecdsa.PrivateKey) (string, []byte, error) {
	var (
		pubkey string
		err    error
	)
	pk := crypto.FromECDSAPub(&priKey.PublicKey)
	pubkey = crypto.PubkeyToAddress(priKey.PublicKey).String()
	if _, err := crypto.UnmarshalPubkey(pk); err != nil {
		log.Fatal("ValidPk error", err)
	}
	return pubkey, pk, err
}

func sendContractTransaction(client *ethclient.Client, from, toAddress common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, input []byte) {
	// Ensure a valid value field and resolve the account nonce
	nonce, err := client.PendingNonceAt(context.Background(), from)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	gasLimit := uint64(2100000) // in units
	// If the contract surely has code (or code is not needed), estimate the transaction
	msg := ethchain.CallMsg{From: from, To: &toAddress, GasPrice: gasPrice, Value: value, Data: input}
	gasLimit, err = client.EstimateGas(context.Background(), msg)
	if err != nil {
		fmt.Println("Contract exec failed", err)
	}
	if gasLimit < 1 {
		gasLimit = 866328
	}

	// Create the transaction, sign it and schedule it for execution
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, input)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("TX data nonce ", nonce, " transfer value ", value, " gasLimit ", gasLimit, " gasPrice ", gasPrice, " chainID ", chainID)
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	txHash := signedTx.Hash()
	count := 0
	for {
		time.Sleep(time.Millisecond * 200)
		_, isPending, err := client.TransactionByHash(context.Background(), txHash)
		if err != nil {
			log.Fatal(err)
		}
		count++
		if !isPending {
			break
		}
	}
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		log.Fatal(err)
	}
	if receipt.Status == types.ReceiptStatusSuccessful {
		_, err := client.BlockByHash(context.Background(), receipt.BlockHash)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Transaction Success", " block Number", receipt.BlockNumber.Uint64())
	} else if receipt.Status == types.ReceiptStatusFailed {
		fmt.Println("Transaction Failed ", " Block Number", receipt.BlockNumber.Uint64())
	}

}

func getAllFile(path string) (string, error) {
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal("path ", err)
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

func packInput(abiMethod string, params ...interface{}) []byte {
	input, err := abiRelayer.Pack(abiMethod, params...)
	if err != nil {
		log.Fatal(abiMethod, " error ", err)
	}
	return input
}

func loadPrivate(ctx *cli.Context) (common.Address, *ecdsa.PrivateKey) {
	key = ctx.GlobalString(KeyFlag.Name)
	store = ctx.GlobalString(KeyStoreFlag.Name)
	keyjson, err := ioutil.ReadFile(store)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to read the keyfile at '%s': %v", store, err))
	}
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil || key == nil {
		log.Fatal(fmt.Errorf("error decrypting key: %v", err))
	}
	priKey := key.PrivateKey
	from := crypto.PubkeyToAddress(priKey.PublicKey)
	//fmt.Println("address ", from.Hex(), "key", hex.EncodeToString(crypto.FromECDSA(priKey)))
	ctx.GlobalSet(PublicAdressFlag.Name, ctx.String(from.String()))

	if priKey == nil {
		log.Fatal("load privateKey failed")
	}
	return from, priKey
}
func dialConn() (*ethclient.Client, string) {
	ip = AtlasRPCListenAddr
	port = AtlasRPCPortFlag
	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	conn, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the atlas chain client: %v", err)
	}
	return conn, url
}

func queryRegisterInfo(conn *ethclient.Client, from common.Address) (bool, bool, *big.Int, error) {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	var input []byte
	input = packInput("getRelayer", from)
	msg := ethchain.CallMsg{From: from, To: &RelayerAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Fatal("method CallContract error", err)
	}

	method, _ := abiRelayer.Methods["getRelayer"]
	ret, err := method.Outputs.Unpack(output)
	if len(ret) != 0 {
		args := struct {
			register bool
			relayer  bool
			epoch    *big.Int
		}{
			ret[0].(bool),
			ret[1].(bool),
			ret[2].(*big.Int),
		}
		return args.register, args.relayer, args.epoch, nil
	} else {
		fmt.Println("Contract query failed result len == 0")
		return false, false, nil, contractQueryFailedErr
	}
}

func queryIsRegister(conn *ethclient.Client, from common.Address) bool {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	var input []byte
	input = packInput("getRelayer", from)
	msg := ethchain.CallMsg{From: from, To: &RelayerAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Fatal("method CallContract error", err)
	}
	method, _ := abiRelayer.Methods["getRelayer"]
	ret, err := method.Outputs.Unpack(output)
	if len(ret) != 0 {
		return ret[1].(bool)
	} else {
		return false
	}
}
func queryRelayerEpoch(conn *ethclient.Client, currentNum uint64, from common.Address) bool {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	input := packInput("getPeriodHeight", from)
	msg := ethchain.CallMsg{From: from, To: &RelayerAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Fatal("method CallContract error", err)
	}

	method, _ := abiRelayer.Methods["getPeriodHeight"]
	ret, err := method.Outputs.Unpack(output)
	start := ret[0].(*big.Int).Uint64()
	end := ret[1].(*big.Int).Uint64()
	remain := ret[2].(*big.Int).Uint64()
	relayer := ret[3].(bool)

	if len(ret) != 0 {
		if relayer {
			fmt.Println("query successfully,your account is relayer")
			fmt.Println("start height in epoch: ", start)
			fmt.Println("end height in epoch:   ", end)
			fmt.Println("remain height in epoch:", remain)
			if start <= currentNum && currentNum <= end {
				return true
			} else {
				return false
			}
		} else {
			fmt.Println("query successfully,your account is not relayer")
			return false
		}
	} else {
		fmt.Println("Contract query failed result len == 0")
		return false
	}
}
func withdraw(conn *ethclient.Client, from common.Address, priKey *ecdsa.PrivateKey) error {

	value := ethToWei(100000)

	input := packInput("withdraw", from, value)

	sendContractTransaction(conn, from, RelayerAddress, new(big.Int).SetInt64(0), priKey, input)

	return nil
}

func Append(conn *ethclient.Client, from common.Address, priKey *ecdsa.PrivateKey) error {

	value := ethToWei(100000)

	input := packInput("append", from, value)

	sendContractTransaction(conn, from, RelayerAddress, nil, priKey, input)

	return nil
}
