package main

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"math/big"
	"os"
	"time"
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

func voterMonitor(ctx *cli.Context, core *listener) error {
	writeChan = make(chan []string)
	xlsFile1, err := initCsv()
	xlsFile = xlsFile1
	defer xlsFile.Close()
	if err != nil {
		panic("initCsv")
	}

	configName := "./Voters2Validator.json"

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
		time.Sleep(3 * time.Second)
		go pollBlocks(ctx, core, v, voterMap[v])
	}
	select {}
	return nil
}
