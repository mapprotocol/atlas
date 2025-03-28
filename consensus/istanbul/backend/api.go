// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package backend

import (
	"errors"
	"fmt"
	"github.com/mapprotocol/atlas/tools"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/mapprotocol/atlas/consensus"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	vet "github.com/mapprotocol/atlas/consensus/istanbul/backend/internal/enodes"
	"github.com/mapprotocol/atlas/consensus/istanbul/backend/internal/replica"
	"github.com/mapprotocol/atlas/consensus/istanbul/core"
	"github.com/mapprotocol/atlas/consensus/istanbul/proxy"
	"github.com/mapprotocol/atlas/consensus/istanbul/uptime"
	"github.com/mapprotocol/atlas/consensus/istanbul/uptime/store"
	"github.com/mapprotocol/atlas/consensus/istanbul/validator"
	"github.com/mapprotocol/atlas/core/types"
	blscrypto "github.com/mapprotocol/atlas/helper/bls"
)

// API is a user facing RPC API to dump Istanbul state
type API struct {
	chain    consensus.ChainHeaderReader
	istanbul *Backend
}

type G1PublicKey struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type Validator struct {
	Weight   string         `json:"weight"`
	Address  common.Address `json:"address"`
	G1PubKey G1PublicKey    `json:"g1_pub_key"`
}

type EpochInfo struct {
	Epoch      string      `json:"epoch"`
	EpochSize  string      `json:"epoch_size"`
	Threshold  string      `json:"threshold"`
	Validators []Validator `json:"validators"`
}

// getHeaderByNumber retrieves the header requested block or current if unspecified.
func (api *API) getHeaderByNumber(number *rpc.BlockNumber) (*types.Header, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else if *number == rpc.PendingBlockNumber {
		return nil, fmt.Errorf("can't use pending block within istanbul")
	} else if *number == rpc.EarliestBlockNumber {
		header = api.chain.GetHeaderByNumber(0)
	} else {
		header = api.chain.GetHeaderByNumber(uint64(*number))
	}

	if header == nil {
		return nil, errUnknownBlock
	}
	return header, nil
}

// getParentHeaderByNumber retrieves the parent header requested block or current if unspecified.
func (api *API) getParentHeaderByNumber(number *rpc.BlockNumber) (*types.Header, error) {
	var parent uint64
	if number == nil || *number == rpc.LatestBlockNumber || *number == rpc.PendingBlockNumber {
		head := api.chain.CurrentHeader()
		if head == nil {
			return nil, errUnknownBlock
		}
		if number == nil || *number == rpc.LatestBlockNumber {
			parent = head.Number.Uint64() - 1
		} else {
			parent = head.Number.Uint64()
		}
	} else if *number == rpc.EarliestBlockNumber {
		return nil, errUnknownBlock
	} else {
		parent = uint64(*number - 1)
	}

	header := api.chain.GetHeaderByNumber(parent)
	if header == nil {
		return nil, errUnknownBlock
	}
	return header, nil
}

// GetSnapshot retrieves the state snapshot at a given block.
func (api *API) GetSnapshot(number *rpc.BlockNumber) (*Snapshot, error) {
	// Retrieve the requested block number (or current if none requested)
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	// Ensure we have an actually valid block and return its snapshot
	if header == nil {
		return nil, errUnknownBlock
	}
	return api.istanbul.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
}

// GetValidators retrieves the list validators that must sign a given block.
func (api *API) GetValidators(number *rpc.BlockNumber) ([]common.Address, error) {
	header, err := api.getParentHeaderByNumber(number)
	if err != nil {
		return nil, err
	}
	validators := api.istanbul.GetValidators(header.Number, header.Hash())
	return istanbul.MapValidatorsToAddresses(validators), nil
}

// GetValidatorsBLSPublicKeys retrieves the list of validators BLS public keys that must sign a given block.
func (api *API) GetValidatorsBLSPublicKeys(number *rpc.BlockNumber) ([]blscrypto.SerializedPublicKey, error) {
	header, err := api.getParentHeaderByNumber(number)
	if err != nil {
		return nil, err
	}
	validators := api.istanbul.GetValidators(header.Number, header.Hash())
	return istanbul.MapValidatorsToPublicKeys(validators), nil
}

