package cmd

import (
	"bytes"
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/new_marker/define"
	"github.com/mapprotocol/atlas/cmd/new_marker/mapprotocol"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/helper/decimal"
	"github.com/mapprotocol/atlas/helper/decimal/fixed"
	"github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	"os"
	"sort"
	"strings"
)

type Voter struct {
	*base
	account                                                                 *Account
	validator                                                               *Validator
	lockGoldTo, electionTo, validatorTo, goldTokenTo, epochRewardsTo        common.Address
	lockedGoldAbi, electionAbi, validatorAbi, goldTokenAbi, epochRewardsAbi *abi.ABI
}

func NewVoter() *Voter {
	return &Voter{
		base:            newBase(),
		account:         NewAccount(),
		validator:       NewValidator(),
		lockGoldTo:      mapprotocol.MustProxyAddressFor("LockedGold"),
		lockedGoldAbi:   mapprotocol.AbiFor("LockedGold"),
		electionTo:      mapprotocol.MustProxyAddressFor("Election"),
		electionAbi:     mapprotocol.AbiFor("Election"),
		validatorTo:     mapprotocol.MustProxyAddressFor("Validators"),
		validatorAbi:    mapprotocol.AbiFor("Validators"),
		goldTokenTo:     mapprotocol.MustProxyAddressFor("GoldToken"),
		goldTokenAbi:    mapprotocol.AbiFor("GoldToken"),
		epochRewardsTo:  mapprotocol.MustProxyAddressFor("EpochRewards"),
		epochRewardsAbi: mapprotocol.AbiFor("EpochRewards"),
	}
}

func (v *Voter) Vote(_ *cli.Context, cfg *define.Config) error {
	greater, lesser, err := v.getGL(cfg, cfg.TargetAddress)
	if err != nil {
		log.Error("vote", "err", err)
		return err
	}
	amount := new(big.Int).Mul(cfg.VoteNum, big.NewInt(1e18))
	log.Info("=== vote Validator ===", "admin", cfg.From, "voteTargetValidator", cfg.TargetAddress.String(), "vote MAP Num", cfg.VoteNum.String())
	v.handleType1Msg(cfg, v.electionTo, nil, v.electionAbi, "vote", cfg.TargetAddress, amount, lesser, greater)
	return nil
}

func (v *Voter) QuicklyVote(ctx *cli.Context, cfg *define.Config) error {
	//---------------------------- create account ----------------
	_ = v.account.CreateAccount(ctx, cfg)
	//---------------------------- lock --------------------------
	_ = v.validator.LockedMAP(ctx, cfg)
	//---------------------------- vote --------------------------
	_ = v.Vote(ctx, cfg)
	log.Info("=== End ===")
	return nil
}

func (v *Voter) Activate(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== activate validator gold ===", "account.Address", cfg.From)
	v.handleType1Msg(cfg, v.electionTo, nil, v.electionAbi, "activate", cfg.TargetAddress)
	return nil
}

func (v *Voter) GetActiveVotesForValidator(_ *cli.Context, cfg *define.Config) error {
	var ret interface{}
	log.Info("=== getActiveVotesForValidator ===", "admin", cfg.From)
	v.handleType3Msg(cfg, &ret, v.electionTo, nil, v.electionAbi, "getActiveVotesForValidator", cfg.TargetAddress)
	log.Info("ActiveVotes", "balance", ret.(*big.Int))
	return nil
}

func (v *Voter) GetPendingVotersForValidator(_ *cli.Context, cfg *define.Config) error {
	var ret interface{}
	log.Info("=== getPendingVotersForValidator ===", "admin", cfg.From)
	v.handleType3Msg(cfg, &ret, v.electionTo, nil, v.electionAbi, "getPendingVotersForValidator", cfg.TargetAddress)
	log.Info("getPendingVotersForValidator", "voters", ret.([]common.Address))
	return nil
}

