package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/cmd/marker/config"
	"github.com/mapprotocol/atlas/cmd/marker/connections"
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
	conn, _ := connections.DialConn(ctx, config)
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
var deregisterValidatorCommand = cli.Command{
	Name:   "deregisterValidator",
	Usage:  "deregister Validator",
	Action: MigrateFlags(deregisterValidator),
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
	Name:   "unlockMap",
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
	Name:   "withdrawMap",
	Usage:  "withdraw MAP",
	Action: MigrateFlags(withdraw),
	Flags:  Flags,
}

//---------- voter -----------------
var voteValidatorCommand = cli.Command{
	Name:   "vote",
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
var revokePendingCommand = cli.Command{
	Name:   "revokePending",
	Usage:  "Revokes `value` pending votes for `validator`",
	Action: MigrateFlags(revokePending),
	Flags:  Flags,
}
var revokeActiveCommand = cli.Command{
	Name:   "revokeActive",
	Usage:  "Revokes `value` active votes for `validator`",
	Action: MigrateFlags(revokeActive),
	Flags:  Flags,
}

//---------- query -----------------
var queryRegisteredValidatorSignersCommand = cli.Command{
	Name:   "getRegisteredValidatorSigners",
	Usage:  "get Registered Validator Signers",
	Action: MigrateFlags(getRegisteredValidatorSigners),
	Flags:  Flags,
}
var queryNumRegisteredValidatorsCommand = cli.Command{
	Name:   "getNumRegisteredValidators",
	Usage:  "get Num RegisteredValidators",
	Action: MigrateFlags(getNumRegisteredValidators),
	Flags:  Flags,
}
var queryTopValidatorsCommand = cli.Command{
	Name:   "getTopValidators",
	Usage:  "get Top Group Validators",
	Action: MigrateFlags(getTopValidators),
	Flags:  Flags,
}
var queryValidatorEligibilityCommand = cli.Command{
	Name:   "getValidatorEligibility",
	Usage:  "Judge whether the verifier`s Eligibility",
	Action: MigrateFlags(getValidatorEligibility),
	Flags:  Flags,
}
var queryTotalVotesForEligibleValidatorsCommand = cli.Command{
	Name:   "getTotalVotesForEligibleValidators",
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
var getPendingVotesForValidatorByAccountCommand = cli.Command{
	Name:   "getPendingVotesForValidatorByAccount",
	Usage:  "Returns the pending votes for `validator` made by `account`",
	Action: MigrateFlags(getPendingVotesForValidatorByAccount),
	Flags:  Flags,
}
var getActiveVotesForValidatorByAccountCommand = cli.Command{
	Name:   "getActiveVotesForValidatorByAccount",
	Usage:  "Returns the active votes for `validator` made by `account`",
	Action: MigrateFlags(getActiveVotesForValidatorByAccount),
	Flags:  Flags,
}
var getValidatorsVotedForByAccountCommand = cli.Command{
	Name:   "getValidatorsVotedForByAccount",
	Usage:  "Returns the validators that `account` has voted for.",
	Action: MigrateFlags(getValidatorsVotedForByAccount),
	Flags:  Flags,
}
var getAccountTotalLockedGoldCommand = cli.Command{
	Name:   "getAccountTotalLockedGold",
	Usage:  "Returns the total amount of locked gold for an account.",
	Action: MigrateFlags(getAccountTotalLockedGold),
	Flags:  Flags,
}
var getAccountNonvotingLockedGoldCommand = cli.Command{
	Name:   "getAccountNonvotingLockedGold",
	Usage:  "Returns the total amount of non-voting locked gold for an account",
	Action: MigrateFlags(getAccountNonvotingLockedGold),
	Flags:  Flags,
}
var getAccountLockedGoldRequirementCommand = cli.Command{
	Name:   "getAccountLockedGoldRequirement",
	Usage:  "Returns the current locked gold balance requirement for the supplied account.",
	Action: MigrateFlags(getAccountLockedGoldRequirement),
	Flags:  Flags,
}
var getPendingWithdrawalsCommand = cli.Command{
	Name:   "getPendingWithdrawals",
	Usage:  "Returns the pending withdrawals from unlocked gold for an account.",
	Action: MigrateFlags(getPendingWithdrawals),
	Flags:  Flags,
}

//-------------- owner --------------------
var setValidatorLockedGoldRequirementsCommand = cli.Command{
	Name:   "setValidatorLockedGoldRequirements",
	Usage:  "Updates the Locked Gold requirements for Validators.",
	Action: MigrateFlags(setValidatorLockedGoldRequirements),
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
	_params := []interface{}{big.NewInt(core.cfg.Commission), core.cfg.Lesser, core.cfg.Greater, core.cfg.PublicKey[1:], core.cfg.BlsPub[:], core.cfg.BLSProof}
	ValidatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	m := NewMessage(SolveType1, core.msgCh, core.cfg, ValidatorAddress, nil, abiValidators, "registerValidator", _params...)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

func createAccount1(_ *cli.Context, core *listener) error {
	createAccount(core, core.cfg.NamePrefix)
	return nil
}

func createAccount(core *listener, namePrefix string) {
	abiAccounts := core.cfg.AccountsParameters.AccountsABI
	accountsAddress := core.cfg.AccountsParameters.AccountsAddress

	logger := log.New("func", "createAccount")
	logger.Info("Create account", "address", core.cfg.From, "name", namePrefix)
	log.Info("=== create Account ===")
	m := NewMessage(SolveType1, core.msgCh, core.cfg, accountsAddress, nil, abiAccounts, "createAccount")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)

	log.Info("=== setName name ===")
	m = NewMessage(SolveType1, core.msgCh, core.cfg, accountsAddress, nil, abiAccounts, "setName", namePrefix)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)

	log.Info("=== setAccountDataEncryptionKey ===")
	m = NewMessage(SolveType1, core.msgCh, core.cfg, accountsAddress, nil, abiAccounts, "setAccountDataEncryptionKey", core.cfg.PublicKey)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
}

func deregisterValidator(ctx *cli.Context, core *listener) error {
	//----------------------------- deregisterValidator ---------------------------------
	log.Info("=== deregisterValidator ===")
	index := core.cfg.ValidatorIndex
	ValidatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	m := NewMessage(SolveType1, core.msgCh, core.cfg, ValidatorAddress, nil, abiValidators, "deregisterValidator", index)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

//---------- voter -----------------
func vote(_ *cli.Context, core *listener) error {
	ElectionsAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElections := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== vote Validator ===")
	amount := new(big.Int).Mul(core.cfg.VoteNum, big.NewInt(1e18))
	m := NewMessage(SolveType1, core.msgCh, core.cfg, ElectionsAddress, nil, abiElections, "vote", core.cfg.TargetAddress, amount, core.cfg.Lesser, core.cfg.Greater)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

func activate(_ *cli.Context, core *listener) error {
	ElectionsAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElections := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== activate validator gold ===", "account.Address", core.cfg.From)
	m := NewMessage(SolveType1, core.msgCh, core.cfg, ElectionsAddress, nil, abiElections, "activate", core.cfg.TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

/**
 * @notice Revokes `value` pending votes for `validator`
 * @param validator The validator to revoke votes from.
 * @param value The number of votes to revoke.
 * @param lesser The validator receiving fewer votes than the validator for which the vote was revoked,
 *   or 0 if that validator has the fewest votes of any validator.
 * @param greater The validator receiving more votes than the validator for which the vote was revoked,
 *   or 0 if that validator has the most votes of any validator.
 * @param index The index of the validator in the account's voting list.
 * @return True upon success.
 * @dev Fails if the account has not voted on a validator.
 */
func revokePending(_ *cli.Context, core *listener) error {
	ElectionsAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElections := core.cfg.ElectionParameters.ElectionABI
	validator := core.cfg.TargetAddress
	LockedNum := new(big.Int).Mul(core.cfg.LockedNum, big.NewInt(1e18))
	lesser := core.cfg.Lesser
	greater := core.cfg.Greater
	index := core.cfg.ValidatorIndex
	_params := []interface{}{validator, LockedNum, lesser, greater, index}
	log.Info("=== revokePending ===", "admin", core.cfg.From)
	m := NewMessage(SolveType1, core.msgCh, core.cfg, ElectionsAddress, nil, abiElections, "revokePending", _params...)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

/**
 * @notice Revokes `value` active votes for `validator`
 * @param validator The validator  to revoke votes from.
 * @param value The number of votes to revoke.
 * @param lesser The validator receiving fewer votes than the validator for which the vote was revoked,
 *   or 0 if that validator has the fewest votes of any validator.
 * @param greater The validator receiving more votes than the validator for which the vote was revoked,
 *   or 0 if that validator has the most votes of any validator.
 * @param index The index of the validator in the account's voting list.
 * @return True upon success.
 * @dev Fails if the account has not voted on a validator.
 */
func revokeActive(_ *cli.Context, core *listener) error {
	ElectionsAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElections := core.cfg.ElectionParameters.ElectionABI
	validator := core.cfg.TargetAddress
	LockedNum := new(big.Int).Mul(core.cfg.LockedNum, big.NewInt(1e18))
	lesser := core.cfg.Lesser
	greater := core.cfg.Greater
	index := core.cfg.ValidatorIndex
	_params := []interface{}{validator, LockedNum, lesser, greater, index}
	log.Info("=== revokeActive ===", "admin", core.cfg.From)
	m := NewMessage(SolveType1, core.msgCh, core.cfg, ElectionsAddress, nil, abiElections, "revokeActive", _params...)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

//---------- query -----------------
func getRegisteredValidatorSigners(_ *cli.Context, core *listener) error {
	var ValidatorSigners interface{}
	validatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidator := core.cfg.ValidatorParameters.ValidatorABI
	log.Info("==== getRegisteredValidatorSigners ===")
	m := NewMessageRet1(SolveType3, core.msgCh, core.cfg, &ValidatorSigners, validatorAddress, nil, abiValidator, "getRegisteredValidatorSigners")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)

	Validators := ValidatorSigners.([]common.Address)
	for i := 0; i < len(Validators); i++ {
		log.Info("Validator:", "index", i, "addr", Validators[i])
	}
	return nil
}
func getNumRegisteredValidators(_ *cli.Context, core *listener) error {
	var NumValidators interface{}
	validatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidator := core.cfg.ValidatorParameters.ValidatorABI
	log.Info("=== getNumRegisteredValidators ===", "admin", core.cfg.From)
	m := NewMessageRet1(SolveType3, core.msgCh, core.cfg, &NumValidators, validatorAddress, nil, abiValidator, "getNumRegisteredValidators")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	ret := NumValidators.(*big.Int)
	log.Info("=== result ===", "num", ret.String())
	return nil
}
func getTopValidators(_ *cli.Context, core *listener) error {
	var TopValidators interface{}
	validatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidator := core.cfg.ValidatorParameters.ValidatorABI
	log.Info("=== getTopValidators ===", "admin", core.cfg.From)
	m := NewMessageRet1(SolveType3, core.msgCh, core.cfg, &TopValidators, validatorAddress, nil, abiValidator, "getTopValidators", core.cfg.TopNum)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	Validators := TopValidators.([]common.Address)
	for i := 0; i < len(Validators); i++ {
		log.Info("Validator:", "index", i, "addr", Validators[i])
	}
	return nil
}

/*
* @notice Returns lists of all validator validators and the number of votes they've received.
* @return Lists of all  validators and the number of votes they've received.
 */
func getTotalVotesForEligibleValidators(_ *cli.Context, core *listener) error {
	type ret struct {
		Validators interface{} // indexed
		Values     interface{}
	}
	var t ret
	electionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	f := func(output []byte) {
		err := abiElection.UnpackIntoInterface(&t, "getTotalVotesForEligibleValidators", output)
		if err != nil {
			log.Error("getTotalVotesForEligibleValidators", "err", err)
		}
	}
	log.Info("=== getTotalVotesForEligibleValidators ===", "admin", core.cfg.From)
	m := NewMessageRet2(SolveType4, core.msgCh, core.cfg, f, electionAddress, nil, abiElection, "getTotalVotesForEligibleValidators")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	Validators := (t.Validators).([]common.Address)
	Values := (t.Values).([]*big.Int)
	for i := 0; i < len(Validators); i++ {
		log.Info("Validator:", "addr", Validators[i], "vote amount", Values[i])
	}
	return nil
}

/**
 * @notice Returns whether or not a validator is eligible to receive votes.
 * @return Whether or not a validator is eligible to receive votes.
 * @dev Eligible validators that have received their maximum number of votes cannot receive more.
 */
func getValidatorEligibility(_ *cli.Context, core *listener) error {
	var ret interface{}
	electionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== getValidatorEligibility ===", "admin", core.cfg.From)
	m := NewMessageRet1(SolveType3, core.msgCh, core.cfg, &ret, electionAddress, nil, abiElection, "getValidatorEligibility", core.cfg.TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	log.Info("=== result ===", "bool", ret.(bool))
	return nil
}

/**
 * @notice Gets the balance of the specified address.
 * @param owner The address to query the balance of.
 * @return The balance of the specified address.
 */
func balanceOf(_ *cli.Context, core *listener) error {
	var ret interface{}
	GoldTokenAddress := core.cfg.GoldTokenParameters.GoldTokenAddress
	abiGoldToken := core.cfg.GoldTokenParameters.GoldTokenABI
	log.Info("=== balanceOf ===", "admin", core.cfg.From)
	m := NewMessageRet1(SolveType3, core.msgCh, core.cfg, &ret, GoldTokenAddress, nil, abiGoldToken, "balanceOf", core.cfg.TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	log.Info("=== result ===", "balance", ret.(*big.Int).String())
	return nil
}

/**
 * @notice Returns the pending votes for `validator` made by `account`.
 * @param validator The address of the validator.
 * @param account The address of the voting account.
 * @return The pending votes for `validator` made by `account`.
 */
func getPendingVotesForValidatorByAccount(_ *cli.Context, core *listener) error {
	var ret interface{}
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== getPendingVotesForValidatorByAccount ===", "admin", core.cfg.From)
	m := NewMessageRet1(SolveType3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getPendingVotesForValidatorByAccount", core.cfg.TargetAddress, core.cfg.From)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	log.Info("PendingVotes", "balance", ret.(*big.Int))
	return nil
}
func getActiveVotesForValidatorByAccount(_ *cli.Context, core *listener) error {
	var ret interface{}
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== getActiveVotesForValidatorByAccount ===", "admin", core.cfg.From)
	m := NewMessageRet1(SolveType3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getActiveVotesForValidatorByAccount", core.cfg.TargetAddress, core.cfg.From)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	log.Info("ActiveVotes", "balance", ret.(*big.Int))
	return nil
}

/*
* @notice Returns the validators that `account` has voted for.
* @param account The address of the account casting votes.
* @return The validators that `account` has voted for.
 */
func getValidatorsVotedForByAccount(_ *cli.Context, core *listener) error {
	var ret interface{}
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== getValidatorsVotedForByAccount ===", "admin", core.cfg.From)
	m := NewMessageRet1(SolveType3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getValidatorsVotedForByAccount", core.cfg.TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	result := ret.([]common.Address)
	for i := 0; i < len(result); i++ {
		log.Info("validator", "Address", result[i])
	}
	return nil
}
func getAccountTotalLockedGold(_ *cli.Context, core *listener) error {
	var ret interface{}
	LockedGoldAddress := core.cfg.LockedGoldParameters.LockedGoldAddress
	abiLockedGold := core.cfg.LockedGoldParameters.LockedGoldABI
	log.Info("=== getAccountTotalLockedGold ===", "admin", core.cfg.From, "target", core.cfg.TargetAddress.String())
	m := NewMessageRet1(SolveType3, core.msgCh, core.cfg, &ret, LockedGoldAddress, nil, abiLockedGold, "getAccountTotalLockedGold", core.cfg.TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	result := ret.(*big.Int)
	log.Info("result", "lockedGold", result)
	return nil
}
func getAccountNonvotingLockedGold(_ *cli.Context, core *listener) error {
	var ret interface{}
	LockedGoldAddress := core.cfg.LockedGoldParameters.LockedGoldAddress
	abiLockedGold := core.cfg.LockedGoldParameters.LockedGoldABI
	log.Info("=== getAccountNonvotingLockedGold ===", "admin", core.cfg.From, "target", core.cfg.TargetAddress.String())
	m := NewMessageRet1(SolveType3, core.msgCh, core.cfg, &ret, LockedGoldAddress, nil, abiLockedGold, "getAccountNonvotingLockedGold", core.cfg.TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	result := ret.(*big.Int)
	log.Info("result", "lockedGold", result)
	return nil
}
func getAccountLockedGoldRequirement(_ *cli.Context, core *listener) error {
	var ret interface{}
	ValidatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	log.Info("=== getAccountLockedGoldRequirement ===", "admin", core.cfg.From, "target", core.cfg.TargetAddress.String())
	m := NewMessageRet1(SolveType3, core.msgCh, core.cfg, &ret, ValidatorAddress, nil, abiValidators, "getAccountLockedGoldRequirement", core.cfg.From)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	result := ret.(*big.Int)
	log.Info("result", "GoldRequirement", result)
	return nil
}
func getPendingWithdrawals(_ *cli.Context, core *listener) error {
	type ret []interface{}
	var Values interface{}
	var Timestamps interface{}
	t := ret{&Values, &Timestamps}
	LockedGoldAddress := core.cfg.LockedGoldParameters.LockedGoldAddress
	abiLockedGold := core.cfg.LockedGoldParameters.LockedGoldABI
	log.Info("=== getPendingWithdrawals ===", "admin", core.cfg.From, "target", core.cfg.TargetAddress.String())
	f := func(output []byte) {
		err := abiLockedGold.UnpackIntoInterface(&t, "getPendingWithdrawals", output)
		if err != nil {
			log.Error("getPendingWithdrawals", "err", err)
		}
	}
	m := NewMessageRet2(SolveType4, core.msgCh, core.cfg, f, LockedGoldAddress, nil, abiLockedGold, "getPendingWithdrawals", core.cfg.TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)

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

//--------------------- locked Map ------------------------
func lockedMAP(_ *cli.Context, core *listener) error {
	lockedGold := new(big.Int).Mul(core.cfg.LockedNum, big.NewInt(1e18))
	log.Info("=== Lock  gold ===")
	log.Info("Lock  gold", "amount", lockedGold.String())
	LockedGoldAddress := core.cfg.LockedGoldParameters.LockedGoldAddress
	abiLockedGold := core.cfg.LockedGoldParameters.LockedGoldABI
	m := NewMessage(SolveType2, core.msgCh, core.cfg, LockedGoldAddress, lockedGold, abiLockedGold, "lock")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}
func unlockedMAP(_ *cli.Context, core *listener) error {
	lockedGold := new(big.Int).Mul(core.cfg.LockedNum, big.NewInt(1e18))
	log.Info("=== unLock validator gold ===")
	log.Info("unLock validator gold", "amount", lockedGold, "admin", core.cfg.From)
	LockedGoldAddress := core.cfg.LockedGoldParameters.LockedGoldAddress
	abiLockedGold := core.cfg.LockedGoldParameters.LockedGoldABI
	m := NewMessage(SolveType1, core.msgCh, core.cfg, LockedGoldAddress, nil, abiLockedGold, "unlock", lockedGold)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}
func relockMAP(_ *cli.Context, core *listener) error {
	lockedGold := new(big.Int).Mul(core.cfg.LockedNum, big.NewInt(1e18))
	index := core.cfg.RelockIndex
	log.Info("=== relockMAP validator gold ===")
	log.Info("relockMAP validator gold", "amount", lockedGold)
	LockedGoldAddress := core.cfg.LockedGoldParameters.LockedGoldAddress
	abiLockedGold := core.cfg.LockedGoldParameters.LockedGoldABI
	m := NewMessage(SolveType1, core.msgCh, core.cfg, LockedGoldAddress, nil, abiLockedGold, "relock", index, lockedGold)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}
func withdraw(_ *cli.Context, core *listener) error {
	index := core.cfg.WithdrawIndex
	LockedGoldAddress := core.cfg.LockedGoldParameters.LockedGoldAddress
	abiLockedGold := core.cfg.LockedGoldParameters.LockedGoldABI
	log.Info("=== withdraw validator gold ===", "admin", core.cfg.From.String())
	m := NewMessage(SolveType1, core.msgCh, core.cfg, LockedGoldAddress, nil, abiLockedGold, "withdraw", index)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

//-------------------------- owner ------------------------
func setValidatorLockedGoldRequirements(_ *cli.Context, core *listener) error {
	value := big.NewInt(int64(core.cfg.Value))
	duration := big.NewInt(core.cfg.Duration)
	ValidatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	log.Info("=== setValidatorLockedGoldRequirements ===", "admin", core.cfg.From.String())
	m := NewMessage(SolveType1, core.msgCh, core.cfg, ValidatorAddress, nil, abiValidators, "setValidatorLockedGoldRequirements", value, duration)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}
