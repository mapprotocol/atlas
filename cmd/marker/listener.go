package main

import (
	"bytes"
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/cmd/marker/config"
	"github.com/mapprotocol/atlas/cmd/marker/connections"
	"github.com/mapprotocol/atlas/cmd/marker/mapprotocol"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/helper/decimal"
	"github.com/mapprotocol/atlas/helper/decimal/fixed"
	"github.com/mapprotocol/atlas/params"
	"sort"

	"gopkg.in/urfave/cli.v1"
	"math/big"
	"strings"
)

var (
	GetIndexError          = errors.New("get Index nil(no Address)")
	NoTargetValidatorError = errors.New("not find target validator")
	bigSubValue            = errors.New("not enough map")
	isContinueError        = true
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

var revertRegisterValidatorCommand = cli.Command{
	Name:   "revertRegister",
	Usage:  "register validator",
	Action: MigrateFlags(revertRegisterValidator),
	Flags:  Flags,
}
var quicklyRegisterValidatorCommand = cli.Command{
	Name:   "quicklyRegister",
	Usage:  "register validator",
	Action: MigrateFlags(quicklyRegisterValidator),
	Flags:  Flags,
}
var deregisterValidatorCommand = cli.Command{
	Name:   "deregister",
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
var quicklyVoteValidatorCommand = cli.Command{
	Name:   "quicklyVote",
	Usage:  "vote validator ",
	Action: MigrateFlags(quicklyVote),
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

var getValidatorCommand = cli.Command{
	Name:   "getValidator",
	Usage:  "Validator Info",
	Action: MigrateFlags(getValidator),
	Flags:  Flags,
}

var getRewardInfoCommand = cli.Command{
	Name:   "getValidatorRewardInfo",
	Usage:  "getValidator Info",
	Action: MigrateFlags(getRewardInfo),
	Flags:  Flags,
}
var getVoterRewardInfoCommand = cli.Command{
	Name:   "getVoterRewardInfo",
	Usage:  "get Voter Reward Information about yourself",
	Action: MigrateFlags(getVoterRewardInfo),
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
	Usage:  "get Top Validators",
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
var getActiveVotesForValidatorCommand = cli.Command{
	Name:   "getActiveVotesForValidator",
	Usage:  "Returns the total active vote units made for `validator`.",
	Action: MigrateFlags(getActiveVotesForValidator),
	Flags:  Flags,
}

var voterMonitorCommand = cli.Command{
	Name:   "voterMonitor",
	Usage:  "Monitor the revenue of voter to a validator",
	Action: MigrateFlags(voterMonitor),
	Flags:  Flags,
}

var getPendingVotersForValidatorCommand = cli.Command{
	Name:   "getPendingVotersForValidator",
	Usage:  "Returns the total pending voters vote for target `validator`.",
	Action: MigrateFlags(getPendingVotersForValidator),
	Flags:  Flags,
}
var getPendingInfoForValidatorCommand = cli.Command{
	Name:   "getPendingInfoForValidator",
	Usage:  "Returns the  pending Info voters vote And Epoch for target `validator`.",
	Action: MigrateFlags(getPendingInfoForValidator),
	Flags:  Flags,
}
var getValidatorsVotedForByAccountCommand = cli.Command{
	Name:   "getValidatorsVotedForByAccount",
	Usage:  "Returns the validators that `account` has voted for.",
	Action: MigrateFlags(getValidatorsVotedForByAccount),
	Flags:  Flags,
}
var getTotalVotesCommand = cli.Command{
	Name:   "getTotalVotes",
	Usage:  "Returns the total votes received across all validators.",
	Action: MigrateFlags(getTotalVotes),
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
var setImplementationCommand = cli.Command{
	Name:   "setImplementation",
	Usage:  "Sets the address of the implementation contract.",
	Action: MigrateFlags(setImplementation),
	Flags:  Flags,
}
var setOwnerCommand = cli.Command{
	Name:   "setContractOwner",
	Usage:  "Transfers ownership of the contract to a new account (`newOwner`).",
	Action: MigrateFlags(setContractOwner),
	Flags:  Flags,
}
var setProxyContractOwnerCommand = cli.Command{
	Name:   "setProxyContractOwner",
	Usage:  "Transfers ownership of the contract to a new account (`newOwner`).",
	Action: MigrateFlags(setProxyContractOwner),
	Flags:  Flags,
}
var getProxyContractOwnerCommand = cli.Command{
	Name:   "getProxyContractOwner",
	Usage:  "Transfers ownership of the contract to a new account (`newOwner`).",
	Action: MigrateFlags(getProxyContractOwner),
	Flags:  Flags,
}
var getContractOwnerCommand = cli.Command{
	Name:   "getContractOwner",
	Usage:  "Transfers ownership of the contract to a new account (`newOwner`).",
	Action: MigrateFlags(getContractOwner),
	Flags:  Flags,
}
var updateBlsPublicKeyCommand = cli.Command{
	Name:   "updateBlsPublicKey",
	Usage:  "updateBlsPublicKey",
	Action: MigrateFlags(updateBlsPublicKey),
	Flags:  Flags,
}
var setNextCommissionUpdateCommand = cli.Command{
	Name:   "setNextCommissionUpdate",
	Usage:  "set Next Commission Update",
	Action: MigrateFlags(setNextCommissionUpdate),
	Flags:  Flags,
}
var updateCommissionCommand = cli.Command{
	Name:   "updateCommission",
	Usage:  "updateCommission",
	Action: MigrateFlags(updateCommission),
	Flags:  Flags,
}
var setTargetValidatorEpochPaymentCommand = cli.Command{
	Name:   "setValidatorEpochPayment",
	Usage:  "Sets the target per-epoch payment in MAP  for validators",
	Action: MigrateFlags(setTargetValidatorEpochPayment),
	Flags:  Flags,
}

var setEpochRelayerPaymentFractionCommand = cli.Command{
	Name:   "setEpochRelayerPaymentFraction",
	Usage:  "set Epoch Relayer PaymentFraction",
	Action: MigrateFlags(setEpochRelayerPaymentFraction),
	Flags:  Flags,
}

//---------- validator -----------------
func registerValidator(ctx *cli.Context, core *listener) error {
	//----------------------------- registerValidator ---------------------------------
	log.Info("=== Register validator ===")
	//commision := fixed.MustNew(core.cfg.Commission).BigInt()
	commision := big.NewInt(0).SetUint64(core.cfg.Commission)
	log.Info("=== commision ===", "commision", commision)
	if isPendingDeRegisterValidator(core) {
		revertRegisterValidator(ctx, core)
		log.Info("the account is in PendingDeRegisterValidator list please use revertRegisterValidator command")
		return nil
	}
	greater, lesser := registerUseFor(core)
	//fmt.Println("=== greater, lesser ===", greater, lesser)
	//_params := []interface{}{commision, lesser, greater,core.cfg.BlsPub[:], core.cfg.BlsG1Pub[:], core.cfg.BLSProof, core.cfg.PublicKey[1:]}
	validatorParams := [4][]byte{core.cfg.BlsPub[:], core.cfg.BlsG1Pub[:], core.cfg.BLSProof, core.cfg.PublicKey[1:]}
	_params := []interface{}{commision, lesser, greater, validatorParams}
	ValidatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, ValidatorAddress, nil, abiValidators, "registerValidator", _params...)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

func isPendingDeRegisterValidator(core *listener) bool {
	//----------------------------- isPendingDeRegisterValidator ---------------------------------
	ValidatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	var ret bool
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ValidatorAddress, nil, abiValidators, "isPendingDeRegisterValidator")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return ret
}
func revertRegisterValidator(_ *cli.Context, core *listener) error {
	if !isPendingDeRegisterValidator(core) {
		log.Info("revert validator", "msg", "not in the deRegister list")
		return nil
	}
	ValidatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, ValidatorAddress, nil, abiValidators, "revertRegisterValidator")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

/*
   Test whether the contract on poc2 can be changed by calling getOrderID() in mapbridgev1
   Contract address 0xb586DC60e9e39F87c9CB8B7D7E30b2f04D40D14c
   By calling getorderid () nonce + + on poc2, we can judge whether the contract can be called successfully.
*/

func TestPoc2_getOrderID(ctx *cli.Context, core *listener) error {
	log.Info("=== Test ===")
	_params := []interface{}{common.HexToAddress("a"), common.HexToAddress("a"), common.HexToAddress("b"), big.NewInt(123), big.NewInt(1)}
	TestAddress := core.cfg.TestPoc2Parameters.Address
	TestAbi := core.cfg.TestPoc2Parameters.ABI
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, TestAddress, nil, TestAbi, "getOrderID", _params...)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

/*
   Test whether the contract on poc2 can be changed by calling getOrderID() in mapbridgev1
   Contract address 0xb586DC60e9e39F87c9CB8B7D7E30b2f04D40D14c
*/

func TestPoc2_getNonce(_ *cli.Context, core *listener) error {
	var ret interface{}
	TestAddress := core.cfg.TestPoc2Parameters.Address
	TestAbi := core.cfg.TestPoc2Parameters.ABI
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, TestAddress, nil, TestAbi, "nonce")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	result := ret
	log.Info("test poc 2", "result ", result)
	return nil
}

func updateBlsPublicKey(ctx *cli.Context, core *listener) error {
	log.Info("=== updateBlsPublicKey ===")
	_params := []interface{}{core.cfg.PublicKey[1:], core.cfg.BlsPub[:], core.cfg.BlsG1Pub[:], core.cfg.BLSProof}
	ValidatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, ValidatorAddress, nil, abiValidators, "updateBlsPublicKey", _params...)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}
func setNextCommissionUpdate(_ *cli.Context, core *listener) error {
	log.Info("=== setNextCommissionUpdate ===", "commission", core.cfg.Commission)
	Commission := core.cfg.Commission
	ValidatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, ValidatorAddress, nil, abiValidators, "setNextCommissionUpdate", big.NewInt(0).SetUint64(Commission))
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

func updateCommission(_ *cli.Context, core *listener) error {
	log.Info("=== setNextCommissionUpdate ===", "commission", core.cfg.Commission)
	ValidatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, ValidatorAddress, nil, abiValidators, "updateCommission")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

func quicklyRegisterValidator(ctx *cli.Context, core *listener) error {
	//---------------------------- create account ----------------------------------
	createAccount(core)
	//---------------------------- lock ----------------------------------
	if isContinueError {
		lockedMAP(ctx, core)
	}
	//----------------------------- registerValidator ---------------------------------
	if isContinueError {
		registerValidator(ctx, core)
	}
	log.Info("=== End ===")
	return nil
}

func createAccount1(_ *cli.Context, core *listener) error {
	createAccount(core)
	return nil
}

func createAccount(core *listener) {
	abiAccounts := core.cfg.AccountsParameters.AccountsABI
	accountsAddress := core.cfg.AccountsParameters.AccountsAddress

	logger := log.New("func", "createAccount")
	logger.Info("Create account", "address", core.cfg.From, "name", core.cfg.NamePrefix)
	log.Info("=== create Account ===")
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, accountsAddress, nil, abiAccounts, "createAccount")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	if !isContinueError {
		return
	}

	log.Info("=== setName name ===")
	m = NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, accountsAddress, nil, abiAccounts, "setName", core.cfg.NamePrefix)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	if !isContinueError {
		return
	}

	log.Info("=== setAccountDataEncryptionKey ===")
	m = NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, accountsAddress, nil, abiAccounts, "setAccountDataEncryptionKey", core.cfg.PublicKey)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
}

