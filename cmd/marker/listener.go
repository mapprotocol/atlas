package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/cmd/marker/config"
	"github.com/mapprotocol/atlas/cmd/marker/connections"
	"github.com/mapprotocol/atlas/params"

	"gopkg.in/urfave/cli.v1"
	"math/big"
)

type Writer interface {
	ResolveMessage(message Message) bool
}
type listener struct {
	cfg    *config.Config
	conn   *ethclient.Client
	writer Writer
	msgCh  chan struct{} // wait for msg handles
}

func NewListener(ctx *cli.Context, config *config.Config) *listener {
	conn, _ := connections.DialConn(ctx)
	return &listener{
		cfg:   config,
		conn:  conn,
		msgCh: make(chan struct{}),
	}
}
func (l *listener) setWriter(w *writer) {
	l.writer = w
}

// waitUntilMsgHandled this function will block untill message is handled
func (l *listener) waitUntilMsgHandled(counter int) {
	log.Debug("waitUntilMsgHandled", "counter", counter)
	for counter > 0 {
		<-l.msgCh
		counter -= 1
	}
}

//---------- validator -----------------
var registerValidatorCommand = cli.Command{
	Name:   "register",
	Usage:  "register validator",
	Action: MigrateFlags(registerValidator),
	Flags:  Flags,
}
var createAccountCommand = cli.Command{
	Name:   "createAccount",
	Usage:  "creat validator account",
	Action: MigrateFlags(createAccount1),
	Flags:  Flags,
}
var lockedMAPCommand = cli.Command{
	Name:   "lockedMAP",
	Usage:  "locked MAP",
	Action: MigrateFlags(lockedMAP),
	Flags:  Flags,
}
var unlockedMAPCommand = cli.Command{
	Name:   "unlockedMAP",
	Usage:  "unlocked MAP",
	Action: MigrateFlags(unlockedMAP),
	Flags:  Flags,
}
var relockMAPCommand = cli.Command{
	Name:   "relockMAP",
	Usage:  "unlocked MAP",
	Action: MigrateFlags(relockMAP),
	Flags:  Flags,
}
var withdrawCommand = cli.Command{
	Name:   "withdraw",
	Usage:  "withdraw MAP",
	Action: MigrateFlags(withdraw),
	Flags:  Flags,
}

//---------- voter -----------------
var voteValidatorCommand = cli.Command{
	Name:   "voteValidator",
	Usage:  "vote validator ",
	Action: MigrateFlags(vote),
	Flags:  Flags,
}
var activateCommand = cli.Command{
	Name:   "activate",
	Usage:  "Converts `account`'s pending votes for `validator` to active votes.",
	Action: MigrateFlags(activate),
	Flags:  Flags,
}

//---------- query -----------------
var queryRegisteredValidatorSignersCommand = cli.Command{
	Name:   "getRegisteredValidatorSigners",
	Usage:  "Registered Validator Signers",
	Action: MigrateFlags(getRegisteredValidatorSigners),
	Flags:  Flags,
}
var queryTopValidatorsCommand = cli.Command{
	Name:   "getTopValidators",
	Usage:  "get Top Group Validators",
	Action: MigrateFlags(getTopValidators),
	Flags:  Flags,
}
var getValidatorEligibilityCommand = cli.Command{
	Name:   "getValidatorEligibility",
	Usage:  "Judge whether the verifier`s Eligibility",
	Action: MigrateFlags(getValidatorEligibility),
	Flags:  Flags,
}
var getTotalVotesForVCommand = cli.Command{
	Name:   "getTotalVotesForV",
	Usage:  "vote validator ",
	Action: MigrateFlags(getTotalVotesForEligibleValidators),
	Flags:  Flags,
}
var getBalanceCommand = cli.Command{
	Name:   "balanceOf",
	Usage:  "Gets the balance of the specified address using the presently stored inflation factor.",
	Action: MigrateFlags(balanceOf),
	Flags:  Flags,
}

