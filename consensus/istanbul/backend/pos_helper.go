package backend

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/accounts/abi"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/params"
	"math"
	"math/big"
)

var (
	myAddresss = common.HexToAddress("0x81f02fd21657df80783755874a92c996749777bf")
)

func sendTranction(height uint64, gen *chain.BlockGen, state *ethdb.Database, from, to common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, signer types.Signer, header *types.Header) {
	if height == 10 {
		nonce, statedb := getNonce(gen, from, "sendTranction")
		balance := statedb.GetBalance(to)
		remaining := new(big.Int).Sub(value, balance)
		printTest("sendTranction ", balance.Uint64(), " remaining ", remaining.Uint64(), " height ", height, " current ", header.Number.Uint64())
		if remaining.Sign() > 0 {
			tx, _ := types.SignTx(types.NewTransaction(nonce, to, remaining, params.DefaultGasLimit, new(big.Int).SetInt64(1000000), nil), signer, privateKey)
			gen.AddTx(tx)
		} else {
			printTest("to ", to.String(), " have balance ", balance.Uint64(), " height ", height, " current ", header.Number.Uint64())
		}
	}
}

func sendContractTransaction(height uint64, gen *chain.BlockGen, from common.Address, value *big.Int, priKey *ecdsa.PrivateKey, signer types.Signer, blockchain *chain.BlockChain, abiStaking abi.ABI) {
	nonce, _ := getNonce(gen, from, "sendDepositTransaction")
	pub := crypto.FromECDSAPub(&priKey.PublicKey)
	input := packInput(abiStaking, "deposit", "sendDepositTransaction", pub, new(big.Int).SetInt64(5000), value)
	addTx(gen, blockchain, nonce, nil, input, priKey, signer)
}

func packInput(abiStaking abi.ABI, abiMethod, method string, params ...interface{}) []byte {
	input, err := abiStaking.Pack(abiMethod, params...)
	if err != nil {
		printTest(method, " error ", err)
	}
	return input
}

func getNonce(gen *chain.BlockGen, from common.Address, method string) (uint64, *state.StateDB) {
	var nonce uint64
	var stateDb *state.StateDB
	nonce = gen.TxNonce(from)
	stateDb = gen.GetStateDB()
	printBalance(stateDb, from, method)
	return nonce, stateDb
}

func addTx(gen *chain.BlockGen, blockchain *chain.BlockChain, nonce uint64, value *big.Int, input []byte, priKey *ecdsa.PrivateKey, signer types.Signer) {

	tx, _ := types.SignTx(types.NewTransaction(nonce, myAddresss, value, 2646392, big.NewInt(1000000000000), input), signer, priKey)

	gen.AddTxWithChain(blockchain, tx)

}
func printBalance(stateDb *state.StateDB, from common.Address, method string) {
	balance := stateDb.GetBalance(myAddresss)
	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	StakinValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))

	printTest(method, " from ", from.String(), " Staking fbalance ", fbalance, " StakinValue ", StakinValue, "from ", from.String())
}

func printTest(a ...interface{}) {
	log.Info("test", "SendTX", a)
}
