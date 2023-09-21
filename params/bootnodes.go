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
	"enode://f50304fcabf3a578d8db883a233046c2673490636cf15d109e61f7d77067465ded8ca11dd097caad39bc5c26a9fffadb1ad835a24e0e461a49f03d20aa6cb378@124.156.200.50:30321",
	"enode://dcefb2c0a75945e4512b56de228f59715b6832ee0c92f6edafa5fd0c0cab31d5c0bb9175c4f43cb8b26b97ddaf72da7daa193937e2034363ca5dff070173a2eb@124.156.200.50:30322",
	"enode://e9ccb2587a9ba8e91a2f01004ff05530a3654907d703227a97a296c03a30db042e94bff768f95330a978841046a72a3bed541dce50bbe22cf2d6f73291da2598@43.134.183.62:30323",
	"enode://4bda7d541e2c7daa348bbbbc69448aeee33d5312b1b4d38777042534776d27537c12dc2937609ce37c7e9cd57c8b75acda15359257d8bd40b3559fde385c2da0@43.134.183.62:30324",
}

const dnsPrefix = "enrtree://AKA3AM6LPBYEUDMVNU3BSVQJ5AD45Y7YPOHJLEF6W26QOE4VTUDPE@"

// KnownDNSNetwork returns the address of a public DNS-based node list for the given
// genesis hash and protocol. See https://github.com/ethereum/discv4-dns-lists for more
// information.
func KnownDNSNetwork(genesis common.Hash, protocol string) string {
	return ""
}
