package cmd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/accounts"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/marker/account"
	"github.com/mapprotocol/atlas/cmd/new_marker/define"
	"github.com/mapprotocol/atlas/cmd/new_marker/mapprotocol"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/helper/bls"
	"github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	"os"
	"sort"
)

type Validator struct {
	*base
	account                         *Account
	to, lockGoldTo, electionTo      common.Address
	abi, lockedGoldAbi, electionAbi *abi.ABI
}

func NewValidator() *Validator {
	return &Validator{
		base:          newBase(),
		account:       NewAccount(),
		to:            mapprotocol.MustProxyAddressFor("Validators"),
		abi:           mapprotocol.AbiFor("Validators"),
		electionTo:    mapprotocol.MustProxyAddressFor("Election"),
		electionAbi:   mapprotocol.AbiFor("Election"),
		lockGoldTo:    mapprotocol.MustProxyAddressFor("LockedGold"),
		lockedGoldAbi: mapprotocol.AbiFor("LockedGold"),
	}
}

func (v *Validator) RegisterValidator(ctx *cli.Context, cfg *define.Config) error {
	log.Info("=== Register validator ===")
	commission := big.NewInt(0).SetUint64(cfg.Commission)
	log.Info("=== commission ===", "commission", commission)
	if v.isPendingDeRegisterValidator(cfg) {
		_ = v.revertRegisterValidator(ctx, cfg)
		log.Info("the account is in PendingDeRegisterValidator list please use revertRegisterValidator command")
		return nil
	}
	greater, lesser := v.registerUseFor(cfg)
	if cfg.PrivateKeyHex != "" {
		SignerPriv := cfg.PrivateKeyHex
		priv, err := crypto.ToECDSA(common.FromHex(SignerPriv))
		if err != nil {
			panic(err)
		}
		publicAddr := crypto.PubkeyToAddress(priv.PublicKey)
		_account := &account.Account{Address: publicAddr, PrivateKey: priv}
		blsPub, err := _account.BLSPublicKey()
		if err != nil {
			panic(err)
		}
		blsG1Pub, err := _account.BLSG1PublicKey()
		if err != nil {
			panic(err)
		}
		cfg.PublicKey = _account.PublicKey()
		cfg.BlsPub = blsPub
		cfg.BlsG1Pub = blsG1Pub
		cfg.BLSProof = _account.MustBLSProofOfPossession()
		cfg.BLSProof = v.makeBLSProofOfPossessionFromSigner_(cfg.From, cfg.PrivateKeyHex).Marshal()
	}
	validatorParams := [4][]byte{cfg.BlsPub[:], cfg.BlsG1Pub[:], cfg.BLSProof, cfg.PublicKey[1:]}

	_params := []interface{}{commission, lesser, greater, validatorParams}
	v.handleType1Msg(cfg, v.to, nil, v.abi, "registerValidator", _params...)
	return nil
}

func (v *Validator) RegisterValidatorByProof(_ *cli.Context, cfg *define.Config) error {
	commission := new(big.Int).SetUint64(cfg.Commission)
	log.Info("registerValidatorByProof", "commission", commission)
	if v.isPendingDeRegisterValidator(cfg) {
		log.Info("the account is in PendingDeRegisterValidator list please use revertRegisterValidator command")
		return nil
	}
	greater, lesser := v.registerUseFor(cfg)
	dec, err := hexutil.Decode(cfg.Proof)
	if err != nil {
		return err
	}
	pf := new(Proof)
	if err := rlp.DecodeBytes(dec, pf); err != nil {
		return err
	}

	validatorParams := [4][]byte{pf.BLSPublicKey[:], pf.BLSG1PublicKey[:], pf.BLSProof, pf.PublicKey[1:]}
	_params := []interface{}{commission, lesser, greater, validatorParams}
	fmt.Println("============================== ", _params)
	v.handleType1Msg(cfg, v.to, nil, v.abi, "registerValidator", _params...)
	return nil
}

func (v *Validator) RevertRegisterValidator(_ *cli.Context, cfg *define.Config) error {
	if !v.isPendingDeRegisterValidator(cfg) {
		log.Info("revert validator", "msg", "not in the deRegister list")
		return nil
	}
	v.handleType1Msg(cfg, v.to, nil, v.abi, "revertRegisterValidator")
	return nil
}

func (v *Validator) DeregisterValidator(_ *cli.Context, cfg *define.Config) error {
	//----------------------------- deregisterValidator ---------------------------------
	log.Info("=== deregisterValidator ===")
	v.handleType1Msg(cfg, v.to, nil, v.abi, "deregisterValidator")
	return nil
}

