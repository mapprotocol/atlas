package genesis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/core/types"
	blscrypto "github.com/mapprotocol/atlas/helper/bls"
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

var ValidatorsAT []AccoutInfo

var AdminAddr common.Address

// CreateCommonGenesisConfig generates a config starting point which templates can then customize further
func CreateCommonGenesisConfig() *Config {
	genesisConfig := BaseConfig()
	genesisConfig.ChainID = params.MainnetChainConfig.ChainID
	genesisConfig.GenesisTimestamp = uint64(time.Now().Unix())
	genesisConfig.Istanbul = *params.MainnetChainConfig.Istanbul
	genesisConfig.Hardforks = HardforkConfig{
		ChurritoBlock: common.Big0,
		DonutBlock:    common.Big0,
	}

	return genesisConfig
}

// GenerateGenesis will create a new genesis block with full atlas blockchain already configured
func GenerateGenesis(_ *cli.Context, accounts *env.AccountsConfig, cfg *Config, contractsBuildPath string) (*chain.Genesis, error) {
	extraData, err := generateGenesisExtraData(ValidatorsAT)
	if err != nil {
		return nil, err
	}
	genesisAlloc, err := generateGenesisState(accounts, cfg, contractsBuildPath)
	if err != nil {
		return nil, err
	}
	genesis := chain.UseForGenesisBlock()
	genesis.ExtraData = extraData
	genesis.Alloc = genesisAlloc
	return genesis, nil
}

func generateGenesisExtraData(validatorAccounts []AccoutInfo) ([]byte, error) {
	addresses := make([]common.Address, len(validatorAccounts))
	blsKeys := make([]blscrypto.SerializedPublicKey, len(validatorAccounts))
	blsG1Keys := make([]blscrypto.SerializedG1PublicKey, len(validatorAccounts))

	for i := 0; i < len(validatorAccounts); i++ {
		var err error
		addresses[i] = validatorAccounts[i].getAddress()
		blsKeys[i], err = validatorAccounts[i].BLSPublicKey()
		blsG1Keys[i], err = validatorAccounts[i].BLSG1PublicKey()
		if err != nil {
			return nil, err
		}
	}

	istExtra := types.IstanbulExtra{
		AddedValidators:             addresses,
		AddedValidatorsPublicKeys:   blsKeys,
		AddedValidatorsG1PublicKeys: blsG1Keys,
		RemovedValidators:           big.NewInt(0),
		Seal:                        []byte{},
		AggregatedSeal:              types.IstanbulAggregatedSeal{},
		ParentAggregatedSeal:        types.IstanbulAggregatedSeal{},
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

//From markerConfig.json used for validators and election contract
type AccoutInfo struct {
	Address              string
	PublicKeyHex         string
	BLSPubKey            string
	BLSG1PubKey          string
	BLSProofOfPossession string
	WalletAddress        string
}

// MustBLSProofOfPossession variant of BLSProofOfPossession that panics on error
func (a *AccoutInfo) MustBLSProofOfPossession() []byte {
	pop, err := a.BLSProofOfPossession_()
	if err != nil {
		panic(err)
	}
	return pop
}
func (a *AccoutInfo) getAddress() common.Address {
	return common.HexToAddress(a.Address)
}

// BLSProofOfPossession generates bls proof of possession
func (a *AccoutInfo) BLSProofOfPossession_() ([]byte, error) {
	b, err := hexutil.Decode(a.BLSProofOfPossession)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// BLSPublicKey returns the bls public key
func (a *AccoutInfo) BLSG1PublicKey() (blscrypto.SerializedG1PublicKey, error) {
	var b blscrypto.SerializedG1PublicKey
	err := b.UnmarshalText([]byte(a.BLSG1PubKey))
	if err != nil {
		return blscrypto.SerializedG1PublicKey{}, err
	}
	return b, nil
}

// BLSPublicKey returns the bls public key
func (a *AccoutInfo) BLSPublicKey() (blscrypto.SerializedPublicKey, error) {
	var b blscrypto.SerializedPublicKey
	err := b.UnmarshalText([]byte(a.BLSPubKey))
	if err != nil {
		return blscrypto.SerializedPublicKey{}, err
	}
	return b, nil
}

// PublicKeyHex hex representation of the public key
func (a *AccoutInfo) PublicKey() []byte {
	b, err := hexutil.Decode(a.PublicKeyHex)
	if err != nil {
		panic(err)
	}
	return b
}

func (a *AccoutInfo) WalletAddress_() common.Address {
	return common.HexToAddress(a.WalletAddress)
}

type MarkerInfo struct {
	AdminAddress string
	Validators   []AccoutInfo
}

func UnmarshalMarkerConfig(ctx *cli.Context) {
	keyDir := fmt.Sprintf("../atlas/marker/config/markerConfig.json")
	if ctx.IsSet("markercfg") {
		keyDir = ctx.String("markercfg")
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
	ValidatorsAT = tt
	AdminAddr = common.HexToAddress(markerCfg.AdminAddress)
}
