package cmd

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/new_marker/define"
	"github.com/mapprotocol/atlas/cmd/new_marker/mapprotocol"
	"github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	"os"
	"sort"
)

type Voter struct {
	*base
	account                         *Account
	validator                       *Validator
	to, lockGoldTo, electionTo      common.Address
	abi, lockedGoldAbi, electionAbi *abi.ABI
}

func NewVoter() *Voter {
	return &Voter{
		base:      newBase(),
		account:   NewAccount(),
		validator: NewValidator(),
		//to:            mapprotocol.MustProxyAddressFor("Voters"),
		//abi:           mapprotocol.AbiFor("Voters"),
		//lockGoldTo:    mapprotocol.MustProxyAddressFor("LockedGold"),
		//lockedGoldAbi: mapprotocol.AbiFor("LockedGold"),
		electionTo:  mapprotocol.MustProxyAddressFor("Election"),
		electionAbi: mapprotocol.AbiFor("Election"),
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
	v.handleType1Msg(cfg, v.electionTo, nil, v.electionAbi, "revokePending", _params)
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
	v.handleType1Msg(cfg, v.electionTo, nil, v.electionAbi, "revokeActive", _params)
	return nil
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