func (v *Voter) GetPendingInfoForValidator(_ *cli.Context, cfg *define.Config) error {
	type ret []interface{}
	var (
		Value interface{}
		Epoch interface{}
	)
	t := ret{&Value, &Epoch}
	f := func(output []byte) {
		err := v.electionAbi.UnpackIntoInterface(&t, "pendingInfo", output)
		if err != nil {
			log.Error("getPendingInfoForValidator", "err", err)
			os.Exit(1)
		}
	}
	log.Info("=== getPendingInfoForValidator ===", "admin", cfg.From)
	v.handleType4Msg(cfg, f, v.electionTo, nil, v.electionAbi, "pendingInfo", cfg.From, cfg.TargetAddress)
	log.Info("getPendingInfoForValidator", "PendingEpoch", Epoch.(*big.Int), "Balance", Value.(*big.Int))
	return nil
}

func (v *Voter) RevokePending(_ *cli.Context, cfg *define.Config) error {
	validator := cfg.TargetAddress
	LockedNum := new(big.Int).Mul(cfg.LockedNum, big.NewInt(1e18))

	greater, lesser, _ := v.getGLSub(cfg, LockedNum, validator)
	list := v._getValidatorsVotedForByAccount(cfg, cfg.From)
	index, err := v.GetIndex(validator, list)
	if err != nil {
		log.Crit("revokePending", "err", err)
	}
	//fmt.Println("=== greater,lesser,index ===", greater, lesser, index)
	_params := []interface{}{validator, LockedNum, lesser, greater, index}
	log.Info("=== revokePending ===", "admin", cfg.From)
	v.handleType1Msg(cfg, v.electionTo, nil, v.electionAbi, "revokePending", _params...)
	return nil
}

func (v *Voter) RevokeActive(_ *cli.Context, cfg *define.Config) error {
	validator := cfg.TargetAddress
	LockedNum := new(big.Int).Mul(cfg.LockedNum, big.NewInt(1e18))
	greater, lesser, _ := v.getGLSub(cfg, LockedNum, validator)

	list := v._getValidatorsVotedForByAccount(cfg, cfg.From)
	index, err := v.GetIndex(validator, list)
	if err != nil {
		log.Crit("revokePending", "err", err)
	}
	_params := []interface{}{validator, LockedNum, lesser, greater, index}
	log.Info("=== revokeActive ===", "admin", cfg.From)
	v.handleType1Msg(cfg, v.electionTo, nil, v.electionAbi, "revokeActive", _params...)
	return nil
}

func (v *Voter) LockedMAP(_ *cli.Context, cfg *define.Config) error {
	lockedGold := new(big.Int).Mul(cfg.LockedNum, big.NewInt(1e18))
	log.Info("=== Lock  gold ===")
	log.Info("Lock  gold", "amount", lockedGold.String())
	v.handleType2Msg(cfg, v.lockGoldTo, lockedGold, v.lockedGoldAbi, "lock")
	return nil
}

func (v *Voter) UnlockedMAP(_ *cli.Context, cfg *define.Config) error {
	lockedGold := new(big.Int).Mul(cfg.LockedNum, big.NewInt(1e18))
	log.Info("=== unLock validator gold ===")
	log.Info("unLock validator gold", "amount", lockedGold, "admin", cfg.From)
	v.handleType1Msg(cfg, v.lockGoldTo, nil, v.lockedGoldAbi, "unlock", lockedGold)
	return nil
}

func (v *Voter) RelockMAP(_ *cli.Context, cfg *define.Config) error {
	lockedGold := new(big.Int).Mul(cfg.LockedNum, big.NewInt(1e18))
	log.Info("=== relockMAP validator gold ===")
	log.Info("relockMAP validator gold", "amount", lockedGold)
	v.handleType1Msg(cfg, v.lockGoldTo, nil, v.lockedGoldAbi, "relock", cfg.RelockIndex, lockedGold)
	return nil
}

func (v *Voter) Withdraw(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== withdraw validator gold ===", "admin", cfg.From.String())
	v.handleType1Msg(cfg, v.lockGoldTo, nil, v.lockedGoldAbi, "withdraw", cfg.WithdrawIndex)
	return nil
}

