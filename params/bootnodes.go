package params

import "github.com/ethereum/go-ethereum/common"

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main network.
var MainnetBootnodes = []string{
	"enode://1f4951d5ac8387c6a0beef775f918a211f119c20ee15dbc98b641219bd093e2ee7a1c2db3f29406193c61f86990c0c0a670e9f423c860741f2140d3f8e87cabf@127.0.0.1:30321",
	"enode://befe1b92db122a9ce3e2b9b961f3c683e851c9c7f39c509906341242e792b049b5e3fcd14504f2bf9ee26e9b26afa9cef5274c17ea680df57c4dc1e7b7d189a6@127.0.0.1:30322",
	"enode://fc54fbd58b1c82af1ddec1d815f97f3865379cd28f66b2e648e234d539951c1bd37fa3f20e72937eb72d14714ce9748769533409b1f8bbaa4b4d211caf032208@127.0.0.1:30323",
	"enode://56bdcd9cee2f8c7a913c0367dae513912d3546df759e22172f7474fb08d06413dc7e75ba9e3e26134f52d7670a105d5b812768fa80ef1654b01e7d761f4f5784@127.0.0.1:30324",
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// test network.
var TestnetBootnodes = []string{
	"enode://023df9b6444a0b0182b06108d6af0ff549368fa377ec1f338531ecc2dec9be8abaca85dda5409e35ec574b7198b70ddb2d5d11ee95f0745742f10cf57719d389@8.219.183.167:30321",
	"enode://0d6541527fb9dc6e79c9f1c3f19f8657b27040527af63d8d062cfb24357b281c5f0bf31c6d74d2e6ce23a3cef42a29e9391f2bc45febd240b25e9f0c0d56c15a@8.219.238.137:30321",
	"enode://1c14593f00017ab08cf2262333c322d490968b618e7b5d28b657ae0d6f62808ab90945017ab34a91fdc16f31eebf21a3885e5e11b53ed76363a80899fe274160@8.219.118.187:30321",
	"enode://16ede4c000bf8110c04cdbb7f81ae9ee2957d9fd2836b662d8e05e2573a7f896590722bc3b523226ccd4bca845d917481c3c130848cb21b85f59f1c8560a8a96@8.219.242.200:30321",
}

var DevnetBootnodes = []string{
	"enode://8aadaff997ddc4cafa63ea88aeb023a29c31e23dd95808fcb0866c0432bc8c367f3850225eec3167595f848a381b658c705d13b8d58080c121b3ee4023703c09@13.250.12.223:30321",
	"enode://d7f54ff377ca8fe0a6df5ace4ce5dd60b36a8c43c89a7a96747a7bc5cb2eb2400a9571e5af2c4a48f874db7564658c80621a42808b4a6dc4424df864124f904a@13.250.12.223:30322",
	"enode://e9ccb2587a9ba8e91a2f01004ff05530a3654907d703227a97a296c03a30db042e94bff768f95330a978841046a72a3bed541dce50bbe22cf2d6f73291da2598@3.0.19.66:30323",
	"enode://4bda7d541e2c7daa348bbbbc69448aeee33d5312b1b4d38777042534776d27537c12dc2937609ce37c7e9cd57c8b75acda15359257d8bd40b3559fde385c2da0@3.0.19.66:30324",
}

const dnsPrefix = "enrtree://AKA3AM6LPBYEUDMVNU3BSVQJ5AD45Y7YPOHJLEF6W26QOE4VTUDPE@"

// KnownDNSNetwork returns the address of a public DNS-based node list for the given
// genesis hash and protocol. See https://github.com/ethereum/discv4-dns-lists for more
// information.
func KnownDNSNetwork(genesis common.Hash, protocol string) string {
	return ""
}