func deregisterValidator(_ *cli.Context, core *listener) error {
	//----------------------------- deregisterValidator ---------------------------------
	log.Info("=== deregisterValidator ===")
	//list := _getRegisteredValidatorSigners(core)
	//index, err := GetIndex(core.cfg.From, list)
	//if err != nil {
	//	log.Crit("deregisterValidator", "err", err)
	//}
	ValidatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, ValidatorAddress, nil, abiValidators, "deregisterValidator")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

//---------- voter -----------------
func vote(_ *cli.Context, core *listener) error {
	ElectionsAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElections := core.cfg.ElectionParameters.ElectionABI
	//greater, lesser := getGreaterLesser(core, core.cfg.TargetAddress)
	greater, lesser, err := getGL(core, core.cfg.TargetAddress)
	if err != nil {
		log.Error("vote", "err", err)
		return err
	}
	amount := new(big.Int).Mul(core.cfg.VoteNum, big.NewInt(1e18))
	log.Info("=== vote Validator ===", "admin", core.cfg.From, "voteTargetValidator", core.cfg.TargetAddress.String(), "vote MAP Num", core.cfg.VoteNum.String())
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, ElectionsAddress, nil, abiElections, "vote", core.cfg.TargetAddress, amount, lesser, greater)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

func quicklyVote(ctx *cli.Context, core *listener) error {
	//---------------------------- create account ----------------
	createAccount(core)
	//---------------------------- lock --------------------------
	if isContinueError {
		lockedMAP(ctx, core)
	}
	//---------------------------- vote --------------------------
	if isContinueError {
		vote(ctx, core)
	}
	log.Info("=== End ===")
	return nil
}

