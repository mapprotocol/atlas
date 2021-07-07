package vm

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	lru "github.com/hashicorp/golang-lru"
	"github.com/mapprotocol/atlas/params"
	"math/big"
	"sort"
)

/////////////////////////////////////////////////////////////////////////////////
var IC *RelayerCache

//var OpQueryReward uint8 = 5
//var OpQueryFine uint8 = 6

func init() {
	IC = newRelayerCache()
}

type RelayerCache struct {
	Cache *lru.Cache
	size  int
}

func newRelayerCache() *RelayerCache {
	cc := &RelayerCache{
		size: 20,
	}
	cc.Cache, _ = lru.New(cc.size)
	return cc
}

/////////////////////////////////////////////////////////////////////////////////
type PairstakingValue struct {
	Amount *big.Int
	Height *big.Int
	State  uint8
}

func (v *PairstakingValue) isElection() bool {
	return (v.State&params.StateRegisterOnce != 0 || v.State&params.StateResgisterAuto != 0)
}

type RewardItem struct {
	EpochID uint64
	Amount  *big.Int
	//Height *big.Int
}
type FineItem struct {
	EpochID uint64
	Amount  *big.Int
	//Height *big.Int
}
type RedeemItem struct {
	Amount  *big.Int
	EpochID uint64
	State   uint8
}

func (r *RedeemItem) toHeight() *big.Int {
	e := GetEpochFromID(r.EpochID + 1)
	return new(big.Int).SetUint64(e.BeginHeight)
}
func (r *RedeemItem) fromHeight(hh *big.Int) {
	e := GetEpochFromHeight(hh.Uint64())
	if e != nil {
		r.EpochID = e.EpochID
	}
}
func (r *RedeemItem) update(o *RedeemItem) {
	if r.EpochID == o.EpochID {
		r.Amount = r.Amount.Add(r.Amount, o.Amount)
	}
}
func (r *RedeemItem) isRedeem(target uint64) bool {
	hh := r.toHeight().Uint64()
	return target > hh+params.MaxRedeemHeight
}
func newRedeemItem(eid uint64, amount *big.Int) *RedeemItem {
	return &RedeemItem{
		Amount:  new(big.Int).Set(amount),
		EpochID: eid,
	}
}

type registerUnit struct {
	Address    common.Address
	Value      []*PairstakingValue // sort by height
	RedeemInof []*RedeemItem
	Reward     []*RewardItem
	Fine       []*FineItem
}

func (s *registerUnit) getReward(epochID uint64) *big.Int {
	for _, v := range s.Reward {
		if v.EpochID == epochID {
			return v.Amount
		}
	}
	return nil
}
func (s *registerUnit) getFine(epochID uint64) *big.Int {
	for _, v := range s.Fine {
		if v.EpochID == epochID {
			return v.Amount
		}
	}
	return nil
}

func (s *registerUnit) getAllStaking(hh uint64) *big.Int {
	all := big.NewInt(0)
	for _, v := range s.Value {
		if v.Height.Uint64() <= hh {
			all = all.Add(all, v.Amount)
		} else {
			break
		}
	}
	return all
}
func (s *registerUnit) getValidStaking(hh uint64) *big.Int {
	all := big.NewInt(0)
	for _, v := range s.Value {
		if v.Height.Uint64() <= hh {
			if v.isElection() {
				all = all.Add(all, v.Amount)
			}
		} else {
			break
		}
	}
	e := GetEpochFromHeight(hh)
	r := s.getRedeemItem(e.EpochID)
	if r != nil {
		res := new(big.Int).Sub(all, r.Amount)
		if res.Sign() >= 0 {
			return res
		} else {
			log.Error("getValidStaking error", "all amount", all, "redeem amount", r.Amount, "height", hh)
			return big.NewInt(0)
		}
	}

	return all
}
func (s *registerUnit) getValidRedeem(hh uint64) *big.Int {
	all := big.NewInt(0)
	for _, v := range s.RedeemInof {
		if v.isRedeem(hh) {
			all = all.Add(all, v.Amount)
		}
	}
	return all
}
func (s *registerUnit) GetRewardAddress() common.Address {
	return s.Address
}
func (s *registerUnit) getRedeemItem(epochID uint64) *RedeemItem {
	for _, v := range s.RedeemInof {
		if v.EpochID == epochID {
			return v
		}
	}
	return nil
}

// stopStakingInfo redeem delay in next epoch + MaxRedeemHeight
func (s *registerUnit) stopStakingInfo(amount, lastHeight *big.Int) error {
	all := s.getValidStaking(lastHeight.Uint64())
	if all.Cmp(amount) < 0 {
		return params.ErrAmountOver
	}
	e := GetEpochFromHeight(lastHeight.Uint64())
	if e == nil {
		return params.ErrNotFoundEpoch
	}
	r := s.getRedeemItem(e.EpochID)
	tmp := &RedeemItem{
		Amount:  new(big.Int).Set(amount),
		EpochID: e.EpochID,
		State:   params.StateRedeem,
	}
	if r == nil {
		s.RedeemInof = append(s.RedeemInof, tmp)
	} else {
		r.update(tmp)
	}
	return nil
}
func (s *registerUnit) redeeming(hh uint64, amount *big.Int) (common.Address, *big.Int, error) {
	if amount.Cmp(s.getValidRedeem(hh)) > 0 {
		return common.Address{}, nil, params.ErrAmountOver
	}
	allAmount := big.NewInt(0)
	s.sortRedeemItems()
	for _, v := range s.RedeemInof {
		if v.isRedeem(hh) {
			allAmount = allAmount.Add(allAmount, v.Amount)
			res := allAmount.Cmp(amount)
			if res <= 0 {
				v.Amount, v.State = v.Amount.Sub(v.Amount, v.Amount), params.StateRedeemed
				if res == 0 {
					break
				}
			} else {
				v.State = params.StateRedeemed
				v.Amount.Set(new(big.Int).Sub(allAmount, amount))
				break
			}
		}
	}
	res := allAmount.Cmp(amount)
	if res >= 0 {
		return s.Address, amount, nil
	} else {
		return s.Address, allAmount, nil
	}
}

// called by user input and it will be execute without wait for the staking be rewarded
func (s *registerUnit) finishRedeemed() {
	pos := -1
	for i, v := range s.RedeemInof {
		if v.Amount.Sign() == 0 && v.State == params.StateRedeemed {
			pos = i
		}
	}
	if len(s.RedeemInof)-1 == pos {
		s.RedeemInof = s.RedeemInof[0:0]
	} else {
		s.RedeemInof = s.RedeemInof[pos+1:]
	}
}

// sort the redeemInof by asc with epochid
func (s *registerUnit) sortRedeemItems() {
	sort.Sort(redeemByID(s.RedeemInof))
}

