package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/mapprotocol/atlas/cmd/marker/mapprotocol"
	"github.com/mapprotocol/atlas/helper/fileutils"
	"github.com/mapprotocol/atlas/params"

	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"io"
	//"fmt"
)

const DefaultGasLimit = 4500000

var (
	url       = "http://127.0.0.1:8545"
	urlSendTx = "http://13.67.118.60:7445"
	msgCh     chan struct{} // wait for msg handles
)

func init() {
	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.LvlInfo)
	log.Root().SetHandler(glogger)
}

type Voter2validatorInfo struct {
	VoterAccount     string
	ValidatorAccount string
	Value            uint64
}

var voter2validator []Voter2validatorInfo

func main() {
	log.Info("start")
	msgCh = make(chan struct{})
	var accounts []*ecdsa.PrivateKey
	for i := 0; i < 10000; i++ {
		priv0, _ := crypto.GenerateKey()
		accounts = append(accounts, priv0)
	}
	conn, err := rpc.Dial(urlSendTx)
	if err != nil {
		log.Error("Failed to connect to the Atlaschain client: ", err)
	}
	from := "0x81f02fd21657df80783755874a92c996749777bf"
	//from := "0xbe27cf1ed3489b6add51a22ce4b25abd92cac3c8"
	var ret bool
	if err := conn.Call(&ret, "personal_unlockAccount", from, "111111"); err != nil {
		log.Error("msg", "err", err)
	}
	fmt.Println("personal_unlockAccount", ret)

	//0xd3c21bcecceda0000000 100万
	//0x21e19e0c9bab2400000  1万
	//0x152d02c7e14af6000000 10万
	//0x43c33c1937564800000  20万
	// 0x3635c9adc5dea00000  1000
	//0x14542ba12a337c00000  6000
	for _, priv := range accounts {
		to := crypto.PubkeyToAddress(priv.PublicKey)
		params := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "eth_sendTransaction",
			"params": []interface{}{map[string]string{
				"from":  from,
				"to":    to.String(),
				"value": "0x14542ba12a337c00000",
			}},
			"id": 1,
		}
		bs, _ := json.Marshal(params)
		ret, err := Post("application/json", bytes.NewBuffer(bs))
		if err != nil {
			panic(err)
		}
		var r struct {
			JSONRPC string `json:"jsonrpc"`
			ID      int    `json:"id"`
			Result  common.Hash
		}
		if err := json.Unmarshal(ret, &r); err != nil {
			panic(err)
		}
		fmt.Println("sendRestlt", r.Result)
	}
	connEth, err := ethclient.Dial(url)
	time.Sleep(100 * time.Second)

	conn, err = rpc.Dial(url)
	if err != nil {
		log.Error("Failed to connect to the Atlaschain client: ", err)
	}

	target := getValidators(conn)
	l := int64(len(target))
	//randIndex := rand.Int63n(l)
	i := 0
	j := 0
	counter := 0
	for index, priv := range accounts {
		counter++
		i %= 5
		i++ //1 - 5

		j++
		j %= int(l) // j ==0

		to := target[j]
		VoteNum := i * 1000
		time.Sleep(time.Second)
		go createVoter(connEth, priv, big.NewInt(int64(index)), big.NewInt(int64(VoteNum)), to)
		from1 := crypto.PubkeyToAddress(priv.PublicKey)
		log.Info("Info ", "from1", from1.String(), "to", to.String(), "VoteNum", uint64(VoteNum))
		voter2validator = append(voter2validator, Voter2validatorInfo{from1.String(), to.String(), uint64(VoteNum)})
	}
	log.Info("WriteJson ", " voter2validator", len(voter2validator))
	fileutils.WriteJson(voter2validator, "./Voters2Validator.json")
	waitUntilMsgHandled(counter)
	log.Info("down")
}

func createVoter(client *ethclient.Client, privateKey *ecdsa.PrivateKey, Name *big.Int, VoteNum *big.Int, target common.Address) {
	from := crypto.PubkeyToAddress(privateKey.PublicKey)
	toAddress := mapprotocol.MustProxyAddressFor("Accounts")
	input := mapprotocol.PackInput(mapprotocol.AbiFor("Accounts"), "createAccount")
	commHash := sendContractTransaction(client, from, toAddress, nil, privateKey, input)
	getResult(client, commHash, false)
	input = mapprotocol.PackInput(mapprotocol.AbiFor("Accounts"), "setName", Name.String())
	commHash = sendContractTransaction(client, from, toAddress, nil, privateKey, input)
	getResult(client, commHash, false)
	input = mapprotocol.PackInput(mapprotocol.AbiFor("Accounts"), "setAccountDataEncryptionKey", crypto.FromECDSAPub(&privateKey.PublicKey))
	commHash = sendContractTransaction(client, from, toAddress, nil, privateKey, input)
	getResult(client, commHash, false)
	// lockedMA
	input = mapprotocol.PackInput(mapprotocol.AbiFor("LockedGold"), "lock")
	lockedGold := new(big.Int).Mul(VoteNum, big.NewInt(1e18))
	toAddress = mapprotocol.MustProxyAddressFor("LockedGold")
	commHash = sendContractTransaction(client, from, toAddress, lockedGold, privateKey, input)
	getResult(client, commHash, false)
	//vote
	g, l, _ := getGL(privateKey, client, target, lockedGold)
	toAddress = mapprotocol.MustProxyAddressFor("Election")
	input = mapprotocol.PackInput(mapprotocol.AbiFor("Election"), "vote", target, lockedGold, l, g)
	commHash = sendContractTransaction(client, from, toAddress, nil, privateKey, input)
	getResult(client, commHash, false)
	msgCh <- struct{}{}
}