func (v *Voter) GetTotalVotesForEligibleValidators(_ *cli.Context, cfg *define.Config) error {
	type ret struct {
		Validators interface{} // indexed
		Values     interface{}
	}
	var t ret
	f := func(output []byte) {
		err := v.electionAbi.UnpackIntoInterface(&t, "getTotalVotesForEligibleValidators", output)
		if err != nil {
			log.Error("getTotalVotesForEligibleValidators", "err", err)
			os.Exit(1)
		}
	}
	log.Info("=== getTotalVotesForEligibleValidators ===", "admin", cfg.From)
	v.handleType4Msg(cfg, f, v.electionTo, nil, v.electionAbi, "getTotalVotesForEligibleValidators")
	Validators := (t.Validators).([]common.Address)
	Values := (t.Values).([]*big.Int)
	for i := 0; i < len(Validators); i++ {
		log.Info("Validator:", "addr", Validators[i], "vote amount", Values[i])
	}
	return nil
}

func (v *Voter) GetRegisteredValidatorSigners(_ *cli.Context, cfg *define.Config) error {
	log.Info("==== getRegisteredValidatorSigners ===")
	Validators := v._getRegisteredValidatorSigners(cfg)
	if len(Validators) == 0 {
		log.Info("nil")
	}
	for i := 0; i < len(Validators); i++ {
		log.Info("Validator:", "index", i, "addr", Validators[i])
	}
	return nil
}

func (v *Voter) GetValidator(_ *cli.Context, cfg *define.Config) error {
	type ret struct {
		EcdsaPublicKey      interface{}
		BlsPublicKey        interface{}
		BlsG1PublicKey      interface{}
		Score               interface{}
		Signer              interface{}
		Commission          interface{}
		NextCommission      interface{}
		NextCommissionBlock interface{}
		SlashMultiplier     interface{}
		LastSlashed         interface{}
	}
	var t ret
	f := func(output []byte) {
		err := v.validatorAbi.UnpackIntoInterface(&t, "getValidator", output)
		if err != nil {
			log.Error("getValidator", "err", err)
			os.Exit(1)
		}
	}

	log.Info("=== getValidator ===", "admin", cfg.From)
	v.handleType4Msg(cfg, f, v.validatorTo, nil, v.validatorAbi, "getValidator", cfg.TargetAddress)
	log.Info("", "ecdsaPublicKey", common.BytesToHash(t.EcdsaPublicKey.([]byte)).String())
	log.Info("", "BlsPublicKey", common.BytesToHash(t.BlsPublicKey.([]byte)).String())
	log.Info("", "BlsG1PublicKey", common.BytesToHash(t.BlsG1PublicKey.([]byte)).String())
	log.Info("", "Score", ConvertToFraction(t.Score))
	log.Info("", "Signer", t.Signer)
	log.Info("", "Commission", ConvertToFraction(t.Commission))
	log.Info("", "NextCommission", ConvertToFraction(t.NextCommission))
	log.Info("", "NextCommissionBlock", t.NextCommissionBlock)
	log.Info("", "SlashMultiplier", ConvertToFraction(t.SlashMultiplier))
	log.Info("", "LastSlashed", ConvertToFraction(t.LastSlashed))
	return nil
}

func (v *Voter) GetRewardInfo(_ *cli.Context, cfg *define.Config) error {
	conn := v.newConn(cfg.RPCAddr)
	curBlockNumber, err := conn.BlockNumber(context.Background())
	epochSize := chain.DefaultGenesisBlock().Config.Istanbul.Epoch
	if err != nil {
		return err
	}
	EpochFirst, err := istanbul.GetEpochFirstBlockGivenBlockNumber(curBlockNumber, epochSize)
	if err != nil {
		return err
	}
	Epoch := istanbul.GetEpochNumber(curBlockNumber, epochSize)
	validatorContractAddress := cfg.ValidatorParameters.ValidatorAddress
	queryBlock := big.NewInt(int64(EpochFirst - 1))
	log.Info("=== getReward ===", "cur_epoch", Epoch, "epochSize", epochSize, "queryBlockNumber", queryBlock, "validatorContractAddress", validatorContractAddress.String(), "admin", cfg.From)
	query := mapprotocol.BuildQuery(validatorContractAddress, mapprotocol.ValidatorEpochPaymentDistributed, queryBlock, queryBlock)
	// querying for logs
	logs, err := conn.FilterLogs(context.Background(), query)
	if err != nil {
		return err
	}
	for _, l := range logs {
		//validator := common.Bytes2Hex(l.Topics[0].Bytes())
		validator := common.BytesToAddress(l.Topics[1].Bytes())
		reward := big.NewInt(0).SetBytes(l.Data[:32])
		log.Info("", "validator", validator, "reward", reward)
	}
	log.Info("=== END ===")
	return nil
}

