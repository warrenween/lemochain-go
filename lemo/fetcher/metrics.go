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

// Contains the metrics collected by the fetcher.

package fetcher

import (
	"github.com/LemoFoundationLtd/lemochain-go/metrics"
)

var (
	propAnnounceInMeter   = metrics.NewRegisteredMeter("lemo/fetcher/prop/announces/in", nil)
	propAnnounceOutTimer  = metrics.NewRegisteredTimer("lemo/fetcher/prop/announces/out", nil)
	propAnnounceDropMeter = metrics.NewRegisteredMeter("lemo/fetcher/prop/announces/drop", nil)
	propAnnounceDOSMeter  = metrics.NewRegisteredMeter("lemo/fetcher/prop/announces/dos", nil)

	propBroadcastInMeter   = metrics.NewRegisteredMeter("lemo/fetcher/prop/broadcasts/in", nil)
	propBroadcastOutTimer  = metrics.NewRegisteredTimer("lemo/fetcher/prop/broadcasts/out", nil)
	propBroadcastDropMeter = metrics.NewRegisteredMeter("lemo/fetcher/prop/broadcasts/drop", nil)
	propBroadcastDOSMeter  = metrics.NewRegisteredMeter("lemo/fetcher/prop/broadcasts/dos", nil)

	headerFetchMeter = metrics.NewRegisteredMeter("lemo/fetcher/fetch/headers", nil)
	bodyFetchMeter   = metrics.NewRegisteredMeter("lemo/fetcher/fetch/bodies", nil)

	headerFilterInMeter  = metrics.NewRegisteredMeter("lemo/fetcher/filter/headers/in", nil)
	headerFilterOutMeter = metrics.NewRegisteredMeter("lemo/fetcher/filter/headers/out", nil)
	bodyFilterInMeter    = metrics.NewRegisteredMeter("lemo/fetcher/filter/bodies/in", nil)
	bodyFilterOutMeter   = metrics.NewRegisteredMeter("lemo/fetcher/filter/bodies/out", nil)
)
