package mempool

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/pkg/errors"
)

type TransactionStatusResponse struct {
	Confirmed   bool   `json:"confirmed"`
	BlockHeight uint64 `json:"block_height"`
	BlockHash   string `json:"block_hash"`
	BlockTime   uint64 `json:"block_time"`
}

type RecommendedFeesResponse struct {
	FastestFee  int64 `json:"fastestFee"`
	HalfHourFee int64 `json:"halfHourFee"`
	HourFee     int64 `json:"hourFee"`
	EconomyFee  int64 `json:"economyFee"`
	MinimumFee  int64 `json:"minimumFee"`
}

func (c *MempoolClient) TransactionStatus(txHash *chainhash.Hash) (*TransactionStatusResponse, error) {
	res, err := c.request(http.MethodGet, fmt.Sprintf("/tx/%s/status", txHash.String()), nil)
	if err != nil {
		return nil, err
	}

	resp := &TransactionStatusResponse{}
	if err := json.Unmarshal(res, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *MempoolClient) GetRawTransaction(txHash *chainhash.Hash) (*wire.MsgTx, error) {
	res, err := c.request(http.MethodGet, fmt.Sprintf("/tx/%s/raw", txHash.String()), nil)
	if err != nil {
		return nil, err
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	if err := tx.Deserialize(bytes.NewReader(res)); err != nil {
		return nil, err
	}
	return tx, nil
}

func (c *MempoolClient) BroadcastTx(tx *wire.MsgTx) (*chainhash.Hash, error) {
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return nil, err
	}

	res, err := c.request(http.MethodPost, "/tx", strings.NewReader(hex.EncodeToString(buf.Bytes())))
	if err != nil {
		return nil, err
	}

	txHash, err := chainhash.NewHashFromStr(string(res))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse tx hash, %s", string(res)))
	}
	return txHash, nil
}

func (c *MempoolClient) RecommendedFees() (*RecommendedFeesResponse, error) {
	res, err := c.request(http.MethodGet, "/v1/fees/recommended", nil)
	if err != nil {
		return nil, err
	}

	resp := &RecommendedFeesResponse{}
	if err := json.Unmarshal(res, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
