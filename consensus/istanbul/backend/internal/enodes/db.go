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

package enodes

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

var (
	errIncorrectEntryType = errors.New("Incorrect entry type")
)

const (
	dbAddressPrefix = "address:" // Identifier to prefix node entries with
	dbNodeIDPrefix  = "nodeid:"  // Identifier to prefix node entries with
)

func addressKey(address common.Address) []byte {
	return append([]byte(dbAddressPrefix), address.Bytes()...)
}

func nodeIDKey(nodeID enode.ID) []byte {
	return append([]byte(dbNodeIDPrefix), nodeID.Bytes()...)
}
