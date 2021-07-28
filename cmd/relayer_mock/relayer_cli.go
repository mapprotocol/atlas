package main

import (
	"context"
	"crypto/ecdsa"
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
	url               string
	fee               uint64
	//baseUnit   = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	RelayerAddress     common.Address = params2.RelayerAddress
	HeaderStoreAddress common.Address = params2.HeaderStoreAddress
	Base                              = new(big.Int).SetUint64(10000)
	impawnValue        int64          = 100000
)
var (
	contractQueryFailedErr        = errors.New("Contract query failed result ")
	SubmitAtDifferentEpochCommand = cli.Command{
		Name:   "SubmitAtDifferentEpoch",
		Usage:  "SubmitAtDifferentEpoch ",
		Action: MigrateFlags(SubmitAtDifferentEpoch),
		Flags:  relayerflags,
	}
	SubmitMultipleTimesAtCurEpochCommand = cli.Command{
		Name:   "SubmitMultipleTimesAtCurEpoch",
		Usage:  "SubmitMultipleTimesAtCurEpoch ",
		Action: MigrateFlags(SubmitMultipleTimesAtCurEpoch),
		Flags:  relayerflags,
	}
	submissionOfDifferentAccountsCommand = cli.Command{
		Name:   "submissionOfDifferentAccounts",
		Usage:  "submissionOfDifferentAccounts ",
		Action: MigrateFlags(submissionOfDifferentAccounts),
		Flags:  relayerflags,
	}
	withdrawAtDifferentEpochCommand = cli.Command{
		Name:   "withdrawAtDifferentEpoch",
		Usage:  "withdrawAtDifferentEpoch ",
		Action: MigrateFlags(withdrawAtDifferentEpoch),
		Flags:  relayerflags,
	}
	withdrawAccordingToDifferentBalanceCommand = cli.Command{
		Name:   "withdrawAccordingToDifferentBalance",
		Usage:  "withdrawAccordingToDifferentBalance ",
		Action: MigrateFlags(withdrawAccordingToDifferentBalance),
		Flags:  relayerflags,
	}
	appendAtDifferentEpochCommand = cli.Command{
		Name:   "appendAtDifferentEpoch",
		Usage:  "appendAtDifferentEpoch ",
		Action: MigrateFlags(appendAtDifferentEpoch),
		Flags:  relayerflags,
	}
)

const (
	datadirPrivateKey      = "key"
	datadirDefaultKeyStore = "keystore"
	RegisterAmount         = 100000
	RewardInterval         = 14
)

func getConn(ctx *cli.Context) *ethclient.Client {
	conn, url1 := dialConn(ctx)
	url = url1
	return conn
}

