package params

import "github.com/ethereum/go-ethereum/common"

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main network.
var MainnetBootnodes = []string{
	"enode://71ffdcb10afd3f105320f4c8e6e85d89c93c55d08723273282ad398570cbec7e2475a286ca491a7eedb4bca84ab587e17a801306671e9fba0e04e12ef64e71f9@13.251.42.33:28360",
	"enode://6aba005b149e0115489829d525ade140ad61a5b86c7517b287c47d9b0a89bc8ae3d335fe207cfa17710fce536f2e918a4ce3f0935ca1423cfc88833642fa1e27@13.214.132.40:28360",
	"enode://51961120e7f5ebc8ee6f559fc1a7f303ab0af6c8e019b8009b77b524e5eb6d53b3a24c1f92ad4b3c2a880e089ef011038041674b29b264c1b24fbde6b125d922@3.0.49.193:28360",
	"enode://6bc28d4c98a8e4af33c9350c699dc2560bbffd3cd4555dfdd100074a4d10c4603fd2d069d8e6abc22aa6d11b018ed251173a723d224452cc1ae95e7c22296292@18.138.231.32:28360",
	"enode://de4df7ce9afbe340a5397d17313fbbad2227e233e1100d40fb8a4a56bb057fa14f445b54dbab520f591285efc7babea020bb7146e3c078b4477d43cc5717a4a7@13.214.151.165:28360",
	"enode://cdac08bd6dc84d55295130a4d0073a6b5e820ca728faf80de5951c0a5c495884bca0bceec52b26a7d6ba9d5f3ea385771200058ce56ec7d08172b858a9dca0f5@18.142.162.63:28360",
	"enode://b68ed0df9dc2306c5cec6ca953fd629d80214d6dbcf5e60e12026e8e3ce42d5eee6d8801d747e110a1c886e5c76e2a526a81f775ef8ac63e9b30b242cae06b44@54.251.159.228:20101",
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// test network.
var TestnetBootnodes = []string{
	"enode://478dace9ed069fdeb2170ca1bbc34314b13dbf2e7273fa15b7e40e63be036a684011bd627486e2d009771ed7e2ad435bb62b9026b87e339d470a1ffe5bd83034@20.205.187.105:36201",
	"enode://530d20cdf6552d9dc65c5abf7c40727254719a42d5e033d391b8c9f77d15ee35e4a43de40804b4649ffca95bf75b6902c65d418fab7c08b07b192ead0cda893d@20.205.189.217:36201",
	"enode://a5cbde4cd043a59dff200882f2663b35cabe753211ee372a1681a9159bdbccac5b08ede298d8681540d1cfaf9e79e1046608e3b61600479e49cb61c95dc842e0@13.67.79.15:25201",
	"enode://1ff54f4b794eba4081bf24da657621051d74edc4fb071c021f44dceb4e7628c030e3bbe6722ccc7902e93d40c018f9a03a99b14c7d7f2d6bfd8955193db46acd@13.67.118.60:25201",
}

const dnsPrefix = "enrtree://AKA3AM6LPBYEUDMVNU3BSVQJ5AD45Y7YPOHJLEF6W26QOE4VTUDPE@"

// KnownDNSNetwork returns the address of a public DNS-based node list for the given
// genesis hash and protocol. See https://github.com/ethereum/discv4-dns-lists for more
// information.
func KnownDNSNetwork(genesis common.Hash, protocol string) string {
	return ""
}