func sendContractTransaction(client *ethclient.Client, from, toAddress common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, input []byte) common.Hash {
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

	msg := ethchain.CallMsg{From: from, To: &toAddress, GasPrice: gasPrice, Value: value, Data: input}
	gasLimit, err = client.EstimateGas(context.Background(), msg)
	if err != nil {
		logger.Error("Contract exec failed", "error", err)
	}
	if gasLimit < 1 {
		//gasLimit = 866328
		gasLimit = 2100000
	}
	gasLimit = uint64(DefaultGasLimit)

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
	i := 0
	for {
		time.Sleep(time.Millisecond * 200)
		i++
		_, isPending, err := conn.TransactionByHash(context.Background(), txHash)
		if err != nil {
			logger.Info("TransactionByHash", "error", err)
		}
		if !isPending {
			break
		}
		if i > 20 {
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
		//isContinueError = false
		logger.Info("Transaction Failed ", "Block Number", receipt.BlockNumber.Uint64())
	}
}
func Post(contentType string, body io.Reader) (result []byte, err error) {
	resp, err := http.Post(urlSendTx, contentType, body)
	if err != nil {
		return result, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("wanted 200 ,get %d", resp.StatusCode)
	}
	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	if result == nil {
		err = fmt.Errorf("result get nil")
		return
	}
	return result, err
}

//Send Transtion (Post Way)
type voteTotal struct {
	Validator common.Address
	Value     *big.Int
}

func getGL(privateKey *ecdsa.PrivateKey, client *ethclient.Client, target common.Address, VoteNum *big.Int) (common.Address, common.Address, error) {
	type ret struct {
		Validators interface{} // indexed
		Values     interface{}
	}
	var t ret
	electionAddress := mapprotocol.MustProxyAddressFor("Election")
	abiElection := mapprotocol.AbiFor("Election")

	from := crypto.PubkeyToAddress(privateKey.PublicKey)
	//vote
	input := mapprotocol.PackInput(mapprotocol.AbiFor("Election"), "getTotalVotesForEligibleValidators")
	sendContractTransaction(client, from, electionAddress, nil, privateKey, input)
	header, err := client.HeaderByNumber(context.Background(), nil)
	msg := ethchain.CallMsg{From: from, To: &electionAddress, Data: input}
	output, err := client.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Error("method CallContract error", "error", err)
	}
	err = abiElection.UnpackIntoInterface(&t, "getTotalVotesForEligibleValidators", output)
	if err != nil {
		log.Error("getTotalVotesForEligibleValidators setLesserGreater", "err", err)
	}

	validators := (t.Validators).([]common.Address)
	votes := (t.Values).([]*big.Int)
	voteTotals := make([]voteTotal, len(validators))
	for i, addr := range validators {
		voteTotals[i] = voteTotal{addr, votes[i]}
	}
	//fmt.Println(target, VoteNum)
	//for i, v := range voteTotals {
	//	fmt.Println("=== ", i, "===", v.Validator.String(), v.Value.String())
	//}

	for _, voteTotal := range voteTotals {
		if bytes.Equal(voteTotal.Validator.Bytes(), target.Bytes()) {
			if big.NewInt(0).Cmp(VoteNum) < 0 {
				voteTotal.Value.Add(voteTotal.Value, VoteNum)
			}
			// Sorting in descending order is necessary to match the order on-chain.
			// TODO: We could make this more efficient by only moving the newly vote member.
			sort.SliceStable(voteTotals, func(j, k int) bool {
				return voteTotals[j].Value.Cmp(voteTotals[k].Value) > 0
			})

			lesser := params.ZeroAddress
			greater := params.ZeroAddress
			for j, voteTotal := range voteTotals {
				if voteTotal.Validator == target {
					if j > 0 {
						greater = voteTotals[j-1].Validator
					}
					if j+1 < len(voteTotals) {
						lesser = voteTotals[j+1].Validator
					}
					break
				}
			}
			return greater, lesser, nil
			break
		}
	}
	return params.ZeroAddress, params.ZeroAddress, nil
}
func getValidators(conn *rpc.Client) []common.Address {
	var ret []common.Address
	if err := conn.Call(&ret, "istanbul_getValidators"); err != nil {
		log.Error("msg", "err", err)
	}
	return ret
}
func waitUntilMsgHandled(counter int) {
	log.Debug("waitUntilMsgHandled", "counter", counter)
	for counter > 0 {
		<-msgCh
		counter -= 1
	}
}