func (v *Voter) getVoterRewardInfo(ctx *cli.Context, cfg *define.Config) error {
	conn := v.newConn(cfg.RPCAddr)
	curBlockNumber, err := conn.BlockNumber(context.Background())
	epochSize := chain.DefaultGenesisBlock().Config.Istanbul.Epoch
	if err != nil {
		return err
	}
	EpochFirst, err := istanbul.GetEpochFirstBlockGivenBlockNumber(curBlockNumber, epochSize)
	if err != nil {
		return err
	}
	Epoch := istanbul.GetEpochNumber(curBlockNumber, epochSize)
	electionContractAddress := cfg.ElectionParameters.ElectionAddress
	firstBlock := big.NewInt(int64(1))
	endBlock := big.NewInt(int64(EpochFirst + 1))
	log.Info("=== get voter Reward ===", "cur_epoch", Epoch, "epochSize", epochSize, "query first BlockNumber", firstBlock, "query end BlockNumber", endBlock, "validatorContractAddress", electionContractAddress.String(), "admin", cfg.From)
	query := mapprotocol.BuildQuery(electionContractAddress, mapprotocol.EpochRewardsDistributedToVoters, firstBlock, endBlock)
	// querying for logs
	logs, err := conn.FilterLogs(context.Background(), query)
	if err != nil {
		return err
	}
	for _, l := range logs {
		validator := common.BytesToAddress(l.Topics[1].Bytes())
		reward := big.NewInt(0).SetBytes(l.Data[:32])
		log.Info("reward to voters", "validator", validator, "reward", reward)
	}
	log.Info("=== END ===")
	return nil
}

func (v *Voter) getNumRegisteredValidators(_ *cli.Context, cfg *define.Config) error {
	var NumValidators interface{}
	v.handleType3Msg(cfg, &NumValidators, v.validatorTo, nil, v.validatorAbi, "getNumRegisteredValidators")
	ret := NumValidators.(*big.Int)
	log.Info("=== result ===", "num", ret.String())
	return nil
}

func (v *Voter) getTopValidators(_ *cli.Context, cfg *define.Config) error {
	var TopValidators interface{}
	v.handleType3Msg(cfg, &TopValidators, v.validatorTo, nil, v.validatorAbi, "getTopValidators", cfg.TopNum)
	Validators := TopValidators.([]common.Address)
	for i := 0; i < len(Validators); i++ {
		log.Info("Validator:", "index", i, "addr", Validators[i])
	}
	return nil
}

func (v *Voter) getValidatorEligibility(_ *cli.Context, cfg *define.Config) error {
	var ret interface{}
	v.handleType3Msg(cfg, &ret, v.electionTo, nil, v.electionAbi, "getValidatorEligibility", cfg.TargetAddress)
	log.Info("=== result ===", "bool", ret.(bool))
	return nil
}

func (v *Voter) balanceOf(_ *cli.Context, cfg *define.Config) error {
	var ret interface{}
	log.Info("=== balanceOf ===", "admin", cfg.From)
	v.handleType3Msg(cfg, &ret, v.goldTokenTo, nil, v.goldTokenAbi, "balanceOf", cfg.TargetAddress)
	log.Info("=== result ===", "balance", ret.(*big.Int).String())
	return nil
}

func (v *Voter) getTotalVotes(_ *cli.Context, cfg *define.Config) error {
	var ret interface{}
	log.Info("=== getAccountLockedGoldRequirement ===", "admin", cfg.From)
	v.handleType3Msg(cfg, &ret, v.electionTo, nil, v.electionAbi, "getTotalVotes")
	result := ret.(*big.Int)
	log.Info("result", "getTotalVotes", result)
	return nil
}

