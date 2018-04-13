// Copyright 2015 The lemochain-go Authors
// This file is part of the lemochain-go library.
//
// The lemochain-go library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The lemochain-go library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the lemochain-go library. If not, see <http://www.gnu.org/licenses/>.

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/LemoFoundationLtd/lemochain-go/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("lemo/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("lemo/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("lemo/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("lemo/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("lemo/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("lemo/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("lemo/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("lemo/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("lemo/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("lemo/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("lemo/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("lemo/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("lemo/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("lemo/downloader/states/drop", nil)
)
