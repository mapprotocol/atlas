package genesis

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/core/rawdb"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/core/vm/runtime"
	"github.com/mapprotocol/atlas/marker/contract"
	"github.com/mapprotocol/atlas/marker/env"
	"github.com/mapprotocol/atlas/params"
)

var (
	proxyByteCode = common.Hex2Bytes("60806040526004361061004a5760003560e01c806303386ba3146101e757806342404e0714610280578063bb913f41146102d7578063d29d44ee14610328578063f7e6af8014610379575b6000600160405180807f656970313936372e70726f78792e696d706c656d656e746174696f6e00000000815250601c019050604051809103902060001c0360001b9050600081549050600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161415610136576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260158152602001807f4e6f20496d706c656d656e746174696f6e20736574000000000000000000000081525060200191505060405180910390fd5b61013f816103d0565b6101b1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f496e76616c696420636f6e74726163742061646472657373000000000000000081525060200191505060405180910390fd5b60405136810160405236600082376000803683855af43d604051818101604052816000823e82600081146101e3578282f35b8282fd5b61027e600480360360408110156101fd57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019064010000000081111561023a57600080fd5b82018360208201111561024c57600080fd5b8035906020019184600183028401116401000000008311171561026e57600080fd5b909192939192939050505061041b565b005b34801561028c57600080fd5b506102956105c1565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b3480156102e357600080fd5b50610326600480360360208110156102fa57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061060d565b005b34801561033457600080fd5b506103776004803603602081101561034b57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506107bd565b005b34801561038557600080fd5b5061038e610871565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b60008060007fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a47060001b9050833f915080821415801561041257506000801b8214155b92505050919050565b610423610871565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146104c3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f73656e64657220776173206e6f74206f776e657200000000000000000000000081525060200191505060405180910390fd5b6104cc8361060d565b600060608473ffffffffffffffffffffffffffffffffffffffff168484604051808383808284378083019250505092505050600060405180830381855af49150503d8060008114610539576040519150601f19603f3d011682016040523d82523d6000602084013e61053e565b606091505b508092508193505050816105ba576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601e8152602001807f696e697469616c697a6174696f6e2063616c6c6261636b206661696c6564000081525060200191505060405180910390fd5b5050505050565b600080600160405180807f656970313936372e70726f78792e696d706c656d656e746174696f6e00000000815250601c019050604051809103902060001c0360001b9050805491505090565b610615610871565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146106b5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f73656e64657220776173206e6f74206f776e657200000000000000000000000081525060200191505060405180910390fd5b6000600160405180807f656970313936372e70726f78792e696d706c656d656e746174696f6e00000000815250601c019050604051809103902060001c0360001b9050610701826103d0565b610773576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f496e76616c696420636f6e74726163742061646472657373000000000000000081525060200191505060405180910390fd5b8181558173ffffffffffffffffffffffffffffffffffffffff167fab64f92ab780ecbf4f3866f57cee465ff36c89450dcce20237ca7a8d81fb7d1360405160405180910390a25050565b6107c5610871565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610865576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f73656e64657220776173206e6f74206f776e657200000000000000000000000081525060200191505060405180910390fd5b61086e816108bd565b50565b600080600160405180807f656970313936372e70726f78792e61646d696e000000000000000000000000008152506013019050604051809103902060001c0360001b9050805491505090565b600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161415610960576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260118152602001807f6f776e65722063616e6e6f74206265203000000000000000000000000000000081525060200191505060405180910390fd5b6000600160405180807f656970313936372e70726f78792e61646d696e000000000000000000000000008152506013019050604051809103902060001c0360001b90508181558173ffffffffffffffffffffffffffffffffffffffff167f50146d0e3c60aa1d17a70635b05494f864e86144a2201275021014fbf08bafe260405160405180910390a2505056fea165627a7a72305820f4f741dbef8c566cb1690ae708b8ef1113bdb503225629cc1f9e86bd47efd1a40029")
)

// deployContext context for deployment
type deployContext struct {
	genesisConfig *Config
	accounts      *env.AccountsConfig
	statedb       *state.StateDB
	runtimeConfig *runtime.Config
	truffleReader contract.TruffleReader
	logger        log.Logger
}

