package params

import "github.com/ethereum/go-ethereum/common"

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main network.
var MainnetBootnodes = []string{
	"enode://2880658365490c34e2a5069a8749aa95e45fae0bc1a237622b2385da946f556fb3499d4fc8f041250fec6cd96ce5c3e84fdb0c111220dd3bc9a4cb93d92006d3@13.67.79.15:21221",
	"enode://f57448354ce6869ac747735d5c584250a59ba590ec4a21f8f5c8d0f882cd4446444923d04890698758bf05ac2b569fc19cda6c5dd1b73e48df197abfcfad2b51@13.67.118.60:21221",
	"enode://67faad96fb84ae2bc9a7a64f1c80bcfd1010097ec8e8e88be1ac1236b867f15a483cfb6c2e116ec519dd7ff3c2e9296236872ee37b933f609d6c076a05e1f3dd@13.76.138.119:21221",
	"enode://b40b6ecac03ac1b8e754e6ac9012f6286a61a4b6f7a685b30e1451e2ee2d404e18c644e58e099ac20e16dab3c87414f210d85a18a616129161c44ed2c620af47@168.63.248.220:21221",
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
