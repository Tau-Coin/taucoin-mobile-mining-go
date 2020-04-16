// Copyright 2015 The go-tau Authors
// This file is part of the go-tau library.
//
// The go-tau library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-tau library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-tau library. If not, see <http://www.gnu.org/licenses/>.

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/Tau-Coin/taucoin-mobile-mining-go/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("tau/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("tau/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("tau/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("tau/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("tau/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("tau/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("tau/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("tau/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("tau/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("tau/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("tau/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("tau/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("tau/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("tau/downloader/states/drop", nil)
)
