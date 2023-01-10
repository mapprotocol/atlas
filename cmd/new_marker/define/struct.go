package define

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Voter2validatorInfo struct {
	VoterAccount     string
	ValidatorAccount string
	Value            uint64
}

var (
	Voter2validator []Voter2validatorInfo
	VoterList       []VoterStruct
)

type VoterStruct struct {
	Voter     common.Address
	Validator common.Address
}

type ValidatorInfo struct {
	EpochNum        uint64
	AllVotes        *big.Int
	ValidatorReward *big.Int
}

type VoterInfo struct {
	EpochNum  uint64
	VActive   *big.Int
	VPending  *big.Int
	Voter     common.Address
	Validator common.Address
}
