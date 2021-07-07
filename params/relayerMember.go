package params

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
	type committee struct {
		Address common.Address `json:"address,omitempty"`
		PubKey  *hexutil.Bytes `json:"publickey,omitempty"`
		Flag    uint32         `json:"flag,omitempty"`
		MType   uint32         `json:"mType,omitempty"`
	}
	var dec committee
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
