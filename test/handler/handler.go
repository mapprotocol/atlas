package handler

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"os"
)

func init() {
	startLogger()
}

func startLogger() {
	var lvl = log.LvlInfo
	logger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(true)))
	logger.Verbosity(lvl)
	log.Root().SetHandler(logger)
}

func getMgrMaintainerAddress(endpoint string) {
	cli := dial(endpoint)
	parsed := parseABI(EpochRewards)
	input := packInput(parsed, "getMgrMaintainerAddress")
	output := CallContract(cli, GenesisAddresses["EpochRewardsProxy"], input)
	var addr common.Address
	if err := parsed.UnpackIntoInterface(&addr, "getMgrMaintainerAddress", output); err != nil {
		log.Crit("unpack failed", "err", err.Error())
	}
	log.Info("getMgrMaintainerAddress", "address", addr)
}

func setMgrMaintainerAddress(endpoint string, from, target common.Address, privateKey *ecdsa.PrivateKey) {
	cli := dial(endpoint)
	input := packInput(parseABI(EpochRewards), "setMgrMaintainerAddress", target)
	txHash := sendContractTransaction(cli, from, GenesisAddresses["EpochRewardsProxy"], nil, privateKey, input, 0)
	getResult(cli, txHash)
	log.Info("setMgrMaintainerAddress", "address", target)
}

func getTargetEpochPayment(endpoint string) {
	cli := dial(endpoint)
	parsed := parseABI(EpochRewards)
	input := packInput(parsed, "epochPayment")
	output := CallContract(cli, GenesisAddresses["EpochRewardsProxy"], input)
	var value *big.Int
	if err := parsed.UnpackIntoInterface(&value, "epochPayment", output); err != nil {
		log.Crit("unpack failed", "err", err.Error())
	}
	log.Info("getTargetEpochPayment", "value", value)
}

func setTargetEpochPayment(endpoint string, from common.Address, target *big.Int, privateKey *ecdsa.PrivateKey) {
	cli := dial(endpoint)
	input := packInput(parseABI(EpochRewards), "setTargetEpochPayment", target)
	txHash := sendContractTransaction(cli, from, GenesisAddresses["EpochRewardsProxy"], nil, privateKey, input, 0)
	getResult(cli, txHash)
	log.Info("setTargetEpochPayment", "value", target)
}
