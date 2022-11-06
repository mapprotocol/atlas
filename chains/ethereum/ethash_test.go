package ethereum

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"log"
	"math/big"
	"testing"
)

func dialSepolia() *ethclient.Client {
	conn, err := ethclient.Dial("https://sepolia.infura.io/v3/508a3c6487bf44cf8ae109840819bec4")
	if err != nil {
		log.Fatalf("Failed to connect to the eth: %v", err)
	}
	return conn
}

func convertETHHeader(header *types.Header) *Header {
	bs, err := rlp.EncodeToBytes(header)
	if err != nil {
		panic(err)
	}
	var headers *Header
	if err := rlp.DecodeBytes(bs, &headers); err != nil {
		panic(err)
	}
	return headers
}

var (
	header1 *Header
	header2 *Header
	header3 *Header
	header4 *Header
	header5 *Header
	header6 *Header
	header7 *Header
	header8 *Header
)

func GetHeaders() {
	h1, err := dialSepolia().HeaderByNumber(context.Background(), big.NewInt(1000001))
	if err != nil {
		panic(err)
	}

	h2, err := dialSepolia().HeaderByNumber(context.Background(), big.NewInt(1000002))
	if err != nil {
		panic(err)
	}
	h2.Number = new(big.Int).Sub(h2.Number, big.NewInt(1))

	h3, err := dialSepolia().HeaderByNumber(context.Background(), big.NewInt(1000003))
	if err != nil {
		panic(err)
	}
	h3.Nonce[7] = 0

	h4, err := dialSepolia().HeaderByNumber(context.Background(), big.NewInt(1000004))
	if err != nil {
		panic(err)
	}
	h4.Difficulty = new(big.Int).Add(h2.Difficulty, big.NewInt(1))

	h5, err := dialSepolia().HeaderByNumber(context.Background(), big.NewInt(1000005))
	if err != nil {
		panic(err)
	}
	h5.Time++

	h6, err := dialSepolia().HeaderByNumber(context.Background(), big.NewInt(1000005))
	if err != nil {
		panic(err)
	}
	h6.GasLimit++

	h7, err := dialSepolia().HeaderByNumber(context.Background(), big.NewInt(1000005))
	if err != nil {
		panic(err)
	}
	h7.GasUsed--

	h8, err := dialSepolia().HeaderByNumber(context.Background(), big.NewInt(1000005))
	if err != nil {
		panic(err)
	}
	h8.BaseFee = new(big.Int).Add(h8.BaseFee, big.NewInt(1))

	header1 = convertETHHeader(h1)
	header2 = convertETHHeader(h2)
	header3 = convertETHHeader(h3)
	header4 = convertETHHeader(h4)
	header5 = convertETHHeader(h5)
	header6 = convertETHHeader(h6)
	header7 = convertETHHeader(h7)
	header8 = convertETHHeader(h8)
}

func TestVerifySeal(t *testing.T) {
	GetHeaders()

	type args struct {
		header *Header
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				header: header1,
			},
			wantErr: false,
		},
		{
			name: "t2",
			args: args{
				header: header2,
			},
			wantErr: true,
		},
		{
			name: "t3",
			args: args{
				header: header3,
			},
			wantErr: true,
		},
		{
			name: "t4",
			args: args{
				header: header4,
			},
			wantErr: true,
		},
		{
			name: "t5",
			args: args{
				header: header5,
			},
			wantErr: true,
		},
		{
			name: "t6",
			args: args{
				header: header6,
			},
			wantErr: true,
		},
		{
			name: "t7",
			args: args{
				header: header7,
			},
			wantErr: true,
		},
		{
			name: "t8",
			args: args{
				header: header8,
			},
			wantErr: true,
		},
	}

	MakeGlobalEthash("/Users/t/data/atlas-unit-test")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := VerifySeal(tt.args.header); (err != nil) != tt.wantErr {
				t.Errorf("VerifySeal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
