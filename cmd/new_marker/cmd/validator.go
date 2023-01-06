package cmd

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/cmd/marker/account"
	"github.com/mapprotocol/atlas/cmd/new_marker/define"
	"github.com/mapprotocol/atlas/cmd/new_marker/mapprotocol"
	"github.com/mapprotocol/atlas/cmd/new_marker/writer"
	"github.com/mapprotocol/atlas/helper/bls"
	"github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	"os"
	"sort"
)

type Validator struct {
	*base
	*writer.Writer
	to, lockGoldTo, electionTo      common.Address
	abi, lockedGoldAbi, electionAbi *abi.ABI
}

func NewValidator() *Validator {
	return &Validator{
		base:        newBase(),
		to:          mapprotocol.MustProxyAddressFor("Validators"),
		abi:         mapprotocol.AbiFor("Validators"),
		electionTo:  mapprotocol.MustProxyAddressFor("Election"),
		electionAbi: mapprotocol.AbiFor("Election"),
	}
}

func (v *Validator) RegisterValidator(ctx *cli.Context, cfg *define.Config) error {
	log.Info("=== Register validator ===")
	commision := big.NewInt(0).SetUint64(cfg.Commission)
	log.Info("=== commision ===", "commision", commision)
	if v.isPendingDeRegisterValidator(cfg) {
		_ = v.revertRegisterValidator(ctx, cfg)
		log.Info("the account is in PendingDeRegisterValidator list please use revertRegisterValidator command")
		return nil
	}
	greater, lesser := v.registerUseFor(cfg)
	if cfg.SignerPriv != "" {
		SignerPriv := cfg.SignerPriv
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
		cfg.BLSProof = v.makeBLSProofOfPossessionFromSigner_(cfg.From, cfg.SignerPriv).Marshal()
	}
	validatorParams := [4][]byte{cfg.BlsPub[:], cfg.BlsG1Pub[:], cfg.BLSProof, cfg.PublicKey[1:]}

	_params := []interface{}{commision, lesser, greater, validatorParams}
	v.handleType1Msg(cfg, v.to, nil, v.abi, "registerValidator", _params)
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
	cfg.VoteNum = result
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
	voteTotals = append(voteTotals, voteTotal{target, cfg.VoteNum})
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
	v.handleType1Msg(cfg, v.to, nil, v.abi, "registerValidator", _params)
	return nil
}