func activate(_ *cli.Context, core *listener) error {
	ElectionsAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElections := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== activate validator gold ===", "account.Address", core.cfg.From)
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, ElectionsAddress, nil, abiElections, "activate", core.cfg.TargetAddress)
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

	greater, lesser, _ := getGLSub(core, LockedNum, validator)
	list := _getValidatorsVotedForByAccount(core, core.cfg.From)
	index, err := GetIndex(validator, list)
	if err != nil {
		log.Crit("revokePending", "err", err)
	}
	//fmt.Println("=== greater,lesser,index ===", greater, lesser, index)
	_params := []interface{}{validator, LockedNum, lesser, greater, index}
	log.Info("=== revokePending ===", "admin", core.cfg.From)
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, ElectionsAddress, nil, abiElections, "revokePending", _params...)
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
	greater, lesser, _ := getGLSub(core, LockedNum, validator)

	list := _getValidatorsVotedForByAccount(core, core.cfg.From)
	index, err := GetIndex(validator, list)
	if err != nil {
		log.Crit("revokePending", "err", err)
	}
	//fmt.Println("=== greater,lesser,index ===", greater, lesser, index)
	_params := []interface{}{validator, LockedNum, lesser, greater, index}
	log.Info("=== revokeActive ===", "admin", core.cfg.From)
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, ElectionsAddress, nil, abiElections, "revokeActive", _params...)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

