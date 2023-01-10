package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/new_marker/connections"
	"github.com/mapprotocol/atlas/cmd/new_marker/define"
	"github.com/mapprotocol/atlas/cmd/new_marker/mapprotocol"
	"github.com/mapprotocol/atlas/cmd/new_marker/writer"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/helper/fileutils"
	"github.com/mapprotocol/atlas/marker/env"
	"github.com/mapprotocol/atlas/marker/genesis"
	"github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"path"
	"strconv"
	"time"
)

type Tool struct {
	*base
	electionTo  common.Address
	electionAbi *abi.ABI
}

func NewTool() *Tool {
	return &Tool{
		base:        newBase(),
		electionTo:  mapprotocol.MustProxyAddressFor("Election"),
		electionAbi: mapprotocol.AbiFor("Election"),
	}
}

func (t *Tool) createGenesis(ctx *cli.Context) error {
	genesis.UnmarshalMarkerConfig(ctx)
	var workdir string
	var err error
	if ctx.IsSet(define.NewEnvFlag.Name) {
		workdir = ctx.String(define.NewEnvFlag.Name)
		if !fileutils.FileExists(workdir) {
			os.MkdirAll(workdir, os.ModePerm)
		}
	} else {
		workdir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	env, genesisConfig, err := envFromTemplate(ctx, workdir)
	if err != nil {
		return err
	}

	buildpath, err := readBuildPath(ctx)
	if err != nil {
		return err
	}

	generatedGenesis, err := genesis.GenerateGenesis(ctx, env.Accounts(), genesisConfig, buildpath)
	if err != nil {
		return err
	}

	if ctx.IsSet(define.NewEnvFlag.Name) {
		if err = env.Save(); err != nil {
			return err
		}
	}

	return env.SaveGenesis(generatedGenesis)
}

func (t *Tool) transfer(_ *cli.Context, cfg *define.Config) error {
	amount, ok := new(big.Int).SetString(cfg.Amount, 10)
	if !ok {
		log.Error("invalid amount", "amount ", cfg.Amount)
		return nil
	}
	conn := t.newConn(cfg.RPCAddr)
	if amount.Cmp(big.NewInt(0)) != 1 {
		log.Error("transfer amount must be greater than 0", "amount", cfg.Amount)
		return nil
	}

	txHash, err := writer.SendContractTransaction(conn, cfg.From, cfg.TargetAddress, amount, cfg.PrivateKey, nil, 0)
	if err != nil {
		return err
	}
	writer.GetResult(conn, txHash, false)
	log.Info("transfer success", "from ", cfg.From, "to", cfg.TargetAddress, "amount", cfg.Amount)
	return nil
}

func (t *Tool) voterMonitor(ctx *cli.Context, cfg *define.Config) error {
	writeChan := make(chan []string)
	xlsFile1, err := initCsv()
	xlsFile := xlsFile1
	defer xlsFile.Close()
	if err != nil {
		panic("initCsv")
	}

	configName := "./Voters2Validator.json"

	data, err := ioutil.ReadFile(configName)
	if err != nil {
		log.Error("compass personInfo config readFile Err", err.Error())
	}
	_ = json.Unmarshal(data, &define.Voter2validator)
	for _, v := range define.Voter2validator {
		define.VoterList = append(define.VoterList, define.VoterStruct{
			Voter:     common.HexToAddress(v.VoterAccount),
			Validator: common.HexToAddress(v.ValidatorAccount),
		})
	}

	validatorMap := map[common.Address]*define.ValidatorInfo{}
	for _, v := range define.VoterList {
		validatorMap[v.Validator] = &define.ValidatorInfo{
			AllVotes:        big.NewInt(0),
			ValidatorReward: big.NewInt(0),
		}
	}
	voterMap := map[define.VoterStruct]*define.VoterInfo{}
	for _, v := range define.VoterList {
		voterMap[v] = &define.VoterInfo{
			VActive:   big.NewInt(0),
			VPending:  big.NewInt(0),
			Voter:     v.Voter,
			Validator: v.Validator,
		}
		time.Sleep(3 * time.Second)
		go t.pollBlocks(ctx, cfg, v, voterMap[v], validatorMap, writeChan)
	}
	select {}
	return nil
}

func (t *Tool) pollBlocks(ctx *cli.Context, cfg *define.Config, key define.VoterStruct, voterInfo *define.VoterInfo,
	validatorMap map[common.Address]*define.ValidatorInfo, writeChan chan []string) error {
	voterInfo.VActive = big.NewInt(0)
	voterInfo.VPending = big.NewInt(0)
	vaInfo := validatorMap[voterInfo.Validator]
	vaInfo.AllVotes = big.NewInt(0)
	From := key.Voter
	TargetAddress := key.Validator
	var (
		epochSize            = params.Epoch
		currentEpoch  uint64 = 0
		rate                 = 5 * time.Second
		f             *big.Float
		validatorList []common.Address
		latestBlock   *big.Int
	)
	conn := t.newConn(cfg.RPCAddr)
	bnum, err := conn.BlockNumber(context.Background())
	if err != nil {
		log.Error("", "", err)
	}
	latestBlock = big.NewInt(0).SetUint64(bnum)
	epochNum := istanbul.GetEpochNumber(latestBlock.Uint64(), epochSize)
	voterInfo.EpochNum = epochNum
	query := func(title string) bool {
		bnum, err = conn.BlockNumber(context.Background())
		if err != nil {
			panic(err)
		}
		latestBlock = big.NewInt(0).SetUint64(bnum)
		log.Info(title, "Epoch", epochNum, "blockNumber", latestBlock, "YourAccount", From, "Validator", TargetAddress, "sign", key.Voter.String()+" to "+key.Validator.String())
		// 查询当前投给validator情况
		t.getActiveVotesForValidatorByAccount_(ctx, cfg, key, voterInfo)
		t.getPendingInfo_(ctx, cfg, key, voterInfo)
		validatorList = t.getValidators(cfg)
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
		t.getActiveVotesForValidator_(ctx, cfg, key, voterInfo, validatorMap)
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
			bnum, err = conn.BlockNumber(context.Background())
			if err != nil {
				log.Error("getlatestBlock", "latestBlock", latestBlock)
				time.Sleep(3 * time.Second)
				continue
			}
			latestBlock = big.NewInt(0).SetUint64(bnum)
			epochNum = istanbul.GetEpochNumber(latestBlock.Uint64(), epochSize)
			if currentEpoch < epochNum && istanbul.GetNumberWithinEpoch(latestBlock.Uint64(), epochSize) == epochSize-1 {
				currentEpoch = epochNum
				query("=== start ====")
			}
			if istanbul.IsLastBlockOfEpoch(latestBlock.Uint64(), epochSize) {
				log.Info("=== Reward ===")
				// getVoterRewardInfo
				t.getVoterRewardInfo_(latestBlock.Uint64(), cfg, key, voterInfo, validatorMap)
				// calculate reward
				calcuR := new(big.Float).SetInt(vaInfo.ValidatorReward)
				calcuR.Mul(calcuR, f)
				nextAVote := new(big.Float).SetInt(voterInfo.VActive)
				nextAVote.Add(nextAVote, calcuR)

				log.Info("=== Reward ===", "Target_Reward", calcuR, "Target_Active_Vote", nextAVote, "sign", key.Voter.String()+" to "+key.Validator.String())
				f1 := new(big.Float).SetInt(big.NewInt(100))
				f.Mul(f, f1)
				t.writeInfo(epochNum, latestBlock.String(), From.String(), TargetAddress.String(),
					voterInfo.VPending, voterInfo.VActive, vaInfo.ValidatorReward, f.String(),
					calcuR, nextAVote, fmt.Sprintf("%v", validatorList), writeChan)

				query("=== Result ===")

				fmt.Println()
				fmt.Println()
				time.Sleep(2 * rate)
			}

		}
	}
}