// merge for move from prev to next epoch,move the staking who was to be voted.
// merge all staking to one staking with the new height(the beginning of next epoch).
// it will remove the staking which was canceled in the prev epoch
// called by move function.
func (s *registerUnit) merge(epochid, hh uint64) {

	all := big.NewInt(0)
	for _, v := range s.Value {
		all = all.Add(all, v.Amount)
	}
	tmp := &PairstakingValue{
		Amount: all,
		Height: new(big.Int).SetUint64(hh),
		State:  params.StateResgisterAuto,
	}
	var val []*PairstakingValue
	redeem := s.getRedeemItem(epochid)
	if redeem != nil {
		left := all.Sub(all, redeem.Amount)
		if left.Sign() >= 0 {
			tmp.Amount = left
		} else {
			panic("big error" + fmt.Sprint("all:", all, "redeem:", redeem.Amount, "epoch:", epochid))
		}
	}
	val = append(val, tmp)
	s.Value = val
}
func (s *registerUnit) update(unit *registerUnit, move bool) {
	sorter := valuesByHeight(s.Value)
	sort.Sort(sorter)
	for _, v := range unit.Value {
		sorter = sorter.update(v)
	}
	s.Value = sorter

	if s.RedeemInof == nil {
		s.RedeemInof = make([]*RedeemItem, 0)
	}
	if move {
		var tmp []*RedeemItem
		s.RedeemInof = append(append(tmp, unit.RedeemInof...), s.RedeemInof...)
	}
}
func (s *registerUnit) clone() *registerUnit {
	tmp := &registerUnit{
		Address:    s.Address,
		Value:      make([]*PairstakingValue, 0),
		RedeemInof: make([]*RedeemItem, 0),
	}
	for _, v := range s.Value {
		tmp.Value = append(tmp.Value, &PairstakingValue{
			Amount: new(big.Int).Set(v.Amount),
			Height: new(big.Int).Set(v.Height),
			State:  v.State,
		})
	}
	for _, v := range s.RedeemInof {
		tmp.RedeemInof = append(tmp.RedeemInof, &RedeemItem{
			Amount:  new(big.Int).Set(v.Amount),
			EpochID: v.EpochID,
			State:   v.State,
		})
	}
	return tmp
}
func (s *registerUnit) sort() {
	sort.Sort(valuesByHeight(s.Value))
	s.sortRedeemItems()
}
func (s *registerUnit) isValid() bool {
	for _, v := range s.Value {
		if v.Amount.Sign() > 0 {
			return true
		}
	}
	for _, v := range s.RedeemInof {
		if v.Amount.Sign() > 0 {
			return true
		}
	}
	return false
}
func (s *registerUnit) valueToMap() map[uint64]*big.Int {
	res := make(map[uint64]*big.Int)
	for _, v := range s.Value {
		res[v.Height.Uint64()] = new(big.Int).Set(v.Amount)
	}
	return res
}
func (s *registerUnit) redeemToMap() map[uint64]*big.Int {
	res := make(map[uint64]*big.Int)
	for _, v := range s.RedeemInof {
		res[v.EpochID] = new(big.Int).Set(v.Amount)
	}
	return res
}

/////////////////////////////////////////////////////////////////////////////////

type DelegationAccount struct {
	SaAddress common.Address
	Unit      *registerUnit
}

func (d *DelegationAccount) update(da *DelegationAccount, move bool) {
	d.Unit.update(da.Unit, move)
}
func (s *DelegationAccount) getAllStaking(hh uint64) *big.Int {
	return s.Unit.getAllStaking(hh)
}
func (s *DelegationAccount) getValidStaking(hh uint64) *big.Int {
	return s.Unit.getValidStaking(hh)
}
func (s *DelegationAccount) stopStakingInfo(amount, lastHeight *big.Int) error {
	return s.Unit.stopStakingInfo(amount, lastHeight)
}
func (s *DelegationAccount) redeeming(hh uint64, amount *big.Int) (common.Address, *big.Int, error) {
	return s.Unit.redeeming(hh, amount)
}
func (s *DelegationAccount) finishRedeemed() {
	s.Unit.finishRedeemed()
}
func (s *DelegationAccount) merge(epochid, hh uint64) {
	s.Unit.merge(epochid, hh)
}
func (s *DelegationAccount) clone() *DelegationAccount {
	return &DelegationAccount{
		SaAddress: s.SaAddress,
		Unit:      s.Unit.clone(),
	}
}
func (s *DelegationAccount) isValid() bool {
	return s.Unit.isValid()
}

type RegisterAccount struct {
	Unit       *registerUnit
	Votepubkey []byte
	Fee        *big.Int
	Relayer    bool
	Delegation []*DelegationAccount
	Modify     *AlterableInfo
}
type AlterableInfo struct {
	Fee        *big.Int
	VotePubkey []byte
}

func (s *RegisterAccount) isInRelayer() bool {
	return s.Relayer
}
func (s *RegisterAccount) addAmount(height uint64, amount *big.Int) {
	unit := &registerUnit{
		Address: s.Unit.Address,
		Value: []*PairstakingValue{&PairstakingValue{
			Amount: new(big.Int).Set(amount),
			Height: new(big.Int).SetUint64(height),
			State:  params.StateResgisterAuto,
		}},
		RedeemInof: make([]*RedeemItem, 0),
	}
	s.Unit.update(unit, false)
}
func (s *RegisterAccount) updateFee(height uint64, fee *big.Int) {
	if height > s.getMaxHeight() {
		s.Modify.Fee = new(big.Int).Set(fee)
	}
}
func (s *RegisterAccount) updatePk(height uint64, pk []byte) {
	if height > s.getMaxHeight() {
		s.Modify.VotePubkey = CopyVotePk(pk)
	}
}
func (s *RegisterAccount) update(sa *RegisterAccount, hh uint64, next, move bool) {
	s.Unit.update(sa.Unit, move)
	dirty := false
	for _, v := range sa.Delegation {
		da := s.getDA(v.Unit.GetRewardAddress())
		if da == nil {
			s.Delegation = append(s.Delegation, v)
			dirty = true
		} else {
			da.update(v, move)
		}
	}
	// ignore the pk param
	if hh > s.getMaxHeight() && s.Modify != nil && sa.Modify != nil {
		if sa.Modify.Fee != nil {
			s.Modify.Fee = new(big.Int).Set(sa.Modify.Fee)
		}
		if sa.Modify.VotePubkey != nil {
			s.Modify.VotePubkey = CopyVotePk(sa.Modify.VotePubkey)
		}
	}
	if next {
		s.changeAlterableInfo()
	}
	if dirty && hh != 0 {
		tmp := toDelegationByAmount(hh, false, s.Delegation)
		sort.Sort(tmp)
		s.Delegation, _ = fromDelegationByAmount(tmp)
	}
}
func (s *RegisterAccount) stopStakingInfo(amount, lastHeight *big.Int) error {
	return s.Unit.stopStakingInfo(amount, lastHeight)
}
func (s *RegisterAccount) redeeming(hh uint64, amount *big.Int) (common.Address, *big.Int, error) {
	return s.Unit.redeeming(hh, amount)
}
func (s *RegisterAccount) finishRedeemed() {
	s.Unit.finishRedeemed()
}
func (s *RegisterAccount) getAllStaking(hh uint64) *big.Int {
	all := s.Unit.getAllStaking(hh)
	for _, v := range s.Delegation {
		all = all.Add(all, v.getAllStaking(hh))
	}
	return all
}
func (s *RegisterAccount) getValidStaking(hh uint64) *big.Int {
	all := s.Unit.getValidStaking(hh)
	for _, v := range s.Delegation {
		all = all.Add(all, v.getValidStaking(hh))
	}
	return all
}
func (s *RegisterAccount) getValidStakingOnly(hh uint64) *big.Int {
	return s.Unit.getValidStaking(hh)
}
func (s *RegisterAccount) merge(epochid, hh, effectHeight uint64) {
	s.Unit.merge(epochid, hh)
	if hh >= effectHeight {
		das := make([]*DelegationAccount, 0, 0)
		for _, v := range s.Delegation {
			v.merge(epochid, hh)
			if v.isValid() {
				das = append(das, v)
			}
		}
		s.Delegation = das
	} else {
		for _, v := range s.Delegation {
			v.merge(epochid, hh)
		}
	}
}
func (s *RegisterAccount) getDA(addr common.Address) *DelegationAccount {
	for _, v := range s.Delegation {
		if bytes.Equal(v.Unit.Address.Bytes(), addr.Bytes()) {
			return v
		}
	}
	return nil
}
func (s *RegisterAccount) getMaxHeight() uint64 {
	l := len(s.Unit.Value)
	return s.Unit.Value[l-1].Height.Uint64()
}
func (s *RegisterAccount) changeAlterableInfo() {
	if s.Modify != nil {
		if s.Modify.Fee != nil && 0 != s.Modify.Fee.Cmp(params.InvalidFee) {
			if s.Modify.Fee.Sign() >= 0 && s.Modify.Fee.Cmp(params.Base) <= 0 {
				preFee := new(big.Int).Set(s.Fee)
				s.Fee = new(big.Int).Set(s.Modify.Fee)
				s.Modify.Fee = new(big.Int).Set(params.InvalidFee)
				log.Info("apply fee", "Address", s.Unit.GetRewardAddress(), "pre-fee", preFee.String(), "fee", s.Fee.String())
			}
		}
		if s.Modify.VotePubkey != nil && len(s.Modify.VotePubkey) >= 64 {
			if err := ValidPk(s.Modify.VotePubkey); err == nil {
				s.Votepubkey = CopyVotePk(s.Modify.VotePubkey)
				s.Modify.VotePubkey = []byte{}
			}
		}
	}
}
func (s *RegisterAccount) clone() *RegisterAccount {
	ss := &RegisterAccount{
		Votepubkey: CopyVotePk(s.Votepubkey),
		Unit:       s.Unit.clone(),
		Fee:        new(big.Int).Set(s.Fee),
		Relayer:    s.Relayer,
		Delegation: make([]*DelegationAccount, 0),
		Modify:     &AlterableInfo{},
	}
	for _, v := range s.Delegation {
		ss.Delegation = append(ss.Delegation, v.clone())
	}
	if s.Modify != nil {
		if s.Modify.Fee != nil {
			ss.Modify.Fee = new(big.Int).Set(s.Modify.Fee)
		}
		if s.Modify.VotePubkey != nil {
			ss.Modify.VotePubkey = CopyVotePk(s.Modify.VotePubkey)
		}
	}
	return ss
}
func (s *RegisterAccount) isvalid() bool {
	for _, v := range s.Delegation {
		if v.isValid() {
			return true
		}
	}
	return s.Unit.isValid()
}

