package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/cmd/marker/connections"
	"strconv"

	"github.com/mapprotocol/atlas/cmd/marker/mapprotocol"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/params"
	"os"
	//"strconv"

	"gopkg.in/urfave/cli.v1"
	"math/big"
	"time"
)

func (c *listener) LatestBlock() (*big.Int, error) {
	bnum, err := c.conn.BlockNumber(context.Background())
	if err != nil {
		return nil, err
	}
	return big.NewInt(0).SetUint64(bnum), nil
}

var VActive *big.Int
var Vpending *big.Int
var allVotes *big.Int
var ValidatorReward *big.Int
var epochNum uint64

func pollBlocks(ctx *cli.Context, core *listener) error {
	xlsFile, err := core.initCsv(ctx, core)
	defer xlsFile.Close()
	if err != nil {
		panic("initCsv")
	}
	VActive = big.NewInt(0)
	Vpending = big.NewInt(0)
	allVotes = big.NewInt(0)

	var epochSize = params.Epoch
	var currentEpoch uint64 = 0
	var rate = 5 * time.Second
	var f *big.Float

	var latestBlock *big.Int
	var epochLast uint64
	latestBlock, err = core.LatestBlock()
	epochNum = istanbul.GetEpochNumber(latestBlock.Uint64(), epochSize)
	query := func() bool {
		latestBlock, err = core.LatestBlock()
		if err != nil {
			panic(err)
		}
		log.Info("", "Epoch", epochNum, "blockNumber", latestBlock, "YourAccount", core.cfg.From, "Validator", core.cfg.TargetAddress)
		// 查询当前投给validator情况
		getActiveVotesForValidatorByAccount_(ctx, core)
		l := core.getValidators()
		exist := false
		for _, e := range l {
			if e == core.cfg.TargetAddress {
				exist = true
			}
		}
		if !exist {
			log.Info("the target validator Not selected ", "current validators", l)
		}
		epochLast = istanbul.GetEpochLastBlockNumber(epochNum, epochSize)
		if !exist {
			log.Info("The validator not be Selected ", "validator ", core.cfg.TargetAddress)
		}

		// 查询validator情况
		getActiveVotesForValidator_(ctx, core)
		f = new(big.Float).SetInt(VActive)
		fSub := new(big.Float).SetInt(allVotes)
		if allVotes.CmpAbs(big.NewInt(0)) > 0 {
			f.Quo(f, fSub)
		} else {
			f = new(big.Float).SetUint64(0)
		}
		log.Info("", "fraction", f)
		return true
	}
	query()
	log.Info("wait for reward...")
	for {
		select {
		default:
			time.Sleep(time.Second)
			latestBlock, err = core.LatestBlock()
			epochNum = istanbul.GetEpochNumber(latestBlock.Uint64(), epochSize)
			if currentEpoch < epochNum && istanbul.GetNumberWithinEpoch(latestBlock.Uint64(), epochSize) == epochSize-1 {
				currentEpoch = epochNum
				log.Info("=== start ====")
				query()
			}
			if istanbul.IsLastBlockOfEpoch(latestBlock.Uint64(), epochSize) {
				log.Info("=== Reward ===")
				// 查询 奖励情况
				getVoterRewardInfo_(latestBlock.Uint64(), core)
				// 应收到奖励
				fReal := new(big.Float).SetInt(ValidatorReward)
				fReal.Mul(fReal, f)
				ft := new(big.Float).SetInt(VActive)
				ft.Add(ft, fReal)

				log.Info("", "Target_Reward", fReal, "Target_Active_Vote", ft)

				wStr := csv.NewWriter(xlsFile)
				//{"Epoch", "BlockNumber", "voter", "validator", "vote", "validatorReward", "targetReward", "target"}
				err := wStr.Write([]string{strconv.FormatUint(epochNum, 10), latestBlock.String(), core.cfg.From.String(), core.cfg.TargetAddress.String(), Vpending.String(), VActive.String(), ValidatorReward.String(), fReal.String(), ft.String()})
				if err != nil {
					log.Error("", "err", err)
				}
				wStr.Flush()
				log.Info("=== Result ===")
				query()

				fmt.Println()
				fmt.Println()
				time.Sleep(2 * rate)
			}

		}
	}
	return nil
}

func getActiveVotesForValidatorByAccount_(_ *cli.Context, core *listener) {
	var ret interface{}
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI

	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getActiveVotesForValidatorByAccount", core.cfg.TargetAddress, core.cfg.From)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	log.Info("", "Active Vote", ret.(*big.Int))
	VActive = ret.(*big.Int)
}

func getPendingInfo_(_ *cli.Context, core *listener) (*big.Int, *big.Int) {
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
	m := NewMessageRet2(SolveQueryResult4, core.msgCh, core.cfg, f, ElectionAddress, nil, abiElection, "pendingInfo", core.cfg.From, core.cfg.TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	p := Epoch.(*big.Int)
	log.Info("", "PendingVotes", Value.(*big.Int), "epoch", p.Add(p, big.NewInt(1)))
	Vpending = big.NewInt(0)
	if Epoch.(*big.Int).CmpAbs(big.NewInt(0).SetUint64(epochNum)) < 0 {
		Vpending = Value.(*big.Int)
	}
	return Epoch.(*big.Int), Value.(*big.Int)
}

func getActiveVotesForValidator_(_ *cli.Context, core *listener) {
	var ret interface{}
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getActiveVotesForValidator", core.cfg.TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	log.Info("", "Validator all Votes", ret.(*big.Int))
	allVotes = ret.(*big.Int)
}

func getVoterRewardInfo_(curBlockNumber uint64, core *listener) {
	//curBlockNumber, err := core.conn.BlockNumber(context.Background())
	epochSize := params.Epoch
	//if err != nil {
	//	log.Error("getVoterRewardInfo_getEpochSize", "err", epochSize)
	//}
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
		if validator == core.cfg.TargetAddress {
			//validator := common.Bytes2Hex(l.Topics[0].Bytes())
			reward := big.NewInt(0).SetBytes(l.Data[:32])
			log.Info("", "Epoch", Epoch, "blockNumber", curBlockNumber, "Reward to Voters", reward)
			ValidatorReward = reward
		}
	}
}

func (l *listener) initCsv(ctx *cli.Context, core *listener) (*os.File, error) {
	strTime := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("Voter_Validator_%s_%s_%s.csv", core.cfg.From, core.cfg.TargetAddress, strTime)
	xlsFile, fErr := os.OpenFile("./"+filename, os.O_RDWR|os.O_CREATE, 0766)
	if fErr != nil {
		fmt.Println("Export:created excel file failed ==", fErr)
		return nil, fErr
	}
	xlsFile.WriteString("\xEF\xBB\xBF")
	wStr := csv.NewWriter(xlsFile)
	wStr.Write([]string{"Epoch", "BlockNumber", "Voter", "Validator", "PendingVote", "ActiveVote", "ValidatorReward", "CalculateReward", "NextActiveVote"})
	wStr.Flush()
	return xlsFile, nil
}

func (l *listener) getValidators() []common.Address {
	client, _ := connections.DialRpc(l.cfg)

	var ret []common.Address
	//blockNrOrHash rpc.BlockNumberOrHash, start []byte, maxResults int, nocode, nostorage, incompletes bool
	if err := client.Call(&ret, "istanbul_getValidators"); err != nil {
		log.Error("msg", "err", err)
	}
	return ret
}
