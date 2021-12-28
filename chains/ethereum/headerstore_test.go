package ethereum

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
)

func TestHeaderStore_delOldHeaders(t *testing.T) {
	type fields struct {
		headers map[string][]byte
		tds     map[string]*big.Int
	}
	tests := []struct {
		name       string
		fields     fields
		fn         func(hs *HeaderStore)
		wantLength int
	}{
		{
			name: "",
			fields: fields{
				headers: make(map[string][]byte),
				tds:     make(map[string]*big.Int),
			},
			fn: func(hs *HeaderStore) {
				for i := 1; i <= 10; i++ {
					key1 := headerKey(uint64(i), common.BigToHash(big.NewInt(int64(i*23))))
					key2 := headerKey(uint64(i), common.BigToHash(big.NewInt(int64(i*654))))
					header := encodeHeader(&Header{
						Number: big.NewInt(int64(i)),
					})
					hs.headers[key1] = header
					hs.headers[key2] = header
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := &HeaderStore{
				headers: tt.fields.headers,
				tds:     tt.fields.tds,
			}

			tt.fn(hs)
			hs.delOldHeaders()
			if len(hs.headers) > MaxHeaderLimit {
				t.Errorf("delOldHeaders() failed, want length: %d, got length: %d", tt.wantLength, len(hs.headers))
			}

			t.Log("header length: ", len(hs.headers))
			for k := range hs.headers {
				t.Log("key-2: ", k)
			}
		})
	}
}