// Helper function to reduce boilerplate, limited to this package on purpose
// Like big.NewInt() except it takes uint64 instead of int64
func newBigInt(x uint64) *big.Int { return new(big.Int).SetUint64(x) }

func generateGenesisState(accounts *env.AccountsConfig, cfg *Config, buildPath string) (chain.GenesisAlloc, error) {
	deployment := newDeployment(cfg, accounts, buildPath)
	return deployment.deploy()
}

// NewDeployment generates a new deployment
func newDeployment(genesisConfig *Config, accounts *env.AccountsConfig, buildPath string) *deployContext {
	logger := log.New("obj", "deployment")
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)

	adminAddress := AdminAddr

	logger.Info("New deployment", "admin_address", adminAddress.Hex())
	return &deployContext{
		genesisConfig: genesisConfig,
		accounts:      accounts,
		logger:        logger,
		statedb:       statedb,
		truffleReader: contract.NewTruffleReader(buildPath),
		runtimeConfig: &runtime.Config{
			ChainConfig: genesisConfig.ChainConfig(),
			Origin:      adminAddress,
			State:       statedb,
			GasLimit:    1000000000000000,
			GasPrice:    big.NewInt(0),
			Value:       big.NewInt(0),
			Time:        newBigInt(genesisConfig.GenesisTimestamp),
			Coinbase:    adminAddress,
			BlockNumber: newBigInt(0),
			EVMConfig: vm.Config{
				Tracer: nil,
				Debug:  false,
			},
		},
	}

}

// Deploy runs the deployment
func (ctx *deployContext) deploy() (chain.GenesisAlloc, error) {
	ctx.fundAdminAccount()

	deploySteps := [](func() error){
		ctx.deployLibraries,
		// 01 Registry
		ctx.deployRegistry,

		// 02 GoldToken
		ctx.deployGoldToken,

		// 03 Accounts
		ctx.deployAccounts,

		// 04 LockedGold
		ctx.deployLockedGold,

		// 05 Validators
		ctx.deployValidators,

		// 06 Election
		ctx.deployElection,
		//
		// 07 EpochRewards
		ctx.deployEpochRewards,

		// 08 deployRandom
		ctx.deployRandom,

		// 09 BlockchainParameters
		ctx.deployBlockchainParameters,

		// 10 Elect Validators
		ctx.electValidators,
	}

	logger := ctx.logger.New()

	for i, step := range deploySteps {
		logger.Info("Running deploy step", "number", i)
		if err := step(); err != nil {
			return nil, err
		}
	}

	// Flush Changes
	_, err := ctx.statedb.Commit(true)
	if err != nil {
		return nil, err
	}
	ctx.statedb.IntermediateRoot(true)

	if err = ctx.verifyState(); err != nil {
		return nil, err
	}
	opts := &state.DumpConfig{
		SkipCode:          false,
		SkipStorage:       false,
		OnlyWithAddresses: true,
	}
	dump := ctx.statedb.RawDump(opts).Accounts
	genesisAlloc := make(map[common.Address]chain.GenesisAccount)
	for acc, dumpAcc := range dump {
		var account chain.GenesisAccount

		if dumpAcc.Balance != "" {
			account.Balance, _ = new(big.Int).SetString(dumpAcc.Balance, 10)
		}

		if dumpAcc.Code != nil {
			account.Code = dumpAcc.Code
		}

		if len(dumpAcc.Storage) > 0 {
			account.Storage = make(map[common.Hash]common.Hash)
			for k, v := range dumpAcc.Storage {
				account.Storage[k] = common.HexToHash(v)
			}
		}

		genesisAlloc[acc] = account

	}

	return genesisAlloc, nil
}

// Initialize AdminAddr  You can add MAP to the designated account here
func (ctx *deployContext) fundAdminAccount() {
}

func (ctx *deployContext) deployLibraries() error {
	for _, name := range env.Libraries() {
		bytecode := ctx.truffleReader.MustReadDeployedBytecodeFor(name)
		ctx.statedb.SetCode(env.MustLibraryAddressFor(name), bytecode)
	}
	return nil
}

