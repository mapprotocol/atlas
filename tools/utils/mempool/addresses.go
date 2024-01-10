package mempool

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/mapprotocol/atlas/tools/utils/btcapi"
)

type UTXO struct {
	Txid   string `json:"txid"`
	Vout   int    `json:"vout"`
	Status struct {
		Confirmed   bool   `json:"confirmed"`
		BlockHeight int    `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		BlockTime   int64  `json:"block_time"`
	} `json:"status"`
	Value int64 `json:"value"`
}

// UTXOs is a slice of UTXO
type UTXOs []UTXO

type VoutItem struct {
	Scriptpubkey         string `json:"scriptpubkey"`
	Scriptpubkey_asm     string `json:"scriptpubkey_asm"`
	Scriptpubkey_type    string `json:"scriptpubkey_type"`
	Scriptpubkey_address string `json:"scriptpubkey_address"`
	Value                int    `json:"value"`
}
type VinItem struct {
	Txid          string   `json:"txid"`
	Vout          int      `json:"vout"`
	Prevout       VoutItem `json:"prevout"`
	Scriptsig     string   `json:"scriptsig"`
	Scriptsig_asm string   `json:"scriptsig_asm"`
	Witness       []string `json:"witness"`
	Is_coinbase   bool     `json:"is_coinbase"`
	Sequence      int64    `json:"sequence"`
}

type TxItem struct {
	Txid     string      `json:"txid"`
	Version  int         `json:"version"`
	Locktime int64       `json:"locktime"`
	Vin      []*VinItem  `json:"vin"`
	Vout     []*VoutItem `json:"vout"`
	Size     int         `json:"size"`
	Weight   int         `json:"weight"`
	Sigops   int         `json:"sigops"`
	Fee      int         `json:"fee"`
	Status   struct {
		Confirmed   bool   `json:"confirmed"`
		BlockHeight int    `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		BlockTime   int64  `json:"block_time"`
	} `json:"status"`
}

type TxItems []TxItem

type SimTx struct {
	Txid    chainhash.Hash
	Sender  string
	OutPuts []*VoutItem
}

func (c *MempoolClient) ListUnspent(address btcutil.Address) ([]*btcapi.UnspentOutput, error) {
	res, err := c.request(http.MethodGet, fmt.Sprintf("/address/%s/utxo", address.EncodeAddress()), nil)
	if err != nil {
		return nil, err
	}

	var utxos UTXOs
	err = json.Unmarshal(res, &utxos)
	if err != nil {
		return nil, err
	}

	unspentOutputs := make([]*btcapi.UnspentOutput, 0)
	for _, utxo := range utxos {
		txHash, err := chainhash.NewHashFromStr(utxo.Txid)
		if err != nil {
			return nil, err
		}
		unspentOutputs = append(unspentOutputs, &btcapi.UnspentOutput{
			Outpoint: wire.NewOutPoint(txHash, uint32(utxo.Vout)),
			Output:   wire.NewTxOut(utxo.Value, address.ScriptAddress()),
		})
	}
	return unspentOutputs, nil
}

func (c *MempoolClient) GetTxsFromAddress(address btcutil.Address) ([]*SimTx, error) {
	res, err := c.request(http.MethodGet, fmt.Sprintf("/address/%s/txs", address.EncodeAddress()), nil)
	if err != nil {
		return nil, err
	}

	var items TxItems
	err = json.Unmarshal(res, &items)
	if err != nil {
		return nil, err
	}

	simTxs := make([]*SimTx, 0)
	for _, item := range items {
		txHash, err := chainhash.NewHashFromStr(item.Txid)
		if err != nil {
			return nil, err
		}
		sender, err := fetchSenderFromItem(&item)
		if err != nil {
			return nil, err
		}
		simTxs = append(simTxs, &SimTx{
			Txid:    *txHash,
			Sender:  sender,
			OutPuts: item.Vout,
		})
	}
	return simTxs, nil
}
func fetchSenderFromItem(item *TxItem) (string, error) {
	sender := ""
	for _, i := range item.Vin {
		sender = i.Prevout.Scriptpubkey_address
		if len(sender) > 0 {
			break
		}
	}
	if sender == "" {
		return "", errors.New("not found sender from the tx" + item.Txid)
	}
	return sender, nil
}
