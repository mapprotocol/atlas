package chain

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"testing"
)

func makeTpAddress(privKey *btcec.PrivateKey, netParams *chaincfg.Params) (btcutil.Address, error) {
	tapKey := txscript.ComputeTaprootKeyNoScript(privKey.PubKey())

	address, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(tapKey),
		netParams,
	)
	if err != nil {
		return nil, err
	}
	return address, nil
}

func TestGeneratePrivateKey(t *testing.T) {
	testnet := true
	netParams := &chaincfg.MainNetParams
	if testnet {
		netParams = &chaincfg.TestNet3Params
	}
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	privateKeyBytes := privateKey.Serialize()
	privateKeyString := hex.EncodeToString(privateKeyBytes)
	t.Logf("private key: %s", privateKeyString)

	privateKeyBytes, err = hex.DecodeString(privateKeyString)
	if err != nil {
		t.Fatal(err)
	}
	privateKey, _ = btcec.PrivKeyFromBytes(privateKeyBytes)
	privateKeyBytes = privateKey.Serialize()
	privateKeyString = hex.EncodeToString(privateKeyBytes)
	t.Logf("private key: %s", privateKeyString)

	sender, err := makeTpAddress(privateKey, netParams)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("sender: ", sender.String())
}