// deployProxiedContract will deploy proxied contract
// It will deploy the proxy contract, the impl contract, and initialize both
func (ctx *deployContext) deployProxiedContract(name string, initialize func(contract *contract.EVMBackend) error) error {
	proxyAddress := env.MustProxyAddressFor(name)
	implAddress := env.MustImplAddressFor(name)
	bytecode := ctx.truffleReader.MustReadDeployedBytecodeFor(name)

	logger := ctx.logger.New("contract", name)
	logger.Info("Start Deploy of Proxied Contract", "proxyAddress", proxyAddress.Hex(), "implAddress", implAddress.Hex())

	logger.Info("Deploy Proxy")
	ctx.statedb.SetCode(proxyAddress, proxyByteCode)
	ctx.statedb.SetState(proxyAddress, params.ProxyOwnerStorageLocation, AdminAddr.Hash())

	//fmt.Println("AdminAddr.Hash()",AdminAddr.Hash())

	logger.Info("Deploy Implementation")
	ctx.statedb.SetCode(implAddress, bytecode)

	logger.Info("Set proxy implementation")
	proxyContract := ctx.proxyContract(name)
	if err := proxyContract.SimpleCall("_setImplementation", implAddress); err != nil {
		return err
	}

	logger.Info("Initialize Contract")
	if err := initialize(ctx.contract(name)); err != nil {
		return err
	}

	return nil
}

// deployCoreContract will deploy a contract + proxy, and add it to the registry
func (ctx *deployContext) deployCoreContract(name string, initialize func(contract *contract.EVMBackend) error) error {
	if err := ctx.deployProxiedContract(name, initialize); err != nil {
		return err
	}

	proxyAddress := env.MustProxyAddressFor(name)
	ctx.logger.Info("Add entry to registry", "name", name, "address", proxyAddress)
	if err := ctx.contract("Registry").SimpleCall("setAddressFor", name, proxyAddress); err != nil {
		return err
	}

	return nil
}

func (ctx *deployContext) deployRegistry() error {
	return ctx.deployCoreContract("Registry", func(contract *contract.EVMBackend) error {
		return contract.SimpleCall("initialize")
	})
}

func (ctx *deployContext) deployBlockchainParameters() error {
	return ctx.deployCoreContract("BlockchainParameters", func(contract *contract.EVMBackend) error {
		return contract.SimpleCall("initialize",
			big.NewInt(ctx.genesisConfig.Blockchain.Version.Major),
			big.NewInt(ctx.genesisConfig.Blockchain.Version.Minor),
			big.NewInt(ctx.genesisConfig.Blockchain.Version.Patch),
			newBigInt(ctx.genesisConfig.Blockchain.GasForNonGoldCurrencies),
			newBigInt(ctx.genesisConfig.Blockchain.BlockGasLimit),
			newBigInt(ctx.genesisConfig.Istanbul.LookbackWindow),
		)
	})
}

func (ctx *deployContext) addSlasher(slasherName string) error {
	ctx.logger.Info("Adding new slasher", "slasher", slasherName)
	return ctx.contract("LockedGold").SimpleCall("addSlasher", slasherName)
}

func (ctx *deployContext) deployGoldToken() error {
	err := ctx.deployCoreContract("GoldToken", func(contract *contract.EVMBackend) error {
		return contract.SimpleCall("initialize", env.MustProxyAddressFor("Registry"))
	})
	if err != nil {
		return err
	}

	//for _, bal := range ctx.genesisConfig.GoldToken.InitialBalances {
	//	ctx.statedb.SetBalance(bal.Account, bal.Amount)
	//}

	return nil
}

func (ctx *deployContext) deployEpochRewards() error {
	err := ctx.deployCoreContract("EpochRewards", func(contract *contract.EVMBackend) error {
		return contract.SimpleCall("initialize",
			env.MustProxyAddressFor("Registry"),
			ctx.genesisConfig.EpochRewards.MaxEpochPayment,
			ctx.genesisConfig.EpochRewards.CommunityRewardFraction.BigInt(),
			ctx.genesisConfig.EpochRewards.EpochRelayerPaymentFraction.BigInt(),
			ctx.genesisConfig.EpochRewards.CommunityPartner,
		)
	})
	if err != nil {
		return err
	}
	return nil
}

