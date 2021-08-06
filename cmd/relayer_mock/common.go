package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/json"
	"fmt"
	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/chains/headers/ethereum"
	"github.com/mapprotocol/atlas/cmd/ethclient"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
	params2 "github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"strings"
	"time"
)

var (
	epochHeight                       = params.NewEpochLength
	keystore1                         = "D:/BaiduNetdiskDownload/test015/atlas/data555/keystore/UTC--2021-07-09T06-27-06.967129500Z--c971f9cec4310cf001ca55078b43a568aaa0366d"
	keystore2                         = "D:/BaiduNetdiskDownload/test015/atlas/data555/keystore/UTC--2021-07-09T06-26-32.960000300Z--78c5285c42572677d3f9dcc27b9ac7b1ff49843c"
	keystore3                         = "D:/BaiduNetdiskDownload/test015/atlas/data555/keystore/UTC--2021-07-11T06-35-36.635750800Z--70bf8d9de50713101992649a4f0d7fa505ebb334"
	keystore4                         = "D:/BaiduNetdiskDownload/test015/atlas/data555/keystore/UTC--2021-07-19T11-51-51.704095400Z--4e0449459f73341f8e9339cb9e49dae3115ec80f"
	keystore5                         = "D:/BaiduNetdiskDownload/test015/atlas/data555/keystore/UTC--2021-07-21T10-26-12.236878500Z--8becddb5fbe6f3d6b08450e2d33e48e63d6c4b29"
	password                          = "123456"
	abiRelayer, _                     = abi.JSON(strings.NewReader(params2.RelayerABIJSON))
	abiHeaderStore, _                 = abi.JSON(strings.NewReader(params2.HeaderStoreABIJSON))
	RelayerAddress     common.Address = params2.RelayerAddress
	HeaderStoreAddress common.Address = params2.HeaderStoreAddress
)

const (
	BALANCE           = "balance"
	REGISTER_BALANCE  = "registerBalance"
	QUERY_RELAYERINFO = "relayerInfo"
	REWARD            = "reward"
	CHAINTYPE_HEIGHT  = "chainTypeHeight"
	NEXT_STEP         = "next step"

	AtlasRPCListenAddr = "localhost"
	AtlasRPCPortFlag   = 7445

	EthRPCListenAddr = "localhost"
	EthRPCPortFlag   = 8083
	ChainTypeETH     = chains.ChainTypeETH
	ChainTypeMAP     = chains.ChainTypeMAP

	// method name
	CurNbrAndHash = vm.CurNbrAndHash
)

type step []int // epoch height
type debugInfo struct {
	atlasBackendCh chan string
	notifyCh       chan uint64
	step           step
	ethData        []ethereum.Header
	client         *ethclient.Client
	relayerData    []*relayerInfo
}
type relayerInfo struct {
	url           string
	from          common.Address
	preBalance    *big.Float
	nowBalance    *big.Float
	registerValue int64
	priKey        *ecdsa.PrivateKey
	fee           uint64
}

func (r *relayerInfo) swapBalance() {
	f, _ := (r.nowBalance).Float64()
	r.preBalance = big.NewFloat(f)
}

func (r *relayerInfo) changeRegisterValue(value int64) {
	r.registerValue = value
}
func (d *debugInfo) changeAllRegisterValue(value int64) {
	for k, _ := range d.relayerData {
		d.relayerData[k].registerValue = value
	}
}

