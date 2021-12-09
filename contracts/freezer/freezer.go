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

package freezer

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/abis"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
)

var (
	isFrozenMethod = contracts.NewRegisteredContractMethod(params.FreezerRegistryId, abis.Freezer, "isFrozen", params.MaxGasForIsFrozen)
)

func IsFrozen(vmRunner vm.EVMRunner, registryId common.Hash) (bool, error) {

	address, err := contracts.GetRegisteredAddress(vmRunner, registryId)
	if err != nil {
		return false, err
	}

	var isFrozen bool
	if err := isFrozenMethod.Query(vmRunner, &isFrozen, address); err != nil {
		return false, err
	}

	return isFrozen, nil
}
