package cmd

import (
	"bytes"
	"context"
	"math/big"
	"os"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/new_marker/define"
	"github.com/mapprotocol/atlas/cmd/new_marker/mapprotocol"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/helper/decimal"
	"github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
)

type Voter struct {
	*base
	account                                                   *Account
	validator                                                 *Validator
	lockGoldTo, electionTo, goldTokenTo, epochRewardsTo       common.Address
	lockedGoldAbi, electionAbi, goldTokenAbi, epochRewardsAbi *abi.ABI
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
		goldTokenTo:     mapprotocol.MustProxyAddressFor("GoldToken"),
		goldTokenAbi:    mapprotocol.AbiFor("GoldToken"),
		epochRewardsTo:  mapprotocol.MustProxyAddressFor("EpochRewards"),
		epochRewardsAbi: mapprotocol.AbiFor("EpochRewards"),
	}
}

func (v *Voter) Vote(_ *cli.Context, cfg *define.Config) error {
	greater, lesser, err := v.getGL(cfg, cfg.Address)
	if err != nil {
		log.Error("vote", "err", err)
		return err
	}
	amount := new(big.Int).Mul(cfg.Value, big.NewInt(1e18))
	log.Info("=== vote Validator ===", "from", cfg.From, "voteTargetValidator", cfg.Address.String(), "vote MAP Num", cfg.Value.String())
	v.handleType1Msg(cfg, v.electionTo, nil, v.electionAbi, "vote", cfg.Address, amount, lesser, greater)
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
	v.handleType1Msg(cfg, v.electionTo, nil, v.electionAbi, "activate", cfg.Address)
	return nil
}

func (v *Voter) RevokePending(_ *cli.Context, cfg *define.Config) error {
	validator := cfg.Address
	LockedNum := new(big.Int).Mul(cfg.Value, big.NewInt(1e18))

	greater, lesser, _ := v.getGLSub(cfg, LockedNum, validator)
	list := v._getValidatorsVotedForByAccount(cfg, cfg.From)
	index, err := v.GetIndex(validator, list)
	if err != nil {
		log.Crit("revokePending", "err", err)
	}
	//fmt.Println("=== greater,lesser,index ===", greater, lesser, index)
	_params := []interface{}{validator, LockedNum, lesser, greater, index}
	log.Info("=== revokePending ===", "from", cfg.From)
	v.handleType1Msg(cfg, v.electionTo, nil, v.electionAbi, "revokePending", _params...)
	return nil
}

func (v *Voter) RevokeActive(_ *cli.Context, cfg *define.Config) error {
	validator := cfg.Address
	LockedNum := new(big.Int).Mul(cfg.Value, big.NewInt(1e18))
	greater, lesser, _ := v.getGLSub(cfg, LockedNum, validator)

	list := v._getValidatorsVotedForByAccount(cfg, cfg.From)
	index, err := v.GetIndex(validator, list)
	if err != nil {
		log.Crit("revokePending", "err", err)
	}
	_params := []interface{}{validator, LockedNum, lesser, greater, index}
	log.Info("=== revokeActive ===", "from", cfg.From)
	v.handleType1Msg(cfg, v.electionTo, nil, v.electionAbi, "revokeActive", _params...)
	return nil
}

func (v *Voter) LockedMAP(_ *cli.Context, cfg *define.Config) error {
	lockedGold := new(big.Int).Mul(cfg.Value, big.NewInt(1e18))
	log.Info("=== Lock  gold ===")
	log.Info("Lock  gold", "amount", lockedGold.String())
	v.handleType2Msg(cfg, v.lockGoldTo, lockedGold, v.lockedGoldAbi, "lock")
	return nil
}

func (v *Voter) UnlockedMAP(_ *cli.Context, cfg *define.Config) error {
	lockedGold := new(big.Int).Mul(cfg.Value, big.NewInt(1e18))
	log.Info("=== unLock validator gold ===")
	log.Info("unLock validator gold", "amount", lockedGold, "from", cfg.From)
	v.handleType1Msg(cfg, v.lockGoldTo, nil, v.lockedGoldAbi, "unlock", lockedGold)
	return nil
}

func (v *Voter) RelockMAP(_ *cli.Context, cfg *define.Config) error {
	lockedGold := new(big.Int).Mul(cfg.Value, big.NewInt(1e18))
	log.Info("=== relockMAP validator gold ===")
	log.Info("relockMAP validator gold", "amount", lockedGold)
	v.handleType1Msg(cfg, v.lockGoldTo, nil, v.lockedGoldAbi, "relock", cfg.Index, lockedGold)
	return nil
}

func (v *Voter) Withdraw(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== withdraw validator gold ===", "from", cfg.From.String())
	v.handleType1Msg(cfg, v.lockGoldTo, nil, v.lockedGoldAbi, "withdraw", cfg.Index)
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
	log.Info("=== get voter Reward ===", "cur_epoch", Epoch, "epochSize", epochSize, "query first BlockNumber", firstBlock, "query end BlockNumber", endBlock, "validatorContractAddress", electionContractAddress.String(), "from", cfg.From)
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

func (v *Voter) getTotalVotes(_ *cli.Context, cfg *define.Config) error {
	var ret interface{}
	log.Info("=== getTotalVotes ===", "from", cfg.From)
	v.handleType3Msg(cfg, &ret, v.electionTo, nil, v.electionAbi, "getTotalVotes")
	result := ret.(*big.Int)
	log.Info("result", "getTotalVotes", result)
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
	voteNum := new(big.Int).Mul(cfg.Value, big.NewInt(1e18))
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

func (v *Voter) GetIndex(target common.Address, list []common.Address) (*big.Int, error) {
	for index, v := range list {
		if bytes.Equal(target.Bytes(), v.Bytes()) {
			return big.NewInt(int64(index)), nil
		}
	}
	return nil, define.GetIndexError
}

func (v *Voter) _getValidatorsVotedForByAccount(cfg *define.Config, target common.Address) []common.Address {
	var ret interface{}
	v.handleType3Msg(cfg, &ret, v.electionTo, nil, v.electionAbi, "getValidatorsVotedForByAccount", target)
	result := ret.([]common.Address)
	return result
}