func (d *debugInfo) preWork(ctx *cli.Context, isRegister bool) {
	conn := getConn11(ctx)
	d.atlasBackendCh = make(chan string)
	d.notifyCh = make(chan uint64)
	d.client = conn

	d.ethData = getEthChains()
	number, err := conn.BlockNumber(context.Background())
	if err != nil {
		log.Fatal("get BlockNumber err ", err)
	}
	currentEpoch := number / epochHeight
	d.step = []int{int(currentEpoch + 1), int(currentEpoch + 2), int(currentEpoch + 3)}
	d.relayerData = append(d.relayerData, &relayerInfo{url: keystore1})
	for k, _ := range d.relayerData {
		Ele := d.relayerData[k]
		priKey, from := loadprivateCommon(Ele.url)
		var acc common.Address
		acc.SetBytes(from.Bytes())
		Ele.registerValue = registerValue
		Ele.from = acc
		Ele.priKey = priKey
		Ele.fee = uint64(0)
		bb := getBalance(conn, Ele.from)
		Ele.preBalance = bb
		Ele.nowBalance = bb
		if isRegister {
			register11(ctx, d.client, *d.relayerData[k])
		}
	}

}
func (d *debugInfo) queryDebuginfo(ss string) {
	conn := d.client
	switch ss {
	case BALANCE:
		for k, _ := range d.relayerData {
			fmt.Println("ADDRESS:", d.relayerData[k].from, " OLD BALANCE :", d.relayerData[k].preBalance, " NOW BALANCE :", getBalance(conn, d.relayerData[k].from))
		}
	case REGISTER_BALANCE:
		for k, _ := range d.relayerData {
			registered, unregistering, unregistered := getRegisterBalance(conn, d.relayerData[k].from)
			fmt.Println("ADDRESS:", d.relayerData[k].from,
				" NOW registerValue BALANCE :", registered, " registerING BALANCE :", unregistering, "registerED BALANCE :", unregistered)
		}
	case QUERY_RELAYERINFO:
		for k, _ := range d.relayerData {
			bool1, bool2, relayerEpoch, _ := queryRegisterInfo(conn, d.relayerData[k].from)
			fmt.Println("ADDRESS:", d.relayerData[k].from, "ISREGISTER:", bool1, " ISRELAYER :", bool2, " RELAYER_EPOCH :", relayerEpoch)
		}
	case REWARD:

	case CHAINTYPE_HEIGHT:
		for k, _ := range d.relayerData {
			currentTypeHeight := getCurrentNumberAbi(conn, ChainTypeETH, d.relayerData[k].from)
			fmt.Println("ADDRESS:", d.relayerData[k].from, " TYPE HEIGHT:", currentTypeHeight)
		}
	}

}
func (d *debugInfo) atlasBackend() {
	canNext := "YES"
	count := 0
	conn := d.client
	var target uint64 // 1 2 3...
	target = uint64(d.step[count]) - 1
	go func() {
		for {
			select {
			case <-d.atlasBackendCh:
				count++
				if count < len(d.step) {
					target = uint64(d.step[count]) - 1
				}

				canNext = "YES"
			}
		}
	}()

	for {
		number, err := conn.BlockNumber(context.Background())
		if err != nil {
			log.Fatal("get BlockNumber err ", err)
		}
		if canNext != "NO" {
			temp := int(number) - int(target*epochHeight)
			if temp > 0 {
				if int((target+1)*epochHeight) < int(number) {
					log.Fatal("Conditions can never be met")
				}
				d.notifyCh <- target + 1
				canNext = "NO"
				if count+1 == len(d.step) {
					return
				}
			}
		}
		time.Sleep(time.Second)
	}
}
func getEthChains() []ethereum.Header {
	Db, err := rawdb.NewLevelDBDatabase("relayerMockStore", 128, 1024, "", false)
	if err != nil {
		log.Fatal(err)
	}
	var key []byte
	key = []byte("ETH_INFO")
	var c []ethereum.Header
	jsonbyte, err := Db.Get(key)
	json.Unmarshal(jsonbyte, &c)
	if len(c) == 1000 {
		return c
	}
	Ethconn, _ := dialEthConn()
	Headers := getChainsCommon(Ethconn)

	rlp, err := json.Marshal(Headers)
	if err != nil {
		log.Fatal("Failed to Marshal block body", "err", err)
	}

	if err := Db.Put(key, rlp); err != nil {
		log.Fatal("Failed to store block body", "err", err)
	}
	return Headers
}
func getChainsCommon(conn *ethclient.Client) []ethereum.Header {
	startNum := 1
	endNum := 1000
	Headers := make([]ethereum.Header, 1000)
	HeaderBytes := make([]bytes.Buffer, 1000)
	for i := startNum; i <= endNum; i++ {
		Header, err := conn.HeaderByNumber(context.Background(), big.NewInt(int64(i)))
		if err != nil {
			log.Fatal(err)
		}
		convertChain(&Headers[i-1], &HeaderBytes[i-1], Header)
	}
	return Headers
}
func ethToWei(registerValue int64) *big.Int {
	baseUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	value := new(big.Int).Mul(big.NewInt(registerValue), baseUnit)
	return value
}