func register(ctx *cli.Context, conn *ethclient.Client, from1 common.Address) {
	loadPrivate(ctx)
	from1 = from
	printBaseInfo(conn)
	PrintBalance(conn, from1)
	value := ethToWei(false)
	if impawnValue < RegisterAmount {
		log.Fatal("Amount must bigger than ", RegisterAmount)
	}
	fee = ctx.GlobalUint64(FeeFlag.Name)
	checkFee(new(big.Int).SetUint64(fee))
	pubkey, pk, _ := getPubKey(priKey)
	fmt.Println("Fee", fee, " Pubkey ", pubkey, " value ", value)
	input := packInput("register", pk, new(big.Int).SetUint64(fee), value)
	txHash := sendContractTransaction(conn, from1, RelayerAddress, nil, priKey, input)
	getResult(conn, txHash, true, from1)
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
			log.Fatal(" getAllFile file name error", err)
		}
		kab, _ := filepath.Abs(datadirPrivateKey)
		path = filepath.Join(kab, file)
	}
	priKey, err = crypto.LoadECDSA(path)
	if err != nil {
		log.Fatal("LoadECDSA error", err)
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

func ethToWei(zero bool) *big.Int {
	if !zero && int(impawnValue) <= 0 {
		log.Fatal("Value must bigger than 0")
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

func getResult(conn *ethclient.Client, txHash common.Hash, contract bool, from common.Address) bool {
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
	queryTx(conn, txHash, contract, from)
	return true
}

func queryTx(conn *ethclient.Client, txHash common.Hash, contract bool, from common.Address) {

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
			queryRegisterInfo(conn, from, "myAccount")
		}
	} else if receipt.Status == types.ReceiptStatusFailed {
		fmt.Println("Transaction Failed ", " Block Number", receipt.BlockNumber.Uint64())
	}
}

func packInput(abiMethod string, params ...interface{}) []byte {
	input, err := abiRelayer.Pack(abiMethod, params...)
	if err != nil {
		log.Fatal(abiMethod, " error ", err)
	}
	return input
}
func PrintBalance(conn *ethclient.Client, from common.Address) *big.Float {
	balance, err := conn.BalanceAt(context.Background(), from, nil)
	if err != nil {
		log.Fatal(err)
	}
	balance2 := new(big.Float)
	balance2.SetString(balance.String())
	Value := new(big.Float).Quo(balance2, big.NewFloat(math.Pow10(18)))
	//fmt.Println("Your wallet balance is ", Value, "'eth ")
	return Value
}

func loadPrivate(ctx *cli.Context) {
	key = ctx.GlobalString(KeyFlag.Name)
	store = ctx.GlobalString(KeyStoreFlag.Name)
	if key != "" {
		loadPrivateKey(ctx, key)
	} else if store != "" {
		loadSigningKey(ctx, store)
	} else {
		log.Fatal("Must specify --key or --keystore")
	}

	if priKey == nil {
		log.Fatal("load privateKey failed")
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

func printBaseInfo(conn *ethclient.Client) *types.Header {
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
		log.Fatal(fmt.Errorf("failed to read the keyfile at '%s': %v", keyfile, err))
	}
	password, _ := prompt.Stdin.PromptPassword("Please enter the password for '" + keyfile + "': ")
	//password := "secret"
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil || key == nil {
		log.Fatal(fmt.Errorf("error decrypting key: %v", err))
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
		log.Fatal("get snail block error", err)
	}
	queryReward := uint64(0)
	currentReward := sheader.Number.Uint64() - RewardInterval
	if number > currentReward {
		log.Fatal("reward no release current reward height ", currentReward)
	} else if number > 0 || start {
		queryReward = number
	} else {
		queryReward = currentReward
	}
	var crc map[string]interface{}
	crc, err = conn.GetChainRewardContent(context.Background(), from, new(big.Int).SetUint64(queryReward))
	if err != nil {
		log.Fatal("get chain reward content error", err)
	}
	if info, ok := crc["stakingReward"]; ok {
		if info, ok := info.([]interface{}); ok {
			fmt.Println("queryRewardInfo", info)
		}
	}
}

func queryRegisterInfo(conn *ethclient.Client, from common.Address, whichAccount string) (bool, bool, *big.Int, error) {
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
	fmt.Println()
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
		if boolPrint {
			fmt.Println("your account relayerInfo:")
			fmt.Printf("%v register:%v \n", whichAccount, args.register)
			fmt.Printf("%v relayer: %v \n", whichAccount, args.relayer)
			fmt.Printf("%v current epoch: %v \n", whichAccount, args.epoch)
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
func queryRelayerEpoch(conn *ethclient.Client, currentNum uint64) bool {
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

	value := ethToWei(false)
	fmt.Println("value:======>", value)
	input := packInput("withdraw", from, value)

	txHash := sendContractTransaction(conn, from, RelayerAddress, new(big.Int).SetInt64(0), priKey, input)

	getResult(conn, txHash, true, from)

	return nil
}

func Append(conn *ethclient.Client, from common.Address, priKey *ecdsa.PrivateKey) error {

	value := ethToWei(false)

	input := packInput("append", from, value)

	txHash := sendContractTransaction(conn, from, RelayerAddress, nil, priKey, input)

	getResult(conn, txHash, true, from)

	return nil
}

// SubmitAtDifferentEpoch
// 测试结果：注册成功直接成为 relayer 当前Epoch为1 到 Epoch为2时候保存失败
// 到了第二阶 就不能同步了 得是relayer  钱没有变化
func SubmitAtDifferentEpoch(ctx *cli.Context) error {
	fmt.Println("==============SubmitAtDifferentEpoch==============")
	conn := getConn(ctx)
	priKey, from = loadprivateCommon(keystore1)
	register(ctx, conn, from)
	_, _, curEpoch, err := queryRegisterInfo(conn, from, "myAccount")
	if err != nil {
		log.Fatal(err)
	}
	oldbalance := PrintBalance(conn, from)
	curEpoch2 := big.NewInt(curEpoch.Int64())
	count := 0
	connEth, _ := dialEthConn()
	chains := getChainsCommon(connEth)
	for {
		curBalance := PrintBalance(conn, from)
		_, _, curEpoch, err = queryRegisterInfo(conn, from, "001:")
		fmt.Println("curEpoch: ", curEpoch, "waitEpoch: ", curEpoch2, "   moneyChange:",
			oldbalance.Abs(oldbalance.Sub(oldbalance, curBalance)))
		if err != nil {
			log.Fatal(err)
		}
		if curEpoch2.Cmp(curEpoch) == 0 {
			fmt.Println("===================save==================")
			aBalance := PrintBalance(conn, from)
			SaveByNum(conn, 10, from, chains)
			curEpoch2.Add(curEpoch2, common.Big1)
			bBalance := PrintBalance(conn, from)
			oldbalance = bBalance
			fmt.Printf("old money:%v  new money %v change %v\n",
				aBalance, bBalance, aBalance.Abs(aBalance.Sub(aBalance, bBalance)))
		}
		time.Sleep(time.Second * 5)
		count++
		if count > 100 {
			fmt.Println("if you want continue please add the count limit")
			os.Exit(1)
		}
	}
	return nil
}

// SubmitMultipleTimesAtCurEpoch
// 同一期多次提交都可以提交 到达下一期不能同步了 下一期的时候钱没增多
func SubmitMultipleTimesAtCurEpoch(ctx *cli.Context) error {
	fmt.Println("==============SubmitMultipleTimesAtCurEpoch==============")
	conn := getConn(ctx)
	priKey, from = loadprivateCommon(keystore1)
	register(ctx, conn, from)
	_, _, curEpoch, err := queryRegisterInfo(conn, from, "001")
	if err != nil {
		log.Fatal(err)
	}
	oldbalance := PrintBalance(conn, from)
	connEth, _ := dialEthConn()
	chains := getChainsCommon(connEth)
	curEpoch2 := big.NewInt(curEpoch.Int64())
	count := 0
	for {
		curBalance := PrintBalance(conn, from)
		_, _, curEpoch, err = queryRegisterInfo(conn, from, "001")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("curEpoch: ", curEpoch, "waitEpoch: ", curEpoch2, "   moneyChange:",
			oldbalance.Abs(oldbalance.Sub(oldbalance, curBalance)))
		if curEpoch2.Cmp(curEpoch) == 0 {
			aBalance := PrintBalance(conn, from)
			SaveByNum(conn, 10, from, chains)
			bBalance := PrintBalance(conn, from)
			fmt.Printf("old money:%v  new money %v change %v\n",
				aBalance, bBalance, aBalance.Abs(aBalance.Sub(aBalance, bBalance)))
			time.Sleep(time.Second)
		}
		count++
		if count > 100 {
			fmt.Println("if you want continue please add the count limit")
			os.Exit(1)
		}
	}
	return nil
}

//submission of different accounts
// 从中间阶开始注册 5个里面有1个会注册失败 全部不是relayer 没有进行同步工作  下一阶段也没有成为ralayer register amount会多出 100000 ETH 其后不变
// 从第一届开始注册 5个会注册失败 查出来注册金额register amount: 10万eth 后来阶段也不会成为 relayer
func submissionOfDifferentAccounts(ctx *cli.Context) error {
	fmt.Println("==============submissionOfDifferentAccounts==============")
	conn := getConn(ctx)

	password = "123456"
	_, _, from1 := registerCommon(conn, keystore1)
	_, _, curEpoch, err := queryRegisterInfo(conn, from1, "from1")
	if err != nil {
		log.Fatal(err)
	}

	password = ""
	_, _, from2 := registerCommon(conn, keystore2)
	_, _, _, err2 := queryRegisterInfo(conn, from2, "from2")
	if err2 != nil {
		log.Fatal(err2)
	}

	password = "123456"
	_, _, from3 := registerCommon(conn, keystore3)
	_, _, _, err3 := queryRegisterInfo(conn, from3, "from3")
	if err3 != nil {
		log.Fatal(err3)
	}

	_, _, from4 := registerCommon(conn, keystore4)
	_, _, _, err4 := queryRegisterInfo(conn, from4, "from4")
	if err4 != nil {
		log.Fatal(err4)
	}

	_, _, from5 := registerCommon(conn, keystore5)
	_, _, _, err5 := queryRegisterInfo(conn, from5, "from5")
	if err5 != nil {
		log.Fatal(err5)
	}

	count := 0
	curEpoch2 := big.NewInt(curEpoch.Int64())
	connEth, _ := dialEthConn()
	chains := getChainsCommon(connEth)

	for {
		boolPrint = false
		_, _, curEpoch, err = queryRegisterInfo(conn, from, "")
		if curEpoch2.Cmp(curEpoch) == 0 {
			fmt.Println("================query==================curEpoch:", curEpoch)
			queryAccountBalance(conn, from1)
			queryAccountBalance(conn, from2)
			queryAccountBalance(conn, from3)
			queryAccountBalance(conn, from4)
			queryAccountBalance(conn, from5)

			fmt.Println("===================== from1 ==============================")
			_, isRelayer1, _, err1 := queryRegisterInfo(conn, from1, "from1:")
			if err1 != nil {
				log.Fatal(err1)
			}
			if isRelayer1 {
				a := PrintBalance(conn, from1)
				SaveByNum(conn, 10, from1, chains)
				b := PrintBalance(conn, from1)
				printChangeBalance(*a, *b)
				queryAccountBalance(conn, from1)
			}
			fmt.Println("================from2====================")
			_, isRelayer2, _, err2 := queryRegisterInfo(conn, from2, "from2:")
			if err2 != nil {
				log.Fatal(err2)
			}
			if isRelayer2 {
				a := PrintBalance(conn, from2)
				SaveByNum(conn, 10, from2, chains)
				b := PrintBalance(conn, from2)
				printChangeBalance(*a, *b)
				queryAccountBalance(conn, from2)
			}
			fmt.Println("==================from3==================")
			_, isRelayer3, _, err3 := queryRegisterInfo(conn, from3, "from3:")
			if err3 != nil {
				log.Fatal(err3)
			}
			if isRelayer3 {
				a := PrintBalance(conn, from3)
				SaveByNum(conn, 10, from3, chains)
				b := PrintBalance(conn, from3)
				printChangeBalance(*a, *b)
				queryAccountBalance(conn, from3)
			}
			fmt.Println("================from4====================")
			_, isRelayer4, _, err4 := queryRegisterInfo(conn, from4, "from4:")
			if err4 != nil {
				log.Fatal(err4)
			}
			if isRelayer4 {
				a := PrintBalance(conn, from4)
				SaveByNum(conn, 10, from4, chains)
				b := PrintBalance(conn, from4)
				printChangeBalance(*a, *b)
				queryAccountBalance(conn, from4)
			}
			fmt.Println("===================from5=================")
			_, isRelayer5, _, err5 := queryRegisterInfo(conn, from5, "from5:")
			if err5 != nil {
				log.Fatal(err5)
			}
			if isRelayer5 {
				a := PrintBalance(conn, from5)
				SaveByNum(conn, 10, from5, chains)
				b := PrintBalance(conn, from5)
				printChangeBalance(*a, *b)
				queryAccountBalance(conn, from5)
			}
			count++
			curEpoch2.Add(curEpoch2, common.Big1)
			if count > 3 {
				break
			}
		}
		time.Sleep(time.Second)
	}

	return nil
}

// withdrawAtDifferentEpoch
// relayer中的总 币值没有变化
// 注册成功 查询的时候 locked amount: 100000000000000000000000 有时候会出来
// 测试结果 当前阶撤销 relayer 用户balance 和 locked amount:没有变化 到下一届 lock->unlock
// 第二阶段 继续撤销 locked amount: 100000000000000000000000 这个会变化
func withdrawAtDifferentEpoch(ctx *cli.Context) error {
	fmt.Println("========================== withdrawAtDifferentEpoch ====================================")
	conn := getConn(ctx)
	priKey, from = loadprivateCommon(keystore1)
	register(ctx, conn, from)
	boolPrint = false
	_, _, curEpoch, err := queryRegisterInfo(conn, from, "001:")
	if err != nil {
		log.Fatal(err)
	}
	boolPrint = true
	oldbalance := PrintBalance(conn, from)
	connEth, _ := dialEthConn()
	chains := getChainsCommon(connEth)
	SaveByNum(conn, 10, from, chains)
	curEpoch2 := big.NewInt(curEpoch.Int64())
	count := 0
	fmt.Println("============= start ==================curEpoch:", curEpoch)
	queryAccountBalance(conn, from)

	if curEpoch2.Cmp(curEpoch) == 0 {
		fmt.Println("===================== withdraw ==============================")
		a := PrintBalance(conn, from)
		err := withdraw(conn, from, priKey)
		if err != nil {
			log.Fatal(err)
		}
		b := PrintBalance(conn, from)
		printChangeBalance(*a, *b)
		oldbalance = b
		curEpoch2.Add(curEpoch2, common.Big1)
		queryAccountBalance(conn, from)
	}
	boolPrint = false
	for {
		_, _, curEpoch, err = queryRegisterInfo(conn, from, "001:")
		if curEpoch2.Cmp(curEpoch) == 0 {
			fmt.Println("================query==================curEpoch:", curEpoch)
			queryAccountBalance(conn, from)
			curBalance := PrintBalance(conn, from)
			printChangeBalance(*oldbalance, *curBalance)

			fmt.Println("===================== withdraw ==============================")
			a := PrintBalance(conn, from)
			err := withdraw(conn, from, priKey)
			if err != nil {
				log.Fatal(err)
			}
			b := PrintBalance(conn, from)
			printChangeBalance(*a, *b)
			queryAccountBalance(conn, from)

			oldbalance = b
			count++
			curEpoch2.Add(curEpoch2, common.Big1)
			if count > 3 {
				break
			}
		}
		time.Sleep(time.Second)
	}

	return nil
}

//Cancellation according to different money
// impawnValue through change impawnValue test
// 123456700000000000000000000000 撤销 最大2000 000 000 000 000 000 000 00
func withdrawAccordingToDifferentBalance(ctx *cli.Context) error {
	fmt.Println("========================== withdrawAtDifferentEpoch ====================================")
	conn := getConn(ctx)
	priKey, from = loadprivateCommon(keystore1)
	register(ctx, conn, from)
	boolPrint = false
	_, _, curEpoch, err := queryRegisterInfo(conn, from, "001:")
	if err != nil {
		log.Fatal(err)
	}

	boolPrint = true
	oldbalance := PrintBalance(conn, from)
	connEth, _ := dialEthConn()
	chains := getChainsCommon(connEth)
	SaveByNum(conn, 10, from, chains)
	curEpoch2 := big.NewInt(curEpoch.Int64())
	count := 0
	fmt.Println("============= start ==================curEpoch:", curEpoch)
	queryAccountBalance(conn, from)

	if curEpoch2.Cmp(curEpoch) == 0 {
		fmt.Println("===================== withdraw ==============================")
		a := PrintBalance(conn, from)
		impawnValue = impawnValue * 1234567 // -------------------------change
		err := withdraw(conn, from, priKey)
		if err != nil {
			log.Fatal(err)
		}
		b := PrintBalance(conn, from)
		printChangeBalance(*a, *b)
		oldbalance = b
		curEpoch2.Add(curEpoch2, common.Big1)
		queryAccountBalance(conn, from)
	}
	boolPrint = false
	for {
		_, _, curEpoch, err = queryRegisterInfo(conn, from, "001:")
		if curEpoch2.Cmp(curEpoch) == 0 {
			fmt.Println("================query==================curEpoch:", curEpoch)
			queryAccountBalance(conn, from)
			curBalance := PrintBalance(conn, from)
			printChangeBalance(*oldbalance, *curBalance)

			fmt.Println("===================== withdraw ==============================")
			a := PrintBalance(conn, from)
			err := withdraw(conn, from, priKey)
			if err != nil {
				log.Fatal(err)
			}
			b := PrintBalance(conn, from)
			printChangeBalance(*a, *b)
			queryAccountBalance(conn, from)

			oldbalance = b
			count++
			curEpoch2.Add(curEpoch2, common.Big1)
			if count > 3 {
				break
			}
		}
		time.Sleep(time.Second)
	}
	return nil
}

// appendAtDifferentEpoch
// 当前阶段追加 币会增加 第二阶段 结算 继续追加 交易会失败
func appendAtDifferentEpoch(ctx *cli.Context) error {
	fmt.Println("========================== appendAtDifferentEpoch ====================================")
	conn := getConn(ctx)
	priKey, from = loadprivateCommon(keystore1)
	fmt.Println("========================== register ==================================== ")
	register(ctx, conn, from)
	boolPrint = false
	_, _, curEpoch, err := queryRegisterInfo(conn, from, "myAccount")
	if err != nil {
		log.Fatal(err)
	}
	boolPrint = true
	oldbalance := PrintBalance(conn, from)
	fmt.Println("========================== save ==================================== ")
	connEth, _ := dialEthConn()
	chains := getChainsCommon(connEth)
	SaveByNum(conn, 10, from, chains)
	curEpoch2 := big.NewInt(curEpoch.Int64())
	count := 0

	fmt.Println("========================== start Append ==================================== Epoch:", curEpoch)
	if curEpoch2.Cmp(curEpoch) == 0 {
		fmt.Println("==========================  Append 1 ==================================== Epoch:", curEpoch)
		a := PrintBalance(conn, from)
		err := Append(conn, from, priKey)
		if err != nil {
			log.Fatal(err)
		}
		b := PrintBalance(conn, from)
		printChangeBalance(*a, *b)
		curEpoch2.Add(curEpoch2, common.Big1)
		queryAccountBalance(conn, from)

		fmt.Println("==========================  Append 2 ==================================== Epoch:", curEpoch)
		a = PrintBalance(conn, from)
		err = Append(conn, from, priKey)
		if err != nil {
			log.Fatal(err)
		}
		b = PrintBalance(conn, from)
		printChangeBalance(*a, *b)
		queryAccountBalance(conn, from)
	}

	boolPrint = false
	for {
		_, _, curEpoch, err = queryRegisterInfo(conn, from, "001:")
		if curEpoch2.Cmp(curEpoch) == 0 {
			fmt.Println("================== query ================curEpoch:", curEpoch)
			queryAccountBalance(conn, from)
			curBalance := PrintBalance(conn, from)
			printChangeBalance(*oldbalance, *curBalance)
			fmt.Println("================== append  ================curEpoch:", curEpoch)

			a := PrintBalance(conn, from)
			err = Append(conn, from, priKey)
			if err != nil {
				log.Fatal(err)
			}
			b := PrintBalance(conn, from)
			printChangeBalance(*a, *b)
			queryAccountBalance(conn, from)

			oldbalance = curBalance
			count++
			curEpoch2.Add(curEpoch2, common.Big1)
			if count > 3 {
				break
			}
		}
		time.Sleep(time.Second)
	}

	return nil
}