// MakeModifyStateByTip10 once called by tip10
func (s *RegisterAccount) makeModifyStateByTip10() {
	if s.Modify == nil {
		s.Modify = &AlterableInfo{
			Fee:        new(big.Int).Set(params.InvalidFee),
			VotePubkey: []byte{},
		}
	} else {
		if s.Modify.Fee == nil || s.Modify.Fee.Sign() == 0 {
			s.Modify.Fee = new(big.Int).Set(params.InvalidFee)
			s.Modify.VotePubkey = []byte{}
		}
	}
}

type SARegister []*RegisterAccount

func (s *SARegister) getAllStaking(hh uint64) *big.Int {
	all := big.NewInt(0)
	for _, val := range *s {
		all = all.Add(all, val.getAllStaking(hh))
	}
	return all
}
func (s *SARegister) getValidStaking(hh uint64) *big.Int {
	all := big.NewInt(0)
	for _, val := range *s {
		all = all.Add(all, val.getValidStaking(hh))
	}
	return all
}
func (s *SARegister) sort(hh uint64, valid bool) {
	for _, v := range *s {
		tmp := toDelegationByAmount(hh, valid, v.Delegation)
		sort.Sort(tmp)
		v.Delegation, _ = fromDelegationByAmount(tmp)
	}
	tmp := toStakingByAmount(hh, valid, *s)
	sort.Sort(tmp)
	*s, _ = fromStakingByAmount(tmp)
}
func (s *SARegister) getSA(addr common.Address) *RegisterAccount {
	for _, val := range *s {
		if bytes.Equal(val.Unit.Address.Bytes(), addr.Bytes()) {
			return val
		}
	}
	return nil
}
func (s *SARegister) update(sa1 *RegisterAccount, hh uint64, next, move bool, effectHeight uint64) {
	sa := s.getSA(sa1.Unit.Address)
	if sa == nil {
		if hh >= effectHeight {
			sa1.changeAlterableInfo()
		}
		*s = append(*s, sa1)
		s.sort(hh, false)
	} else {
		sa.update(sa1, hh, next, move)
	}
}

/////////////////////////////////////////////////////////////////////////////////
// be thread-safe for caller locked
type RegisterImpl struct {
	accounts   map[uint64]SARegister // key is epoch id,value is SA set
	curEpochID uint64                // the new epochid of the current state
	lastReward uint64                // the curnent reward height block
}

func NewRegisterImpl() *RegisterImpl {
	pre := GetPreFirstEpoch()
	return &RegisterImpl{
		curEpochID: pre.EpochID,
		lastReward: 0,
		accounts:   make(map[uint64]SARegister),
	}
}
func CloneRegisterImpl(ori *RegisterImpl) *RegisterImpl {
	if ori == nil {
		return nil
	}
	tmp := &RegisterImpl{
		curEpochID: ori.curEpochID,
		lastReward: ori.lastReward,
		accounts:   make(map[uint64]SARegister),
	}
	for k, val := range ori.accounts {
		items := SARegister{}
		for _, v := range val {
			vv := v.clone()
			items = append(items, vv)
		}
		tmp.accounts[k] = items
	}
	return tmp
}

/////////////////////////////////////////////////////////////////////////////////
///////////  auxiliary function ////////////////////////////////////////////
func (i *RegisterImpl) getCurrentEpoch() uint64 {
	return i.curEpochID
}
func (i *RegisterImpl) SetCurrentEpoch(eid uint64) {
	i.curEpochID = eid
}
func (i *RegisterImpl) getMinEpochID() uint64 {
	eid := i.curEpochID
	for k, _ := range i.accounts {
		if eid > k {
			eid = k
		}
	}
	return eid
}
func (i *RegisterImpl) isInCurrentEpoch(hh uint64) bool {
	return i.curEpochID == GetEpochFromHeight(hh).EpochID
}
func (i *RegisterImpl) getCurrentEpochInfo() []*EpochIDInfo {
	var epochs []*EpochIDInfo
	var eids []float64
	for k, _ := range i.accounts {
		eids = append(eids, float64(k))
	}
	sort.Float64s(eids)
	for _, v := range eids {
		e := GetEpochFromID(uint64(v))
		if e != nil {
			epochs = append(epochs, e)
		}
	}
	return epochs
}

func (i *RegisterImpl) GetCurrentEpochInfo() ([]*EpochIDInfo, uint64) {
	return i.getCurrentEpochInfo(), i.getCurrentEpoch()
}