//---------- query -----------------
func getRegisteredValidatorSigners(_ *cli.Context, core *listener) error {
	log.Info("==== getRegisteredValidatorSigners ===")
	Validators := _getRegisteredValidatorSigners(core)
	if len(Validators) == 0 {
		log.Info("nil")
	}
	for i := 0; i < len(Validators); i++ {
		log.Info("Validator:", "index", i, "addr", Validators[i])
	}
	return nil
}
func getValidator(_ *cli.Context, core *listener) error {
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
	validatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidator := core.cfg.ValidatorParameters.ValidatorABI
	f := func(output []byte) {
		err := abiValidator.UnpackIntoInterface(&t, "getValidator", output)
		if err != nil {
			isContinueError = false
			log.Error("getValidator", "err", err)
		}
	}

	log.Info("=== getValidator ===", "admin", core.cfg.From)
	m := NewMessageRet2(SolveQueryResult4, core.msgCh, core.cfg, f, validatorAddress, nil, abiValidator, "getValidator", core.cfg.TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	if !isContinueError {
		return nil
	}
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
func getRewardInfo(_ *cli.Context, core *listener) error {
	curBlockNumber, err := core.conn.BlockNumber(context.Background())
	epochSize := chain.DefaultGenesisBlock().Config.Istanbul.Epoch
	if err != nil {
		return err
	}
	EpochFirst, err := istanbul.GetEpochFirstBlockGivenBlockNumber(curBlockNumber, epochSize)
	if err != nil {
		return err
	}
	Epoch := istanbul.GetEpochNumber(curBlockNumber, epochSize)
	validatorContractAddress := core.cfg.ValidatorParameters.ValidatorAddress
	queryBlock := big.NewInt(int64(EpochFirst - 1))
	log.Info("=== getReward ===", "cur_epoch", Epoch, "epochSize", epochSize, "queryBlockNumber", queryBlock, "validatorContractAddress", validatorContractAddress.String(), "admin", core.cfg.From)
	query := mapprotocol.BuildQuery(validatorContractAddress, mapprotocol.ValidatorEpochPaymentDistributed, queryBlock, queryBlock)
	// querying for logs
	logs, err := core.conn.FilterLogs(context.Background(), query)
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

func getVoterRewardInfo(ctx *cli.Context, core *listener) error {

	////3.
	//{
	//	f := new(big.Float).SetInt(myVotes)
	//	fSub := new(big.Float).SetInt(allVotes)
	//	f.Quo(f, fSub)
	//	log.Info("getExpectFraction", "balance", f)
	//}

	curBlockNumber, err := core.conn.BlockNumber(context.Background())
	epochSize := chain.DefaultGenesisBlock().Config.Istanbul.Epoch
	//if curBlockNumber < epochSize {
	//	log.Info("=== current block number less than first epoch number ===", "current", curBlockNumber, "epochSize", epochSize)
	//}
	if err != nil {
		return err
	}
	EpochFirst, err := istanbul.GetEpochFirstBlockGivenBlockNumber(curBlockNumber, epochSize)
	if err != nil {
		return err
	}
	Epoch := istanbul.GetEpochNumber(curBlockNumber, epochSize)
	electionContractAddress := core.cfg.ElectionParameters.ElectionAddress
	firstBlock := big.NewInt(int64(1))
	endBlock := big.NewInt(int64(EpochFirst + 1))
	log.Info("=== get voter Reward ===", "cur_epoch", Epoch, "epochSize", epochSize, "query first BlockNumber", firstBlock, "query end BlockNumber", endBlock, "validatorContractAddress", electionContractAddress.String(), "admin", core.cfg.From)
	query := mapprotocol.BuildQuery(electionContractAddress, mapprotocol.EpochRewardsDistributedToVoters, firstBlock, endBlock)
	// querying for logs
	logs, err := core.conn.FilterLogs(context.Background(), query)
	if err != nil {
		return err
	}
	for _, l := range logs {
		//validator := common.Bytes2Hex(l.Topics[0].Bytes())
		validator := common.BytesToAddress(l.Topics[1].Bytes())
		reward := big.NewInt(0).SetBytes(l.Data[:32])
		log.Info("reward to voters", "validator", validator, "reward", reward)
	}
	log.Info("=== END ===")

	return nil
}

func _getRegisteredValidatorSigners(core *listener) []common.Address {
	var ValidatorSigners interface{}
	validatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidator := core.cfg.ValidatorParameters.ValidatorABI
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ValidatorSigners, validatorAddress, nil, abiValidator, "getRegisteredValidatorSigners")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return ValidatorSigners.([]common.Address)
}
func getNumRegisteredValidators(_ *cli.Context, core *listener) error {
	var NumValidators interface{}
	validatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidator := core.cfg.ValidatorParameters.ValidatorABI
	log.Info("=== getNumRegisteredValidators ===", "admin", core.cfg.From)
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &NumValidators, validatorAddress, nil, abiValidator, "getNumRegisteredValidators")
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
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &TopValidators, validatorAddress, nil, abiValidator, "getTopValidators", core.cfg.TopNum)
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
			isContinueError = false
			log.Error("getTotalVotesForEligibleValidators", "err", err)
		}
	}
	log.Info("=== getTotalVotesForEligibleValidators ===", "admin", core.cfg.From)
	m := NewMessageRet2(SolveQueryResult4, core.msgCh, core.cfg, f, electionAddress, nil, abiElection, "getTotalVotesForEligibleValidators")
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
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, electionAddress, nil, abiElection, "getValidatorEligibility", core.cfg.TargetAddress)
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
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, GoldTokenAddress, nil, abiGoldToken, "balanceOf", core.cfg.TargetAddress)
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
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getPendingVotesForValidatorByAccount", core.cfg.TargetAddress, core.cfg.From)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	log.Info("PendingVotes", "balance", ret.(*big.Int))
	return nil
}

