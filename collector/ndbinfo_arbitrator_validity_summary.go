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

// Scrape `ndbinfo.arbitrator_validity_summary`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoArbValSumQuery = `
	SELECT
		arbitrator,
		arb_ticket,
		arb_connected,
		consensus_count
	FROM ndbinfo.arbitrator_validity_summary;
`

// Metric descriptors.
var (
	ndbinfoArbitratorValiditySummaryStateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "arbitrator_validity_summary_consensus_count"),
		"The number of data nodes that see this node as arbitrator by arbitrator/arb_ticket/arb_connected",
		[]string{"arbitrator", "arb_ticket", "arb_connected"}, nil,
	)
)

// ScrapeNDBInfoArbSumDet collects from `ndbinfo.arbitrator_validity_summary`.
type ScrapeNDBInfoArbSumDet struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoArbSumDet) Name() string {
	return NDBInfo + ".arbitrator_validity_summary"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoArbSumDet) Help() string {
	return "Collect metrics from ndbinfo.arbitrator_validity_summary"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoArbSumDet) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoArbSumDet) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoArbValSumRows, err := db.QueryContext(ctx, ndbinfoArbValSumQuery)
	if err != nil {
		return err
	}
	defer ndbinfoArbValSumRows.Close()

	var (
		arbitrator, arbTicket, connected string
		consensusCount                   uint64
	)
	for ndbinfoArbValSumRows.Next() {
		if err := ndbinfoArbValSumRows.Scan(
			&arbitrator, &arbTicket, &connected,
			&consensusCount,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoArbitratorValiditySummaryStateDesc, prometheus.GaugeValue, float64(consensusCount),
			arbitrator, arbTicket, connected,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoArbSumDet{}
