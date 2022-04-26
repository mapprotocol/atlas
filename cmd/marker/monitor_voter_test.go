package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/mapprotocol/atlas/helper/fileutils"
	exec "golang.org/x/sys/execabs"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"testing"
)

//creat new keystore at remote
func Test_NewAccount(t *testing.T) {
	var voterAccounts []common.Address
	url := "http://13.67.118.60:7445" //测试网 validator voter
	conn, err := rpc.Dial(url)
	if err != nil {
		t.Error("Failed to connect to the Atlaschain client: ", err)
	}
	for i := 0; i < 100; i++ {
		var ret common.Address
		if err := conn.Call(&ret, "personal_newAccount", "111111"); err != nil {
			log.Error("msg", "err", err)
		}
		voterAccounts = append(voterAccounts, ret)
	}
	err = fileutils.WriteJson(voterAccounts, "./VoterAccounts")
	if err != nil {
		t.Error("WriteJson err: ", err)
	}
}

type VoterInfoT struct {
	Account common.Address
	PrivHex string
}

var voterAccounts []VoterInfoT

// creat new address use crypto.GenerateKey()
func Test_NewAccount2(t *testing.T) {
	priv, _ := crypto.GenerateKey()
	privHex := hex.EncodeToString(crypto.FromECDSA(priv))
	fmt.Println(privHex)
	addr := crypto.PubkeyToAddress(priv.PublicKey)
	fmt.Println(addr.String())
	voterAccounts = append(voterAccounts, VoterInfoT{addr, privHex})
	fmt.Println(voterAccounts)
	err := fileutils.WriteJson(voterAccounts, "./VoterAccounts")
	if err != nil {
		t.Error("WriteJson err: ", err)
	}
}

//creat new address use exec.Command()
func Test_newAccount3(t *testing.T) {
	cmd := exec.Command("D:/root/atlas", "account", "new")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("combined out:\n%s\n", string(out))
		log.Info("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("combined out:\n%s\n", string(out))
}

func Post(url, contentType string, body io.Reader) (result []byte, err error) {
	resp, err := http.Post(url, contentType, body)
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
func Test_SendTranstion(t *testing.T) {
	//url := "http://localhost:8545"
	url := "http://13.67.118.60:7445"
	conn, err := rpc.Dial(url)
	if err != nil {
		t.Error("Failed to connect to the Atlaschain client: ", err)
	}
	//from := "0x81f02fd21657df80783755874a92c996749777bf"
	from := "0xbe27cf1ed3489b6add51a22ce4b25abd92cac3c8"
	var ret bool
	if err := conn.Call(&ret, "personal_unlockAccount", from, ""); err != nil {
		log.Error("msg", "err", err)
	}
	fmt.Println("personal_unlockAccount", ret)

	data, err := ioutil.ReadFile("D:\\work\\zhangwei812\\atlas\\zw_config\\Voters.json")
	if err != nil {
		log.Crit("compass personInfo config readFile Err", err.Error())
	}
	type Info struct {
		Account string
		Path    string
	}
	var accounts []Info
	_ = json.Unmarshal(data, &accounts)
	//0xd3c21bcecceda0000000 100万
	//0x21e19e0c9bab2400000  1万
	//0x152d02c7e14af6000000 10万
	//0x43c33c1937564800000  20万
	for index, v := range accounts {
		to := v.Account
		fmt.Println("from  ", from, "to    ", to)
		params := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "eth_sendTransaction",
			"params": []interface{}{map[string]string{
				"from":  from,
				"to":    to,
				"value": "0x43c33c1937564800000",
			}},
			"id": 1,
		}
		bs, _ := json.Marshal(params)
		ret, err := Post(url, "application/json", bytes.NewBuffer(bs))
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
		fmt.Println(r.Result, "   ", index)
	}
}

// test write to .csv file
func Test_writeInfo(t *testing.T) {
	xlsFile1, _ := initCsv()
	xlsFile = xlsFile1
	writeChan = make(chan []string)
	writeInfo(0, "1", "1", "1", big.NewInt(1234561111111111111), big.NewInt(1234561111111111111), big.NewInt(1234561111111111111), "1", big.NewFloat(1234561111111111111), big.NewFloat(1234561111111111111), "1")
}

func Test_f(t *testing.T) {
	f := new(big.Float).SetInt(big.NewInt(50))
	fSub := new(big.Float).SetInt(big.NewInt(150))
	f1 := new(big.Float).SetInt(big.NewInt(100))
	f.Quo(f, fSub)
	f.Mul(f, f1)
	fs := f.String()
	fs += "%"
	fmt.Println(fs)
}