func (v *Voter) getPendingWithdrawals(_ *cli.Context, cfg *define.Config) error {
	type ret []interface{}
	var (
		Values     interface{}
		Timestamps interface{}
	)
	t := ret{&Values, &Timestamps}
	log.Info("=== getPendingWithdrawals ===", "admin", cfg.From, "target", cfg.TargetAddress.String())
	f := func(output []byte) {
		err := v.lockedGoldAbi.UnpackIntoInterface(&t, "getPendingWithdrawals", output)
		if err != nil {
			log.Error("getPendingWithdrawals", "err", err)
			os.Exit(1)
		}
	}
	v.handleType4Msg(cfg, f, v.lockGoldTo, nil, v.lockedGoldAbi, "getPendingWithdrawals", cfg.TargetAddress)
	Values1 := (Values).([]*big.Int)
	Timestamps1 := (Timestamps).([]*big.Int)
	if len(Values1) == 0 {
		log.Info("nil")
		return nil
	}
	for i := 0; i < len(Values1); i++ {
		log.Info("result:", "index", i, "values", Values1[i], "timestamps", Timestamps1[i])
	}
	return nil
}

func (v *Voter) setValidatorLockedGoldRequirements(_ *cli.Context, cfg *define.Config) error {
	value := new(big.Int).Mul(big.NewInt(int64(cfg.Value)), big.NewInt(1e18))
	duration := big.NewInt(cfg.Duration)
	log.Info("=== setValidatorLockedGoldRequirements ===", "admin", cfg.From.String())
	v.handleType1Msg(cfg, v.validatorTo, nil, v.validatorAbi, "setValidatorLockedGoldRequirements", value, duration)
	return nil
}

func (v *Voter) setImplementation(_ *cli.Context, cfg *define.Config) error {
	implementation := cfg.ImplementationAddress
	ContractAddress := cfg.ContractAddress
	ProxyAbi := mapprotocol.AbiFor("Proxy")
	log.Info("=== setImplementation ===", "admin", cfg.From.String())
	v.handleType1Msg(cfg, ContractAddress, nil, ProxyAbi, "_setImplementation", implementation)
	return nil
}

func (v *Voter) setContractOwner(_ *cli.Context, cfg *define.Config) error {
	NewOwner := cfg.TargetAddress
	ContractAddress := cfg.ContractAddress // 代理地址
	abiValidators := cfg.ValidatorParameters.ValidatorABI
	log.Info("ProxyAddress", "ContractAddress", ContractAddress, "NewOwner", NewOwner.String())
	log.Info("=== setOwner ===", "admin", cfg.From.String())
	v.handleType1Msg(cfg, ContractAddress, nil, abiValidators, "transferOwnership", NewOwner)
	return nil
}

func (v *Voter) setProxyContractOwner(_ *cli.Context, cfg *define.Config) error {
	NewOwner := cfg.TargetAddress
	ContractAddress := cfg.ContractAddress //代理地址
	log.Info("ProxyAddress", "ContractAddress", ContractAddress, "NewOwner", NewOwner.String())
	ProxyAbi := mapprotocol.AbiFor("Proxy") //代理ABI
	log.Info("=== setOwner ===", "admin", cfg.From.String())
	v.handleType1Msg(cfg, ContractAddress, nil, ProxyAbi, "_transferOwnership", NewOwner)
	return nil
}

func (v *Voter) getProxyContractOwner(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== getOwner ===", "admin", cfg.From.String())
	var ret interface{}
	ProxyAbi := mapprotocol.AbiFor("Proxy")
	v.handleType3Msg(cfg, &ret, cfg.ContractAddress, nil, ProxyAbi, "_getOwner")
	result := ret
	log.Info("getOwner", "Owner ", result)
	return nil
}

func (v *Voter) getContractOwner(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== getOwner ===", "admin", cfg.From.String())
	var ret interface{}
	v.handleType3Msg(cfg, &ret, cfg.ContractAddress, nil, v.validatorAbi, "owner")
	result := ret
	log.Info("getOwner", "Owner ", result)
	return nil
}

