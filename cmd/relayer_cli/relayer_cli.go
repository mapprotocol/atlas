package main

import (
	"context"
	"crypto/ecdsa"
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
	"github.com/mapprotocol/atlas/core/vm"
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
	abiRelayer, _ = abi.JSON(strings.NewReader(vm.RelayerABIJSON))
	priKey        *ecdsa.PrivateKey
	from          common.Address
	Value         uint64
	fee           uint64
	holder        common.Address
	//baseUnit   = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	RelayerAddress common.Address = common.BytesToAddress([]byte("relayeraddress"))
	Base                          = new(big.Int).SetUint64(10000)
)

const (
	datadirPrivateKey      = "key"
	datadirDefaultKeyStore = "keystore"
	RegisterAmount         = 100000
	RewardInterval         = 14
)

func register(ctx *cli.Context) error {

	loadPrivate(ctx)

	conn, url := dialConn(ctx)

	printBaseInfo(conn, url)

	PrintBalance(conn, from)

	value := ethToWei(ctx, false)

	if Value < RegisterAmount {
		printError("Amount must bigger than ", RegisterAmount)
	}

	fee = ctx.GlobalUint64(FeeFlag.Name)
	checkFee(new(big.Int).SetUint64(fee))

	pubkey, pk, _ := getPubKey(ctx)

	fmt.Println("Fee", fee, " Pubkey ", pubkey, " value ", value)
	input := packInput("register", pk, new(big.Int).SetUint64(fee), value)
	txHash := sendContractTransaction(conn, from, RelayerAddress, nil, priKey, input)

	getResult(conn, txHash, true, false)

	return nil
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

	if ctx.GlobalIsSet(PubKeyKeyFlag.Name) {
		pubkey = ctx.GlobalString(PubKeyKeyFlag.Name)
	} else if ctx.GlobalIsSet(BFTKeyKeyFlag.Name) {
		bftKey, err := crypto.HexToECDSA(ctx.GlobalString(BFTKeyKeyFlag.Name))
		if err != nil {
			printError("bft key error", err)
		}
		pk := crypto.FromECDSAPub(&bftKey.PublicKey)
		pubkey = common.Bytes2Hex(pk)
	} else {
		//pubkey, err = conn.Pubkey(context.Background())
		//if err != nil {
		printError("get pubkey error", err)
		//}
	}

	pk := common.Hex2Bytes(pubkey)
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
	log.Fatal(error)
}

func ethToWei(ctx *cli.Context, zero bool) *big.Int {
	Value = ctx.GlobalUint64(ValueFlag.Name)
	if !zero && Value <= 0 {
		printError("Value must bigger than 0")
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

func getResult(conn *ethclient.Client, txHash common.Hash, contract bool, delegate bool) {
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
			queryRegisterInfo(conn, false, delegate)
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

func queryRegisterInfo(conn *ethclient.Client, query bool, delegate bool) {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	var input []byte
	if delegate {
		input = packInput("getDelegate", from, holder)
	} else {
		input = packInput("getDeposit", from)
	}
	msg := ethchain.CallMsg{From: from, To: &RelayerAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		printError("method CallContract error", err)
	}
	if len(output) != 0 {
		args := struct {
			Staked   *big.Int
			Locked   *big.Int
			Unlocked *big.Int
		}{}
		err = abiRelayer.UnpackIntoInterface(&args, "getDeposit", output)
		if err != nil {
			printError("abi error", err)
		}
		fmt.Println("Staked ", args.Staked.String(), "wei =", weiToEth(args.Staked), "true Locked ",
			args.Locked.String(), " wei =", weiToEth(args.Locked), "true",
			"Unlocked ", args.Unlocked.String(), " wei =", weiToEth(args.Unlocked), "true")
		if query && args.Locked.Sign() > 0 {
			lockAssets, err := conn.GetLockedAsset(context.Background(), from, header.Number)
			if err != nil {
				printError("GetLockedAsset error", err)
			}
			for k, v := range lockAssets {
				for m, n := range v.LockValue {
					if !n.Locked {
						fmt.Println("Your can instant withdraw", " count value ", n.Amount, " true")
					} else {
						if n.EpochID > 0 || n.Amount != "0" {
							fmt.Println("Your can withdraw after height", n.Height.Uint64(), " count value ", n.Amount, " true  index", k+m, " lock ", n.Locked)
						}
					}
				}
			}
		}
	} else {
		fmt.Println("Contract query failed result len == 0")
	}
}
