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
var AdminAT env.Account

// CreateCommonGenesisConfig generates a config starting point which templates can then customize further
func CreateCommonGenesisConfig(chainID *big.Int, adminAccountAddress common.Address, istanbulConfig params.IstanbulConfig) *Config {
	genesisConfig := BaseConfig()
	genesisConfig.ChainID = chainID
	genesisConfig.GenesisTimestamp = uint64(time.Now().Unix())
	genesisConfig.Istanbul = istanbulConfig
	genesisConfig.Hardforks = HardforkConfig{
		ChurritoBlock: common.Big0,
		DonutBlock:    common.Big0,
	}

	// Make admin account manager of Governance & Reserve
	adminMultisig := MultiSigParameters{
		Signatories:                      []common.Address{adminAccountAddress},
		NumRequiredConfirmations:         1,
		NumInternalRequiredConfirmations: 1,
	}

	genesisConfig.ReserveSpenderMultiSig = adminMultisig
	genesisConfig.GovernanceApproverMultiSig = adminMultisig

	// Ensure nothing is frozen
	genesisConfig.GoldToken.Frozen = false
	genesisConfig.EpochRewards.Frozen = false

	return genesisConfig
}

func FundAccounts(genesisConfig *Config, accounts []env.Account) {
	cusdBalances := make([]Balance, len(accounts))
	ceurBalances := make([]Balance, len(accounts))
	goldBalances := make([]Balance, len(accounts))
	for i, acc := range accounts {
		cusdBalances[i] = Balance{Account: acc.Address, Amount: (*big.Int)(token.MustNew("50000"))} // 50k cUSD
		ceurBalances[i] = Balance{Account: acc.Address, Amount: (*big.Int)(token.MustNew("50000"))} // 50k cEUR
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
		Coinbase:  AdminAT.Address,
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
	AdminInfo  AccoutInfo
	Validators []AccoutInfo
}

func UnmarshalMarkerConfig() {
	keyDir := fmt.Sprintf("../atlas/marker/config/markerConfig.json")
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
	AdminAT = loadPrivate([]AccoutInfo{markerCfg.AdminInfo})[0]
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