func (i *RegisterImpl) repeatPK(addr common.Address, pk []byte) bool {
	for _, v := range i.accounts {
		for _, vv := range v {
			if !bytes.Equal(addr.Bytes(), vv.Unit.Address.Bytes()) && bytes.Equal(pk, vv.Votepubkey) {
				return true
			}
		}
	}
	return false
}
func (i *RegisterImpl) GetStakingAccount(epochid uint64, addr common.Address) (*RegisterAccount, error) {
	if v, ok := i.accounts[epochid]; !ok {
		return nil, params.ErrInvalidStaking
	} else {
		for _, val := range v {
			if bytes.Equal(val.Unit.Address.Bytes(), addr.Bytes()) {
				return val, nil
			}
		}
	}
	return nil, params.ErrInvalidStaking
}
func (i *RegisterImpl) getDAfromSA(sa *RegisterAccount, addr common.Address) (*DelegationAccount, error) {
	for _, ii := range sa.Delegation {
		if bytes.Equal(ii.Unit.Address.Bytes(), addr.Bytes()) {
			return ii, nil
		}
	}
	return nil, nil
}
func (i *RegisterImpl) getElections(epochid uint64) []common.Address {
	if accounts, ok := i.accounts[epochid]; !ok {
		return nil
	} else {
		var addrs []common.Address
		for _, v := range accounts {
			if v.isInRelayer() {
				addrs = append(addrs, v.Unit.GetRewardAddress())
			}
		}
		return addrs
	}
}
func (i *RegisterImpl) getElections2(epochid uint64) []*RegisterAccount {
	if accounts, ok := i.accounts[epochid]; !ok {
		return nil
	} else {
		var sas []*RegisterAccount
		for _, v := range accounts {
			if v.isInRelayer() {
				sas = append(sas, v)
			}
		}
		return sas
	}
}
func (i *RegisterImpl) getElections3(epochid uint64) []*RegisterAccount {
	eid := epochid
	if eid >= params.FirstNewEpochID {
		eid = eid - 1
	}
	return i.getElections2(eid)
}
func (i *RegisterImpl) fetchAccountsInEpoch(epochid uint64, addrs []*RegisterAccount) []*RegisterAccount {
	if accounts, ok := i.accounts[epochid]; !ok {
		return addrs
	} else {
		find := func(addrs []*RegisterAccount, addr common.Address) bool {
			for _, v := range addrs {
				if bytes.Equal(v.Unit.Address.Bytes(), addr.Bytes()) {
					return true
				}
			}
			return false
		}
		items := make([]*RegisterAccount, 0, 0)
		for _, val := range accounts {
			if find(addrs, val.Unit.GetRewardAddress()) {
				items = append(items, val)
			}
		}
		return items
	}
}
func (i *RegisterImpl) redeemBySa(sa *RegisterAccount, height uint64, amount *big.Int) error {
	// can be redeem in the SA
	_, all, err1 := sa.redeeming(height, amount)
	if err1 != nil {
		return err1
	}
	if all.Cmp(amount) != 0 {
		return errors.New(fmt.Sprint(params.ErrRedeemAmount, "request amount", amount, "redeem amount", all))
	}
	sa.finishRedeemed()
	// fmt.Println("SA redeemed amount:[", all.String(), "],addr:[", addr.String())
	return nil
}
func (i *RegisterImpl) redeemByDa(da *DelegationAccount, height uint64, amount *big.Int) error {
	// can be redeem in the DA
	_, all, err1 := da.redeeming(height, amount)
	if err1 != nil {
		return err1
	}
	if all.Cmp(amount) != 0 {
		return errors.New(fmt.Sprint(params.ErrRedeemAmount, "request amount", amount, "redeem amount", all))
	}
	da.finishRedeemed()
	// fmt.Println("DA redeemed amount:[", all.String(), "],addr:[", addr.String())
	return nil
}
func (i *RegisterImpl) calcRewardInSa(target uint64, sa *RegisterAccount, allReward, allStaking *big.Int, item *RewardInfo) ([]*RewardInfo, error) {
	if sa == nil || allReward == nil || item == nil || allStaking == nil {
		return nil, params.ErrInvalidParam
	}
	var items []*RewardInfo
	fee := new(big.Int).Quo(new(big.Int).Mul(allReward, sa.Fee), params.Base)
	all, left, left2 := new(big.Int).Sub(allReward, fee), big.NewInt(0), big.NewInt(0)
	for _, v := range sa.Delegation {
		daAll := v.getAllStaking(target)
		if daAll.Sign() <= 0 {
			continue
		}
		v1 := new(big.Int).Quo(new(big.Int).Mul(all, daAll), allStaking)
		left = left.Add(left, v1)
		left2 = left2.Add(left2, daAll)
		var ii RewardInfo
		ii.Address, ii.Amount, ii.Register = v.Unit.GetRewardAddress(), new(big.Int).Set(v1), new(big.Int).Set(daAll)
		items = append(items, &ii)
	}
	item.Amount = new(big.Int).Add(new(big.Int).Sub(all, left), fee)
	item.Register = new(big.Int).Sub(allStaking, left2)
	return items, nil
}
func (i *RegisterImpl) calcReward(target, effectid uint64, allAmount *big.Int, einfo *EpochIDInfo) ([]*SARewardInfos, error) {
	if _, ok := i.accounts[einfo.EpochID]; !ok {
		return nil, params.ErrInvalidParam
	} else {
		sas := i.getElections3(einfo.EpochID)
		if sas == nil {
			return nil, errors.New(fmt.Sprint(params.ErrMatchEpochID, "epochid:", einfo.EpochID))
		}
		sas = i.fetchAccountsInEpoch(einfo.EpochID, sas)
		if len(sas) == 0 {
			return nil, errors.New(fmt.Sprint(params.ErrMatchEpochID, "epochid:", einfo.EpochID, "sas=0"))
		}
		impawns := SARegister(sas)
		impawns.sort(target, false)
		var res []*SARewardInfos
		allValidatorStaking := impawns.getAllStaking(target)
		sum := len(impawns)
		left := big.NewInt(0)

		for pos, v := range impawns {
			var info SARewardInfos
			var item RewardInfo
			item.Address = v.Unit.GetRewardAddress()
			allStaking := v.getAllStaking(target)
			if allStaking.Sign() <= 0 {
				continue
			}

			v2 := new(big.Int).Quo(new(big.Int).Mul(allStaking, allAmount), allValidatorStaking)
			if pos == sum-1 {
				v2 = new(big.Int).Sub(allAmount, left)
			}
			left = left.Add(left, v2)

			if ii, err := i.calcRewardInSa(target, v, v2, allStaking, &item); err != nil {
				return nil, err
			} else {
				info.Items = append(info.Items, &item)
				info.Items = append(info.Items, ii[:]...)
			}
			res = append(res, &info)
		}
		return res, nil
	}
}
func (i *RegisterImpl) reward(begin, end, effectid uint64, allAmount *big.Int) ([]*SARewardInfos, error) {
	ids := GetEpochFromRange(begin, end)
	if ids == nil || len(ids) > 2 {
		return nil, errors.New(fmt.Sprint(params.ErrMatchEpochID, "more than 2 epochid:", begin, end))
	}

	if len(ids) == 2 {
		tmp := new(big.Int).Quo(new(big.Int).Mul(allAmount, new(big.Int).SetUint64(ids[0].EndHeight-begin+1)), new(big.Int).SetUint64(end-begin+1))
		amount1, amount2 := tmp, new(big.Int).Sub(allAmount, tmp)
		// log.Info("*****reward", "begin", begin, "end", end, "allAmount", allAmount,"amount1",amount1,"amount2",
		// amount2,"ids[0]",ids[0].String(),"ids[1]",ids[1].String())
		if items, err := i.calcReward(ids[0].EndHeight, effectid, amount1, ids[0]); err != nil {
			return nil, err
		} else {
			if items1, err2 := i.calcReward(end, effectid, amount2, ids[1]); err2 != nil {
				return nil, err2
			} else {
				items = append(items, items1[:]...)
			}
			return items, nil
		}
	} else {
		return i.calcReward(end, effectid, allAmount, ids[0])
	}
}

