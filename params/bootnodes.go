package params

import "github.com/ethereum/go-ethereum/common"

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main network.
var MainnetBootnodes = []string{
	"enode://71ffdcb10afd3f105320f4c8e6e85d89c93c55d08723273282ad398570cbec7e2475a286ca491a7eedb4bca84ab587e17a801306671e9fba0e04e12ef64e71f9@13.251.42.33:28360",
	"enode://6aba005b149e0115489829d525ade140ad61a5b86c7517b287c47d9b0a89bc8ae3d335fe207cfa17710fce536f2e918a4ce3f0935ca1423cfc88833642fa1e27@13.214.132.40:28360",
	"enode://51961120e7f5ebc8ee6f559fc1a7f303ab0af6c8e019b8009b77b524e5eb6d53b3a24c1f92ad4b3c2a880e089ef011038041674b29b264c1b24fbde6b125d922@3.0.49.193:28360",
	"enode://6bc28d4c98a8e4af33c9350c699dc2560bbffd3cd4555dfdd100074a4d10c4603fd2d069d8e6abc22aa6d11b018ed251173a723d224452cc1ae95e7c22296292@18.138.231.32:28360",
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
