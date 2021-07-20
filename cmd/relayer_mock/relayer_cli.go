package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	params2 "github.com/mapprotocol/atlas/params"

	"errors"
	"fmt"
	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mapprotocol/atlas/cmd/ethclient"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	key   string
	store string
	ip    string
	port  int
)

var (
	abiRelayer, _     = abi.JSON(strings.NewReader(params2.RelayerABIJSON))
	abiHeaderStore, _ = abi.JSON(strings.NewReader(params2.HeaderStoreABIJSON))
	priKey            *ecdsa.PrivateKey
	from              common.Address
	Value             uint64
	fee               uint64
	holder            common.Address
	//baseUnit   = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	RelayerAddress     common.Address = params2.RelayerAddress
	HeaderStoreAddress common.Address = params2.HeaderStoreAddress
	Base                              = new(big.Int).SetUint64(10000)
)

const (
	datadirPrivateKey            = "key"
	datadirDefaultKeyStore       = "keystore"
	RegisterAmount               = 100000
	RewardInterval               = 14
	impawnValue            int64 = 100000
)

func register(ctx *cli.Context) *ethclient.Client {

	loadPrivate(ctx)

	conn, url := dialConn(ctx)

	printBaseInfo(conn, url)

	PrintBalance(conn, from)

	value := ethToWei(ctx, false)

	if impawnValue < RegisterAmount {
		printError("Amount must bigger than ", RegisterAmount)
	}

	fee = ctx.GlobalUint64(FeeFlag.Name)
	checkFee(new(big.Int).SetUint64(fee))

	pubkey, pk, _ := getPubKey(ctx)

	fmt.Println("Fee", fee, " Pubkey ", pubkey, " value ", value)
	input := packInput("register", pk, new(big.Int).SetUint64(fee), value)
	txHash := sendContractTransaction(conn, from, RelayerAddress, nil, priKey, input)

	getResult(conn, txHash, true, false)

	return conn
}

func checkFee(fee *big.Int) {
	if fee.Sign() < 0 || fee.Cmp(Base) > 0 {
		printError("Please set correct fee value")
	}
}

func getPubKey(ctx *cli.Context) (string, []byte, error) {
	var (
		pubkey string
		err    error
	)

	pk := crypto.FromECDSAPub(&priKey.PublicKey)
	pubkey = hex.EncodeToString(pk)
	pk = common.Hex2Bytes(pubkey)
	if _, err := crypto.UnmarshalPubkey(pk); err != nil {
		printError("ValidPk error", err)
	}
	return pubkey, pk, err
}

func sendContractTransaction(client *ethclient.Client, from, toAddress common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, input []byte) common.Hash {
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
	fmt.Println("TX data nonce ", nonce, " transfer value ", value, " gasLimit ", gasLimit, " gasPrice ", gasPrice, " chainID ", chainID)
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	return signedTx.Hash()
}

func loadPrivateKey(ctx *cli.Context, path string) common.Address {
	var err error
	if path == "" {
		file, err := getAllFile(datadirPrivateKey)
		if err != nil {
			printError(" getAllFile file name error", err)
		}
		kab, _ := filepath.Abs(datadirPrivateKey)
		path = filepath.Join(kab, file)
	}
	priKey, err = crypto.LoadECDSA(path)
	if err != nil {
		printError("LoadECDSA error", err)
	}
	from = crypto.PubkeyToAddress(priKey.PublicKey)
	if ctx.IsSet(PublicAdressFlag.Name) {
		ctx.GlobalSet(PublicAdressFlag.Name, ctx.String(from.String()))
	}
	return from
}

func getAllFile(path string) (string, error) {
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		printError("path ", err)
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

func printError(error ...interface{}) {
	log.Fatal(error)
}

func ethToWei(ctx *cli.Context, zero bool) *big.Int {
	if !zero && int(impawnValue) <= 0 {
		printError("Value must bigger than 0")
	}
	baseUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	value := new(big.Int).Mul(big.NewInt(impawnValue), baseUnit)
	return value
}

func weiToEth(value *big.Int) uint64 {
	baseUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	valueT := new(big.Int).Div(value, baseUnit).Uint64()
	return valueT
}

func getResult(conn *ethclient.Client, txHash common.Hash, contract bool, delegate bool) bool {
	fmt.Println("Please waiting ", " txHash ", txHash.String())

	count := 0
	for {
		time.Sleep(time.Millisecond * 200)
		_, isPending, err := conn.TransactionByHash(context.Background(), txHash)
		if err != nil {
			log.Fatal(err)
		}
		count++
		if !isPending {
			break
		}
		if count >= 40 {
			fmt.Println("Please use querytx sub command query later.")
			os.Exit(0)
		}
	}

	queryTx(conn, txHash, contract, false, delegate)
	return true
}

func queryTx(conn *ethclient.Client, txHash common.Hash, contract bool, pending bool, delegate bool) {

	if pending {
		_, isPending, err := conn.TransactionByHash(context.Background(), txHash)
		if err != nil {
			log.Fatal(err)
		}
		if isPending {
			println("In tx_pool no validator  process this, please query later")
			os.Exit(0)
		}
	}

	receipt, err := conn.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		log.Fatal(err)
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		block, err := conn.BlockByHash(context.Background(), receipt.BlockHash)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Transaction Success", " block Number", receipt.BlockNumber.Uint64(), " block txs", len(block.Transactions()), "blockhash", block.Hash().Hex())
		if contract && common.IsHexAddress(from.Hex()) {
			queryRegisterInfo(conn)
		}
	} else if receipt.Status == types.ReceiptStatusFailed {
		fmt.Println("Transaction Failed ", " Block Number", receipt.BlockNumber.Uint64())
	}
}

