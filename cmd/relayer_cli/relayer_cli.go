package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mapprotocol/atlas/chains/headers/ethereum"
	"github.com/mapprotocol/atlas/cmd/ethclient"
	params2 "github.com/mapprotocol/atlas/params"
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
	abiRelayer, _  = abi.JSON(strings.NewReader(params2.RelayerABIJSON))
	priKey         *ecdsa.PrivateKey
	from           common.Address
	Value          uint64
	fee            uint64
	RelayerAddress common.Address = params2.RelayerAddress
	Base                          = new(big.Int).SetUint64(10000)
)

const (
	datadirPrivateKey      = "key"
	datadirDefaultKeyStore = "keystore"
	RegisterAmount         = 100000
	RewardInterval         = 14
)

func register(ctx *cli.Context) error {

	fmt.Println(abiRelayer.Methods)

	loadPrivate(ctx)

	conn, url := dialConn(ctx)

	printBaseInfo(conn, url)

	//PrintBalance(conn, from)

	value := ethToWei(ctx, false)

	if Value < RegisterAmount {
		printError("Amount must bigger than ", RegisterAmount)
	}

	fmt.Println("Fee", fee, " value ", value)
	input := packInput("register", value)
	txHash := sendContractTransaction(conn, from, RelayerAddress, nil, priKey, input)

	getResult(conn, txHash, true)

	return nil
}

func checkFee(fee *big.Int) {
	if fee.Sign() < 0 || fee.Cmp(Base) > 0 {
		printError("Please set correct fee value")
	}
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
	//msg := ethchain.CallMsg{From: from, To: &toAddress, GasPrice: gasPrice, Value: value, Data: input}
	//gasLimit, err = client.EstimateGas(context.Background(), msg)
	//if err != nil {
	//	fmt.Println("Contract exec failed", err)
	//}
	//if gasLimit < 1 {
	//	gasLimit = 866328
	//}

	// Create the transaction, sign it and schedule it for execution
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, input)

	chainID, _ := client.ChainID(context.Background())
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

func loadPrivateKey(path string) common.Address {
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
	log.Fatal("!", error)
}

func ethToWei(ctx *cli.Context, zero bool) *big.Int {
	Value = ctx.GlobalUint64(ValueFlag.Name)
	if !zero && Value <= 0 {
		printError("value must bigger than 0")
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

	queryTx(conn, txHash, contract, false)
}

func queryTx(conn *ethclient.Client, txHash common.Hash, contract bool, pending bool) {

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
			queryAccountBalance(conn)
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
		printError("Must specify --key or --keystore")
	}

	if priKey == nil {
		printError("load privateKey failed")
	}
}

