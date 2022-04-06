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
	"enode://7b77fc2322c2006d60880ae55cf5725c19f5022ab384e33cf77a6a348575f57fd7ff968bbab8872a91d51030ffe7062053463dcaee8269dc7ca1409f83b916cd@20.205.187.105:36201",
	"enode://565da6757eabc7d9dbd12687b2bc12be9edb07a6ad5e113afa08e7826555c65c45e4fa360d966e402670cdaeb17a4a0b33a3861fa6eaf82587d0404656021dbd@20.205.187.105:36202",
	"enode://8acf7a294c439d1a2c040e268b1d9758f73369e136428072ad8f29f9616fddf6b327282a3c72c457116cbcc0cc607c1f6129319129c1f2fac6dd78bc83116342@20.205.189.217:36203",
	"enode://65252ae7d160be483d54a20637be2ab1e30cb411e42cfeb58fa7a559f85ed5baeab7639de06110ebe8e6389681afb2a4d543f574a172830f97eeb71cf4ce92e1@20.205.189.217:36204",
}

const dnsPrefix = "enrtree://AKA3AM6LPBYEUDMVNU3BSVQJ5AD45Y7YPOHJLEF6W26QOE4VTUDPE@"

// KnownDNSNetwork returns the address of a public DNS-based node list for the given
// genesis hash and protocol. See https://github.com/ethereum/discv4-dns-lists for more
// information.
func KnownDNSNetwork(genesis common.Hash, protocol string) string {
	return ""
}
