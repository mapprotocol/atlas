package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/cmd/marker/connections"
	"github.com/mapprotocol/atlas/cmd/marker/mapprotocol"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/params"
	"os"
	"strconv"

	//"strconv"

	"gopkg.in/urfave/cli.v1"
	"math/big"
	"time"
)

var (
	baseUnit  = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	fbaseUnit = new(big.Float).SetFloat64(float64(baseUnit.Int64()))
)

func (c *listener) LatestBlock() (*big.Int, error) {
	bnum, err := c.conn.BlockNumber(context.Background())
	if err != nil {
		return nil, err
	}
	return big.NewInt(0).SetUint64(bnum), nil
}

var epochNum uint64

func pollBlocks(ctx *cli.Context, core *listener, key VoterStruct, voterInfo *VoterInfo) error {
	voterInfo.VActive = big.NewInt(0)
	voterInfo.VPending = big.NewInt(0)
	vaInfo := validatorMap[voterInfo.Validator]
	vaInfo.AllVotes = big.NewInt(0)
	From := key.Voter
	TargetAddress := key.Validator
	var epochSize = params.Epoch
	var currentEpoch uint64 = 0
	var rate = 5 * time.Second
	var f *big.Float
	var validatorList []common.Address
	var latestBlock *big.Int
	//var epochLast uint64
	latestBlock, err := core.LatestBlock()
	epochNum = istanbul.GetEpochNumber(latestBlock.Uint64(), epochSize)
	voterInfo.epochNum = epochNum
	query := func(title string) bool {
		latestBlock, err = core.LatestBlock()
		if err != nil {
			panic(err)
		}
		log.Info(title, "Epoch", epochNum, "blockNumber", latestBlock, "YourAccount", From, "Validator", TargetAddress, "sign", key.Voter.String()+" to "+key.Validator.String())
		// 查询当前投给validator情况
		getActiveVotesForValidatorByAccount_(ctx, core, key, voterInfo)
		getPendingInfo_(ctx, core, key, voterInfo)
		validatorList = core.getValidators()
		exist := false
		for _, e := range validatorList {
			if e == TargetAddress {
				exist = true
			}
		}
		if !exist {
			log.Info(title+" the target validator Not selected ", "current validators", validatorList, "sign", key.Voter.String()+" to "+key.Validator.String())
		}
		//epochLast = istanbul.GetEpochLastBlockNumber(epochNum, epochSize)
		if !exist {
			log.Info(title+"The validator not be Selected ", "validator ", TargetAddress, "sign", key.Voter.String()+" to "+key.Validator.String())
		}

		// 查询validator情况
		getActiveVotesForValidator_(ctx, core, key, voterInfo)
		f = new(big.Float).SetInt(voterInfo.VActive)
		fSub := new(big.Float).SetInt(vaInfo.AllVotes)
		if vaInfo.AllVotes.CmpAbs(big.NewInt(0)) > 0 {
			f.Quo(f, fSub)
		} else {
			f = new(big.Float).SetUint64(0)
		}
		log.Info(title, "fraction", f, "sign", key.Voter.String()+" to "+key.Validator.String())
		return true
	}
	query("=== init ===")
	log.Info("wait for reward...", "sign", key.Voter.String()+" to "+key.Validator.String())
	for {
		select {
		default:
			time.Sleep(time.Second)
			latestBlock, err = core.LatestBlock()
			epochNum = istanbul.GetEpochNumber(latestBlock.Uint64(), epochSize)
			if currentEpoch < epochNum && istanbul.GetNumberWithinEpoch(latestBlock.Uint64(), epochSize) == epochSize-1 {
				currentEpoch = epochNum
				query("=== start ====")
			}
			if istanbul.IsLastBlockOfEpoch(latestBlock.Uint64(), epochSize) {
				log.Info("=== Reward ===")
				// getVoterRewardInfo
				getVoterRewardInfo_(latestBlock.Uint64(), core, key, voterInfo)
				// calculate reward
				calcuR := new(big.Float).SetInt(vaInfo.ValidatorReward)
				calcuR.Mul(calcuR, f)
				nextAVote := new(big.Float).SetInt(voterInfo.VActive)
				nextAVote.Add(nextAVote, calcuR)

				log.Info("=== Reward ===", "Target_Reward", calcuR, "Target_Active_Vote", nextAVote, "sign", key.Voter.String()+" to "+key.Validator.String())
				f1 := new(big.Float).SetInt(big.NewInt(100))
				f.Mul(f, f1)
				//{"Epoch", "BlockNumber", "voter", "validator", "vote", "validatorReward", "targetReward", "target"}
				writeInfo(epochNum, latestBlock.String(), From.String(), TargetAddress.String(), voterInfo.VPending, voterInfo.VActive, vaInfo.ValidatorReward, f.String(), calcuR, nextAVote, fmt.Sprintf("%v", validatorList))

				query("=== Result ===")

				fmt.Println()
				fmt.Println()
				time.Sleep(2 * rate)
			}

		}
	}
	return nil
}

func getActiveVotesForValidatorByAccount_(_ *cli.Context, core *listener, key VoterStruct, voterInfo *VoterInfo) {
	From := key.Voter
	TargetAddress := key.Validator
	var ret interface{}
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getActiveVotesForValidatorByAccount", TargetAddress, From)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	log.Info("", "Active Vote", ret.(*big.Int))
	voterInfo.VActive = ret.(*big.Int)
}