//---------- validator -----------------
func registerValidator(ctx *cli.Context, core *listener) error {
	//---------------------------- create account ----------------------------------
	createAccount(core, "validator")
	//---------------------------- lock ----------------------------------
	lockedMAP(ctx, core)
	//----------------------------- registerValidator ---------------------------------
	log.Info("=== Register validator ===")
	_params := []interface{}{big.NewInt(core.cfg.Commission), core.cfg.Lesser, core.cfg.Greater, core.cfg.PublicKey, core.cfg.BlsPub[:], core.cfg.BLSProof}
	ValidatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	m := NewMessage(SolveType1, core.msgCh, core.cfg, ValidatorAddress, nil, abiValidators, "registerValidator", _params...)
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}
func lockedMAP(_ *cli.Context, core *listener) error {
	groupRequiredGold := params.MustBigInt("10000000000000000000000") // 10k Map
	log.Info("=== Lock  gold ===")
	log.Info("Lock  gold", "amount", groupRequiredGold)
	LockedGoldAddress := core.cfg.LockedGoldParameters.LockedGoldAddress
	abiLockedGold := core.cfg.LockedGoldParameters.LockedGoldABI
	m := NewMessage(SolveType1, core.msgCh, core.cfg, LockedGoldAddress, groupRequiredGold, abiLockedGold, "lock")
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)

	return nil
}
func unlockedMAP(_ *cli.Context, core *listener) error {
	groupRequiredGold := params.MustBigInt("10000000000000000000000") // 10k Map
	log.Info("=== unLock validator gold ===")
	log.Info("unLock validator gold", "amount", groupRequiredGold)
	LockedGoldAddress := core.cfg.LockedGoldParameters.LockedGoldAddress
	abiLockedGold := core.cfg.LockedGoldParameters.LockedGoldABI
	m := NewMessage(SolveType1, core.msgCh, core.cfg, LockedGoldAddress, nil, abiLockedGold, "unlock", groupRequiredGold)
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}
func relockMAP(_ *cli.Context, core *listener) error {
	groupRequiredGold := params.MustBigInt("10000000000000000000000") // 10k Map
	log.Info("=== relockMAP validator gold ===")
	log.Info("relockMAP validator gold", "amount", groupRequiredGold)
	LockedGoldAddress := core.cfg.LockedGoldParameters.LockedGoldAddress
	abiLockedGold := core.cfg.LockedGoldParameters.LockedGoldABI
	m := NewMessage(SolveType1, core.msgCh, core.cfg, LockedGoldAddress, nil, abiLockedGold, "relock", groupRequiredGold)
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}
func withdraw(_ *cli.Context, core *listener) error {
	index := big.NewInt(0)
	LockedGoldAddress := core.cfg.LockedGoldParameters.LockedGoldAddress
	abiLockedGold := core.cfg.LockedGoldParameters.LockedGoldABI
	log.Info("=== withdraw validator gold ===")
	m := NewMessage(SolveType1, core.msgCh, core.cfg, LockedGoldAddress, nil, abiLockedGold, "withdraw", index)
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}
func createAccount1(_ *cli.Context, core *listener) error {
	createAccount(core, "validator")
	return nil
}
func createAccount(core *listener, namePrefix string) {
	abiAccounts := core.cfg.AccountsParameters.AccountsABI
	accountsAddress := core.cfg.AccountsParameters.AccountsAddress

	logger := log.New("func", "createAccount")
	logger.Info("Create account", "address", core.cfg.From, "name", namePrefix)
	log.Info("=== create Account ===")
	m := NewMessage(SolveType1, core.msgCh, core.cfg, accountsAddress, nil, abiAccounts, "createAccount")
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)

	log.Info("=== setName name ===")
	m = NewMessage(SolveType1, core.msgCh, core.cfg, accountsAddress, nil, abiAccounts, "setName", namePrefix)
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)

	log.Info("=== setAccountDataEncryptionKey ===")
	m = NewMessage(SolveType1, core.msgCh, core.cfg, accountsAddress, nil, abiAccounts, "setAccountDataEncryptionKey", core.cfg.PublicKey)
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)

}

