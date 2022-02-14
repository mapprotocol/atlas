package genesis

import (
	"github.com/mapprotocol/atlas/helper/decimal/fixed"
	"math/big"
	"math/rand"

	"github.com/mapprotocol/atlas/marker/env"
	"github.com/mapprotocol/atlas/marker/genesis"
	"github.com/mapprotocol/atlas/params"
)

const (
	Epoch uint64 = 20
)

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

	genesisConfig := genesis.CreateCommonGenesisConfig(env.Config.ChainID, genesis.AdminAT.Address, params.IstanbulConfig{
		Epoch:          Epoch,
		ProposerPolicy: 2,
		LookbackWindow: 3,
		BlockPeriod:    1,
		RequestTimeout: 3000,
	})

	// Add balances to developer accounts
	genesis.FundAccounts(genesisConfig, env.Accounts().DeveloperAccounts())

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
	genesisConfig := genesis.CreateCommonGenesisConfig(env.Config.ChainID, genesis.AdminAT.Address, params.IstanbulConfig{
		Epoch:          1000,
		ProposerPolicy: 2,
		LookbackWindow: 3,
		BlockPeriod:    5,
		RequestTimeout: 3000,
	})

	// 10 billion gas limit, set super high on purpose
	genesisConfig.Blockchain.BlockGasLimit = 1000000000

	// Add balances to developer accounts
	genesis.FundAccounts(genesisConfig, env.Accounts().DeveloperAccounts())

	// Disable gas price min being updated
	genesisConfig.GasPriceMinimum.TargetDensity = fixed.MustNew("0.9999")
	genesisConfig.GasPriceMinimum.AdjustmentSpeed = fixed.MustNew("0")

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
	genesisConfig := genesis.CreateCommonGenesisConfig(env.Config.ChainID, genesis.AdminAT.Address, params.IstanbulConfig{
		Epoch:          Epoch,
		ProposerPolicy: 2,
		LookbackWindow: 3,
		BlockPeriod:    1,
		RequestTimeout: 3000,
	})

	// Add balances to validator accounts instead of developer accounts
	genesis.FundAccounts(genesisConfig, env.Accounts().ValidatorAccounts())

	return genesisConfig, nil
}
