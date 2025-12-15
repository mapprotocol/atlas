package cmd

import (
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"

	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/new_marker/define"
	"github.com/mapprotocol/atlas/cmd/new_marker/mapprotocol"
)

type Viewer struct {
	*base
	electionTo, validatorTo   common.Address
	electionAbi, validatorAbi *abi.ABI
}

func NewViewer() *Viewer {
	return &Viewer{
		base:         newBase(),
		electionTo:   mapprotocol.MustProxyAddressFor("Election"),
		electionAbi:  mapprotocol.AbiFor("Election"),
		validatorTo:  mapprotocol.MustProxyAddressFor("Validators"),
		validatorAbi: mapprotocol.AbiFor("Validators"),
	}
}

func (v *Viewer) GetValidator(_ *cli.Context, cfg *define.Config) error {
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

	log.Info("=== getValidator ===", "from", cfg.From)
	v.handleType4Msg(cfg, f, v.validatorTo, nil, v.validatorAbi, "getValidator", cfg.Address)
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

func (v *Viewer) getNumRegisteredValidators(_ *cli.Context, cfg *define.Config) error {
	var NumValidators interface{}
	v.handleType3Msg(cfg, &NumValidators, v.validatorTo, nil, v.validatorAbi, "getNumRegisteredValidators")
	ret := NumValidators.(*big.Int)
	log.Info("=== result ===", "num", ret.String())
	return nil
}

func (v *Viewer) getTopValidators(_ *cli.Context, cfg *define.Config) error {
	var TopValidators interface{}
	v.handleType3Msg(cfg, &TopValidators, v.validatorTo, nil, v.validatorAbi, "getTopValidators", cfg.Value)
	Validators := TopValidators.([]common.Address)
	for i := 0; i < len(Validators); i++ {
		log.Info("Validator:", "index", i, "addr", Validators[i])
	}
	return nil
}

func (v *Viewer) getValidatorEligibility(_ *cli.Context, cfg *define.Config) error {
	var ret interface{}
	v.handleType3Msg(cfg, &ret, v.electionTo, nil, v.electionAbi, "c", cfg.Address)
	log.Info("=== result ===", "bool", ret.(bool))
	return nil
}

func (v *Viewer) GetActiveVotesForValidator(_ *cli.Context, cfg *define.Config) error {
	var ret interface{}
	log.Info("=== getActiveVotesForValidator ===", "from", cfg.From)
	v.handleType3Msg(cfg, &ret, v.electionTo, nil, v.electionAbi, "getActiveVotesForValidator", cfg.Address)
	log.Info("ActiveVotes", "balance", ret.(*big.Int))
	return nil
}

func (v *Viewer) GetPendingVotersForValidator(_ *cli.Context, cfg *define.Config) error {
	var ret interface{}
	log.Info("=== getPendingVotersForValidator ===", "from", cfg.From)
	v.handleType3Msg(cfg, &ret, v.electionTo, nil, v.electionAbi, "getPendingVotersForValidator", cfg.Address)
	log.Info("getPendingVotersForValidator", "voters", ret.([]common.Address))
	return nil
}

func (v *Viewer) GetPendingInfoForValidator(_ *cli.Context, cfg *define.Config) error {
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
	log.Info("=== getPendingInfoForValidator ===", "from", cfg.From)
	v.handleType4Msg(cfg, f, v.electionTo, nil, v.electionAbi, "pendingInfo", cfg.From, cfg.Address)
	log.Info("getPendingInfoForValidator", "PendingEpoch", Epoch.(*big.Int), "Balance", Value.(*big.Int))
	return nil
}

func (v *Viewer) GetTotalVotesForEligibleValidators(_ *cli.Context, cfg *define.Config) error {
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
	log.Info("=== getTotalVotesForEligibleValidators ===", "from", cfg.From)
	v.handleType4Msg(cfg, f, v.electionTo, nil, v.electionAbi, "getTotalVotesForEligibleValidators")
	Validators := (t.Validators).([]common.Address)
	Values := (t.Values).([]*big.Int)
	for i := 0; i < len(Validators); i++ {
		log.Info("Validator:", "addr", Validators[i], "vote amount", Values[i])
	}
	return nil
}

func (v *Viewer) GetRegisteredValidatorSigners(_ *cli.Context, cfg *define.Config) error {
	var ValidatorSigners interface{}
	log.Info("==== getRegisteredValidatorSigners ===")
	v.handleType3Msg(cfg, &ValidatorSigners, v.validatorTo, nil, v.validatorAbi, "getRegisteredValidatorSigners")
	Validators := ValidatorSigners.([]common.Address)
	if len(Validators) == 0 {
		log.Info("nil")
	}
	for i := 0; i < len(Validators); i++ {
		log.Info("Validator:", "index", i, "addr", Validators[i])
	}
	return nil
}

func (v *Viewer) getPendingWithdrawals(_ *cli.Context, cfg *define.Config) error {
	type ret []interface{}
	var (
		Values     interface{}
		Timestamps interface{}
	)
	t := ret{&Values, &Timestamps}
	log.Info("=== getPendingWithdrawals ===", "from", cfg.From, "target", cfg.Address.String())
	f := func(output []byte) {
		err := v.lockedGoldAbi.UnpackIntoInterface(&t, "getPendingWithdrawals", output)
		if err != nil {
			log.Error("getPendingWithdrawals", "err", err)
			os.Exit(1)
		}
	}
	v.handleType4Msg(cfg, f, v.lockGoldTo, nil, v.lockedGoldAbi, "getPendingWithdrawals", cfg.Address)
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

func (v *Viewer) balanceOf(_ *cli.Context, cfg *define.Config) error {
	var ret interface{}
	log.Info("=== balanceOf ===", "from", cfg.From)
	v.handleType3Msg(cfg, &ret, v.goldTokenTo, nil, v.goldTokenAbi, "balanceOf", cfg.Address)
	log.Info("=== result ===", "balance", ret.(*big.Int).String())
	return nil
}
