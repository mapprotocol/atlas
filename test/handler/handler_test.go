package handler

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var endpoint = "http://18.142.54.137:7445"

func Test_getMgrMaintainerAddress(t *testing.T) {
	getMgrMaintainerAddress(endpoint)
}

func Test_setMgrMaintainerAddress(t *testing.T) {
	from := common.HexToAddress("")
	target := common.HexToAddress("")
	privateKey, err := crypto.ToECDSA(common.FromHex(""))
	if err != nil {
		t.Fatal(err)
	}

	setMgrMaintainerAddress(endpoint, from, target, privateKey)
}

func Test_getTargetEpochPayment(t *testing.T) {
	getTargetEpochPayment(endpoint)
}

// INFO [08-18|15:32:31.365] getTargetEpochPayment                    value=50,000,000,000,000,000,000,000
// INFO [08-18|15:47:45.804] getTargetEpochPayment                    value=60,000,000,000,000,000,000,000
// INFO [08-18|15:49:36.350] getTargetEpochPayment                    value=50,000,000,000,000,000,000,000

func Test_setTargetEpochPayment(t *testing.T) {
	from := common.HexToAddress("")
	target := new(big.Int).Mul(big.NewInt(50000), big.NewInt(1e18))
	privateKey, err := crypto.ToECDSA(common.FromHex(""))
	if err != nil {
		t.Fatal(err)
	}

	setTargetEpochPayment(endpoint, from, target, privateKey)
}

func Test_getElectableValidators(t *testing.T) {
	getElectableValidators(endpoint)
}

// INFO [08-26|16:55:35.641] getElectableValidators                   minElectableValidators=1 maxElectableValidators=100
// INFO [08-26|17:00:42.247] getElectableValidators                   minElectableValidators=1 maxElectableValidators=50

func Test_setElectableValidators(t *testing.T) {
	from := common.HexToAddress("")
	privateKey, err := crypto.ToECDSA(common.FromHex(""))
	if err != nil {
		t.Fatal(err)
	}
	min := big.NewInt(1)
	max := big.NewInt(100)
	setElectableValidators(endpoint, from, privateKey, min, max)
}