func dialConn(ctx *cli.Context) (*ethclient.Client, string) {
	ip = ctx.GlobalString("rpcaddr") //utils.RPCListenAddrFlag.Name)
	port = ctx.GlobalInt("rpcport")  //utils.RPCPortFlag.Name)

	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	// Create an IPC based RPC connection to a remote node
	// "http://39.100.97.129:8545"
	conn, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the Atlaschain client: %v", err)
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
func loadSigningKey(keyfile string) common.Address {
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
	return from
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

func queryAccountBalance(conn *ethclient.Client) {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	input := packInput("getRelayerBalance", from)
	msg := ethchain.CallMsg{From: from, To: &RelayerAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		printError("method CallContract error", err)
	}

	//fmt.Println()
	//PrintBalance(conn, from)

	method, _ := abiRelayer.Methods["getRelayerBalance"]
	ret, err := method.Outputs.Unpack(output)
	if len(ret) != 0 {
		args := struct {
			registered    *big.Int
			unregistering *big.Int
			unregistered  *big.Int
		}{
			ret[0].(*big.Int),
			ret[1].(*big.Int),
			ret[2].(*big.Int),
		}
		fmt.Println("query successfully,your account(uint eth):")
		fmt.Println("registered amount:    ", weiToEth(args.registered))
		fmt.Println("unregistering amount: ", weiToEth(args.unregistering))
		fmt.Println("unregistered amount:  ", weiToEth(args.unregistered))
		//fmt.Println("reward amount:        ", weiToEth(args.reward))
		//fmt.Println("fine amount:          ", weiToEth(args.fine))
	} else {
		fmt.Println("Contract query failed result len == 0")
	}
}

func queryRelayerEpoch(conn *ethclient.Client) {
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
	if len(ret) != 0 {
		args := struct {
			start   *big.Int
			end     *big.Int
			relayer bool
		}{
			ret[0].(*big.Int),
			ret[1].(*big.Int),
			ret[2].(bool),
		}
		if args.relayer {
			fmt.Println("query successfully, your account is relayer")
			fmt.Println("start height in epoch: ", args.start)
			fmt.Println("end height in epoch:   ", args.end)
			//fmt.Println("remain height in epoch:", args.remain)
		} else {
			fmt.Println("query successfully, your account isn't in current epoch")
		}

	} else {
		fmt.Println("Contract query failed result len == 0")
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//under func be used to test only, not formal

var syncCommand = cli.Command{
	Name:   "sync",
	Usage:  "sync transactions from eth, not formal tool",
	Action: MigrateFlags(syncTransactonsFromEth),
	Flags:  RegisterFlags,
}

func syncTransactonsFromEth(ctx *cli.Context) error {
	ip = "localhost" //utils.RPCListenAddrFlag.Name)
	port = 7445      //utils.RPCPortFlag.Name)
	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	conn, err := ethclient.Dial(url)
	if err != nil {
		return err
	}

	loadPrivate(ctx)
	from = crypto.PubkeyToAddress(priKey.PublicKey)

	for start := int64(1001); start <= 20000; start += 20 {
		chains := getChains(start, start+19)
		if chains == nil {
			fmt.Println("header is nil")
			continue
		}
		marshal, err := json.Marshal(chains)
		if err != nil {
			return err
		}
		input := packInputStore("save", "ETH", "MAP", marshal)
		hash := sendContractTransaction(conn, from, params2.HeaderStoreAddress, nil, priKey, input)
		fmt.Println("transaction hash:", hash)
		fmt.Println()
		time.Sleep(1e9)
	}
	return nil
}

func getChains(startNum, endNum int64) []*ethereum.Header {
	conn, _ := dialEthConn()
	Headers := make([]*ethereum.Header, 0, endNum-startNum+1)
	for i := startNum; i <= endNum; i++ {
		Header, _ := conn.HeaderByNumber(context.Background(), big.NewInt(i))
		covertHeader := convertChain(Header)
		if Header != nil {
			Headers = append(Headers, covertHeader)
		}
	}
	return Headers
}

func convertChain(e *types.Header) *ethereum.Header {
	header := new(ethereum.Header)
	header.ParentHash = e.ParentHash
	header.UncleHash = e.UncleHash
	header.Coinbase = e.Coinbase
	header.Root = e.Root
	header.TxHash = e.TxHash
	header.ReceiptHash = e.ReceiptHash
	header.GasLimit = e.GasLimit
	header.GasUsed = e.GasUsed
	header.Time = e.Time
	header.MixDigest = e.MixDigest
	header.Nonce = e.Nonce
	header.Bloom = e.Bloom
	if header.Difficulty = new(big.Int); e.Difficulty != nil {
		header.Difficulty = e.Difficulty
	}
	if header.Number = new(big.Int); e.Number != nil {
		header.Number = e.Number
	}
	if len(e.Extra) > 0 {
		header.Extra = make([]byte, len(e.Extra))
		header.Extra = e.Extra
	}
	return header
}

func dialEthConn() (*ethclient.Client, string) {
	ip = "localhost" //utils.RPCListenAddrFlag.Name)
	port = 8545      //utils.RPCPortFlag.Name)
	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	conn, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the Atlaschain client: %v", err)
	}
	return conn, url
}

func packInputStore(abiMethod string, params ...interface{}) []byte {
	abiHeaderStore, _ := abi.JSON(strings.NewReader(params2.HeaderStoreABIJSON))
	input, err := abiHeaderStore.Pack(abiMethod, params...)
	if err != nil {
		log.Fatal(abiMethod, " error ", err)
	}
	return input
}