//---------- voter -----------------
func vote(_ *cli.Context, core *listener) error {
	ElectionsAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElections := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== vote Validator ===")
	amount := new(big.Int).Mul(core.cfg.VoteNum, big.NewInt(1e18))
	m := NewMessage(SolveType1, core.msgCh, core.cfg, ElectionsAddress, nil, abiElections, "vote", core.cfg.From, amount, core.cfg.Lesser, core.cfg.Greater)
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}
func activate(_ *cli.Context, core *listener) error {
	ElectionsAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElections := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== activate validator gold ===", "account.Address", core.cfg.From)
	m := NewMessage(SolveType1, core.msgCh, core.cfg, ElectionsAddress, nil, abiElections, "activate", core.cfg.TargetAddress)
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

//---------- query -----------------
func getRegisteredValidatorSigners(_ *cli.Context, core *listener) error {
	var ValidatorSigners interface{}
	validatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidator := core.cfg.ValidatorParameters.ValidatorABI
	log.Info("==== getRegisteredValidatorSigners ===")
	m := NewMessageRet(SolveType3, core.msgCh, core.cfg, &ValidatorSigners, validatorAddress, nil, abiValidator, "getRegisteredValidatorSigners")
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	fmt.Println("getRegisteredValidatorSigners:", ValidatorSigners.([]common.Address))
	return nil
}
func getTopValidators(ctx *cli.Context, core *listener) error {
	var TopValidators interface{}
	electionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== getTopValidators admin", "obj", core.cfg.From)
	m := NewMessageRet(SolveType3, core.msgCh, core.cfg, &TopValidators, electionAddress, nil, abiElection, "getTopValidators", core.cfg.TopNum)
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	fmt.Println("getTopValidators:", TopValidators.([]common.Address))
	return nil
}
func getTotalVotesForEligibleValidators(ctx *cli.Context, core *listener) error {
	type Ret struct {
		Validators interface{} // indexed
		Values     interface{}
	}
	var t Ret
	electionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== getTotalVotesForEligibleValidators admin", "obj", core.cfg.From)
	m := NewMessageRet(SolveType3, core.msgCh, core.cfg, &t, electionAddress, nil, abiElection, "getTotalVotesForEligibleValidators", core.cfg.TopNum)
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	fmt.Println((t.Validators).([]common.Address))
	fmt.Println((t.Values).([]*big.Int))
	return nil
}
func getValidatorEligibility(ctx *cli.Context, core *listener) error {
	var ret interface{}
	electionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== getValidatorEligibility admin", "obj", core.cfg.From)
	m := NewMessageRet(SolveType3, core.msgCh, core.cfg, &ret, electionAddress, nil, abiElection, "getValidatorEligibility", core.cfg.TargetAddress)
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	fmt.Println(ret.(bool))
	return nil
}
func balanceOf(ctx *cli.Context, core *listener) error {
	var ret interface{}
	GoldTokenAddress := core.cfg.GoldTokenParameters.GoldTokenAddress
	abiGoldToken := core.cfg.GoldTokenParameters.GoldTokenABI
	log.Info("=== balanceOf admin", "obj", core.cfg.From)
	m := NewMessageRet(SolveType3, core.msgCh, core.cfg, &ret, GoldTokenAddress, nil, abiGoldToken, "balanceOf", core.cfg.TargetAddress)
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	fmt.Println(ret.(big.Int))
	return nil
}
func getPendingVotesForValidatorByAccount(ctx *cli.Context, core *listener) error {
	var ret interface{}
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== getPendingVotesForValidatorByAccount ===", "account.Address", core.cfg.From)
	m := NewMessageRet(SolveType3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getPendingVotesForValidatorByAccount", core.cfg.TargetAddress, core.cfg.From)
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	fmt.Println(ret.(big.Int))
	return nil
}
func getValidatorsVotedForByAccount(ctx *cli.Context, core *listener) error {
	var ret interface{}
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== getValidatorsVotedForByAccount ===", "account.Address", core.cfg.From)
	m := NewMessageRet(SolveType3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getValidatorsVotedForByAccount", core.cfg.TargetAddress)
	core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	fmt.Println(ret.([]common.Address))
	return nil
}
