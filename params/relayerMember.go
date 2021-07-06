package params

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type CommitteeMember struct {
	Coinbase      common.Address `json:"coinbase`
	CommitteeBase common.Address `json:"committeebase`
	Publickey     []byte
	Flag          uint32
	MType         uint32
}

func GetCommitteeMember() []CommitteeMember {
	return []CommitteeMember{}
}

func (c *CommitteeMember) Compared(d *CommitteeMember) bool {
	if c.MType == d.MType && c.Coinbase == d.Coinbase && c.CommitteeBase == d.CommitteeBase && bytes.Equal(c.Publickey, d.Publickey) {
		return true
	}
	return false
}

func (c *CommitteeMember) String() string {
	return fmt.Sprintf("F:%d,T:%d,C:%s,P:%s,A:%s", c.Flag, c.MType, hexutil.Encode(c.Coinbase[:]),
		hexutil.Encode(c.Publickey), hexutil.Encode(c.CommitteeBase[:]))
}

func (c *CommitteeMember) UnmarshalJSON(input []byte) error {
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