func getPendingVotersForValidator(_ *cli.Context, core *listener) error {
	var ret interface{}
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== getPendingVotersForValidator ===", "admin", core.cfg.From)
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getPendingVotersForValidator", core.cfg.TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	log.Info("getPendingVotersForValidator", "voters", ret.([]common.Address))
	return nil
}

func getPendingInfoForValidator(_ *cli.Context, core *listener) error {
	type ret []interface{}
	var Value interface{}
	var Epoch interface{}
	t := ret{&Value, &Epoch}
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	f := func(output []byte) {
		err := abiElection.UnpackIntoInterface(&t, "pendingInfo", output)
		if err != nil {
			isContinueError = false
			log.Error("getPendingInfoForValidator", "err", err)
		}
	}
	log.Info("=== getPendingInfoForValidator ===", "admin", core.cfg.From)
	m := NewMessageRet2(SolveQueryResult4, core.msgCh, core.cfg, f, ElectionAddress, nil, abiElection, "pendingInfo", core.cfg.From, core.cfg.TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	log.Info("getPendingInfoForValidator", "PendingEpoch", Epoch.(*big.Int), "Balance", Value.(*big.Int))
	return nil
}

func getActiveVotesForValidatorByAccount(_ *cli.Context, core *listener) error {
	var ret interface{}
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== getActiveVotesForValidatorByAccount ===", "admin", core.cfg.From)
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getActiveVotesForValidatorByAccount", core.cfg.TargetAddress, core.cfg.From)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	log.Info("ActiveVotes", "balance", ret.(*big.Int))
	return nil
}

func getActiveVotesForValidator(_ *cli.Context, core *listener) error {
	var ret interface{}
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	log.Info("=== getActiveVotesForValidator ===", "admin", core.cfg.From)
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getActiveVotesForValidator", core.cfg.TargetAddress)
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
	log.Info("=== getValidatorsVotedForByAccount ===", "admin", core.cfg.From)
	result := _getValidatorsVotedForByAccount(core, core.cfg.TargetAddress)
	if len(result) == 0 {
		log.Info("nil")
	}
	for i := 0; i < len(result); i++ {
		log.Info("validator", "Address", result[i])
	}
	return nil
}
func _getValidatorsVotedForByAccount(core *listener, target common.Address) []common.Address {
	var ret interface{}
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getValidatorsVotedForByAccount", target)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	result := ret.([]common.Address)
	return result
}
func getAccountTotalLockedGold(_ *cli.Context, core *listener) error {
	var ret interface{}
	LockedGoldAddress := core.cfg.LockedGoldParameters.LockedGoldAddress
	abiLockedGold := core.cfg.LockedGoldParameters.LockedGoldABI
	log.Info("=== getAccountTotalLockedGold ===", "admin", core.cfg.From, "target", core.cfg.TargetAddress.String())
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, LockedGoldAddress, nil, abiLockedGold, "getAccountTotalLockedGold", core.cfg.TargetAddress)
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
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, LockedGoldAddress, nil, abiLockedGold, "getAccountNonvotingLockedGold", core.cfg.TargetAddress)
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
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ValidatorAddress, nil, abiValidators, "getAccountLockedGoldRequirement", core.cfg.TargetAddress)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	result := ret.(*big.Int)
	log.Info("result", "GoldRequirement", result)
	return nil
}
func getTotalVotes(_ *cli.Context, core *listener) error {
	var ret interface{}
	log.Info("=== getAccountLockedGoldRequirement ===", "admin", core.cfg.From)
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getTotalVotes")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	result := ret.(*big.Int)
	log.Info("result", "getTotalVotes", result)
	//updatetime := big.NewInt(0).Mul(big.NewInt(1000000),big.NewInt(1))
	//var ret interface{}
	//Validator := core.cfg.ValidatorParameters.Validator
	//abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	//m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, Validator, nil, abiValidators, "calculateEpochScore", updatetime)
	//go core.writer.ResolveMessage(m)
	//core.waitUntilMsgHandled(1)
	//result := ret.(*big.Int)
	//log.Info("111111", "calculateEpochScore1 ", result)
	//log.Info("111111", "calculateEpochScore1 ", updatetime)
	//log.Info("222222", "calculateEpochScore2 ", params.MustBigInt("1000000000000000000000000"))
	//log.Info("222222", "calculateEpochScore2 ", params.MustBigInt("271000000000000000000000"))
	//log.Info("333333", "calculateEpochScore3 ", fixed.MustNew("0.271").BigInt())
	//a:=params.MustBigInt("1000000000000000000000000")
	//fmt.Println(result.Div(result,a))
	//INFO [01-19|11:00:16.269] 111111                                   calculateEpochScore1 =1,000,000,000,000,000,000,000,000
	//INFO [01-19|11:00:16.289] 111111                                   calculateEpochScore1 =1,000,000,000,000,000,000,000,000
	//INFO [01-19|11:00:16.289] 222222                                   calculateEpochScore2 =1,000,000,000,000,000,000,000,000
	//INFO [01-19|11:00:16.289] 222222                                   calculateEpochScore2 =271,000,000,000,000,000,000,000
	//INFO [01-19|11:00:16.289] 333333                                   calculateEpochScore3 =271,000,000,000,000,000,000,000

	//updatetime := big.NewInt(90)
	//var ret interface{}
	//Validator := mapprotocol.MustProxyAddressFor("Random")
	//abiValidators := mapprotocol.AbiFor("Random")
	//m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, Validator, nil, abiValidators, "getBlockRandomness", updatetime)
	//go core.writer.ResolveMessage(m)
	//core.waitUntilMsgHandled(1)
	//result := ret.(common.Hash)
	//fmt.Println(result.String())
	return nil
}
func getTotalVotesForValidator(_ *cli.Context, core *listener) error {
	var ret interface{}
	log.Info("=== getTotalVotesForValidator ===", "admin", core.cfg.From)
	ElectionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ElectionAddress, nil, abiElection, "getTotalVotesForValidator")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	result := ret.(*big.Int)
	log.Info("=== getTotalVotesForValidator ===", "result", result)
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
			isContinueError = false
			log.Error("getPendingWithdrawals", "err", err)
		}
	}
	m := NewMessageRet2(SolveQueryResult4, core.msgCh, core.cfg, f, LockedGoldAddress, nil, abiLockedGold, "getPendingWithdrawals", core.cfg.TargetAddress)
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
	m := NewMessage(SolveSendTranstion2, core.msgCh, core.cfg, LockedGoldAddress, lockedGold, abiLockedGold, "lock")
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
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, LockedGoldAddress, nil, abiLockedGold, "unlock", lockedGold)
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
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, LockedGoldAddress, nil, abiLockedGold, "relock", index, lockedGold)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}
func withdraw(_ *cli.Context, core *listener) error {
	index := core.cfg.WithdrawIndex
	LockedGoldAddress := core.cfg.LockedGoldParameters.LockedGoldAddress
	abiLockedGold := core.cfg.LockedGoldParameters.LockedGoldABI
	log.Info("=== withdraw validator gold ===", "admin", core.cfg.From.String())
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, LockedGoldAddress, nil, abiLockedGold, "withdraw", index)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