//func (i *RegisterImpl) Reward(begin, end, effectid uint64, allAmount *big.Int) ([]*types.SARewardInfos, error) {
//	return i.reward(begin,effectid,abiStaking)
//}
///////////auxiliary function ////////////////////////////////////////////

/////////////////////////////////////////////////////////////////////////////////
// move the accounts from prev to next epoch and keeps the prev account still here
func (i *RegisterImpl) move(prev, next, effectHeight uint64) error {
	nextEpoch := GetEpochFromID(next)
	if nextEpoch == nil {
		return params.ErrOverEpochID
	}
	prevInfos, ok := i.accounts[prev]
	nextInfos, ok2 := i.accounts[next]
	if !ok {
		return errors.New(fmt.Sprintln("the epoch is nil", prev, "err:", params.ErrNotMatchEpochInfo))
	}
	if !ok2 {
		nextInfos = SARegister{}
	}
	for _, v := range prevInfos {
		vv := v.clone()
		vv.merge(prev, nextEpoch.BeginHeight, effectHeight)
		if vv.isvalid() {
			vv.Relayer = false
			nextInfos.update(vv, nextEpoch.BeginHeight, true, true, effectHeight)
		}
	}
	i.accounts[next] = nextInfos
	return nil
}

/////////////////////////////////////////////////////////////////////////////////
////////////// external function //////////////////////////////////////////

// DoElections called by consensus while it closer the end of epoch,have 500~1000 fast block
func (i *RegisterImpl) DoElections(state StateDB, epochid, height uint64) ([]*RegisterAccount, error) {
	if epochid < params.FirstNewEpochID && epochid != i.getCurrentEpoch()+1 {
		return nil, params.ErrOverEpochID
	}
	cur := GetEpochFromID(i.curEpochID)
	if cur.EndHeight != height+params.ElectionPoint && i.curEpochID >= params.FirstNewEpochID {
		return nil, params.ErrNotElectionTime
	}
	// e := types.GetEpochFromID(epochid)
	eid := epochid
	if eid >= params.FirstNewEpochID {
		eid = eid - 1
	}
	if val, ok := i.accounts[eid]; ok {
		val.sort(height, true)
		var ee []*RegisterAccount
		for _, v := range val {
			validStaking := v.getValidStakingOnly(height)
			num, _ := HistoryWorkEfficiency(state, epochid, v.Unit.Address)
			if validStaking.Cmp(params.ElectionMinLimitForRegister) < 0 && num < params.MinWorkEfficiency && i.curEpochID > 1 {
				continue
			}
			v.Relayer = true
			ee = append(ee, v)
			if len(ee) >= params.CountInEpoch {
				break
			}
		}
		return ee, nil
	} else {
		return nil, params.ErrMatchEpochID
	}
}

// Shift will move the staking account which has election flag to the next epoch
// it will be save the whole state in the current epoch end block after it called by consensus
func (i *RegisterImpl) Shift(epochid, effectHeight uint64) error {
	lastReward := i.lastReward
	minEpoch := GetEpochFromHeight(lastReward)
	min := i.getMinEpochID()
	// fmt.Println("*** move min:", min, "minEpoch:", minEpoch.EpochID, "lastReward:", i.lastReward)
	for ii := min; minEpoch.EpochID > 1 && ii < minEpoch.EpochID-1; ii++ {
		delete(i.accounts, ii)
		// fmt.Println("delete epoch:", ii)
	}

	if epochid != i.getCurrentEpoch()+1 {
		return params.ErrOverEpochID
	}
	i.SetCurrentEpoch(epochid)
	prev := epochid - 1
	return i.move(prev, epochid, effectHeight)
}

// CancelSAccount cancel amount of asset for staking account,it will be work in next epoch
func (i *RegisterImpl) CancelSAccount(curHeight uint64, addr common.Address, amount *big.Int) error {
	if amount.Sign() <= 0 || curHeight <= 0 {
		return params.ErrInvalidParam
	}
	curEpoch := GetEpochFromHeight(curHeight)
	if curEpoch == nil || curEpoch.EpochID != i.curEpochID {
		return params.ErrInvalidParam
	}
	sa, err := i.GetStakingAccount(curEpoch.EpochID, addr)
	if err != nil {
		return err
	}
	err2 := sa.stopStakingInfo(amount, new(big.Int).SetUint64(curHeight))
	// fmt.Println("[SA]insert a redeem,address:[", addr.String(), "],amount:[", amount.String(), "],height:", curHeight, "]err:", err2)
	return err2
}

// CancelDAccount cancel amount of asset for delegation account,it will be work in next epoch
func (i *RegisterImpl) CancelDAccount(curHeight uint64, addrSA, addrDA common.Address, amount *big.Int) error {
	if amount.Sign() <= 0 || curHeight <= 0 {
		return params.ErrInvalidParam
	}
	curEpoch := GetEpochFromHeight(curHeight)
	if curEpoch == nil || curEpoch.EpochID != i.curEpochID {
		return params.ErrInvalidParam
	}
	sa, err := i.GetStakingAccount(curEpoch.EpochID, addrSA)
	if err != nil {
		return err
	}
	da, err2 := i.getDAfromSA(sa, addrDA)
	if err2 != nil {
		return err
	}
	if da == nil {
		log.Error("CancelDAccount error", "height", curHeight, "SA", addrSA, "DA", addrDA)
		return params.ErrNotDelegation
	}
	err3 := da.stopStakingInfo(amount, new(big.Int).SetUint64(curHeight))
	// fmt.Println("[DA]insert a redeem,address:[", addrSA.String(), "],DA address:[", addrDA.String(), "],amount:[", amount.String(), "],height:", curHeight, "]err:", err3)
	return err3
}

// RedeemSAccount redeem amount of asset for staking account,it will locked for a certain time
func (i *RegisterImpl) RedeemSAccount(curHeight uint64, addr common.Address, amount *big.Int) error {
	if amount.Sign() <= 0 || curHeight <= 0 {
		return params.ErrInvalidParam
	}
	curEpoch := GetEpochFromHeight(curHeight)
	if curEpoch == nil || curEpoch.EpochID != i.curEpochID {
		return params.ErrInvalidParam
	}
	sa, err := i.GetStakingAccount(curEpoch.EpochID, addr)
	if err != nil {
		return err
	}
	return i.redeemBySa(sa, curHeight, amount)
}

// RedeemDAccount redeem amount of asset for delegation account,it will locked for a certain time
func (i *RegisterImpl) RedeemDAccount(curHeight uint64, addrSA, addrDA common.Address, amount *big.Int) error {
	if amount.Sign() <= 0 || curHeight <= 0 {
		return params.ErrInvalidParam
	}
	curEpoch := GetEpochFromHeight(curHeight)
	if curEpoch == nil || curEpoch.EpochID != i.curEpochID {
		return params.ErrInvalidParam
	}
	sa, err := i.GetStakingAccount(curEpoch.EpochID, addrSA)
	if err != nil {
		return err
	}
	da, err2 := i.getDAfromSA(sa, addrDA)
	if err2 != nil {
		return err
	}
	if da == nil {
		log.Error("RedeemDAccount error", "height", curHeight, "SA", addrSA, "DA", addrDA)
		return params.ErrNotDelegation
	}
	return i.redeemByDa(da, curHeight, amount)
}

