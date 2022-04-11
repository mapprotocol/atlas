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
	"enode://6c91db56d9e48357f9f9a1c8bd3ccaf0f198680eefe2bb481370022467dfc623bda498ae8784324373f1bb60f346b5bc470efea3d2f5d50a9f52fd11cae6bd1b@20.205.187.105:36201",
	"enode://1dfc3bb6baa953e278f2c0cf1beeed5c46ab41cd766da4e0d985095c36669d5bbfd4c3b21860c4a7b8f2827b374c7ffd5167833f8b66acf1fcf1df988e45ec62@20.205.189.217:36202",
	"enode://429639a3100f178b0c469cc9c909f7eb45f7a1269f398cb577822216d273aeb503e49d4da57d632e326b0210eb53c421eae96a0698c14b15f5b9de307cf84f27@13.67.79.15:36203",
	"enode://6426516ad565d6cdd7b2882b306f428a7871f9af61598e5d42ccceb606cd8fa5c316a4287ac3b9ed277f3a981364ee49c9ab76301dbdbf346ad10aa14b942eff@13.67.118.60:36204",
}

const dnsPrefix = "enrtree://AKA3AM6LPBYEUDMVNU3BSVQJ5AD45Y7YPOHJLEF6W26QOE4VTUDPE@"

// KnownDNSNetwork returns the address of a public DNS-based node list for the given
// genesis hash and protocol. See https://github.com/ethereum/discv4-dns-lists for more
// information.
func KnownDNSNetwork(genesis common.Hash, protocol string) string {
	return ""
}