func (v *Voter) updateBlsPublicKey(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== updateBlsPublicKey ===")
	_params := []interface{}{cfg.PublicKey[1:], cfg.BlsPub[:], cfg.BlsG1Pub[:], cfg.BLSProof}
	v.handleType1Msg(cfg, v.validatorTo, nil, v.validatorAbi, "updateBlsPublicKey", _params...)
	return nil
}

func (v *Voter) setNextCommissionUpdate(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== setNextCommissionUpdate ===", "commission", cfg.Commission)
	Commission := cfg.Commission
	v.handleType1Msg(cfg, v.validatorTo, nil, v.validatorAbi, "setNextCommissionUpdate", big.NewInt(0).SetUint64(Commission))
	return nil
}

func (v *Voter) updateCommission(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== updateCommission ===")
	v.handleType1Msg(cfg, v.validatorTo, nil, v.validatorAbi, "updateCommission")
	return nil
}

func (v *Voter) setTargetValidatorEpochPayment(_ *cli.Context, cfg *define.Config) error {
	value := new(big.Int).Mul(big.NewInt(int64(cfg.Value)), big.NewInt(1e18))
	log.Info("=== setTargetValidatorEpochPayment ===", "admin", cfg.From.String())
	v.handleType1Msg(cfg, v.epochRewardsTo, nil, v.epochRewardsAbi, "setTargetValidatorEpochPayment", value)
	return nil
}

func (v *Voter) setEpochMaintainerPaymentFraction(_ *cli.Context, cfg *define.Config) error {
	fixed := fixed.MustNew(cfg.Fixed).BigInt()
	log.Info("=== setEpochMaintainerPaymentFraction ===", "admin", cfg.From.String())
	v.handleType1Msg(cfg, v.epochRewardsTo, nil, v.epochRewardsAbi, "setEpochMaintainerPaymentFraction", fixed)
	return nil
}

func (v *Voter) setMgrMaintainerAddress(_ *cli.Context, cfg *define.Config) error {
	address := cfg.TargetAddress
	log.Info("=== setMgrMaintainerAddress ===", "admin", cfg.From.String())
	v.handleType1Msg(cfg, v.epochRewardsTo, nil, v.epochRewardsAbi, "setMgrMaintainerAddress", address)
	return nil
}

func (v *Voter) getMgrMaintainerAddress(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== getMgrMaintainerAddress ===", "admin", cfg.From.String())
	var ret interface{}
	v.handleType3Msg(cfg, &ret, v.epochRewardsTo, nil, v.epochRewardsAbi, "getMgrMaintainerAddress")
	result := ret
	log.Info("getMgrMaintainerAddress", "address ", result)
	return nil
}

func ConvertToFraction(num interface{}) string {
	s := num.(*big.Int)
	p := decimal.Precision(24)
	b, err := decimal.ToJSON(s, p)
	if err != nil {
		log.Error("ConvertToFraction", "err", err)
	}
	str := (string)(b)
	str = strings.Replace(str, "\"", "", -1)
	return str
}

func (v *Voter) _getRegisteredValidatorSigners(cfg *define.Config) []common.Address {
	var ValidatorSigners interface{}
	v.handleType3Msg(cfg, &ValidatorSigners, v.validatorTo, nil, v.validatorAbi, "getRegisteredValidatorSigners")
	return ValidatorSigners.([]common.Address)
}

