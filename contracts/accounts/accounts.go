package accounts

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

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/abis"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
)

var (
	signerToAccountMethod = contracts.NewRegisteredContractMethod(params.AccountsId, abis.Accounts, "signerToAccount", params.MaxGasForGetRegisteredValidators)
)

func GetSignerToAccountMethod(vmRunner vm.EVMRunner, signer common.Address) (common.Address, error) {
	var regVals common.Address
	if err := signerToAccountMethod.Query(vmRunner, &regVals, signer); err != nil {
		return regVals, err
	}
	return regVals, nil
}
