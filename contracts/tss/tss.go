// Copyright 2021 MAP Protocol Authors.
// This file is part of MAP Protocol.

// MAP Protocol is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// MAP Protocol is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with MAP Protocol.  If not, see <http://www.gnu.org/licenses/>.

package tss

import (
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/abis"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
	"math/big"
)

var (
	getCurrentEpochMethod   = contracts.NewRegisteredContractMethod(params.TSSManagerRegistryId, abis.TSSManager, "currentEpoch", params.MaxGasForReadBlockchainParameter)
	getEpochPublicKeyMethod = contracts.NewRegisteredContractMethod(params.TSSManagerRegistryId, abis.TSSManager, "getEpochPubkey", params.MaxGasForReadBlockchainParameter)
)

func GetCurrentEpoch(vmRunner vm.EVMRunner) (*big.Int, error) {
	var epoch *big.Int
	err := getCurrentEpochMethod.Query(vmRunner, &epoch)
	if err != nil {
		return big.NewInt(0), err
	}
	return epoch, nil
}

func GetEpochPublicKey(vmRunner vm.EVMRunner, epoch *big.Int) ([]byte, error) {
	var pk []byte
	err := getEpochPublicKeyMethod.Query(vmRunner, &pk, epoch)
	if err != nil {
		return []byte{}, err
	}
	return pk, nil
}
