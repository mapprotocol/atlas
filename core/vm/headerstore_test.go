package vm

import (
	"testing"
)

func TestSavePackAndUnpack(t *testing.T) {
	args := struct {
		From    string
		To      string
		Headers []byte
	}{}

	method, _ := abiHeaderStore.Methods["save"]
	pack, err := abiHeaderStore.Pack("save", "ETH", "MAP", []byte("1234"))
	if err != nil {
		panic(err)
	}

	unpack, err := method.Inputs.Unpack(pack[4:])
	if err != nil {
		panic(err)
	}
	if err := method.Inputs.Copy(&args, unpack); err != nil {
		panic(err)
	}
}