func (i *RegisterImpl) insertDAccount(height uint64, da *DelegationAccount) error {
	if da == nil {
		return params.ErrInvalidParam
	}
	epochInfo := GetEpochFromHeight(height)
	if epochInfo == nil || epochInfo.EpochID > i.getCurrentEpoch() {
		return params.ErrOverEpochID
	}
	sa, err := i.GetStakingAccount(epochInfo.EpochID, da.SaAddress)
	if err != nil {
		return err
	}
	if ds, err := i.getDAfromSA(sa, da.Unit.Address); err != nil {
		return err
	} else {
		if ds == nil {
			sa.Delegation = append(sa.Delegation, da)
			log.Debug("Insert delegation account", "staking account", sa.Unit.GetRewardAddress(), "account", da.Unit.GetRewardAddress())
		} else {
			ds.update(da, false)
			log.Debug("Update delegation account", "staking account", sa.Unit.GetRewardAddress(), "account", da.Unit.GetRewardAddress())
		}
	}
	return nil
}
func (i *RegisterImpl) InsertDAccount2(height uint64, addrSA, addrDA common.Address, val *big.Int) error {
	if val.Sign() <= 0 || height < 0 {
		return params.ErrInvalidParam
	}
	if bytes.Equal(addrSA.Bytes(), addrDA.Bytes()) {
		return params.ErrDelegationSelf
	}
	state := uint8(0)
	state |= params.StateResgisterAuto
	da := &DelegationAccount{
		SaAddress: addrSA,
		Unit: &registerUnit{
			Address: addrDA,
			Value: []*PairstakingValue{&PairstakingValue{
				Amount: new(big.Int).Set(val),
				Height: new(big.Int).SetUint64(height),
				State:  state,
			}},
			RedeemInof: make([]*RedeemItem, 0),
		},
	}
	return i.insertDAccount(height, da)
}
func (i *RegisterImpl) insertSAccount(height uint64, sa *RegisterAccount) error {
	if sa == nil {
		return params.ErrInvalidParam
	}
	epochInfo := GetEpochFromHeight(height)
	if epochInfo == nil || epochInfo.EpochID > i.getCurrentEpoch() {
		log.Error("insertSAccount", "eid", epochInfo.EpochID, "height", height, "eid2", i.getCurrentEpoch())
		return params.ErrOverEpochID
	}
	if val, ok := i.accounts[epochInfo.EpochID]; !ok {
		var accounts []*RegisterAccount
		accounts = append(accounts, sa)
		i.accounts[epochInfo.EpochID] = SARegister(accounts)
		log.Debug("Insert staking account", "epoch", epochInfo, "account", sa.Unit.GetRewardAddress())
	} else {
		for _, ii := range val {
			if bytes.Equal(ii.Unit.Address.Bytes(), sa.Unit.Address.Bytes()) {
				ii.update(sa, height, false, false)
				log.Debug("Update staking account", "account", sa.Unit.GetRewardAddress())
				return nil
			}
		}
		i.accounts[epochInfo.EpochID] = append(val, sa)
		log.Debug("Insert staking account", "epoch", epochInfo, "account", sa.Unit.GetRewardAddress())
	}
	return nil
}
func (i *RegisterImpl) InsertSAccount2(height, effectHeight uint64, addr common.Address, pk []byte, val *big.Int, fee *big.Int, auto bool) error {
	if val.Sign() <= 0 || height < 0 || fee.Sign() < 0 || fee.Cmp(params.Base) > 0 {
		return params.ErrInvalidParam
	}
	if err := ValidPk(pk); err != nil {
		return err
	}
	if i.repeatPK(addr, pk) {
		log.Error("Insert SA account repeat pk", "addr", addr, "pk", pk)
		return params.ErrRepeatPk
	}
	state := uint8(0)
	if auto {
		state |= params.StateResgisterAuto
	}
	sa := &RegisterAccount{
		Votepubkey: append([]byte{}, pk...),
		Fee:        new(big.Int).Set(fee),
		Unit: &registerUnit{
			Address: addr,
			Value: []*PairstakingValue{&PairstakingValue{
				Amount: new(big.Int).Set(val),
				Height: new(big.Int).SetUint64(height),
				State:  state,
			}},
			RedeemInof: make([]*RedeemItem, 0),
		},
		Modify: &AlterableInfo{},
	}
	if height >= effectHeight {
		sa.Modify = &AlterableInfo{
			Fee:        new(big.Int).Set(params.InvalidFee),
			VotePubkey: []byte{},
		}
	}
	return i.insertSAccount(height, sa)
}
func (i *RegisterImpl) AppendSAAmount(height uint64, addr common.Address, val *big.Int) error {
	if val.Sign() <= 0 || height < 0 {
		return params.ErrInvalidParam
	}
	epochInfo := GetEpochFromHeight(height)
	if epochInfo.EpochID > i.getCurrentEpoch() {
		log.Debug("insertSAccount", "eid", epochInfo.EpochID, "height", height, "eid2", i.getCurrentEpoch())
		return params.ErrOverEpochID
	}
	sa, err := i.GetStakingAccount(epochInfo.EpochID, addr)
	if err != nil {
		return err
	}
	sa.addAmount(height, val)
	return nil
}
func (i *RegisterImpl) UpdateSAFee(height uint64, addr common.Address, fee *big.Int) error {
	if height < 0 || fee.Sign() < 0 || fee.Cmp(params.Base) > 0 {
		return params.ErrInvalidParam
	}
	epochInfo := GetEpochFromHeight(height)
	if epochInfo.EpochID > i.getCurrentEpoch() {
		log.Info("UpdateSAFee", "eid", epochInfo.EpochID, "height", height, "eid2", i.getCurrentEpoch())
		return params.ErrOverEpochID
	}
	sa, err := i.GetStakingAccount(epochInfo.EpochID, addr)
	if err != nil {
		return err
	}
	sa.updateFee(height, fee)
	return nil
}
func (i *RegisterImpl) UpdateSAPK(height uint64, addr common.Address, pk []byte) error {
	if height < 0 {
		return params.ErrInvalidParam
	}
	if err := ValidPk(pk); err != nil {
		return err
	}
	if i.repeatPK(addr, pk) {
		log.Error("UpdateSAPK repeat pk", "addr", addr, "pk", pk)
		return params.ErrRepeatPk
	}
	epochInfo := GetEpochFromHeight(height)
	if epochInfo.EpochID > i.getCurrentEpoch() {
		log.Info("UpdateSAPK", "eid", epochInfo.EpochID, "height", height, "eid2", i.getCurrentEpoch())
		return params.ErrOverEpochID
	}
	sa, err := i.GetStakingAccount(epochInfo.EpochID, addr)
	if err != nil {
		return err
	}
	sa.updatePk(height, pk)
	return nil
}

//func (i *RegisterImpl) Reward(block *types.SnailBlock, allAmount *big.Int, effectid uint64) ([]*SARewardInfos, error) {
//	begin, end := types.FromBlock(block)
//	res, err := i.reward(begin, end, effectid, allAmount)
//	if err == nil {
//		i.lastReward = end
//	}
//	return res, err
//}
func (i *RegisterImpl) Reward2(begin, end, effectid uint64, allAmount *big.Int) ([]*SARewardInfos, error) {

	res, err := i.reward(begin, end, effectid, allAmount)
	if err == nil {
		i.lastReward = end
	}
	return res, err
}

