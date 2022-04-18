package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/mapprotocol/atlas/accounts/keystore"
	accountTool "github.com/mapprotocol/atlas/cmd/marker/account"
	"github.com/mapprotocol/atlas/cmd/utils"
	atlaschain "github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/helper/fileutils"
)

func Test_dumpGenesis(t *testing.T) {
	genesis := atlaschain.DefaultGenesisBlock()
	b, err := json.MarshalIndent(genesis, " ", " ")
	if err != nil {
		t.Fatalf("could not encode genesis")
	}
	path, err := os.Getwd()
	if err != nil {
		t.Fatalf("get path err%s", err)
	}
	fmt.Println(path)
	err = ioutil.WriteFile(filepath.Join(path, "/genesis.json"), b, 0644)
	if err != nil {
		t.Fatalf("could not encode genesis")
	}
}

func Test_CreatAccount(T *testing.T) {
	type Info struct {
		Account string
		Path    string
	}
	var voterAccounts []Info
	for i := 0; i < 900; i++ {
		cfg := atlasConfig{Node: defaultNodeConfig()}
		scryptN, scryptP, keydir, err := cfg.Node.AccountConfig()
		keydir = "./voters"
		if err != nil {
			utils.Fatalf("Failed to read configuration: %v", err)
		}

		password := "111111"
		account, err := keystore.StoreKey(keydir, password, scryptN, scryptP)
		if err != nil {
			utils.Fatalf("Failed to create account: %v", err)
		}
		accountBls, err := accountTool.LoadAccount(account.URL.Path, password)
		if err != nil {
			utils.Fatalf("Failed to create account: %v", err)
		}
		blsProofOfPossession := accountBls.MustBLSProofOfPossession()
		blsPubKey, err := accountBls.BLSPublicKey()
		if err != nil {
			utils.Fatalf("Failed to create account: %v", err)
		}
		blsPubKeyText, err := blsPubKey.MarshalText()

		if err != nil {
			utils.Fatalf("Failed to create account: %v", err)
		}
		fmt.Printf("\nYour new key was generated\n\n")
		fmt.Printf("Public address of the key:   %s\n", account.Address.Hex())
		fmt.Printf("PublicKeyHex:   %s\n", hexutil.Encode(accountBls.PublicKey()))
		fmt.Printf("BLS Public address of the key:   %s\n", blsPubKeyText)
		fmt.Printf("BLSProofOfPossession:   %s\n", hexutil.Encode(blsProofOfPossession))
		fmt.Printf("Path of the secret key file: %s\n\n", account.URL.Path)
		fmt.Printf("- You can share your public address with anyone. Others need it to interact with you.\n")
		fmt.Printf("- You must NEVER share the secret key with anyone! The key controls access to your funds!\n")
		fmt.Printf("- You must BACKUP your key file! Without the key, it's impossible to access account funds!\n")
		fmt.Printf("- You must REMEMBER your password! Without the password, it's impossible to decrypt the key!\n\n")
		voterAccounts = append(voterAccounts, Info{account.Address.String(), account.URL.Path})
	}
	fileutils.WriteJson(voterAccounts, "./voters/Voters.json")
}