func getPendingInfo_(_ *cli.Context, core *listener, key VoterStruct, voterInfo *VoterInfo) (*big.Int, *big.Int) {

	From := key.Voter
	TargetAddress := key.Validator
	type ret []interface{}
	var Value interface{}
	var Epoch interface{}
	t := ret{&Value, &Epoch}

	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	f := func(output []byte) {
		err := abiElection.UnpackIntoInterface(&t, "pendingInfo", output)
		if err != nil {
			isContinueError = false
			log.Error("getPendingInfoForValidator", "err", err)
		}

	}
	m := NewMessageRet2(SolveQueryResult4, core.msgCh, core.cfg, f, ElectionAddress, nil, abiElection, "pendingInfo", From, TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	p := Epoch.(*big.Int)
	log.Info("", "PendingVotes", Value.(*big.Int), "epoch", p.Add(p, big.NewInt(1)), "sign", key.Voter.String()+" to "+key.Validator.String())
	voterInfo.VPending = big.NewInt(0)
	if Epoch.(*big.Int).CmpAbs(big.NewInt(0).SetUint64(epochNum)) < 0 {
		voterInfo.VPending = Value.(*big.Int)
	}
	return Epoch.(*big.Int), Value.(*big.Int)
}

func getActiveVotesForValidator_(_ *cli.Context, core *listener, key VoterStruct, voterInfo *VoterInfo) {
	TargetAddress := key.Validator
	valiInfo := validatorMap[voterInfo.Validator]
	var ret interface{}
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getActiveVotesForValidator", TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	log.Info("", "Validator all Votes", ret.(*big.Int), "sign", key.Voter.String()+" to "+key.Validator.String())
	valiInfo.AllVotes = ret.(*big.Int)
}

func getVoterRewardInfo_(curBlockNumber uint64, core *listener, key VoterStruct, info *VoterInfo) {
	TargetAddress := key.Validator
	epochSize := params.Epoch
	epochNum := istanbul.GetEpochNumber(curBlockNumber, epochSize)
	EpochLast := big.NewInt(0).SetUint64(istanbul.GetEpochLastBlockNumber(epochNum, epochSize))
	Epoch := istanbul.GetEpochNumber(curBlockNumber, epochSize)
	electionContractAddress := core.cfg.ElectionParameters.ElectionAddress
	query := mapprotocol.BuildQuery(electionContractAddress, mapprotocol.EpochRewardsDistributedToVoters, EpochLast, EpochLast)
	// querying for logs
	logs, err := core.conn.FilterLogs(context.Background(), query)
	if err != nil {
		log.Error("getVoterRewardInfo_FilterLogs ", "err", err)
	}
	for _, l := range logs {
		validator := common.BytesToAddress(l.Topics[1].Bytes())
		if validator == TargetAddress {
			reward := big.NewInt(0).SetBytes(l.Data[:32])
			log.Info("=== Reward ===", "Epoch", Epoch, "blockNumber", curBlockNumber, "Reward to Voters", reward, "sign", key.Voter.String()+" to "+key.Validator.String())
			validatorMap[info.Validator].ValidatorReward = reward
		}
	}
}

func initCsv() (*os.File, error) {
	strTime := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("VoterInfo_%s.csv", strTime)
	xlsFile, fErr := os.OpenFile("./"+filename, os.O_RDWR|os.O_CREATE, 0766)
	if fErr != nil {
		fmt.Println("Export:created excel file failed ==", fErr)
		return nil, fErr
	}
	xlsFile.WriteString("\xEF\xBB\xBF")
	wStr := csv.NewWriter(xlsFile)
	wStr.Write([]string{"Epoch", "BlockNumber", "Voter", "Validator", "PendingVote", "ActiveVote", "ValidatorReward", "Proportion", "CalculateReward", "NextActiveVote", "Validators"})
	wStr.Flush()
	return xlsFile, nil
}

// voterInfo.VPending, voterInfo.VActive , vaInfo.ValidatorReward, f.String(), calcuR, nextAVote,
func writeInfo(epochNum uint64, latestBlock string, From string, TargetAddress string, VPending *big.Int, VActive *big.Int, ValidatorReward *big.Int, f string, calcuR *big.Float, nextAVote *big.Float, validators string) {
	//wStr := csv.NewWriter(xlsFile)
	f += "%"

	wStr := csv.NewWriter(xlsFile)
	go func() {
		s := <-writeChan
		err := wStr.Write(s)
		if err != nil {
			log.Error("", "err", err)
		}
		wStr.Flush()
	}()

	VPending1 := fmt.Sprintf("%.4f", ToMapI(VPending)) + " Map"
	VActive1 := fmt.Sprintf("%.4f", ToMapI(VActive)) + " Map"
	ValidatorReward1 := fmt.Sprintf("%.4f", ToMapI(ValidatorReward)) + " Map"
	calcuR1 := fmt.Sprintf("%.4f", ToMapF(calcuR)) + " Map"
	nextAVote1 := fmt.Sprintf("%.4f", ToMapF(nextAVote)) + " Map"
	s0 := []string{strconv.FormatUint(epochNum, 10), latestBlock, From, TargetAddress, VPending1, VActive1, ValidatorReward1, f, calcuR1, nextAVote1, validators}
	writeChan <- s0

}

func (l *listener) getValidators() []common.Address {
	client, _ := connections.DialRpc(l.cfg)
	var ret []common.Address
	if err := client.Call(&ret, "istanbul_getValidators"); err != nil {
		log.Error("msg", "err", err)
	}
	return ret
}
func ToMapI(val *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(val), fbaseUnit)
}
func ToMapF(val *big.Float) *big.Float {
	return new(big.Float).Quo(val, fbaseUnit)
}