var xlsFile *os.File

func (t *Tool) writeInfo(epochNum uint64, latestBlock string, From string, TargetAddress string,
	VPending *big.Int, VActive *big.Int, ValidatorReward *big.Int, f string, calcuR *big.Float,
	nextAVote *big.Float, validators string, writeChan chan []string) {
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

var (
	baseUnit  = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	fbaseUnit = new(big.Float).SetFloat64(float64(baseUnit.Int64()))
)

func ToMapI(val *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(val), fbaseUnit)
}
func ToMapF(val *big.Float) *big.Float {
	return new(big.Float).Quo(val, fbaseUnit)
}

func (t *Tool) getVoterRewardInfo_(curBlockNumber uint64, cfg *define.Config, key define.VoterStruct,
	info *define.VoterInfo, validatorMap map[common.Address]*define.ValidatorInfo) {
	TargetAddress := key.Validator
	epochSize := params.Epoch
	epochNum := istanbul.GetEpochNumber(curBlockNumber, epochSize)
	EpochLast := big.NewInt(0).SetUint64(istanbul.GetEpochLastBlockNumber(epochNum, epochSize))
	Epoch := istanbul.GetEpochNumber(curBlockNumber, epochSize)
	query := mapprotocol.BuildQuery(t.electionTo, mapprotocol.EpochRewardsDistributedToVoters, EpochLast, EpochLast)
	// querying for logs
	conn := t.newConn(cfg.RPCAddr)
	logs, err := conn.FilterLogs(context.Background(), query)
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

func (t *Tool) getActiveVotesForValidator_(_ *cli.Context, cfg *define.Config, key define.VoterStruct,
	voterInfo *define.VoterInfo, validatorMap map[common.Address]*define.ValidatorInfo) {
	TargetAddress := key.Validator
	valiInfo := validatorMap[voterInfo.Validator]
	var ret interface{}
	ElectionAddress := cfg.ElectionParameters.ElectionAddress
	abiElection := cfg.ElectionParameters.ElectionABI
	t.handleType3Msg(cfg, &ret, ElectionAddress, nil, abiElection, "getActiveVotesForValidator", TargetAddress)
	log.Info("", "Validator all Votes", ret.(*big.Int), "sign", key.Voter.String()+" to "+key.Validator.String())
	valiInfo.AllVotes = ret.(*big.Int)
}

func (t *Tool) getActiveVotesForValidatorByAccount_(_ *cli.Context, cfg *define.Config, key define.VoterStruct, voterInfo *define.VoterInfo) *big.Int {
	From := key.Voter
	TargetAddress := key.Validator
	var ret interface{}
	t.handleType3Msg(cfg, &ret, t.electionTo, nil, t.electionAbi, "getActiveVotesForValidatorByAccount", TargetAddress, From)
	log.Info("", "Active Vote", ret.(*big.Int))
	voterInfo.VActive = ret.(*big.Int)
	return voterInfo.VActive
}

func (t *Tool) getValidators(cfg *define.Config) []common.Address {
	client, _ := connections.DialRpc(cfg)
	var ret []common.Address
	if err := client.Call(&ret, "istanbul_getValidators"); err != nil {
		log.Error("msg", "err", err)
	}
	return ret
}

func (t *Tool) getPendingInfo_(_ *cli.Context, cfg *define.Config, key define.VoterStruct, voterInfo *define.VoterInfo) (*big.Int, *big.Int) {
	From := key.Voter
	TargetAddress := key.Validator
	type ret []interface{}
	var (
		Value    interface{}
		Epoch    interface{}
		epochNum uint64
	)
	result := ret{&Value, &Epoch}

	f := func(output []byte) {
		err := t.electionAbi.UnpackIntoInterface(&result, "pendingInfo", output)
		if err != nil {
			log.Error("getPendingInfoForValidator", "err", err)
			os.Exit(1)
		}
	}
	t.handleType4Msg(cfg, f, t.electionTo, nil, t.electionAbi, "pendingInfo", From, TargetAddress)
	p := Epoch.(*big.Int)
	log.Info("", "PendingVotes", Value.(*big.Int), "epoch", p.Add(p, big.NewInt(1)), "sign", key.Voter.String()+" to "+key.Validator.String())
	voterInfo.VPending = big.NewInt(0)
	if Epoch.(*big.Int).CmpAbs(big.NewInt(0).SetUint64(epochNum)) < 0 {
		voterInfo.VPending = Value.(*big.Int)
	}
	return Epoch.(*big.Int), Value.(*big.Int)
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

func readBuildPath(ctx *cli.Context) (string, error) {
	buildpath := ctx.String(define.BuildpathFlag.Name)
	if buildpath == "" {
		buildpath = path.Join(os.Getenv("ATLAS_MONOREPO"), "packages/protocol/build/contracts")
		if fileutils.FileExists(buildpath) {
			log.Info("Missing --buildpath flag, using ATLAS_MONOREPO derived path", "buildpath", buildpath)
		} else {
			return "", fmt.Errorf("Missing --buildpath flag")
		}
	}
	return buildpath, nil
}

func init() {
	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.LvlInfo)
	log.Root().SetHandler(glogger)
}

func envFromTemplate(ctx *cli.Context, workdir string) (*env.Environment, *genesis.Config, error) {
	templateString := ctx.String("template")
	template := templateFromString(templateString)
	env, err := template.createEnv(workdir)
	if err != nil {
		return nil, nil, err
	}
	// Env overrides
	if ctx.IsSet("validators") {
		env.Accounts().NumValidators = ctx.Int("validators")
	}

	// Genesis config
	genesisConfig, err := template.createGenesisConfig(env)
	if err != nil {
		return nil, nil, err
	}

	return env, genesisConfig, nil
}

type template interface {
	createEnv(workdir string) (*env.Environment, error)
	createGenesisConfig(*env.Environment) (*genesis.Config, error)
}

func templateFromString(templateStr string) template {
	switch templateStr {
	case "local":
		return localEnv{}
	case "loadtest":
		return loadtestEnv{}
	case "monorepo":
		return monorepoEnv{}
	}
	return localEnv{}
}

type localEnv struct{}

func (e localEnv) createEnv(workdir string) (*env.Environment, error) {
	envCfg := &env.Config{
		Accounts: env.AccountsConfig{
			Mnemonic:             env.MustNewMnemonic(),
			NumValidators:        4,
			NumDeveloperAccounts: 10,
		},
		ChainID: big.NewInt(211),
	}
	env, err := env.New(workdir, envCfg)
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (e localEnv) createGenesisConfig(env *env.Environment) (*genesis.Config, error) {
	genesisConfig := genesis.CreateCommonGenesisConfig()
	return genesisConfig, nil
}

type loadtestEnv struct{}

func (e loadtestEnv) createEnv(workdir string) (*env.Environment, error) {
	envCfg := &env.Config{
		Accounts: env.AccountsConfig{
			Mnemonic:             "miss fire behind decide egg buyer honey seven advance uniform profit renew",
			NumValidators:        1,
			NumDeveloperAccounts: 10000,
		},
		ChainID: big.NewInt(9099000),
	}

	env, err := env.New(workdir, envCfg)
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (e loadtestEnv) createGenesisConfig(env *env.Environment) (*genesis.Config, error) {
	genesisConfig := genesis.CreateCommonGenesisConfig()
	// 10 billion gas limit, set super high on purpose
	genesisConfig.Blockchain.BlockGasLimit = 1000000000

	return genesisConfig, nil
}

type monorepoEnv struct{}

func (e monorepoEnv) createEnv(workdir string) (*env.Environment, error) {
	envCfg := &env.Config{
		Accounts: env.AccountsConfig{
			Mnemonic:             env.MustNewMnemonic(),
			NumValidators:        3,
			NumDeveloperAccounts: 0,
			UseValidatorAsAdmin:  true, // monorepo doesn't use the admin account type, uses first validator instead
		},
		ChainID: big.NewInt(1000 * (1 + rand.Int63n(9999))),
	}
	env, err := env.New(workdir, envCfg)
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (e monorepoEnv) createGenesisConfig(env *env.Environment) (*genesis.Config, error) {
	genesisConfig := genesis.CreateCommonGenesisConfig()
	return genesisConfig, nil
}