func (v *Validator) GenerateSignerProof(_ *cli.Context, cfg *define.Config) error {
	log.Info("generateBLSProof", "validator", cfg.Address, "signerPrivate", cfg.PrivateKeyHex)
	private, err := crypto.ToECDSA(common.FromHex(cfg.PrivateKeyHex))
	if err != nil {
		return err
	}
	publicAddr := crypto.PubkeyToAddress(private.PublicKey)
	_account := &account.Account{Address: publicAddr, PrivateKey: private}
	blsPub, err := _account.BLSPublicKey()
	if err != nil {
		return err
	}
	blsG1Pub, err := _account.BLSG1PublicKey()
	if err != nil {
		return err
	}

	args := Proof{
		PublicKey:      _account.PublicKey(),
		BLSPublicKey:   blsPub,
		BLSG1PublicKey: blsG1Pub,
		BLSProof:       v.makeBLSProofOfPossessionFromSigner_(cfg.Address, cfg.PrivateKeyHex).Marshal(),
	}
	enc, err := rlp.EncodeToBytes(args)
	if err != nil {
		return err
	}
	log.Info("generateBLSProof", "proof", hexutil.Encode(enc))
	return nil
}

func (v *Validator) QuicklyRegisterValidator(ctx *cli.Context, cfg *define.Config) error {
	//---------------------------- create account ----------------------------------
	_ = v.account.CreateAccount(ctx, cfg)

	if cfg.PrivateKeyHex != "" {
		_ = v.AuthorizeValidatorSigner(ctx, cfg)
	}
	//---------------------------- lock ----------------------------------
	_ = v.LockedMAP(ctx, cfg)

	//----------------------------- registerValidator ---------------------------------
	_ = v.RegisterValidator(ctx, cfg)
	log.Info("=== End ===")
	return nil
}

func (v *Validator) LockedMAP(_ *cli.Context, cfg *define.Config) error {
	lockedGold := new(big.Int).Mul(cfg.Value, big.NewInt(1e18))
	log.Info("=== Lock  gold ===")
	log.Info("Lock  gold", "amount", lockedGold.String())
	v.handleType2Msg(cfg, v.lockGoldTo, lockedGold, v.lockedGoldAbi, "lock")
	return nil
}

/*
AuthorizeValidatorSigner
note:account function before become to be a validator
signer sign account
need signer private
*/
func (v *Validator) AuthorizeValidatorSigner(_ *cli.Context, cfg *define.Config) error {
	SignatureStr, signer := v.makeECDSASignatureFromSigner_(cfg.From, cfg.PrivateKeyHex) // signer sign account
	Signature, err := hexutil.Decode(SignatureStr)
	if err != nil {
		panic(err)
	}
	all := uint8(new(big.Int).SetBytes([]byte{Signature[64] + 27}).Uint64())
	r := common.BytesToHash(Signature[:32])
	s := common.BytesToHash(Signature[32:64])

	logger := log.New("func", "authorizeValidatorSigner")
	logger.Info("authorizeValidatorSigner", "validator", cfg.From, "signer", signer)
	v.handleType1Msg(cfg, v.account.to, nil, v.account.abi, "authorizeValidatorSigner", signer, all, r, s)
	return nil
}

func (v *Validator) AuthorizeValidatorSignerBySignature(_ *cli.Context, cfg *define.Config) error {
	Signature, err := hexutil.Decode(cfg.Signature)
	if err != nil {
		panic(err)
	}
	all := uint8(new(big.Int).SetBytes([]byte{Signature[64] + 27}).Uint64())
	r := common.BytesToHash(Signature[:32])
	s := common.BytesToHash(Signature[32:64])

	log.Info("authorizeValidatorSignerBySignature", "signer", cfg.Address, "signature", cfg.Signature)
	v.handleType1Msg(cfg, v.account.to, nil, v.account.abi, "authorizeValidatorSigner", cfg.Address, all, r, s)
	return nil
}

func (v *Validator) MakeECDSASignatureFromSigner(_ *cli.Context, cfg *define.Config) error {
	v.makeECDSASignatureFromSigner_(cfg.Address, cfg.PrivateKeyHex)
	return nil
}

func (v *Validator) MakeBLSProofOfPossessionFromsigner(_ *cli.Context, cfg *define.Config) error {
	signature := v.makeBLSProofOfPossessionFromSigner_(cfg.Address, cfg.PrivateKeyHex)
	log.Info("=== pop ===", "result", hexutil.Encode(signature.Marshal()))
	return nil
}

func (v *Validator) isPendingDeRegisterValidator(cfg *define.Config) bool {
	//----------------------------- isPendingDeRegisterValidator ---------------------------------
	var ret bool
	v.handleType3Msg(cfg, &ret, v.to, nil, v.abi, "isPendingDeRegisterValidator")
	return ret
}

