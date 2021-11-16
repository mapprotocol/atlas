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

package istanbul

import (
	"github.com/mapprotocol/atlas/p2p"
)

// Message codes for istanbul related messages
// If you want to add a code, you need to increment the protocolLengths Array size
// and update the IsIstanbulMsg function below!
const (
	ConsensusMsg           = 0x11
	QueryEnodeMsg          = 0x12
	ValEnodesShareMsg      = 0x13
	FwdMsg                 = 0x14
	DelegateSignMsg        = 0x15
	VersionCertificatesMsg = 0x16
	EnodeCertificateMsg    = 0x17
	ValidatorHandshakeMsg  = 0x18
)

func IsIstanbulMsg(msg p2p.Msg) bool {
	return msg.Code >= ConsensusMsg && msg.Code <= ValidatorHandshakeMsg
}

// IsGossipedMsg specifies which messages should be gossiped throughout the network (as opposed to directly sent to a peer).
func IsGossipedMsg(msgCode uint64) bool {
	return msgCode == QueryEnodeMsg || msgCode == VersionCertificatesMsg
}