func (v *Voter) getGL(cfg *define.Config, target common.Address) (common.Address, common.Address, error) {
	type ret struct {
		Validators interface{} // indexed
		Values     interface{}
	}
	var t ret
	electionAddress := cfg.ElectionParameters.ElectionAddress
	abiElection := cfg.ElectionParameters.ElectionABI
	f := func(output []byte) {
		err := abiElection.UnpackIntoInterface(&t, "getTotalVotesForEligibleValidators", output)
		if err != nil {
			log.Error("getTotalVotesForEligibleValidators setLesserGreater", "err", err)
			os.Exit(1)
		}
	}
	v.handleType4Msg(cfg, f, electionAddress, nil, abiElection, "getTotalVotesForEligibleValidators")
	validators := (t.Validators).([]common.Address)
	votes := (t.Values).([]*big.Int)
	voteTotals := make([]voteTotal, len(validators))
	for i, addr := range validators {
		voteTotals[i] = voteTotal{addr, votes[i]}
	}
	voteNum := new(big.Int).Mul(cfg.VoteNum, big.NewInt(1e18))
	for _, voteTotal := range voteTotals {
		if bytes.Equal(voteTotal.Validator.Bytes(), target.Bytes()) {
			if big.NewInt(0).Cmp(voteNum) < 0 {
				voteTotal.Value.Add(voteTotal.Value, voteNum)
			}
			// Sorting in descending order is necessary to match the order on-chain.
			sort.SliceStable(voteTotals, func(j, k int) bool {
				return voteTotals[j].Value.Cmp(voteTotals[k].Value) > 0
			})

			lesser := params.ZeroAddress
			greater := params.ZeroAddress
			for j, voteTotal := range voteTotals {
				if voteTotal.Validator == target {
					if j > 0 {
						greater = voteTotals[j-1].Validator
					}
					if j+1 < len(voteTotals) {
						lesser = voteTotals[j+1].Validator
					}
					break
				}
			}
			return greater, lesser, nil
		}
	}
	return params.ZeroAddress, params.ZeroAddress, define.NoTargetValidatorError
}

func (v *Voter) getGLSub(cfg *define.Config, SubValue *big.Int, target common.Address) (common.Address, common.Address, error) {
	type ret struct {
		Validators interface{} // indexed
		Values     interface{}
	}
	var t ret
	f := func(output []byte) {
		err := v.electionAbi.UnpackIntoInterface(&t, "getTotalVotesForEligibleValidators", output)
		if err != nil {
			log.Error("getTotalVotesForEligibleValidators setLesserGreater", "err", err)
			os.Exit(1)
		}
	}
	v.handleType4Msg(cfg, f, v.electionTo, nil, v.electionAbi, "getTotalVotesForEligibleValidators")
	validators := (t.Validators).([]common.Address)
	votes := (t.Values).([]*big.Int)
	voteTotals := make([]voteTotal, len(validators))
	for i, addr := range validators {
		voteTotals[i] = voteTotal{addr, votes[i]}
	}
	for _, voteTotal := range voteTotals {
		if bytes.Equal(voteTotal.Validator.Bytes(), target.Bytes()) {
			if big.NewInt(0).Cmp(SubValue) < 0 {
				if voteTotal.Value.Cmp(SubValue) > 0 {
					voteTotal.Value.Sub(voteTotal.Value, SubValue)
				} else {
					return params.ZeroAddress, params.ZeroAddress, define.BigSubValue
				}
			}
			// Sorting in descending order is necessary to match the order on-chain.

			sort.SliceStable(voteTotals, func(j, k int) bool {
				return voteTotals[j].Value.Cmp(voteTotals[k].Value) > 0
			})

			lesser := params.ZeroAddress
			greater := params.ZeroAddress
			for j, voteTotal := range voteTotals {
				if voteTotal.Validator == target {
					if j > 0 {
						greater = voteTotals[j-1].Validator
					}
					if j+1 < len(voteTotals) {
						lesser = voteTotals[j+1].Validator
					}
					break
				}
			}
			return greater, lesser, nil
		}
	}
	return params.ZeroAddress, params.ZeroAddress, define.NoTargetValidatorError
}

func (v *Voter) _getValidatorsVotedForByAccount(cfg *define.Config, target common.Address) []common.Address {
	var ret interface{}
	v.handleType3Msg(cfg, &ret, v.electionTo, nil, v.electionAbi, "getValidatorsVotedForByAccount", target)
	result := ret.([]common.Address)
	return result
}

func (v *Voter) GetIndex(target common.Address, list []common.Address) (*big.Int, error) {
	for index, v := range list {
		if bytes.Equal(target.Bytes(), v.Bytes()) {
			return big.NewInt(int64(index)), nil
		}
	}
	return nil, define.GetIndexError
}