//-------------------------- owner ------------------------
func setValidatorLockedGoldRequirements(_ *cli.Context, core *listener) error {
	value := new(big.Int).Mul(big.NewInt(int64(core.cfg.Value)), big.NewInt(1e18))
	duration := big.NewInt(core.cfg.Duration)
	ValidatorAddress := core.cfg.ValidatorParameters.ValidatorAddress
	abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	log.Info("=== setValidatorLockedGoldRequirements ===", "admin", core.cfg.From.String())
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, ValidatorAddress, nil, abiValidators, "setValidatorLockedGoldRequirements", value, duration)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

func setImplementation(_ *cli.Context, core *listener) error {
	//implementation := common.HexToAddress("0x000000000000000000000000000000000000F012")
	implementation := core.cfg.ImplementationAddress
	ContractAddress := core.cfg.ContractAddress
	ProxyAbi := mapprotocol.AbiFor("Proxy")
	log.Info("=== setImplementation ===", "admin", core.cfg.From.String())
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, ContractAddress, nil, ProxyAbi, "_setImplementation", implementation)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

func setContractOwner(_ *cli.Context, core *listener) error {
	NewOwner := core.cfg.TargetAddress
	ContractAddress := core.cfg.ContractAddress //代理地址
	abiValidators := core.cfg.ValidatorParameters.ValidatorABI
	log.Info("ProxyAddress", "ContractAddress", ContractAddress, "NewOwner", NewOwner.String())
	log.Info("=== setOwner ===", "admin", core.cfg.From.String())
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, ContractAddress, nil, abiValidators, "transferOwnership", NewOwner)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

