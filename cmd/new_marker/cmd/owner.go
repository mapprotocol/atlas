package cmd

import (
	"gopkg.in/urfave/cli.v1"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/new_marker/define"
	"github.com/mapprotocol/atlas/cmd/new_marker/mapprotocol"
	"github.com/mapprotocol/atlas/helper/decimal/fixed"
)

type Owner struct {
	*base
	account                                                                 *Account
	validator                                                               *Validator
	lockGoldTo, electionTo, validatorTo, goldTokenTo, epochRewardsTo        common.Address
	lockedGoldAbi, electionAbi, validatorAbi, goldTokenAbi, epochRewardsAbi *abi.ABI
}

func NewOwner() *Owner {
	return &Owner{
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

func (o *Owner) setImplementation(_ *cli.Context, cfg *define.Config) error {
	implementation := cfg.ImplementationAddress
	ContractAddress := cfg.Address
	ProxyAbi := mapprotocol.AbiFor("Proxy")
	log.Info("=== setImplementation ===", "from", cfg.From.String())
	o.handleType1Msg(cfg, ContractAddress, nil, ProxyAbi, "_setImplementation", implementation)
	return nil
}

func (o *Owner) getContractOwner(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== getOwner ===", "admin", cfg.From.String())
	var ret interface{}
	o.handleType3Msg(cfg, &ret, cfg.Address, nil, o.validatorAbi, "owner") // todo method ???
	result := ret
	log.Info("getOwner", "Owner ", result)
	return nil
}

func (o *Owner) setContractOwner(_ *cli.Context, cfg *define.Config) error {
	NewOwner := cfg.Address
	ContractAddress := cfg.Address // 代理地址
	abiValidators := cfg.ValidatorParameters.ValidatorABI
	log.Info("ProxyAddress", "Address", ContractAddress, "NewOwner", NewOwner.String())
	log.Info("=== setOwner ===", "from", cfg.From.String())
	o.handleType1Msg(cfg, ContractAddress, nil, abiValidators, "transferOwnership", NewOwner)
	return nil
}

func (o *Owner) getProxyContractOwner(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== getOwner ===", "from", cfg.From.String())
	var ret interface{}
	ProxyAbi := mapprotocol.AbiFor("Proxy")
	o.handleType3Msg(cfg, &ret, cfg.Address, nil, ProxyAbi, "_getOwner")
	result := ret
	log.Info("getOwner", "Owner ", result)
	return nil
}

func (o *Owner) setProxyContractOwner(_ *cli.Context, cfg *define.Config) error {
	NewOwner := cfg.Address
	ContractAddress := cfg.Address //代理地址
	log.Info("ProxyAddress", "Address", ContractAddress, "NewOwner", NewOwner.String())
	ProxyAbi := mapprotocol.AbiFor("Proxy") //代理ABI
	log.Info("=== setOwner ===", "from", cfg.From.String())
	o.handleType1Msg(cfg, ContractAddress, nil, ProxyAbi, "_transferOwnership", NewOwner)
	return nil
}

func (o *Owner) setValidatorLockedGoldRequirements(_ *cli.Context, cfg *define.Config) error {
	value := new(big.Int).Mul(cfg.Value, big.NewInt(1e18))
	duration := big.NewInt(cfg.Duration)
	log.Info("=== setValidatorLockedGoldRequirements ===", "from", cfg.From.String())
	o.handleType1Msg(cfg, o.validatorTo, nil, o.validatorAbi, "setValidatorLockedGoldRequirements", value, duration)
	return nil
}

func (o *Owner) setTargetValidatorEpochPayment(_ *cli.Context, cfg *define.Config) error {
	value := new(big.Int).Mul(cfg.Value, big.NewInt(1e18))
	log.Info("=== setTargetValidatorEpochPayment ===", "admin", cfg.From.String())
	o.handleType1Msg(cfg, o.epochRewardsTo, nil, o.epochRewardsAbi, "setTargetValidatorEpochPayment", value) // todo method
	return nil
}

func (o *Owner) setEpochMaintainerPaymentFraction(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== setEpochMaintainerPaymentFraction ===", "from", cfg.From.String())
	o.handleType1Msg(cfg, o.epochRewardsTo, nil, o.epochRewardsAbi, "setEpochMaintainerPaymentFraction", fixed.MustNew(cfg.Fixed).BigInt())
	return nil
}

func (o *Owner) getMgrMaintainerAddress(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== getMgrMaintainerAddress ===", "from", cfg.From.String())
	var ret interface{}
	o.handleType3Msg(cfg, &ret, o.epochRewardsTo, nil, o.epochRewardsAbi, "getMgrMaintainerAddress")
	result := ret
	log.Info("getMgrMaintainerAddress", "address ", result)
	return nil
}

func (o *Owner) setMgrMaintainerAddress(_ *cli.Context, cfg *define.Config) error {
	address := cfg.Address
	log.Info("=== setMgrMaintainerAddress ===", "from", cfg.From.String())
	o.handleType1Msg(cfg, o.epochRewardsTo, nil, o.epochRewardsAbi, "setMgrMaintainerAddress", address)
	return nil
}
