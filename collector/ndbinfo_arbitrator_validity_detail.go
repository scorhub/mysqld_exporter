// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Scrape `ndbinfo.arbitrator_validity_detail`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoArbValDetQuery = `
	SELECT 
		node_id,
		arbitrator,
		arb_ticket,
		arb_connected,
		arb_state
	FROM ndbinfo.arbitrator_validity_detail;
`

// Metric descriptors.
var (
	ndbinfoArbitratorValidityDetailArbitratorDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "arbitrator_validity_detail_arbitrator"),
		"The arbitrator id by node_id/arb_ticket/arb_connected/arb_state.",
		[]string{"node_id", "arb_ticket", "arb_connected", "arb_state"}, nil,
	)
)

// ScrapeNDBInfoArbValDet collects from `ndbinfo.arbitrator_validity_detail`.
type ScrapeNDBInfoArbValDet struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoArbValDet) Name() string {
	return NDBInfo + ".arbitrator_validity_detail"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoArbValDet) Help() string {
	return "Collect metrics from ndbinfo.arbitrator_validity_detail"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoArbValDet) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoArbValDet) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoArbValDetRows, err := db.QueryContext(ctx, ndbinfoArbValDetQuery)
	if err != nil {
		return err
	}
	defer ndbinfoArbValDetRows.Close()

	var (
		nodeID                      string
		arbitrator                  uint64
		arbTicket, connected, state string
	)
	for ndbinfoArbValDetRows.Next() {
		if err := ndbinfoArbValDetRows.Scan(
			&nodeID,
			&arbitrator,
			&arbTicket, &connected, &state,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoArbitratorValidityDetailArbitratorDesc, prometheus.GaugeValue, float64(arbitrator),
			nodeID, arbTicket, connected, state,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoArbValDet{}