/////////////////////////////////////////////////////////////////////////////////
// GetStakings return all staking accounts of the current epoch
func (i *RegisterImpl) GetAllStakingAccount() SARegister {
	if val, ok := i.accounts[i.curEpochID]; ok {
		return val
	} else {
		return nil
	}
}

// GetStakingAsset returns a map for all staking amount of the address, the key is the SA address
func (i *RegisterImpl) GetStakingAsset(addr common.Address) map[common.Address]*RelayerValue {
	epochid := i.curEpochID
	res, _, _, _ := i.getAsset(addr, epochid, params.OpQueryStaking)
	return res
}

// GetLockedAsset returns a group canceled asset from the state of the addr,it includes redemption on
// maturity and unmaturity asset
func (i *RegisterImpl) GetLockedAsset(addr common.Address) map[common.Address]*RelayerValue {
	epochid := i.curEpochID
	res, _, _, _ := i.getAsset(addr, epochid, params.OpQueryLocked)
	return res
}
func (i *RegisterImpl) GetLockedAsset2(addr common.Address, height uint64) map[common.Address]*LockedValue {
	epochid := i.curEpochID
	items, _, _, _ := i.getAsset(addr, epochid, params.OpQueryLocked)
	res := make(map[common.Address]*LockedValue)
	for k, v := range items {
		res[k] = v.ToLockedValue(height)
	}
	return res
}

func (i *RegisterImpl) GetBalance(addr common.Address) (*big.Int, *big.Int, *big.Int, *big.Int, *big.Int) {
	epochid := i.curEpochID
	locked, _, _, _ := i.getAsset(addr, epochid, params.OpQueryLocked)
	staked, _, _, _ := i.getAsset(addr, epochid, params.OpQueryStaking)
	_, unlock, _, _ := i.getAsset(addr, epochid, params.OpQueryCancelable)
	_, _, reward, _ := i.getAsset(addr, epochid, params.OpQueryReward)
	_, _, _, fine := i.getAsset(addr, epochid, params.OpQueryFine)
	var l *big.Int
	var s *big.Int
	var r *big.Int
	var f *big.Int
	if locked[addr] != nil {
		l = locked[addr].Value[epochid]
	}
	if staked[addr] != nil {
		s = staked[addr].Value[epochid]
	}
	if reward[addr] != nil {
		r = reward[addr].Amount
	}
	if fine[addr] != nil {
		f = fine[addr].Amount
	}
	return l, s, unlock[addr], r, f
}

func (i *RegisterImpl) GetReward(addr common.Address) map[common.Address]*RewardItem {
	epochid := i.curEpochID
	_, _, res, _ := i.getAsset(addr, epochid, params.OpQueryReward) //types.OpQueryRewarded
	return res
}

func (i *RegisterImpl) GetFine(addr common.Address) map[common.Address]*FineItem {
	epochid := i.curEpochID
	_, _, _, res := i.getAsset(addr, epochid, params.OpQueryFine) //types.OpQueryFine
	return res
}

// GetAllCancelableAsset returns all asset on addr it can be canceled
func (i *RegisterImpl) GetAllCancelableAsset(addr common.Address) map[common.Address]*big.Int {
	epochid := i.curEpochID
	_, res, _, _ := i.getAsset(addr, epochid, params.OpQueryCancelable)
	return res
}
func (i *RegisterImpl) getAsset(addr common.Address, epoch uint64, op uint8) (map[common.Address]*RelayerValue, map[common.Address]*big.Int, map[common.Address]*RewardItem, map[common.Address]*FineItem) {
	epochid := epoch
	end := GetEpochFromID(epochid).EndHeight
	if val, ok := i.accounts[epochid]; ok {
		res := make(map[common.Address]*RelayerValue)
		res2 := make(map[common.Address]*big.Int)
		res3 := make(map[common.Address]*RewardItem)
		res4 := make(map[common.Address]*FineItem)
		for _, v := range val {
			if bytes.Equal(v.Unit.Address.Bytes(), addr.Bytes()) {
				if op&params.OpQueryStaking != 0 || op&params.OpQueryLocked != 0 {
					if _, ok := res[addr]; !ok {
						if op&params.OpQueryLocked != 0 {
							res[addr] = &RelayerValue{
								Value: v.Unit.redeemToMap(),
							}
						} else {
							res[addr] = &RelayerValue{
								Value: v.Unit.valueToMap(),
							}
						}

					} else {
						log.Error("getAsset", "repeat staking account", addr, "epochid", epochid, "op", op)
					}
				}
				if op&params.OpQueryCancelable != 0 {
					all := v.Unit.getValidStaking(end)
					if all.Sign() >= 0 {
						res2[addr] = all
					}
				}
				if op&params.OpQueryReward != 0 {
					res3[addr] = &RewardItem{
						Amount: v.Unit.getReward(epochid),
					}
				}
				if op&params.OpQueryFine != 0 {
					res4[addr] = &FineItem{
						Amount: v.Unit.getFine(epochid),
					}
				}

				continue
			}
		}
		return res, res2, res3, res4
	} else {
		log.Error("getAsset", "wrong epoch in current", epochid)
	}
	return nil, nil, nil, nil
}
func (i *RegisterImpl) MakeModifyStateByTip10() {
	if val, ok := i.accounts[i.curEpochID]; ok {
		for _, v := range val {
			v.makeModifyStateByTip10()
		}
		log.Info("relayerCLI: MakeModifyStateByTip10")
	}
}

/////////////////////////////////////////////////////////////////////////////////
// storage layer
func (i *RegisterImpl) GetRoot() common.Hash {
	return common.Hash{}
}
func (i *RegisterImpl) Save(state StateDB, preAddress common.Address) error {
	key := common.BytesToHash(preAddress[:])
	data, err := rlp.EncodeToBytes(i)

	if err != nil {
		log.Crit("Failed to RLP encode RegisterImpl", "err", err)
	}
	hash := RlpHash(data)
	//state.SetPOSState(preAddress, key, data)
	state.SetState(preAddress, key, hash)
	tmp := CloneRegisterImpl(i)
	if tmp != nil {
		IC.Cache.Add(hash, tmp)
	}
	return err
}
func (i *RegisterImpl) Load(state StateDB, preAddress common.Address) error {
	key := common.BytesToHash(preAddress[:])
	//data := state.GetPOSState(preAddress, key)
	hash := state.GetState(preAddress, key)
	data, err := rlp.EncodeToBytes(hash)
	if err != nil {
		return errors.New("EncodeToBytes failed")
	}
	//lenght := len(data)
	//if lenght == 0 {
	//	return errors.New("Load data = 0")
	//}
	// cache := true
	//hash := RlpHash(data)
	var temp RegisterImpl
	if cc, ok := IC.Cache.Get(hash); ok {
		impawn := cc.(*RegisterImpl)
		temp = *(CloneRegisterImpl(impawn))
	} else {
		if err := rlp.DecodeBytes(data, &temp); err != nil {
			log.Error("Invalid RegisterImpl entry RLP", "err", err)
			return errors.New(fmt.Sprintf("Invalid RegisterImpl entry RLP %s", err.Error()))
		}
		tmp := CloneRegisterImpl(&temp)
		if tmp != nil {
			IC.Cache.Add(hash, tmp)
		}
		// cache = false
	}
	// log.Info("-----Load relayerCLI---","len:",lenght,"count:",temp.Counts(),"cache",cache)
	i.curEpochID, i.accounts, i.lastReward = temp.curEpochID, temp.accounts, temp.lastReward
	return nil
}

