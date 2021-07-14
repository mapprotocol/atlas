package vm

import (
	"fmt"
	"github.com/mapprotocol/atlas/chains/chainsdb"
	"reflect"
	"testing"
)

func headerStorePack(method string, args ...interface{}) []byte {
	pack, err := abiHeaderStore.Pack(method, args...)
	if err != nil {
		panic(err)
	}
	return pack[4:]
}

func TestSavePackAndUnpack(t *testing.T) {
	args := struct {
		From    string
		To      string
		Headers []byte
	}{}

	//pack := headerStorePack(Save, "ETH", "MAP", []byte("1234"))
	//method, _ := abiHeaderStore.Methods[Save]

	pack := headerStorePack(CurrentHeaderNumber, "ETH")
	method, _ := abiHeaderStore.Methods[CurrentHeaderNumber]

	unpack, err := method.Inputs.Unpack(pack[4:])
	if err != nil {
		panic(err)
	}
	if err := method.Inputs.Copy(&args, unpack); err != nil {
		panic(err)
	}
	fmt.Printf("============================ unpack: %#v\n", unpack)
}

func Test_currentHeaderNumber(t *testing.T) {
	type args struct {
		evm      *EVM
		contract *Contract
		input    []byte
	}
	tests := []struct {
		name    string
		args    args
		before  func()
		wantRet []byte
		wantErr bool
	}{
		{
			name: "",
			args: args{
				evm:      &EVM{},
				contract: &Contract{},
				input:    headerStorePack(CurrentHeaderNumber, "ETH"),
			},
			before: func() {
				chainsdb.NewStoreDb(nil, 10, 2)
			},
			wantRet: nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt.before()
		t.Run(tt.name, func(t *testing.T) {
			gotRet, err := currentHeaderNumber(tt.args.evm, tt.args.contract, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("currentHeaderNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("currentHeaderNumber() gotRet = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}