// GetProposer retrieves the proposer for a given block number (i.e. sequence) and round.
func (api *API) GetProposer(sequence *rpc.BlockNumber, round *uint64) (common.Address, error) {
	header, err := api.getParentHeaderByNumber(sequence)
	if err != nil {
		return common.Address{}, err
	}

	valSet := api.istanbul.getOrderedValidators(header.Number.Uint64(), header.Hash())
	if valSet == nil {
		return common.Address{}, err
	}
	previousProposer, err := api.istanbul.Author(header)
	if err != nil {
		return common.Address{}, err
	}
	if round == nil {
		round = new(uint64)
	}
	proposer := validator.GetProposerSelector(api.istanbul.config.ProposerPolicy)(valSet, previousProposer, *round)
	return proposer.Address(), nil
}

// AddProxy peers with a remote node that acts as a proxy, even if slots are full
func (api *API) AddProxy(url, externalUrl string) (bool, error) {
	if !api.istanbul.config.Proxied {
		api.istanbul.logger.Error("Add proxy node failed: this node is not configured to be proxied")
		return false, errors.New("Can't add proxy for node that is not configured to be proxied")
	}

	node, err := enode.ParseV4(url)
	if err != nil {
		return false, fmt.Errorf("invalid enode: %v", err)
	}

	externalNode, err := enode.ParseV4(externalUrl)
	if err != nil {
		return false, fmt.Errorf("invalid external enode: %v", err)
	}

	err = api.istanbul.AddProxy(node, externalNode)
	return true, err
}

// RemoveProxy removes a node from acting as a proxy
func (api *API) RemoveProxy(url string) (bool, error) {
	// Try to remove the url as a proxy and return
	node, err := enode.ParseV4(url)
	if err != nil {
		return false, fmt.Errorf("invalid enode: %v", err)
	}
	if err = api.istanbul.RemoveProxy(node); err != nil {
		return false, err
	}

	return true, nil
}

// Retrieve the Validator Enode Table
func (api *API) GetValEnodeTable() (map[string]*vet.ValEnodeEntryInfo, error) {
	return api.istanbul.valEnodeTable.ValEnodeTableInfo()
}

func (api *API) GetVersionCertificateTableInfo() (map[string]*vet.VersionCertificateEntryInfo, error) {
	return api.istanbul.announceManager.GetVersionCertificateTableInfo()
}

// GetCurrentRoundState retrieves the current IBFT RoundState
func (api *API) GetCurrentRoundState() (*core.RoundStateSummary, error) {
	api.istanbul.coreMu.RLock()
	defer api.istanbul.coreMu.RUnlock()

	if !api.istanbul.isCoreStarted() {
		return nil, istanbul.ErrStoppedEngine
	}
	return api.istanbul.core.CurrentRoundState().Summary(), nil
}

func (api *API) ForceRoundChange() (bool, error) {
	api.istanbul.coreMu.RLock()
	defer api.istanbul.coreMu.RUnlock()

	if !api.istanbul.isCoreStarted() {
		return false, istanbul.ErrStoppedEngine
	}
	api.istanbul.core.ForceRoundChange()
	return true, nil
}

// GetProxiesInfo retrieves all the proxied validator's proxies' info
func (api *API) GetProxiesInfo() ([]*proxy.ProxyInfo, error) {
	if api.istanbul.IsProxiedValidator() {
		proxies, valAssignments, err := api.istanbul.proxiedValidatorEngine.GetProxiesAndValAssignments()

		if err != nil {
			return nil, err
		}

		proxyInfoArray := make([]*proxy.ProxyInfo, 0, len(proxies))

		for _, proxyObj := range proxies {
			proxyInfoArray = append(proxyInfoArray, proxy.NewProxyInfo(proxyObj, valAssignments[proxyObj.ID()]))
		}

		return proxyInfoArray, nil
	} else {
		return nil, proxy.ErrNodeNotProxiedValidator
	}
}

// ProxiedValidators retrieves all of the proxies connected proxied validators.
// Note that we plan to support validators per proxy in the future, so this function
// is plural and returns an array of proxied validators.  This is to prevent
// future backwards compatibility issues.
func (api *API) GetProxiedValidators() ([]*proxy.ProxiedValidatorInfo, error) {
	if api.istanbul.IsProxy() {
		return api.istanbul.proxyEngine.GetProxiedValidatorsInfo()
	} else {
		return nil, proxy.ErrNodeNotProxy
	}
}

// StartValidating starts the consensus engine
func (api *API) StartValidating() error {
	return api.istanbul.MakePrimary()
}

// StopValidating stops the consensus engine from participating in consensus
func (api *API) StopValidating() error {
	return api.istanbul.MakeReplica()
}

// StartValidatingAtBlock starts the consensus engine on the given
// block number.
func (api *API) StartValidatingAtBlock(blockNumber int64) error {
	seq := big.NewInt(blockNumber)
	return api.istanbul.SetStartValidatingBlock(seq)
}