func (ctx *deployContext) deployAccounts() error {
	return ctx.deployCoreContract("Accounts", func(contract *contract.EVMBackend) error {
		return contract.SimpleCall("initialize", env.MustProxyAddressFor("Registry"))
	})
}

func (ctx *deployContext) deployRandom() error {
	return ctx.deployCoreContract("Random", func(contract *contract.EVMBackend) error {
		return contract.SimpleCall("initialize",
			newBigInt(ctx.genesisConfig.Random.RandomnessBlockRetentionWindow),
		)
	})
}

func (ctx *deployContext) deployLockedGold() error {
	return ctx.deployCoreContract("LockedGold", func(contract *contract.EVMBackend) error {
		return contract.SimpleCall("initialize",
			env.MustProxyAddressFor("Registry"),
			newBigInt(ctx.genesisConfig.LockedGold.UnlockingPeriod),
		)
	})
}

func (ctx *deployContext) deployValidators() error {
	return ctx.deployCoreContract("Validators", func(contract *contract.EVMBackend) error {
		return contract.SimpleCall("initialize",
			env.MustProxyAddressFor("Registry"),
			ctx.genesisConfig.Validators.ValidatorLockedGoldRequirements.Value,
			newBigInt(ctx.genesisConfig.Validators.ValidatorLockedGoldRequirements.Duration),
			newBigInt(ctx.genesisConfig.Validators.ValidatorScoreExponent),
			ctx.genesisConfig.Validators.ValidatorScoreAdjustmentSpeed.BigInt(),
			newBigInt(ctx.genesisConfig.Validators.SlashingPenaltyResetPeriod),
			newBigInt(ctx.genesisConfig.Validators.CommissionUpdateDelay),
			ctx.genesisConfig.Validators.PledgeMultiplierInReward.BigInt(),
			newBigInt(ctx.genesisConfig.Validators.DowntimeGracePeriod),
		)
	})
}

func (ctx *deployContext) deployElection() error {
	return ctx.deployCoreContract("Election", func(contract *contract.EVMBackend) error {
		return contract.SimpleCall("initialize",
			env.MustProxyAddressFor("Registry"),
			newBigInt(ctx.genesisConfig.Election.MinElectableValidators),
			newBigInt(ctx.genesisConfig.Election.MaxElectableValidators),
			ctx.genesisConfig.Election.MaxVotesPerAccount,
			ctx.genesisConfig.Election.ElectabilityThreshold.BigInt(),
		)
	})
}

func (ctx *deployContext) createAccounts(accs []AccoutInfo, namePrefix string) error {
	accounts := ctx.contract("Accounts")

	for i, acc := range accs {
		name := fmt.Sprintf("%s %03d", namePrefix, i)
		ctx.logger.Info("Create account", "address", acc.Address, "name", name)

		if err := accounts.SimpleCallFrom(acc.getAddress(), "createAccount"); err != nil {
			return err
		}

		if err := accounts.SimpleCallFrom(acc.getAddress(), "setName", name); err != nil {
			return err
		}

		if err := accounts.SimpleCallFrom(acc.getAddress(), "setAccountDataEncryptionKey", acc.PublicKey()); err != nil {
			return err
		}

		// Missing: authorizeAttestationSigner
	}
	return nil
}

