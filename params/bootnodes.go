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
	"enode://1457e5fc1cc9a1584aacbc860922563d049a2efb9aa0fc73089c85769f1e39e53c3419984af0af75cd29177af1b62e90ac86912562c05583b6cd158fd532f0d5@20.205.187.105:36201",
	"enode://ec9e7dee840333b72addfb6d1b76020bd9a72f0cb113dbd18ff3f51d70e03d80b23c63e102fbeb4ecad36da03d2d4fb6afaa6178849d499ba653d2c55013497d@20.205.187.105:36202",
	"enode://b02d993aa6f81012e532f9b494d1ab3036cabbbd6401a40fb0c3133ed3997887e41f800de8e9da9f3a94665270a1b14c7346d9c13332397bca953bf65538e715@20.205.189.217:36203",
	"enode://dcbe8b57c5c265668a33e4f41214b1161860602d22d710a713e4cc2005a3cf2bc417be10a8a1d251ffe7e55e566319d368826103dba159d734b3ee4fa7c02c76@20.205.189.217:36204",
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
