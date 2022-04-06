package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	atlaschain "github.com/mapprotocol/atlas/core/chain"
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
