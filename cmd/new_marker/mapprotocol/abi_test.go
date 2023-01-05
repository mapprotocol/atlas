package mapprotocol

import (
	"fmt"
	"testing"
)

func TestProxyAddress(t *testing.T) {
	for name, value := range genesisAddresses {
		fmt.Println(name, value)
	}
}
