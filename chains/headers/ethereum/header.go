package ethereum

import (
	"encoding/json"
	"fmt"
	"github.com/opentracing/opentracing-go/log"
	"io/ioutil"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Header struct {
	ParentHash  common.Hash      `json:"parentHash"       gencodec:"required"`
	UncleHash   common.Hash      `json:"sha3Uncles"       gencodec:"required"`
	Coinbase    common.Address   `json:"miner"            gencodec:"required"`
	Root        common.Hash      `json:"stateRoot"        gencodec:"required"`
	TxHash      common.Hash      `json:"transactionsRoot" gencodec:"required"`
	ReceiptHash common.Hash      `json:"receiptsRoot"     gencodec:"required"`
	Bloom       types.Bloom      `json:"logsBloom"        gencodec:"required"`
	Difficulty  *big.Int         `json:"difficulty"       gencodec:"required"`
	Number      *big.Int         `json:"number"           gencodec:"required"`
	GasLimit    uint64           `json:"gasLimit"         gencodec:"required"`
	GasUsed     uint64           `json:"gasUsed"          gencodec:"required"`
	Time        uint64           `json:"timestamp"        gencodec:"required"`
	Extra       []byte           `json:"extraData"        gencodec:"required"`
	MixDigest   common.Hash      `json:"mixHash"`
	Nonce       types.BlockNonce `json:"nonce"`
}

func (eh *Header) Hash() common.Hash {
	return rlpHash(eh)
}

func (eh *Header) Genesis(chainID uint64) *Header {
	genesis := &Header{}
	g := GetGenesisByChainID(chainID)
	_ = json.Unmarshal([]byte(g), genesis)
	return genesis
}
func configGenesis(name string) {
	data, err := ioutil.ReadFile(fmt.Sprintf("config/%v_config.json", name))
	if err != nil {
		log.Error(err)
	}
	var genesis = &Header{}
	json.Unmarshal(data, genesis)
	fmt.Println(genesis.Hash())
}
