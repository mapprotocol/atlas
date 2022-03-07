package genesis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/accounts/keystore"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/core/types"
	blscrypto "github.com/mapprotocol/atlas/helper/bls"
	"github.com/mapprotocol/atlas/helper/decimal/token"
	"github.com/mapprotocol/atlas/marker/env"
	"github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"math/big"
	"time"
)

// Keccak256 of "The Times 09/Apr/2020 With $2.3 Trillion Injection, Fed’s Plan Far Exceeds Its 2008 Rescue"
var genesisMsgHash = bytes.Repeat([]byte{0x00}, 32)

//var genesisMsgHash = common.HexToHash("ecc833a7747eaa8327335e8e0c6b6d8aa3a38d0063591e43ce116ccf5c89753e")

var ValidatorsAT []env.Account
var AdminAddr common.Address

// CreateCommonGenesisConfig generates a config starting point which templates can then customize further
func CreateCommonGenesisConfig(chainID *big.Int, istanbulConfig params.IstanbulConfig) *Config {
	genesisConfig := BaseConfig()
	genesisConfig.ChainID = chainID
	genesisConfig.GenesisTimestamp = uint64(time.Now().Unix())
	genesisConfig.Istanbul = istanbulConfig
	genesisConfig.Hardforks = HardforkConfig{
		ChurritoBlock: common.Big0,
		DonutBlock:    common.Big0,
	}

	return genesisConfig
}

func FundAccounts(genesisConfig *Config, accounts []env.Account) {
	goldBalances := make([]Balance, len(accounts))
	for i, acc := range accounts {
		goldBalances[i] = Balance{Account: acc.Address, Amount: (*big.Int)(token.MustNew("50000"))} // 50k Atlas
	}
	genesisConfig.GoldToken.InitialBalances = goldBalances
}

// GenerateGenesis will create a new genesis block with full atlas blockchain already configured
func GenerateGenesis(ctx *cli.Context, accounts *env.AccountsConfig, cfg *Config, contractsBuildPath string) (*chain.Genesis, error) {
	extraData, err := generateGenesisExtraData(ValidatorsAT)
	//fmt.Println("extraData: ", hexutil.Encode(extraData))
	if err != nil {
		return nil, err
	}

	genesisAlloc, err := generateGenesisState(accounts, cfg, contractsBuildPath)
	if err != nil {
		return nil, err
	}

	return &chain.Genesis{
		Config:    cfg.ChainConfig(),
		ExtraData: extraData,
		Coinbase:  AdminAddr,
		Timestamp: cfg.GenesisTimestamp,
		Alloc:     genesisAlloc,
	}, nil
}

func generateGenesisExtraData(validatorAccounts []env.Account) ([]byte, error) {
	addresses := make([]common.Address, len(validatorAccounts))
	blsKeys := make([]blscrypto.SerializedPublicKey, len(validatorAccounts))

	for i := 0; i < len(validatorAccounts); i++ {
		var err error
		addresses[i] = validatorAccounts[i].Address
		blsKeys[i], err = validatorAccounts[i].BLSPublicKey()
		if err != nil {
			return nil, err
		}
	}

	istExtra := types.IstanbulExtra{
		AddedValidators:           addresses,
		AddedValidatorsPublicKeys: blsKeys,
		RemovedValidators:         big.NewInt(0),
		Seal:                      []byte{},
		AggregatedSeal:            types.IstanbulAggregatedSeal{},
		ParentAggregatedSeal:      types.IstanbulAggregatedSeal{},
	}

	payload, err := rlp.EncodeToBytes(&istExtra)
	if err != nil {
		return nil, err
	}

	var extraBytes []byte
	extraBytes = append(extraBytes, genesisMsgHash...)
	extraBytes = append(extraBytes, payload...)

	return extraBytes, nil
}

////////////////////////////////////////////////////////////////////////

type AccoutInfo struct {
	Account  string
	Password string
}

type MarkerInfo struct {
	AdminAddress string
	Validators   []AccoutInfo
}

func UnmarshalMarkerConfig(ctx *cli.Context) {
	keyDir := fmt.Sprintf("../atlas/marker/config/markerConfig.json")
	if ctx.IsSet("markerCfg") {
		keyDir = ctx.String("markerCfg")
	}

	data, err := ioutil.ReadFile(keyDir)
	if err != nil {
		log.Crit(" readFile Err:", "err:", err.Error())
	}

	markerCfg := &MarkerInfo{}
	_ = json.Unmarshal(data, markerCfg)

	var tt []AccoutInfo
	for _, v := range (*markerCfg).Validators {
		tt = append(tt, v)
	}
	ValidatorsAT = loadPrivate(tt)
	AdminAddr = common.HexToAddress(markerCfg.AdminAddress)
}
func loadPrivate(paths []AccoutInfo) []env.Account {
	num := len(paths)
	accounts := make([]env.Account, num)
	for i, v := range paths {
		path := v.Account
		keyjson, err := ioutil.ReadFile(path)
		if err != nil {
			log.Error("loadPrivate ReadFile", fmt.Errorf("failed to read the keyfile at '%s': %v", path, err))
		}
		key, err := keystore.DecryptKey(keyjson, v.Password)
		if err != nil {
			log.Error("loadPrivate DecryptKey", fmt.Errorf("error decrypting key: %v", err))
		}
		priKey1 := key.PrivateKey
		publicAddr := crypto.PubkeyToAddress(priKey1.PublicKey)
		var addr common.Address
		addr.SetBytes(publicAddr.Bytes())
		accounts[i] = env.Account{
			Address:    addr,
			PrivateKey: priKey1,
		}
	}
	return accounts
}
