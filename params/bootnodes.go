package params

import "github.com/ethereum/go-ethereum/common"

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main network.
var MainnetBootnodes = []string{
	"enode://e6b83c8894b9af84520d8578145e467e5f99cf8b0b15df874c9005e6bb337d19a08c6943d69357913bce8f29a413cd0b2dc17c354cd506f8b824a46599003b33@13.67.79.15:21221",
	"enode://a1ba132ef1ff7e6fea9200a0e9520c44f46498be2f692f0ace0d40ecc3afe968d995e3989242d956cd7180136eb217ce2153468e8fc8f0d9a9d72eaa22a44a9b@13.67.118.60:21221",
	"enode://b5d1c1af9b2329d7e50624fc73c53dcfdb6d16781268887cf98f49d17a0fc83701ceb38d00a6a0d9861247d05da359f1edb9ffd8d6edb0685db20d63d8b591f1@13.76.138.119:21221",
	"enode://0e2edf827a7520f16c20d4ff9d18c269de5d48f73f28a3c284edc7462df096853d2b9e8a1131e937061b25e4fb1c0a98913cc92a483af2e1d7f3d62f674f1da9@168.63.248.220:21221",
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// test network.
var TestnetBootnodes = []string{
	//"enode://8ff2e770fdfab6234cd2cd31147c69faaf69daf0c0473e2033ee532a77db1c3838616183f9c53f212ffaeae151f96d847c184f29371829b281087979fd7f7db4@119.8.165.158:20101",
}

const dnsPrefix = "enrtree://AKA3AM6LPBYEUDMVNU3BSVQJ5AD45Y7YPOHJLEF6W26QOE4VTUDPE@"

// KnownDNSNetwork returns the address of a public DNS-based node list for the given
// genesis hash and protocol. See https://github.com/ethereum/discv4-dns-lists for more
// information.
func KnownDNSNetwork(genesis common.Hash, protocol string) string {
	var net string
	switch genesis {
	case MainnetGenesisHash:
		net = "mainnet"
	case TestnetGenesisHash:
		net = "testnet"
	default:
		return ""
	}
	return dnsPrefix + protocol + "." + net + ".ethdisco.net"
}
