package tools

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

const hexPrefix = "0x"

func RlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	if e := rlp.Encode(hw, x); e != nil {
		log.Warn("RlpHash", "error", e.Error())
	}
	hw.Sum(h[:0])
	return h
}

func DeepCopy(src, dst interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func Bytes2Hex(bs []byte) string {
	return hexPrefix + hex.EncodeToString(bs)
}