// StopValidatingAtBlock stops the consensus engine from participating in consensus
// on the given block number.
func (api *API) StopValidatingAtBlock(blockNumber int64) error {
	seq := big.NewInt(blockNumber)
	return api.istanbul.SetStopValidatingBlock(seq)
}

// IsValidating returns true if this node is participating in the consensus protocol
func (api *API) IsValidating() bool {
	return api.istanbul.IsValidating()
}

// GetCurrentReplicaState retrieves the current replica state
func (api *API) GetCurrentReplicaState() (*replica.ReplicaStateSummary, error) {
	if api.istanbul.replicaState != nil {
		return api.istanbul.replicaState.Summary(), nil
	}
	return &replica.ReplicaStateSummary{State: "Not a validator"}, nil
}

// GetLookbackWindow retrieves the current replica state
func (api *API) GetLookbackWindow(number *rpc.BlockNumber) (uint64, error) {
	header, err := api.getHeaderByNumber(number)
	if err != nil {
		return 0, err
	}

	state, err := api.istanbul.stateAt(header.Hash())
	if err != nil {
		return 0, err
	}

	return api.istanbul.LookbackWindow(header, state), nil
}

func (api *API) Activity() (map[string]interface{}, error) {
	header := api.chain.CurrentHeader()
	if header == nil {
		return nil, errUnknownBlock
	}

	state, err := api.istanbul.stateAt(header.Hash())
	if err != nil {
		return nil, err
	}

	signers := api.istanbul.GetValidators(big.NewInt(header.Number.Int64()-1), header.ParentHash)
	if len(signers) == 0 {
		return nil, errors.New("unable to fetch validators")
	}
	vmRunner := api.istanbul.chain.NewEVMRunner(header, state)
	accounts, err := api.istanbul.GetAccountsFromSigners(vmRunner, signers)
	if err != nil {
		return nil, err
	}

	epochNum := istanbul.GetEpochNumber(header.Number.Uint64(), api.istanbul.EpochSize())
	numberWithinEpoch := istanbul.GetNumberWithinEpoch(header.Number.Uint64(), api.istanbul.EpochSize())

	ret := make(map[string]interface{})
	ret["epoch"] = epochNum
	ret["number"] = header.Number
	ret["hash"] = header.Hash()
	us := make(map[string]interface{})
	if numberWithinEpoch <= 12 {
		for _, acc := range accounts {
			us[acc.Hex()] = map[string]interface{}{
				"uptime":   1,
				"upBlocks": numberWithinEpoch,
			}
		}
		ret["uptimes"] = us
		return ret, nil
	}

	lookBackWindow := api.istanbul.LookbackWindow(header, state)
	monitor := uptime.NewMonitor(store.New(api.istanbul.db), api.istanbul.EpochSize(), lookBackWindow)

	entries, uptimes, err := monitor.GetValidatorsActivity(epochNum, numberWithinEpoch, len(accounts))
	if err != nil {
		return nil, err
	}

	for i, u := range uptimes {
		us[accounts[i].Hex()] = map[string]interface{}{
			"uptime":   u,
			"upBlocks": entries[i].UpBlocks + lookBackWindow,
		}
	}
	ret["uptimes"] = us
	return ret, nil
}

// GetEpochInfo retrieves the epoch info
func (api *API) GetEpochInfo(epochNumber uint64) *EpochInfo {
	number, _ := istanbul.GetEpochFirstBlockNumber(epochNumber, api.istanbul.config.Epoch)
	header := api.chain.GetHeaderByNumber(number)
	// Ensure we have an actually valid block
	if header == nil {
		return nil
	}

	ss, err := api.istanbul.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil
	}

	validatorNum := len(ss.validators())
	validators := make([]Validator, 0, validatorNum)
	for _, v := range ss.validators() {
		validators = append(validators, Validator{
			Weight:  strconv.Itoa(1),
			Address: v.Address,
			G1PubKey: G1PublicKey{
				X: tools.Bytes2Hex(v.BLSG1PublicKey[:32]),
				Y: tools.Bytes2Hex(v.BLSG1PublicKey[32:]),
			},
		})
	}

	epochInfo := &EpochInfo{
		Epoch:      strconv.FormatUint(epochNumber, 10),
		EpochSize:  strconv.FormatUint(api.istanbul.config.Epoch, 10),
		Threshold:  strconv.Itoa(ss.ValSet.MinQuorumSize()),
		Validators: validators,
	}
	return epochInfo
}