func (ctx *deployContext) registerValidators() error {
	validatorAccounts := ValidatorsAT
	requiredAmount := ctx.genesisConfig.Validators.ValidatorLockedGoldRequirements.Value
	if err := ctx.createAccounts(validatorAccounts, "validator"); err != nil {
		return err
	}

	lockedGold := ctx.contract("LockedGold")
	validators := ctx.contract("Validators")
	commission := ctx.genesisConfig.Validators.Commission
	for validatorIdx, validator := range validatorAccounts {
		address := validator.getAddress()
		logger := ctx.logger.New("validator", address)
		prevValidatorAddress := params.ZeroAddress
		if validatorIdx > 0 {
			prevValidatorAddress = validatorAccounts[validatorIdx-1].getAddress()
		}
		ctx.statedb.AddBalance(address, requiredAmount)

		logger.Info("Lock validator gold", "amount", requiredAmount)
		if _, err := lockedGold.Call(contract.CallOpts{Origin: address, Value: requiredAmount}, "lock"); err != nil {
			return err
		}

		logger.Info("Register validator")
		blsPub, err := validator.BLSPublicKey()
		if err != nil {
			return err
		}
		blsG1Pub, err := validator.BLSG1PublicKey()
		if err != nil {
			return err
		}
		// remove the 0x04 prefix from the pub key (we need the 64 bytes variant)
		pubKey := validator.PublicKey()[1:]
		//err = validators.SimpleCallFrom(address, "registerValidatorPre", blsPub[:], blsG1Pub[:], validator.MustBLSProofOfPossession(), pubKey)
		//if err != nil {
		//	return err
		//}
		validatorParams := [4][]byte{blsPub[:], blsG1Pub[:], validator.MustBLSProofOfPossession()[:], pubKey[:]}
		//err = validators.SimpleCallFrom(address, "registerValidator", commission, params.ZeroAddress, prevValidatorAddress, blsPub[:], blsG1Pub[:], validator.MustBLSProofOfPossession(), pubKey)
		err = validators.SimpleCallFrom(address, "registerValidator", commission, params.ZeroAddress, prevValidatorAddress, validatorParams)
		if err != nil {
			return err
		}
	}
	return nil
}

//each validator votes for themselves.
func (ctx *deployContext) voteForValidators() error {
	election := ctx.contract("Election")

	// value previously locked on registerValidatorGroups()
	lockedGoldOnValidator := ctx.genesisConfig.Validators.ValidatorLockedGoldRequirements.Value

	// current validator order (see `addFirstMember` on addValidatorsToGroup) is:
	// [ validatorZero, validatorOne, ..., lastvalidator]

	// each validator votes for themselves.
	// each validator votes the SAME AMOUNT
	// each validator starts with 0 votes

	// so, everytime we vote, that validator becomes the one with most votes (most or equal)
	// hence, we use:
	//    greater = zero (we become the one with most votes)
	//    lesser = currentLeader
	validatorAddress := ValidatorsAT[0].getAddress()
	// special case: only one validator (no lesser or greater)
	if len(ValidatorsAT) == 1 {
		voterAddress := ValidatorsAT[0].getAddress()
		ctx.logger.Info("Vote for validator", "validator", validatorAddress, "amount", lockedGoldOnValidator)
		return election.SimpleCallFrom(voterAddress, "vote", validatorAddress, lockedGoldOnValidator, params.ZeroAddress, params.ZeroAddress)
	}

	// first to vote is validator 0, which is already the leader. Hence lesser should go to validator 1
	currentLeader := ValidatorsAT[1].getAddress()
	for _, voter := range ValidatorsAT {
		validatorAddress = voter.getAddress()
		ctx.logger.Info("Vote for validator", "voter", voter.Address, "validator", validatorAddress, "amount", lockedGoldOnValidator)
		if err := election.SimpleCallFrom(voter.getAddress(), "vote", validatorAddress, lockedGoldOnValidator, currentLeader, params.ZeroAddress); err != nil {
			return err
		}
		// we now become the currentLeader
		currentLeader = voter.getAddress()
	}
	return nil
}

func (ctx *deployContext) electValidators() error {
	if err := ctx.registerValidators(); err != nil {
		return err
	}

	if err := ctx.voteForValidators(); err != nil {
		return err
	}

	// TODO: Config checks
	// check that we have enough validators (validators >= election.minElectableValidators)
	// check that validatorsPerGroup is <= valdiators.maxGroupSize

	return nil
}

func (ctx *deployContext) contract(contractName string) *contract.EVMBackend {
	return contract.CoreContract(ctx.runtimeConfig, contractName, env.MustProxyAddressFor(contractName))
}

func (ctx *deployContext) proxyContract(contractName string) *contract.EVMBackend {
	return contract.ProxyContract(ctx.runtimeConfig, contractName, env.MustProxyAddressFor(contractName))
}

func (ctx *deployContext) verifyState() error {
	snapshotVersion := ctx.statedb.Snapshot()
	defer ctx.statedb.RevertToSnapshot(snapshotVersion)
	return nil
}
