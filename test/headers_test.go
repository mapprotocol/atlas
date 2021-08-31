package test

import (
	"context"
	"log"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/mapprotocol/atlas/chains/chainsdb"
	"github.com/mapprotocol/atlas/chains/headers/ethereum"
	ve "github.com/mapprotocol/atlas/chains/validates/ethereum"
)

func dialEthConn() (*ethclient.Client, string) {
	url := "http://127.0.0.1:8545"
	conn, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the MAP chain client: %v", err)
	}
	return conn, url
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func convertChain(e *types.Header) *ethereum.Header {
	header := new(ethereum.Header)
	header.ParentHash = e.ParentHash
	header.UncleHash = e.UncleHash
	header.Coinbase = e.Coinbase
	header.Root = e.Root
	header.TxHash = e.TxHash
	header.ReceiptHash = e.ReceiptHash
	header.GasLimit = e.GasLimit
	header.GasUsed = e.GasUsed
	header.Time = e.Time
	header.MixDigest = e.MixDigest
	header.Nonce = types.EncodeNonce(e.Nonce.Uint64())
	header.Bloom.SetBytes(e.Bloom.Bytes())
	if header.Difficulty = new(big.Int); e.Difficulty != nil {
		header.Difficulty.Set(e.Difficulty)
	}
	if header.Number = new(big.Int); e.Number != nil {
		header.Number.Set(e.Number)
	}
	if len(e.Extra) > 0 {
		header.Extra = make([]byte, len(e.Extra))
		copy(header.Extra, e.Extra)
	}
	return header
}

func getChains(startNum, endNum uint64) []*ethereum.Header {
	conn, _ := dialEthConn()

	Headers := make([]*ethereum.Header, 0, endNum-startNum+1)
	for i := startNum; i <= endNum; i++ {
		Header, _ := conn.HeaderByNumber(context.Background(), big.NewInt(int64(i)))
		if Header != nil {
			Headers = append(Headers, convertChain(Header))
		}
	}
	return Headers
}

func TestHeader_ValidateHeaderChain(t *testing.T) {

	type args struct {
		chain []*ethereum.Header
	}
	tests := []struct {
		name    string
		args    args
		before  func()
		want    int
		wantErr bool
	}{
		{
			name: "t-1",
			args: args{
				chain: getChains(0, 300),
			},
			before: func() {
				chainsdb.NewStoreDb(nil, 10, 2)
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()
			eh := &ve.Validate{}
			got, err := eh.ValidateHeaderChain(tt.args.chain)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHeaderChain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateHeaderChain() got = %v, want %v", got, tt.want)
			}
		})
	}
}