func (v *Validator) revertRegisterValidator(_ *cli.Context, cfg *define.Config) error {
	if !v.isPendingDeRegisterValidator(cfg) {
		log.Info("revert validator", "msg", "not in the deRegister list")
		return nil
	}
	v.handleType1Msg(cfg, v.to, nil, v.abi, "revertRegisterValidator")
	return nil
}

func (v *Validator) makeBLSProofOfPossessionFromSigner_(message common.Address, signerPrivate string) *bls.UnsafeSignature {
	log.Info("=== makeBLSProofOfPossessionFromSigner ===")
	privECDSA, err := crypto.ToECDSA(common.FromHex(signerPrivate))
	if err != nil {
		panic(err)
	}

	blsPrivateKey, _ := bls.CryptoType().ECDSAToBLS(privECDSA)
	privateKey, _ := bls.DeserializePrivateKey(blsPrivateKey)
	serializedPrivateKey, _ := privateKey.Serialize()
	publicKey, _ := bls.CryptoType().PrivateToPublic(serializedPrivateKey)
	pk, err := bls.UnmarshalPk(publicKey[:])
	if err != nil {
		panic(err)
	}
	signature, err := bls.UnsafeSign2(privateKey, message.Bytes())
	if err != nil {
		panic(err)
	}
	if err := bls.VerifyUnsafe2(pk, message.Bytes(), signature); err != nil {
		panic(err)
	}
	return signature
}

func (v *Validator) registerUseFor(cfg *define.Config) (common.Address, common.Address) {
	var ret1 interface{}
	v.handleType3Msg(cfg, &ret1, v.electionTo, nil, v.electionAbi, "getTotalVotesForValidator", cfg.From)
	result := ret1.(*big.Int)
	log.Info("=== getTotalVotesForValidator ===", "result", result)
	cfg.Value = result
	G, L, _ := v.getGL2(cfg, cfg.From)
	return G, L
}

type voteTotal struct {
	Validator common.Address
	Value     *big.Int
}

type Proof struct {
	PublicKey      []byte
	BLSPublicKey   [128]byte
	BLSG1PublicKey [64]byte
	BLSProof       []byte
}

func (v *Validator) getGL2(cfg *define.Config, target common.Address) (common.Address, common.Address, error) {
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
	voteTotals = append(voteTotals, voteTotal{target, cfg.Value})
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
		}
	}
	return params.ZeroAddress, params.ZeroAddress, define.NoTargetValidatorError
}

func (v *Validator) makeECDSASignatureFromSigner_(validator common.Address, signerPrivate string) (string, common.Address) {
	log.Info("=== makeECDSASignatureFromSigner ===")
	// PrivateKeyHex := cfg.PrivateKeyHex
	priv, err := crypto.ToECDSA(common.FromHex(signerPrivate))
	if err != nil {
		panic(err)
	}
	signer := crypto.PubkeyToAddress(priv.PublicKey)
	// account_ := cfg.From
	hash := accounts.TextHash(crypto.Keccak256(validator[:]))
	sig, err := crypto.Sign(hash, priv)
	if err != nil {
		panic(err)
	}
	//for test
	recoverPubKey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		panic(err)
	}
	log.Info("=== signer  ===", "account", crypto.PubkeyToAddress(*recoverPubKey))
	log.Info("ECDSASignature", "result", hexutil.Encode(sig))
	return hexutil.Encode(sig), signer
}

func (v *Validator) updateBlsPublicKey(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== updateBlsPublicKey ===")
	_params := []interface{}{cfg.PublicKey[1:], cfg.BlsPub[:], cfg.BlsG1Pub[:], cfg.BLSProof}
	v.handleType1Msg(cfg, v.to, nil, v.abi, "updateBlsPublicKey", _params...)
	return nil
}

func (v *Validator) updateCommission(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== updateCommission ===")
	v.handleType1Msg(cfg, v.to, nil, v.abi, "updateCommission")
	return nil
}

func (v *Validator) setNextCommissionUpdate(_ *cli.Context, cfg *define.Config) error {
	log.Info("=== setNextCommissionUpdate ===", "commission", cfg.Commission)
	Commission := cfg.Commission
	v.handleType1Msg(cfg, v.to, nil, v.abi, "setNextCommissionUpdate", big.NewInt(0).SetUint64(Commission))
	return nil
}

func (v *Validator) GetRewardInfo(_ *cli.Context, cfg *define.Config) error {
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
	log.Info("=== getReward ===", "cur_epoch", Epoch, "epochSize", epochSize, "queryBlockNumber", queryBlock, "validatorContractAddress", validatorContractAddress.String(), "from", cfg.From)
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