func GetCurrentRelayer(state StateDB) []*params.RelayerMember {
	i := NewRegisterImpl()
	i.Load(state, params.RelayerAddress)
	eid := i.getCurrentEpoch()
	accs := i.getElections3(eid)
	var vv []*params.RelayerMember
	for _, v := range accs {
		pubkey, _ := crypto.UnmarshalPubkey(v.Votepubkey)
		vv = append(vv, &params.RelayerMember{
			RelayerBase: crypto.PubkeyToAddress(*pubkey),
			Coinbase:    v.Unit.GetRewardAddress(),
			Publickey:   CopyVotePk(v.Votepubkey),
			Flag:        params.StateUsedFlag,
			MType:       params.TypeWorked,
		})
	}
	return vv
}

func GetRelayersByEpoch(state StateDB, eid, hh uint64) []*params.RelayerMember {
	i := NewRegisterImpl()
	err := i.Load(state, params.RelayerAddress)
	accs := i.getElections3(eid)
	first := GetFirstEpoch()
	if hh == first.EndHeight-params.ElectionPoint {
		fmt.Println("****** accounts len:", len(i.accounts), "election:", len(accs), " err ", err)
	}
	var vv []*params.RelayerMember
	for _, v := range accs {
		pubkey, _ := crypto.UnmarshalPubkey(v.Votepubkey)
		vv = append(vv, &params.RelayerMember{
			RelayerBase: crypto.PubkeyToAddress(*pubkey),
			Coinbase:    v.Unit.GetRewardAddress(),
			Publickey:   CopyVotePk(v.Votepubkey),
			Flag:        params.StateUsedFlag,
			MType:       params.TypeWorked,
		})
	}
	return vv
}
func (i *RegisterImpl) Counts() int {
	pos := 0
	for _, val := range i.accounts {
		for _, vv := range val {
			pos = pos + len(vv.Delegation)
		}
		pos = pos + len(val)
	}
	return pos
}
func (i *RegisterImpl) Summay() *RegisterSummay {
	summay := &RegisterSummay{
		LastReward: i.lastReward,
		Infos:      make([]*SummayEpochInfo, 0, 0),
	}
	sumAccount := 0
	for k, val := range i.accounts {
		info := GetEpochFromID(k)
		item := &SummayEpochInfo{
			EpochID:     info.EpochID,
			BeginHeight: info.BeginHeight,
			EndHeight:   info.EndHeight,
		}
		item.AllAmount = val.getValidStaking(info.EndHeight)
		daSum, saSum := 0, len(val)
		for _, vv := range val {
			daSum = daSum + len(vv.Delegation)
		}
		item.DaCount, item.SaCount = uint64(daSum), uint64(saSum)
		summay.Infos = append(summay.Infos, item)
		sumAccount = sumAccount + daSum + saSum
		if i.curEpochID == k {
			summay.AllAmount = new(big.Int).Set(item.AllAmount)
		}
	}
	summay.Accounts = uint64(sumAccount)
	return summay
}

/////////////////////////////////////////////////////////////////////////////////
type valuesByHeight []*PairstakingValue

func (vs valuesByHeight) Len() int {
	return len(vs)
}
func (vs valuesByHeight) Less(i, j int) bool {
	return vs[i].Height.Cmp(vs[j].Height) == -1
}
func (vs valuesByHeight) Swap(i, j int) {
	it := vs[i]
	vs[i] = vs[j]
	vs[j] = it
}
func (vs valuesByHeight) find(hh uint64) (*PairstakingValue, int) {
	low, height := 0, len(vs)-1
	mid := 0
	for low <= height {
		mid = (height + low) / 2
		if hh == vs[mid].Height.Uint64() {
			return vs[mid], mid
		} else if hh > vs[mid].Height.Uint64() {
			low = mid + 1
			if low > height {
				return nil, low
			}
		} else {
			height = mid - 1
		}
	}
	return nil, mid
}
func (vs valuesByHeight) update(val *PairstakingValue) valuesByHeight {
	item, pos := vs.find(val.Height.Uint64())
	if item != nil {
		item.Amount = item.Amount.Add(item.Amount, val.Amount)
		item.State |= val.State
	} else {
		rear := append([]*PairstakingValue{}, vs[pos:]...)
		vs = append(append(vs[:pos], val), rear...)
	}
	return vs
}

type redeemByID []*RedeemItem

func (vs redeemByID) Len() int {
	return len(vs)
}
func (vs redeemByID) Less(i, j int) bool {
	return vs[i].EpochID < vs[j].EpochID
}
func (vs redeemByID) Swap(i, j int) {
	it := vs[i]
	vs[i] = vs[j]
	vs[j] = it
}

type stakingItem struct {
	item   *RegisterAccount
	height uint64
	valid  bool
}

func (s *stakingItem) getAll() *big.Int {
	if s.valid {
		return s.item.getValidStaking(s.height)
	} else {
		return s.item.getAllStaking(s.height)
	}
}

type stakingByAmount []*stakingItem

func toStakingByAmount(hh uint64, valid bool, items []*RegisterAccount) stakingByAmount {
	var tmp []*stakingItem
	for _, v := range items {
		v.Unit.sort()
		tmp = append(tmp, &stakingItem{
			item:   v,
			height: hh,
			valid:  valid,
		})
	}
	return stakingByAmount(tmp)
}
func fromStakingByAmount(items stakingByAmount) ([]*RegisterAccount, uint64) {
	var tmp []*RegisterAccount
	var vv uint64
	for _, v := range items {
		tmp = append(tmp, v.item)
		vv = v.height
	}
	return tmp, vv
}
func (vs stakingByAmount) Len() int {
	return len(vs)
}
func (vs stakingByAmount) Less(i, j int) bool {
	return vs[i].getAll().Cmp(vs[j].getAll()) > 0
}
func (vs stakingByAmount) Swap(i, j int) {
	it := vs[i]
	vs[i] = vs[j]
	vs[j] = it
}

type delegationItem struct {
	item   *DelegationAccount
	height uint64
	valid  bool
}

func (d *delegationItem) getAll() *big.Int {
	if d.valid {
		return d.item.getValidStaking(d.height)
	} else {
		return d.item.getAllStaking(d.height)
	}
}

type delegationItemByAmount []*delegationItem

func toDelegationByAmount(hh uint64, valid bool, items []*DelegationAccount) delegationItemByAmount {
	var tmp []*delegationItem
	for _, v := range items {
		v.Unit.sort()
		tmp = append(tmp, &delegationItem{
			item:   v,
			height: hh,
			valid:  valid,
		})
	}
	return delegationItemByAmount(tmp)
}
func fromDelegationByAmount(items delegationItemByAmount) ([]*DelegationAccount, uint64) {
	var tmp []*DelegationAccount
	var vv uint64
	for _, v := range items {
		tmp = append(tmp, v.item)
		vv = v.height
	}
	return tmp, vv
}
func (vs delegationItemByAmount) Len() int {
	return len(vs)
}
func (vs delegationItemByAmount) Less(i, j int) bool {
	return vs[i].getAll().Cmp(vs[j].getAll()) > 0
}
func (vs delegationItemByAmount) Swap(i, j int) {
	it := vs[i]
	vs[i] = vs[j]
	vs[j] = it
}

func GetCurrentEpochID(evm *EVM) (uint64, error) {
	impawn := NewRegisterImpl()
	err := impawn.Load(evm.StateDB, HeaderStoreAddress)
	if err != nil {
		log.Error("relayerCLI load error", "error", err)
		return 0, err
	}
	return impawn.getCurrentEpoch(), nil
}
