// Copyright 2015 The go-ethereum Authors
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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/ethereum/go-ethereum/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("atlas/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("atlas/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("atlas/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("atlas/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("atlas/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("atlas/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("atlas/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("atlas/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("atlas/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("atlas/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("atlas/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("atlas/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("atlas/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("atlas/downloader/states/drop", nil)

	throttleCounter = metrics.NewRegisteredCounter("atlas/downloader/throttle", nil)
)
