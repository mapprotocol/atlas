package vm

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	lru "github.com/hashicorp/golang-lru"
	"github.com/mapprotocol/atlas/params"
	"io"
	"math/big"
	"sort"
)

/////////////////////////////////////////////////////////////////////////////////
var IC *RelayerCache

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
type PairRegisterValue struct {
	Amount *big.Int
	Height *big.Int
	State  uint8
}

func (v *PairRegisterValue) isElection() bool {
	return (v.State&params.StateRegisterOnce != 0 || v.State&params.StateResgisterAuto != 0)
}

type RewardItem struct {
	EpochID uint64
	Amount  *big.Int
	Height  *big.Int
}
type FineItem struct {
	EpochID uint64
	Amount  *big.Int
	Height  *big.Int
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
	Value      []*PairRegisterValue // sort by height
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

func (s *registerUnit) getAllRegister(hh uint64) *big.Int {
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
func (s *registerUnit) getValidRegister(hh uint64) *big.Int {
	all := big.NewInt(0)
	for _, v := range s.Value {
		if v.Height.Uint64() <= hh-params.ElectionPoint {
			//if v.isElection() {
			all = all.Add(all, v.Amount)
			//}
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
			log.Error("getValidRegister error", "all amount", all, "redeem amount", r.Amount, "height", hh)
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

// stopRegisterInfo redeem delay in next epoch + MaxRedeemHeight
func (s *registerUnit) stopRegisterInfo(amount, lastHeight *big.Int) error {
	all := s.getValidRegister(lastHeight.Uint64())
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
		State:   params.StateUnregister,
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
				v.Amount, v.State = v.Amount.Sub(v.Amount, v.Amount), params.StateUnregistered
				if res == 0 {
					break
				}
			} else {
				v.State = params.StateUnregistered
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

// called by user input and it will be execute without wait for the relayer be rewarded
func (s *registerUnit) finishRedeemed() {
	pos := -1
	for i, v := range s.RedeemInof {
		if v.Amount.Sign() == 0 && v.State == params.StateUnregistered {
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

// merge for move from prev to next epoch,move the register who was to be voted.
// merge all register to one register with the new height(the beginning of next epoch).
// it will remove the register which was canceled in the prev epoch
// called by move function.
func (s *registerUnit) merge(epochid, hh uint64) {

	all := big.NewInt(0)
	for _, v := range s.Value {
		all = all.Add(all, v.Amount)
	}
	tmp := &PairRegisterValue{
		Amount: all,
		Height: new(big.Int).SetUint64(hh),
		State:  params.StateResgisterAuto,
	}
	var val []*PairRegisterValue
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
		Value:      make([]*PairRegisterValue, 0),
		RedeemInof: make([]*RedeemItem, 0),
	}
	for _, v := range s.Value {
		tmp.Value = append(tmp.Value, &PairRegisterValue{
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

type RegisterAccount struct {
	Unit    *registerUnit
	Relayer bool
	Modify  *AlterableInfo
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
		Value: []*PairRegisterValue{&PairRegisterValue{
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

	// ignore the pk param
	if hh > s.getMaxHeight() && s.Modify != nil && sa.Modify != nil {
		if sa.Modify.Fee != nil {
			s.Modify.Fee = new(big.Int).Set(sa.Modify.Fee)
		}
		if sa.Modify.VotePubkey != nil {
			s.Modify.VotePubkey = CopyVotePk(sa.Modify.VotePubkey)
		}
	}

}
func (s *RegisterAccount) stopRegisterInfo(amount, lastHeight *big.Int) error {
	return s.Unit.stopRegisterInfo(amount, lastHeight)
}
func (s *RegisterAccount) redeeming(hh uint64, amount *big.Int) (common.Address, *big.Int, error) {
	return s.Unit.redeeming(hh, amount)
}
func (s *RegisterAccount) finishRedeemed() {
	s.Unit.finishRedeemed()
}
func (s *RegisterAccount) getAllRegister(hh uint64) *big.Int {
	all := s.Unit.getAllRegister(hh)
	return all
}
func (s *RegisterAccount) getValidRegister(hh uint64) *big.Int {
	all := s.Unit.getValidRegister(hh)
	return all
}
func (s *RegisterAccount) getValidRegisterOnly(hh uint64) *big.Int {
	return s.Unit.getValidRegister(hh)
}
func (s *RegisterAccount) merge(epochid, hh, effectHeight uint64) {
	s.Unit.merge(epochid, hh)
}

func (s *RegisterAccount) getMaxHeight() uint64 {
	l := len(s.Unit.Value)
	return s.Unit.Value[l-1].Height.Uint64()
}

func (s *RegisterAccount) clone() *RegisterAccount {
	ss := &RegisterAccount{
		Unit:    s.Unit.clone(),
		Relayer: s.Relayer,
		Modify:  &AlterableInfo{},
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

type Register []*RegisterAccount

func (s *Register) getAllRegister(hh uint64) *big.Int {
	all := big.NewInt(0)
	for _, val := range *s {
		all = all.Add(all, val.getAllRegister(hh))
	}
	return all
}
func (s *Register) getValidRegister(hh uint64) *big.Int {
	all := big.NewInt(0)
	for _, val := range *s {
		all = all.Add(all, val.getValidRegister(hh))
	}
	return all
}
func (s *Register) sort(hh uint64, valid bool) {
	tmp := toRegisterByAmount(hh, valid, *s)
	sort.Sort(tmp)
	*s, _ = fromRegisterByAmount(tmp)
}
func (s *Register) getRA(addr common.Address) *RegisterAccount {
	for _, val := range *s {
		if bytes.Equal(val.Unit.Address.Bytes(), addr.Bytes()) {
			return val
		}
	}
	return nil
}
func (s *Register) update(sa1 *RegisterAccount, hh uint64, next, move bool, effectHeight uint64) {
	sa := s.getRA(sa1.Unit.Address)
	if sa == nil {
		*s = append(*s, sa1)
		s.sort(hh, false)
	} else {
		sa.update(sa1, hh, next, move)
	}
}

/////////////////////////////////////////////////////////////////////////////////
// be thread-safe for caller locked
type RegisterImpl struct {
	accounts   map[uint64]Register // key is epoch id,value is SA set
	curEpochID uint64              // the new epochid of the current state
	lastReward uint64              // the curnent reward height block
}

func NewRegisterImpl() *RegisterImpl {
	return &RegisterImpl{
		curEpochID: params.FirstNewEpochID,
		lastReward: 0,
		accounts:   make(map[uint64]Register),
	}
}
func CloneRegisterImpl(ori *RegisterImpl) *RegisterImpl {
	if ori == nil {
		return nil
	}
	tmp := &RegisterImpl{
		curEpochID: ori.curEpochID,
		lastReward: ori.lastReward,
		accounts:   make(map[uint64]Register),
	}
	for k, val := range ori.accounts {
		items := Register{}
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

func (i *RegisterImpl) GetRegisterAccount(epochid uint64, addr common.Address) (*RegisterAccount, error) {
	if v, ok := i.accounts[epochid]; !ok {
		return nil, params.ErrInvalidRegister
	} else {
		for _, val := range v {
			if bytes.Equal(val.Unit.Address.Bytes(), addr.Bytes()) {
				return val, nil
			}
		}
	}
	return nil, params.ErrInvalidRegister
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
		var ra []*RegisterAccount
		for _, v := range accounts {
			if v.isInRelayer() {
				ra = append(ra, v)
			}
		}
		return ra
	}
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
func (i *RegisterImpl) redeem(sa *RegisterAccount, height uint64, amount *big.Int) error {
	// can be redeem in the SA
	_, all, err1 := sa.redeeming(height, amount)
	if err1 != nil {
		return err1
	}
	if all.Cmp(amount) != 0 {
		return errors.New(fmt.Sprint(params.ErrRedeemAmount, "request amount", amount, "redeem amount", all))
	}
	sa.finishRedeemed()
	return nil
}

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
		nextInfos = Register{}
	}
	for _, v := range prevInfos {
		vv := v.clone()
		vv.merge(prev, nextEpoch.BeginHeight, effectHeight)
		if vv.isvalid() {
			nextInfos.update(vv, nextEpoch.BeginHeight, true, true, effectHeight)
		}
	}
	i.accounts[next] = nextInfos
	return nil
}

/////////////////////////////////////////////////////////////////////////////////
////////////// external function //////////////////////////////////////////

// DoElections called by consensus while it closer the end of epoch
func (i *RegisterImpl) DoElections(state StateDB, epochid, height uint64) ([]*RegisterAccount, error) {

	cur := GetEpochFromID(i.curEpochID)
	if i.curEpochID == params.FirstNewEpochID && height == 0 {
		if val, ok := i.accounts[1]; ok {
			var ee []*RegisterAccount
			for _, v := range val {
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
	//height = 10000x + 9900 and Epoch > 1
	if cur.EndHeight != height+params.ElectionPoint && i.curEpochID >= params.FirstNewEpochID {
		return nil, params.ErrNotElectionTime
	}
	old_relayers := i.getElections2(epochid)
	if val, ok := i.accounts[epochid]; ok {
		val.sort(height, true)
		var ee []*RegisterAccount
		for _, v := range val {
			validRegister := v.getValidRegisterOnly(height)
			num, _ := HistoryWorkEfficiency(state, epochid, v.Unit.Address)
			if v.Relayer == true && num < params.MinWorkEfficiency {
				v.Relayer = false
				continue
			}
			if validRegister.Cmp(params.ElectionMinLimitForRegister) < 0 {
				continue
			}
			v.Relayer = true
			ee = append(ee, v)
			if len(ee) >= params.CountInEpoch {
				break
			}
		}
		if len(ee) == 0 {
			for _, v := range old_relayers {
				v.Relayer = true
			}
			return old_relayers, nil
		}
		return ee, nil
	} else {
		return nil, params.ErrMatchEpochID
	}
}

// Shift will move the register account which has election flag to the next epoch
// it will be save the whole state in the current epoch end block after it called by consensus
func (i *RegisterImpl) Shift(epochid, effectHeight uint64) error {
	lastReward := i.lastReward
	minEpoch := GetEpochFromHeight(lastReward)
	min := i.getMinEpochID()
	for ii := min; minEpoch.EpochID > 1 && ii < minEpoch.EpochID-1; ii++ {
		delete(i.accounts, ii)
	}

	if epochid != i.getCurrentEpoch()+1 {
		return params.ErrOverEpochID
	}
	i.SetCurrentEpoch(epochid)
	prev := epochid - 1
	return i.move(prev, epochid, effectHeight)
}

// CancelAccount cancel amount of asset for register account,it will be work in next epoch
func (i *RegisterImpl) CancelAccount(curHeight uint64, addr common.Address, amount *big.Int) error {
	if amount.Sign() <= 0 || curHeight <= 0 {
		return params.ErrInvalidParam
	}
	curEpoch := GetEpochFromHeight(curHeight)
	if curEpoch == nil || curEpoch.EpochID != i.curEpochID {
		return params.ErrInvalidParam
	}
	sa, err := i.GetRegisterAccount(curEpoch.EpochID, addr)
	if err != nil {
		return err
	}
	err2 := sa.stopRegisterInfo(amount, new(big.Int).SetUint64(curHeight))
	return err2
}

// RedeemAccount redeem amount of asset for register account,it will locked for a certain time
func (i *RegisterImpl) RedeemAccount(curHeight uint64, addr common.Address, amount *big.Int) error {
	if amount.Sign() <= 0 || curHeight <= 0 {
		return params.ErrInvalidParam
	}
	curEpoch := GetEpochFromHeight(curHeight)
	if curEpoch == nil || curEpoch.EpochID < i.curEpochID {
		return params.ErrInvalidParam
	}
	sa, err := i.GetRegisterAccount(curEpoch.EpochID, addr)
	if err != nil {
		return err
	}
	return i.redeem(sa, curHeight, amount)
}

func (i *RegisterImpl) insertAccount(height uint64, sa *RegisterAccount) error {
	if sa == nil {
		return params.ErrInvalidParam
	}
	epochInfo := GetEpochFromHeight(height)
	if epochInfo == nil || epochInfo.EpochID > i.getCurrentEpoch() {
		log.Error("insertAccount", "eid", epochInfo.EpochID, "height", height, "eid2", i.getCurrentEpoch())
		return params.ErrOverEpochID
	}
	if val, ok := i.accounts[epochInfo.EpochID]; !ok {
		var accounts []*RegisterAccount
		accounts = append(accounts, sa)
		i.accounts[epochInfo.EpochID] = Register(accounts)
		log.Debug("Insert register account", "epoch", epochInfo, "account", sa.Unit.GetRewardAddress())
	} else {
		for _, ii := range val {
			if bytes.Equal(ii.Unit.Address.Bytes(), sa.Unit.Address.Bytes()) {
				ii.update(sa, height, false, false)
				log.Debug("Update register account", "account", sa.Unit.GetRewardAddress())
				return nil
			}
		}
		i.accounts[epochInfo.EpochID] = append(val, sa)
		log.Debug("Insert register account", "epoch", epochInfo, "account", sa.Unit.GetRewardAddress())
	}
	return nil
}
func (i *RegisterImpl) InsertAccount2(height uint64, addr common.Address, val *big.Int) error {
	ra := &RegisterAccount{
		Unit: &registerUnit{
			Address: addr,
			Value: []*PairRegisterValue{&PairRegisterValue{
				Amount: new(big.Int).Set(val),
				Height: new(big.Int).SetUint64(height),
			}},
			RedeemInof: make([]*RedeemItem, 0),
		},
		Modify: &AlterableInfo{
			Fee:        new(big.Int).Set(params.InvalidFee),
			VotePubkey: []byte{},
		},
	}
	return i.insertAccount(height, ra)
}
func (i *RegisterImpl) AppendAmount(height uint64, addr common.Address, val *big.Int) error {
	if val.Sign() <= 0 || height < 0 {
		return params.ErrInvalidParam
	}
	epochInfo := GetEpochFromHeight(height)
	if epochInfo.EpochID > i.getCurrentEpoch() {
		log.Debug("insertAccount", "eid", epochInfo.EpochID, "height", height, "eid2", i.getCurrentEpoch())
		return params.ErrOverEpochID
	}
	sa, err := i.GetRegisterAccount(epochInfo.EpochID, addr)
	if err != nil {
		return err
	}
	sa.addAmount(height, val)
	return nil
}
func (i *RegisterImpl) UpdateFee(height uint64, addr common.Address, fee *big.Int) error {
	if height < 0 || fee.Sign() < 0 || fee.Cmp(params.Base) > 0 {
		return params.ErrInvalidParam
	}
	epochInfo := GetEpochFromHeight(height)
	if epochInfo.EpochID > i.getCurrentEpoch() {
		log.Info("UpdateFee", "eid", epochInfo.EpochID, "height", height, "eid2", i.getCurrentEpoch())
		return params.ErrOverEpochID
	}
	sa, err := i.GetRegisterAccount(epochInfo.EpochID, addr)
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
	epochInfo := GetEpochFromHeight(height)
	if epochInfo.EpochID > i.getCurrentEpoch() {
		log.Info("UpdateSAPK", "eid", epochInfo.EpochID, "height", height, "eid2", i.getCurrentEpoch())
		return params.ErrOverEpochID
	}
	sa, err := i.GetRegisterAccount(epochInfo.EpochID, addr)
	if err != nil {
		return err
	}
	sa.updatePk(height, pk)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////
//return all register accounts of the current epoch
func (i *RegisterImpl) GetAllRegisterAccount() Register {
	if val, ok := i.accounts[i.curEpochID]; ok {
		return val
	} else {
		return nil
	}
}

func (i *RegisterImpl) GetBalance(addr common.Address, height uint64) (*big.Int, *big.Int, *big.Int, *big.Int) {
	epochid := i.curEpochID
	unlocking, _, _ := i.getAsset(addr, epochid, params.OpQueryUnlocking)
	_, lock, _ := i.getAsset(addr, epochid, params.OpQueryLocked)
	_, _, fine := i.getAsset(addr, epochid, params.OpQueryFine)
	var u1 = big.NewInt(0)
	var u2 = big.NewInt(0)
	var f *big.Int
	if unlocking[addr] != nil {
		u1.Add(u1, unlocking[addr].ToUnlockedValue(height))
		u2.Add(u2, unlocking[addr].ToUnlockingValue(height))
	}

	if fine[addr] != nil {
		f = fine[addr].Amount
	}
	//locked,unlocked,fine
	return u1, u2, lock, f
}

func (i *RegisterImpl) getAsset(addr common.Address, epoch uint64, op uint8) (map[common.Address]*RelayerValue, *big.Int, map[common.Address]*FineItem) {
	epochid := epoch
	end := GetEpochFromID(epochid).EndHeight
	if val, ok := i.accounts[epochid]; ok {
		res := make(map[common.Address]*RelayerValue)
		res2 := big.NewInt(0)
		res3 := make(map[common.Address]*FineItem)
		for _, v := range val {
			if bytes.Equal(v.Unit.Address.Bytes(), addr.Bytes()) {
				if op&params.OpQueryRegister != 0 || op&params.OpQueryUnlocking != 0 {
					if _, ok := res[addr]; !ok {
						if op&params.OpQueryUnlocking != 0 {
							res[addr] = &RelayerValue{
								Value: v.Unit.redeemToMap(),
							}
						} else {
							res[addr] = &RelayerValue{
								Value: v.Unit.valueToMap(),
							}
						}

					} else {
						log.Error("getAsset", "repeat register account", addr, "epochid", epochid, "op", op)
					}
				}
				if op&params.OpQueryLocked != 0 {
					all := big.NewInt(0)
					for _, v := range v.Unit.Value {
						if v.Height.Uint64() <= end {
							all = all.Add(all, v.Amount)
						} else {
							break
						}
					}
					e := GetEpochFromHeight(end)
					r := v.Unit.getRedeemItem(e.EpochID)
					if r != nil {
						balance := new(big.Int).Sub(all, r.Amount)
						res2 = balance
					} else {
						res2 = all
					}
				}
				if op&params.OpQueryFine != 0 {
					res3[addr] = &FineItem{
						Amount: v.Unit.getFine(epochid),
					}
				}

				continue
			}
		}
		return res, res2, res3
	} else {
		log.Error("getAsset", "wrong epoch in current", epochid)
	}
	return nil, nil, nil
}
func (i *RegisterImpl) MakeModifyStateByTip10() {
	if val, ok := i.accounts[i.curEpochID]; ok {
		for _, v := range val {
			v.makeModifyStateByTip10()
		}
		log.Info("relayer_cli: MakeModifyStateByTip10")
	}
}

/////////////////////////////////////////////////////////////////////////////////
// storage layer

func (i *RegisterImpl) Save(state StateDB, preAddress common.Address) error {
	key := common.BytesToHash(preAddress[:])
	data, err := rlp.EncodeToBytes(i)
	if err != nil {
		log.Crit("Failed to RLP encode RegisterImpl", "err", err)
	}
	hash := RlpHash(data)
	state.SetPOWState(preAddress, key, data)
	tmp := CloneRegisterImpl(i)
	if tmp != nil {
		IC.Cache.Add(hash, tmp)
	}
	return err
}
func (i *RegisterImpl) Load(state StateDB, preAddress common.Address) error {
	var temp RegisterImpl
	key := common.BytesToHash(preAddress[:])
	data := state.GetPOWState(preAddress, key)
	hash := RlpHash(data)
	if cc, ok := IC.Cache.Get(hash); ok {
		register := cc.(*RegisterImpl)
		temp = *(CloneRegisterImpl(register))
	} else {
		if err := rlp.DecodeBytes(data, &temp); err != nil {
			log.Error("Invalid RegisterImpl entry RLP", "err", err)
			return errors.New(fmt.Sprintf("Invalid RegisterImpl entry RLP %s", err.Error()))
		}
		tmp := CloneRegisterImpl(&temp)
		if tmp != nil {
			IC.Cache.Add(hash, tmp)
		}
	}
	i.curEpochID, i.accounts, i.lastReward = temp.curEpochID, temp.accounts, temp.lastReward
	return nil
}

func GetCurrentRelayer(state StateDB) []*params.RelayerMember {
	i := NewRegisterImpl()
	i.Load(state, params.RelayerAddress)
	eid := i.getCurrentEpoch()
	accs := i.getElections2(eid)
	var vv []*params.RelayerMember
	for _, v := range accs {
		//pubkey, _ := crypto.UnmarshalPubkey(v.Votepubkey)
		vv = append(vv, &params.RelayerMember{
			//RelayerBase: crypto.PubkeyToAddress(*pubkey),
			Coinbase: v.Unit.GetRewardAddress(),
			Flag:     params.StateUsedFlag,
			MType:    params.TypeWorked,
		})
	}
	return vv
}

func IsInCurrentEpoch(state StateDB, relayer common.Address) bool {
	relayers := GetCurrentRelayer(state)
	for _, r := range relayers {
		if bytes.Equal(r.Coinbase.Bytes(), relayer.Bytes()) {
			return true
		}
	}
	return false
}

func GetRelayersByEpoch(state StateDB, eid, hh uint64) []*params.RelayerMember {
	i := NewRegisterImpl()
	err := i.Load(state, params.RelayerAddress)
	accs := i.getElections2(eid)
	first := GetFirstEpoch()
	if hh == first.EndHeight-params.ElectionPoint {
		fmt.Println("****** accounts len:", len(i.accounts), "election:", len(accs), " err ", err)
	}
	var vv []*params.RelayerMember
	for _, v := range accs {
		vv = append(vv, &params.RelayerMember{
			Coinbase: v.Unit.GetRewardAddress(),
			Flag:     params.StateUsedFlag,
			MType:    params.TypeWorked,
		})
	}
	return vv
}
func (i *RegisterImpl) Counts() int {
	pos := 0
	for _, val := range i.accounts {
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
		item.AllAmount = val.getValidRegister(info.EndHeight)
		daSum, saSum := 0, len(val)
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
type valuesByHeight []*PairRegisterValue

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
func (vs valuesByHeight) find(hh uint64) (*PairRegisterValue, int) {
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
func (vs valuesByHeight) update(val *PairRegisterValue) valuesByHeight {
	item, pos := vs.find(val.Height.Uint64())
	if item != nil {
		item.Amount = item.Amount.Add(item.Amount, val.Amount)
		item.State |= val.State
	} else {
		rear := append([]*PairRegisterValue{}, vs[pos:]...)
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

type registerItem struct {
	item   *RegisterAccount
	height uint64
	valid  bool
}

func (s *registerItem) getAll() *big.Int {
	if s.valid {
		return s.item.getValidRegister(s.height)
	} else {
		return s.item.getAllRegister(s.height)
	}
}

type registerByAmount []*registerItem

func toRegisterByAmount(hh uint64, valid bool, items []*RegisterAccount) registerByAmount {
	var tmp []*registerItem
	for _, v := range items {
		v.Unit.sort()
		tmp = append(tmp, &registerItem{
			item:   v,
			height: hh,
			valid:  valid,
		})
	}
	return registerByAmount(tmp)
}
func fromRegisterByAmount(items registerByAmount) ([]*RegisterAccount, uint64) {
	var tmp []*RegisterAccount
	var vv uint64
	for _, v := range items {
		tmp = append(tmp, v.item)
		vv = v.height
	}
	return tmp, vv
}
func (vs registerByAmount) Len() int {
	return len(vs)
}
func (vs registerByAmount) Less(i, j int) bool {
	return vs[i].getAll().Cmp(vs[j].getAll()) > 0
}
func (vs registerByAmount) Swap(i, j int) {
	it := vs[i]
	vs[i] = vs[j]
	vs[j] = it
}

func GetCurrentEpochID(evm *EVM) (uint64, error) {
	register := NewRegisterImpl()
	err := register.Load(evm.StateDB, params.RelayerAddress)
	if err != nil {
		log.Error("relayer_cli load error", "error", err)
		return 0, err
	}
	return register.getCurrentEpoch(), nil
}

type extRegisterImpl struct {
	Accounts   []Register
	CurEpochID uint64
	Array      []uint64
	LastReward uint64
}

func (i *RegisterImpl) DecodeRLP(s *rlp.Stream) error {
	var ei extRegisterImpl
	if err := s.Decode(&ei); err != nil {
		return err
	}
	accounts := make(map[uint64]Register)
	for i, account := range ei.Accounts {
		accounts[ei.Array[i]] = account
	}

	i.curEpochID, i.accounts, i.lastReward = ei.CurEpochID, accounts, ei.LastReward
	return nil
}

// EncodeRLP serializes b into the atlaschain RLP RegisterImpl format.
func (i *RegisterImpl) EncodeRLP(w io.Writer) error {
	var accounts []Register
	var order []uint64
	for i, _ := range i.accounts {
		order = append(order, i)
	}
	for m := 0; m < len(order)-1; m++ {
		for n := 0; n < len(order)-1-m; n++ {
			if order[n] > order[n+1] {
				order[n], order[n+1] = order[n+1], order[n]
			}
		}
	}
	for _, epoch := range order {
		accounts = append(accounts, i.accounts[epoch])
	}
	return rlp.Encode(w, extRegisterImpl{
		CurEpochID: i.curEpochID,
		Accounts:   accounts,
		Array:      order,
		LastReward: i.lastReward,
	})
}
