package handler

import "github.com/ethereum/go-ethereum/common"

var addr = common.HexToAddress

var GenesisAddresses = map[string]common.Address{
	"RegistryProxy":             addr("0xce10"),
	"GoldTokenProxy":            addr("0xd003"),
	"StableTokenProxy":          addr("0xd008"),
	"AccountsProxy":             addr("0xd010"),
	"LockedGoldProxy":           addr("0xd011"),
	"ValidatorsProxy":           addr("0xd012"),
	"ElectionProxy":             addr("0xd013"),
	"EpochRewardsProxy":         addr("0xd014"),
	"RandomProxy":               addr("0xd015"),
	"BlockchainParametersProxy": addr("0xd018"),
}
