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
	"enode://0db552cc51d743b817457111be8341300f48f1e2c9859d795bbdbb1c3d13d45da333ec2307580cd360e814c48b7e46055334739590f833cd3ed718edaefe861d@47.236.181.176:20301",
	"enode://0443f7e7c525a50a87abb633da5aa5d1c394087cacc1330f48dc70ae1d0eacd00a94a51ce71707264a78b25102c09c023e786198a754010038862aa58ad2eddc@47.236.178.133:20301",
	"enode://67ed15af0dcda70f34f3338c7300b1625d5fe3e237dcb0fb2ff2319c82871b6fb3a4189a1ac4be85715a65db6976616ecb22878811b31cce63efd18a6e231532@47.236.182.75:20301",
	"enode://a93467563321c97fdf831e0bf75199e43bf0af3e698824726768606005d90248ff25408d6b0ff9e19d812a1bef335a6df0196eb40cff593be61d36ffdc1e1a36@47.236.182.75:9101",
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