func setProxyContractOwner(_ *cli.Context, core *listener) error {
	NewOwner := core.cfg.TargetAddress
	ContractAddress := core.cfg.ContractAddress //代理地址
	log.Info("ProxyAddress", "ContractAddress", ContractAddress, "NewOwner", NewOwner.String())
	ProxyAbi := mapprotocol.AbiFor("Proxy") //代理ABI
	log.Info("=== setOwner ===", "admin", core.cfg.From.String())
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, ContractAddress, nil, ProxyAbi, "_transferOwnership", NewOwner)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}
func getProxyContractOwner(_ *cli.Context, core *listener) error {
	log.Info("=== getOwner ===", "admin", core.cfg.From.String())
	var ret interface{}
	ContractAddress := core.cfg.TargetAddress   //代理地址
	ValidatorAbi := mapprotocol.AbiFor("Proxy") //代理ABI
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ContractAddress, nil, ValidatorAbi, "_getOwner")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	result := ret
	log.Info("getOwner", "Owner ", result)
	return nil
}

func getContractOwner(_ *cli.Context, core *listener) error {
	log.Info("=== getOwner ===", "admin", core.cfg.From.String())
	var ret interface{}
	ContractAddress := core.cfg.TargetAddress                 //代理地址
	ValidatorAbi := core.cfg.ValidatorParameters.ValidatorABI //代理地址
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret, ContractAddress, nil, ValidatorAbi, "owner")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	result := ret
	log.Info("getOwner", "Owner ", result)
	return nil
}

func setTargetValidatorEpochPayment(_ *cli.Context, core *listener) error {
	value := new(big.Int).Mul(big.NewInt(int64(core.cfg.Value)), big.NewInt(1e18))
	EpochRewardAddress := core.cfg.EpochRewardParameters.EpochRewardsAddress
	abiEpochReward := core.cfg.EpochRewardParameters.EpochRewardsABI
	log.Info("=== setTargetValidatorEpochPayment ===", "admin", core.cfg.From.String())
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, EpochRewardAddress, nil, abiEpochReward, "setTargetValidatorEpochPayment", value)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

func setEpochRelayerPaymentFraction(_ *cli.Context, core *listener) error {
	fixed := fixed.MustNew(core.cfg.Fixed).BigInt()
	EpochRewardAddress := core.cfg.EpochRewardParameters.EpochRewardsAddress
	abiEpochReward := core.cfg.EpochRewardParameters.EpochRewardsABI
	log.Info("=== setEpochRelayerPaymentFraction ===", "admin", core.cfg.From.String())
	m := NewMessage(SolveSendTranstion1, core.msgCh, core.cfg, EpochRewardAddress, nil, abiEpochReward, "setEpochRelayerPaymentFraction", fixed)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	return nil
}

