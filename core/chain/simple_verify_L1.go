package chain

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/tools/utils/mempool"
	"strings"
)

var (
	// for test
	l1Address        = "tb1pkepxd60wx4z33qdgz5vad5dvtus6syv3m5m6xc3kthdfar9jmmvq3a8mp7"
	checkPointLength = 200000
	testnet          = true
)

type CheckPoint struct {
	Height uint64 `json:"height"`
	Root   string `json:"root"`
}

func FromBytes(data []byte) (*CheckPoint, error) {
	if len(data) != 40 {
		return nil, errors.New("invalid data length in from bytes")
	}
	v := &CheckPoint{}
	v.Height = binary.LittleEndian.Uint64(data[0:8])
	v.Root = hexutil.Encode(data[8:])
	return v, nil
}
func (c *CheckPoint) String() string {
	return fmt.Sprintf("root=%s, height:%v", c.Root, c.Height)
}
func (c *CheckPoint) Equal(ck *CheckPoint) bool {
	if ck == nil || c == nil {
		return false
	}
	if c.Root == ck.Root && c.Height == ck.Height {
		return true
	}
	return false
}

func IsCheckPointData(script []byte) bool {
	if len(script) == 42 {
		return script[0] == 0x6a
	}
	return false
}

func getCheckPointByHeight(height uint64, bc *BlockChain) (*CheckPoint, error) {
	head := bc.GetHeaderByNumber(height)
	if head != nil {
		return &CheckPoint{
			Height: height,
			Root:   head.Root.String(),
		}, nil
	}
	return nil, errors.New("not found the head")
}
func verifyBlockWithCheckPoint(ck *CheckPoint, bc *BlockChain) error {
	ck0, err := getCheckPointByHeight(ck.Height, bc)
	if err != nil {
		return err
	}
	if ck0.Equal(ck) {
		return nil
	}
	return errors.New("verify checkpoint failed" + ck.String() + ck0.String())
}
func fetchLatestCheckPoint(sender string, network *chaincfg.Params) (*CheckPoint, error) {
	client := mempool.NewClient(network)
	cCheckPoint := &CheckPoint{}
	simTxs, err := client.GetTxsFromAddress(sender)
	if err != nil {
		return nil, err
	}
	// check latest checkpoint match with the config checkpoint
	for i := range simTxs {
		tx := simTxs[len(simTxs)-i-1]
		if sender == tx.Sender && len(tx.OutPuts) == 2 {
			//str := tx.OutPuts[0].Scriptpubkey_asm
			script, err := hex.DecodeString(tx.OutPuts[0].Scriptpubkey)
			if err != nil {
				log.Error("decode the Scriptpubkey failed", "err", err, "txid", tx.Txid.String())
				continue
			}
			if !IsCheckPointData(script) {
				fmt.Println(tx.Txid.String(), "is a op_return tx")
			}
			cc, err := checkPointFromScript(script)
			if err != nil {
				log.Error("not a OP_RETURN tx", "txhash", tx.Txid, "error", err)
				continue
			}
			if cCheckPoint != nil {
				if cc.Height < cCheckPoint.Height {
					continue
				}
			}
			cCheckPoint = cc
		} else {
			log.Info("fetch the latest checkpoint, invalid tx", "txid", tx.Txid)
		}
	}
	return cCheckPoint, nil
}
func checkPointFromScript(script []byte) (*CheckPoint, error) {
	if len(script) == 42 {
		b0 := script[2:]
		return FromBytes(b0)
	}
	return nil, errors.New("invalid script in parse checkpoint")
}
func checkPointFromAsm(str string) (*CheckPoint, error) {
	result := strings.Split(str, " ")
	// op_return op_pushbytes data
	if len(result) != 3 {
		return nil, errors.New("invalid script length")
	}
	strScript := result[len(result)-1]
	b0, err := hex.DecodeString(strScript)
	if err != nil {
		return nil, err
	}
	return FromBytes(b0)
}

func VerifyCheckPoint(testnet bool, bc *BlockChain) error {
	netParams := &chaincfg.MainNetParams
	if testnet {
		netParams = &chaincfg.TestNet3Params
	}
	sender := l1Address
	//sender, err := btcutil.DecodeAddress(l1Address, netParams)
	//if err != nil {
	//	log.Error("check point verify failed [DecodeAddress]", "error", err)
	//	// make it pass
	//	return nil
	//}
	checkpoint, err := fetchLatestCheckPoint(sender, netParams)
	if err != nil {
		log.Error("check point verify failed [fetchLatestCheckPoint]", "error", err)
		// make it pass
		return nil
	}
	cur := bc.CurrentBlock()
	if cur == nil {
		log.Info("current block is nil....,not verify")
		return nil
	}
	curHeight := cur.Number().Uint64()
	if curHeight < checkpoint.Height {
		log.Info("current height is low than checkpoint, not verify")
		return nil
	}
	err = verifyBlockWithCheckPoint(checkpoint, bc)
	if err != nil {
		log.Error("check point verify failed [verifyBlockWithCheckPoint]", "error", err)
		return err
	}
	return nil
}