func packInput(abiMethod string, params ...interface{}) []byte {
	input, err := abiRelayer.Pack(abiMethod, params...)
	if err != nil {
		printError(abiMethod, " error ", err)
	}
	return input
}

func PrintBalance(conn *ethclient.Client, from common.Address) {
	balance, err := conn.BalanceAt(context.Background(), from, nil)
	if err != nil {
		log.Fatal(err)
	}
	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	trueValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))

	sbalance, err := conn.LockBalanceAt(context.Background(), from, nil)
	fmt.Println("Your wallet valid balance is ", trueValue, "'true ", " lock balance is ", sbalance, "'true ")
}

func loadPrivate(ctx *cli.Context) {
	key = ctx.GlobalString(KeyFlag.Name)
	store = ctx.GlobalString(KeyStoreFlag.Name)
	if key != "" {
		loadPrivateKey(ctx, key)
	} else if store != "" {
		loadSigningKey(ctx, store)
	} else {
		printError("Must specify --key or --keystore")
	}

	if priKey == nil {
		printError("load privateKey failed")
	}
}

func dialConn(ctx *cli.Context) (*ethclient.Client, string) {
	ip = ctx.GlobalString(RPCListenAddrFlag.Name) //utils.RPCListenAddrFlag.Name)
	port = ctx.GlobalInt(RPCPortFlag.Name)        //utils.RPCPortFlag.Name)

	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	// Create an IPC based RPC connection to a remote node
	// "http://39.100.97.129:8545"
	conn, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the Abeychain client: %v", err)
	}
	return conn, url
}

func printBaseInfo(conn *ethclient.Client, url string) *types.Header {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	if common.IsHexAddress(from.Hex()) {
		fmt.Println("Connect url ", url, " current number ", header.Number.String(), " address ", from.Hex())
	} else {
		fmt.Println("Connect url ", url, " current number ", header.Number.String())
	}

	return header
}

// loadSigningKey loads a private key in Ethereum keystore format.
func loadSigningKey(ctx *cli.Context, keyfile string) common.Address {
	keyjson, err := ioutil.ReadFile(keyfile)
	if err != nil {
		printError(fmt.Errorf("failed to read the keyfile at '%s': %v", keyfile, err))
	}
	password, _ := prompt.Stdin.PromptPassword("Please enter the password for '" + keyfile + "': ")
	//password := "secret"
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		printError(fmt.Errorf("error decrypting key: %v", err))
	}
	priKey = key.PrivateKey
	from = crypto.PubkeyToAddress(priKey.PublicKey)
	//fmt.Println("address ", from.Hex(), "key", hex.EncodeToString(crypto.FromECDSA(priKey)))
	ctx.GlobalSet(PublicAdressFlag.Name, ctx.String(from.String()))
	return from
}

func queryRewardInfo(conn *ethclient.Client, number uint64, start bool) {
	sheader, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		printError("get snail block error", err)
	}
	queryReward := uint64(0)
	currentReward := sheader.Number.Uint64() - RewardInterval
	if number > currentReward {
		printError("reward no release current reward height ", currentReward)
	} else if number > 0 || start {
		queryReward = number
	} else {
		queryReward = currentReward
	}
	var crc map[string]interface{}
	crc, err = conn.GetChainRewardContent(context.Background(), from, new(big.Int).SetUint64(queryReward))
	if err != nil {
		printError("get chain reward content error", err)
	}
	if info, ok := crc["stakingReward"]; ok {
		if info, ok := info.([]interface{}); ok {
			fmt.Println("queryRewardInfo", info)
		}
	}
}

func queryRegisterInfo(conn *ethclient.Client) {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	var input []byte
	input = packInput("getRelayer", from)
	msg := ethchain.CallMsg{From: from, To: &RelayerAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		printError("method CallContract error", err)
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
		fmt.Println("query successfully,your account:")
		fmt.Println("register: ", args.register)
		fmt.Println("relayer:", args.relayer)
		fmt.Println("current epoch:", args.epoch)
	} else {
		fmt.Println("Contract query failed result len == 0")
	}
}

func getRelayerRange() (a, b uint64) {
	return 1, 100
}
func queryIsRegister(conn *ethclient.Client) bool {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	var input []byte
	input = packInput("getRelayer", from)
	msg := ethchain.CallMsg{From: from, To: &RelayerAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		printError("method CallContract error", err)
	}
	method, _ := abiRelayer.Methods["getRelayer"]
	ret, err := method.Outputs.Unpack(output)
	if len(ret) != 0 {
		return ret[1].(bool)
	} else {
		return false
	}
}
func queryRelayerEpoch(conn *ethclient.Client, currentNum uint64) bool {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	input := packInput("getPeriodHeight", from)
	msg := ethchain.CallMsg{From: from, To: &RelayerAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		printError("method CallContract error", err)
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