//-------------------- getLesser getGreater -------
//Sub todo judge locked and withdrawal comparison
func getGLSub(core *listener, SubValue *big.Int, target common.Address) (common.Address, common.Address, error) {
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
			isContinueError = false
			log.Error("getTotalVotesForEligibleValidators setLesserGreater", "err", err)
		}
	}
	m := NewMessageRet2(SolveQueryResult4, core.msgCh, core.cfg, f, electionAddress, nil, abiElection, "getTotalVotesForEligibleValidators")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	validators := (t.Validators).([]common.Address)
	votes := (t.Values).([]*big.Int)
	voteTotals := make([]voteTotal, len(validators))
	for i, addr := range validators {
		voteTotals[i] = voteTotal{addr, votes[i]}
	}
	//for i, v := range voteTotals {
	//	fmt.Println("=== ", i, "===", v.Validator.String(), v.Value.String())
	//}
	//fmt.Println("=== target ===", target.String())
	for _, voteTotal := range voteTotals {
		if bytes.Equal(voteTotal.Validator.Bytes(), target.Bytes()) {
			if big.NewInt(0).Cmp(SubValue) < 0 {
				if voteTotal.Value.Cmp(SubValue) > 0 {
					voteTotal.Value.Sub(voteTotal.Value, SubValue)
				} else {
					return params.ZeroAddress, params.ZeroAddress, bigSubValue
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
			break
		}
	}
	return params.ZeroAddress, params.ZeroAddress, NoTargetValidatorError
}

func GetIndex(target common.Address, list []common.Address) (*big.Int, error) {
	//fmt.Println("=== target ===", target.String())
	for index, v := range list {
		//fmt.Println("=== list ===", v.String())
		if bytes.Equal(target.Bytes(), v.Bytes()) {
			return big.NewInt(int64(index)), nil
		}
	}
	return nil, GetIndexError
}

type voteTotal struct {
	Validator common.Address
	Value     *big.Int
}

//add
func getGL(core *listener, target common.Address) (common.Address, common.Address, error) {
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
			isContinueError = false
			log.Error("getTotalVotesForEligibleValidators setLesserGreater", "err", err)
		}
	}
	m := NewMessageRet2(SolveQueryResult4, core.msgCh, core.cfg, f, electionAddress, nil, abiElection, "getTotalVotesForEligibleValidators")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	validators := (t.Validators).([]common.Address)
	votes := (t.Values).([]*big.Int)
	voteTotals := make([]voteTotal, len(validators))
	for i, addr := range validators {
		voteTotals[i] = voteTotal{addr, votes[i]}
	}
	//for i, v := range voteTotals {
	//	fmt.Println("=== ", i, "===", v.Validator.String(), v.Value.String())
	//}

	voteNum := new(big.Int).Mul(core.cfg.VoteNum, big.NewInt(1e18))
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
			break
		}
	}
	return params.ZeroAddress, params.ZeroAddress, NoTargetValidatorError
}

func registerUseFor(core *listener) (common.Address, common.Address) {
	//type ret struct {
	//	Validators interface{} // indexed
	//	Values     interface{}
	//}
	//var t ret
	electionAddress := core.cfg.ElectionParameters.ElectionAddress
	abiElection := core.cfg.ElectionParameters.ElectionABI
	//f := func(output []byte) {
	//	err := abiElection.UnpackIntoInterface(&t, "getTotalVotesForEligibleValidators", output)
	//	if err != nil {
	//		isContinueError = false
	//		log.Error("getTotalVotesForEligibleValidators setLesserGreater", "err", err)
	//	}
	//}
	//m := NewMessageRet2(SolveQueryResult4, core.msgCh, core.cfg, f, electionAddress, nil, abiElection, "getTotalVotesForEligibleValidators")
	//go core.writer.ResolveMessage(m)
	//core.waitUntilMsgHandled(1)
	//Validators := (t.Validators).([]common.Address)

	var ret1 interface{}
	log.Info("=== getTotalVotesForValidator ===", "admin", core.cfg.From)
	m := NewMessageRet1(SolveQueryResult3, core.msgCh, core.cfg, &ret1, electionAddress, nil, abiElection, "getTotalVotesForValidator", core.cfg.From)
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	result := ret1.(*big.Int)
	log.Info("=== getTotalVotesForValidator ===", "result", result)
	core.cfg.VoteNum = result
	G, L, _ := getGL2(core, core.cfg.From)
	return G, L
}
func getGL2(core *listener, target common.Address) (common.Address, common.Address, error) {
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
			isContinueError = false
			log.Error("getTotalVotesForEligibleValidators setLesserGreater", "err", err)
		}
	}
	m := NewMessageRet2(SolveQueryResult4, core.msgCh, core.cfg, f, electionAddress, nil, abiElection, "getTotalVotesForEligibleValidators")
	go core.writer.ResolveMessage(m)
	core.waitUntilMsgHandled(1)
	validators := (t.Validators).([]common.Address)
	votes := (t.Values).([]*big.Int)
	voteTotals := make([]voteTotal, len(validators))
	for i, addr := range validators {
		voteTotals[i] = voteTotal{addr, votes[i]}
	}
	voteTotals = append(voteTotals, voteTotal{target, core.cfg.VoteNum})
	for _, voteTotal := range voteTotals {
		if bytes.Equal(voteTotal.Validator.Bytes(), target.Bytes()) {

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
			break
		}
	}
	return params.ZeroAddress, params.ZeroAddress, NoTargetValidatorError
}
