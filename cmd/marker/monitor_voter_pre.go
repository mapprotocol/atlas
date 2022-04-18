package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/cmd/marker/config"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"math/big"
	"os"
)

// voter key
type VoterStruct struct {
	Voter     common.Address
	Validator common.Address
}

// voter value
type VoterInfo struct {
	epochNum  uint64
	VActive   *big.Int
	VPending  *big.Int
	Voter     common.Address
	Validator common.Address
}

//  validator value
type ValidatorInfo struct {
	EpochNum        uint64
	AllVotes        *big.Int
	ValidatorReward *big.Int
}

var voterMap map[VoterStruct]*VoterInfo

var voterList []VoterStruct

var validatorMap map[common.Address]*ValidatorInfo
var writeChan chan []string
var xlsFile *os.File

type Voter2validatorInfo struct {
	VoterAccount     string
	ValidatorAccount string
	Value            uint64
}

var voter2validator []Voter2validatorInfo

func start(ctx *cli.Context, core *listener) error {

	writeChan = make(chan []string)
	xlsFile1, err := initCsv()
	xlsFile = xlsFile1
	defer xlsFile.Close()
	if err != nil {
		panic("initCsv")
	}
	//voterList = []VoterStruct{
	//	{common.HexToAddress("0x81f02fd21657df80783755874a92c996749777bf"), common.HexToAddress("0x1c0eDab88dbb72B119039c4d14b1663525b3aC15")},
	//	{common.HexToAddress("0xdf945e6ffd840ed5787d367708307bd1fa3d40f4"), common.HexToAddress("0x16FdBcAC4D4Cc24DCa47B9b80f58155a551ca2aF")},
	//}

	configName := "D:\\work\\zhangwei812\\atlas\\zw_config\\Voters2Validator.json"

	data, err := ioutil.ReadFile(configName)
	if err != nil {
		log.Error("compass personInfo config readFile Err", err.Error())
	}
	_ = json.Unmarshal(data, &voter2validator)
	for _, v := range voter2validator {
		voterList = append(voterList, VoterStruct{common.HexToAddress(v.VoterAccount), common.HexToAddress(v.ValidatorAccount)})
	}

	validatorMap = map[common.Address]*ValidatorInfo{}
	for _, v := range voterList {
		validatorMap[v.Validator] = &ValidatorInfo{0, big.NewInt(0), big.NewInt(0)}
	}
	voterMap = map[VoterStruct]*VoterInfo{}
	for _, v := range voterList {
		voterMap[v] = &VoterInfo{0, big.NewInt(0), big.NewInt(0), v.Voter, v.Validator}
		core := NewListener(ctx, core.cfg)
		writer := NewWriter(ctx, core.cfg)
		core.setWriter(writer)
		go pollBlocks(ctx, core, v, voterMap[v])
	}
	select {}
	return nil
}

type Voters_my struct {
	Account string
	Path    string
}

var voters []Voters_my
var validators []common.Address

//vote Automatic
func voteAutomatic(ctx *cli.Context, coreA *listener) error {
	configName := "D:\\work\\zhangwei812\\atlas\\zw_config\\Voters.json"

	data, err := ioutil.ReadFile(configName)
	if err != nil {
		log.Error("compass personInfo config readFile Err", err.Error())
	}
	_ = json.Unmarshal(data, &voters)
	log.Info("voters length:", "l", len(voters))

	validatorsName := "D:\\work\\zhangwei812\\atlas\\zw_config\\Validators.json"

	data, err = ioutil.ReadFile(validatorsName)
	if err != nil {
		log.Error("compass personInfo config readFile Err", err.Error())
	}
	_ = json.Unmarshal(data, &validators)
	log.Info("validators length:", "l", len(validators))
	LenValidator := len(validators)
	i := 0
	for index, v := range voters {
		_config, err := config.AssemblyConfig2(ctx, v.Path, "111111")
		i %= 5
		i++ //1 - 5
		VoteNum := i * 1000
		_config.LockedNum = big.NewInt(int64(VoteNum))
		TargetValidatorIndex := index % LenValidator
		_config.VoteNum = big.NewInt(int64(VoteNum))
		_config.TargetAddress = validators[TargetValidatorIndex]
		_config.From = common.HexToAddress(v.Account)
		if err != nil {
			return err
		}
		coreA.cfg = _config
		core := NewListener(ctx, _config)
		writer := NewWriter(ctx, _config)
		core.setWriter(writer)
		quicklyVote(ctx, core)
		voter2validator = append(voter2validator, Voter2validatorInfo{core.cfg.From.String(), core.cfg.TargetAddress.String(), core.cfg.VoteNum.Uint64()})
		log.Info("vote ", "Index", index, "From", coreA.cfg.From, "TargetAddress", coreA.cfg.TargetAddress)
	}
	log.Info("WriteJson  voter2validator", len(voter2validator))
	WriteJson(voter2validator, "D:\\work\\zhangwei812\\atlas\\zw_config\\Voters2Validator.json")
	return nil
}

func pre(ctx *cli.Context, core *listener) error {
	start(ctx, core)
	return nil
}

// request validators from remote
func getValidatorsJson(ctx *cli.Context, core *listener) error {
	validators := core.getValidators()
	WriteJson(validators, "D:\\work\\zhangwei812\\atlas\\zw_config\\Validators.json")
	fmt.Println("getValidatorsJson", len(validators))
	return nil
}

func WriteJson(in interface{}, filepath string) error {
	byteValue, err := json.MarshalIndent(in, " ", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath, byteValue, 0644)
}
