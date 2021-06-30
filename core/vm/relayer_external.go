package vm

import (
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
)

type LockedAsset struct {
	LockValue []*LockValue   `json:"lockValue"`
	Address   common.Address `json:"address"`
}

// MarshalJSON marshals as JSON.
func (l LockedAsset) MarshalJSON() ([]byte, error) {
	type LockedAsset struct {
		LockValue []*LockValue   `json:"lockValue"`
		Address   common.Address `json:"address"`
	}
	var enc LockedAsset
	enc.LockValue = l.LockValue
	enc.Address = l.Address
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (l *LockedAsset) UnmarshalJSON(input []byte) error {
	type LockedAsset struct {
		LockValue []*LockValue    `json:"lockValue"`
		Address   *common.Address `json:"address"`
	}
	var dec LockedAsset
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.LockValue == nil {
		return errors.New("missing required field 'lockValue' for LockedAsset")
	}
	l.LockValue = dec.LockValue
	if dec.Address != nil {
		l.Address = *dec.Address
	}
	return nil
}

type LockValue struct {
	EpochID uint64
	Amount  string
	Height  *big.Int
	Locked  bool
}

// MarshalJSON marshals as JSON.
func (l LockValue) MarshalJSON() ([]byte, error) {
	type LockValue struct {
		EpochID hexutil.Uint64 `json:"epochID"`
		Amount  string         `json:"amount"`
		Height  *hexutil.Big   `json:"height"`
		Locked  bool           `json:"locked"`
	}
	var enc LockValue
	enc.EpochID = hexutil.Uint64(l.EpochID)

	enc.Amount = l.Amount
	enc.Height = (*hexutil.Big)(l.Height)
	enc.Locked = l.Locked
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (l *LockValue) UnmarshalJSON(input []byte) error {
	type LockValue struct {
		EpochID *hexutil.Uint64 `json:"epochID"`
		Amount  *string         `json:"amount"`
		Height  *hexutil.Big    `json:"height"`
		Locked  *bool           `json:"locked"`
	}
	var dec LockValue
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.EpochID != nil {
		l.EpochID = uint64(*dec.EpochID)
	}
	if dec.Amount != nil {
		l.Amount = *dec.Amount
	}
	if dec.Height != nil {
		l.Height = (*big.Int)(dec.Height)
	}
	if dec.Locked != nil {
		l.Locked = *dec.Locked
	}
	return nil
}
