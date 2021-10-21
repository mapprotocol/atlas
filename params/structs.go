package params

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/mapprotocol/atlas/core/types"
	"math/big"
)

type RelayerMember struct {
	Coinbase    common.Address `json:"coinbase`
	RelayerBase common.Address `json:"relayerbase`
	Publickey   []byte
	Flag        uint32
	MType       uint32
}

func (c *RelayerMember) Compared(d *RelayerMember) bool {
	if c.MType == d.MType && c.Coinbase == d.Coinbase && c.RelayerBase == d.RelayerBase && bytes.Equal(c.Publickey, d.Publickey) {
		return true
	}
	return false
}

func (c *RelayerMember) String() string {
	return fmt.Sprintf("F:%d,T:%d,C:%s,P:%s,A:%s", c.Flag, c.MType, hexutil.Encode(c.Coinbase[:]),
		hexutil.Encode(c.Publickey), hexutil.Encode(c.RelayerBase[:]))
}

func (c *RelayerMember) UnmarshalJSON(input []byte) error {
	type relayer struct {
		Address common.Address `json:"address,omitempty"`
		PubKey  *hexutil.Bytes `json:"publickey,omitempty"`
		Flag    uint32         `json:"flag,omitempty"`
		MType   uint32         `json:"mType,omitempty"`
	}
	var dec relayer
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}

	c.Coinbase = dec.Address
	c.Flag = dec.Flag
	c.MType = dec.MType
	if dec.PubKey != nil {
		c.Publickey = *dec.PubKey
	}
	/*var err error
	if dec.PubKey != nil {
		_, err = crypto.UnmarshalPubkey(*dec.PubKey)
		if err != nil {
			return err
		}
	}*/
	return nil
}

type CallMsg struct {
	From      common.Address  // the sender of the 'transaction'
	To        *common.Address // the destination contract (nil for contract creation)
	Gas       uint64          // if 0, the call executes with near-infinite gas
	GasPrice  *big.Int        // wei <-> gas exchange ratio
	GasFeeCap *big.Int        // EIP-1559 fee cap per gas.
	GasTipCap *big.Int        // EIP-1559 tip per gas.
	Value     *big.Int        // amount of wei sent along with the call
	Data      []byte          // input data, usually an ABI-encoded contract method invocation

	AccessList types.AccessList // EIP-2930 access list.
}