func weiToEth(value *big.Int) uint64 {
	baseUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	valueT := new(big.Int).Div(value, baseUnit).Uint64()
	return valueT
}
func printChangeBalance(old, new big.Float) {
	f, _ := old.Float64()
	old1 := big.NewFloat(f)
	f2, _ := new.Float64()
	new1 := big.NewFloat(f2)
	f3, _ := old1.Float64()
	c := big.NewFloat(f3)
	fmt.Printf("old balance:%v  new balance %v  change %v\n",
		old1, new1, c.Abs(c.Sub(c, new1)))
}
func getBalance(conn *ethclient.Client, address common.Address) *big.Float {
	balance, err := conn.BalanceAt(context.Background(), address, nil)
	if err != nil {
		log.Fatal(err)
	}
	balance2 := new(big.Float)
	balance2.SetString(balance.String())
	Value := new(big.Float).Quo(balance2, big.NewFloat(math.Pow10(18)))
	return Value
}
func getRegisterBalance(conn *ethclient.Client, from common.Address) (uint64, uint64, uint64) {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	input := packInput("getRelayerBalance", from)
	msg := ethchain.CallMsg{From: from, To: &RelayerAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Fatal("method CallContract error", err)
	}
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
		return weiToEth(args.registered), weiToEth(args.unregistering), weiToEth(args.unregistered)
	}
	log.Fatal("Contract query failed result len == 0")
	return 0, 0, 0
}
func dialEthConn() (*ethclient.Client, string) {
	ip = EthRPCListenAddr //utils.RPCListenAddrFlag.Name)
	port = EthRPCPortFlag //utils.RPCPortFlag.Name)
	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	conn, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the AtlasChain client: %v", err)
	}
	return conn, url
}
func register11(ctx *cli.Context, conn *ethclient.Client, info relayerInfo) {
	value := ethToWei(info.registerValue)
	if info.registerValue < RegisterAmount {
		log.Fatal("Amount must bigger than ", RegisterAmount)
	}
	fee := ctx.GlobalUint64(FeeFlag.Name)
	checkFee(new(big.Int).SetUint64(fee))
	input := packInput("register", value)
	sendContractTransaction(conn, info.from, RelayerAddress, nil, info.priKey, input)
}
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
func convertChain(header *ethereum.Header, headerbyte *bytes.Buffer, e *types.Header) (*ethereum.Header, *bytes.Buffer) {
	if header == nil || e == nil {
		fmt.Println("header:", header, "e:", e)
		return header, headerbyte
	}
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
	header.Nonce = types.EncodeNonce(e.Nonce.Uint64())
	header.Bloom.SetBytes(e.Bloom.Bytes())
	if header.Difficulty = new(big.Int); e.Difficulty != nil {
		header.Difficulty.Set(e.Difficulty)
	}
	if header.Number = new(big.Int); e.Number != nil {
		header.Number.Set(e.Number)
	}
	if len(e.Extra) > 0 {
		header.Extra = make([]byte, len(e.Extra))
		copy(header.Extra, e.Extra)
	}
	binary.Write(headerbyte, binary.BigEndian, header)
	return header, headerbyte
}
func registerCommon(conn *ethclient.Client, keystore string) (*big.Float, common.Address) {
	fee := uint64(0)
	value := ethToWei(100000)
	priKey, from := loadprivateCommon(keystore)

	pkey, pk, _ := getPubKey(priKey)
	aBalance := getBalance(conn, from)
	fmt.Printf("Fee: %v \nPub key:%v\nvalue:%v\n \n", fee, pkey, value)
	input := packInput("register", pk, new(big.Int).SetUint64(fee), value)
	sendContractTransaction(conn, from, RelayerAddress, nil, priKey, input)
	return aBalance, from
}

func loadprivateCommon(keyfile string) (*ecdsa.PrivateKey, common.Address) {
	keyjson, err := ioutil.ReadFile(keyfile)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to read the keyfile at '%s': %v", keyfile, err))
	}
	password = "123456"
	if keyfile == keystore2 {
		password = ""
	}
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		log.Fatal(fmt.Errorf("error decrypting key: %v", err))
	}
	priKey1 := key.PrivateKey
	return priKey1, crypto.PubkeyToAddress(priKey1.PublicKey)
}
