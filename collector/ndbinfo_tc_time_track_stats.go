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

// Scrape `ndbinfo.tc_time_track_stats`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoTCTimeTrackStatsQuery = `
	SELECT
		node_id,
		comm_node_id,
		upper_bound,
		scans,
		scan_errors,
		scan_fragments,
		scan_fragment_errors,
		transactions,
		transaction_errors,
		read_key_ops,
		write_key_ops,
		index_key_ops,
		key_op_errors
	FROM ndbinfo.tc_time_track_stats;
`

// Metric descriptors.
var (
	ndbinfoTCTimeTrackStatsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "tc_time_track_stats_count"),
		"The amount of actions by node_id/comm_node_id/upper_bound/action.",
		[]string{"node_id", "comm_node_id", "upper_bound", "action"}, nil,
	)
)

// ScrapeNDBInfoTCTimeTrackStats collects from `ndbinfo.tc_time_track_stats`.
type ScrapeNDBInfoTCTimeTrackStats struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoTCTimeTrackStats) Name() string {
	return NDBInfo + ".tc_time_track_stats"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoTCTimeTrackStats) Help() string {
	return "Collect metrics from ndbinfo.tc_time_track_stats"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoTCTimeTrackStats) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoTCTimeTrackStats) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoTCTimeTrackStatsRows, err := db.QueryContext(ctx, ndbinfoTCTimeTrackStatsQuery)
	if err != nil {
		return err
	}
	defer ndbinfoTCTimeTrackStatsRows.Close()

	var (
		nodeID, commNodeID, upperBound                       string
		scans, scanErrors, scanFragments, scanFragmentErrors uint64
		transactions, transactionErrors                      uint64
		readKeyOps, writeKeyOps, indexKeyOps, keyOpErrors    uint64
	)

	for ndbinfoTCTimeTrackStatsRows.Next() {
		if err := ndbinfoTCTimeTrackStatsRows.Scan(
			&nodeID, &commNodeID, &upperBound,
			&scans, &scanErrors, &scanFragments, &scanFragmentErrors,
			&transactions, &transactionErrors,
			&readKeyOps, &writeKeyOps, &indexKeyOps, &keyOpErrors,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTCTimeTrackStatsDesc, prometheus.CounterValue, float64(scans),
			nodeID, commNodeID, upperBound, "scans",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTCTimeTrackStatsDesc, prometheus.CounterValue, float64(scanErrors),
			nodeID, commNodeID, upperBound, "scan_errors",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTCTimeTrackStatsDesc, prometheus.CounterValue, float64(scanFragments),
			nodeID, commNodeID, upperBound, "scan_fragments",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTCTimeTrackStatsDesc, prometheus.CounterValue, float64(scanFragmentErrors),
			nodeID, commNodeID, upperBound, "scan_fragment_errors",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTCTimeTrackStatsDesc, prometheus.CounterValue, float64(transactions),
			nodeID, commNodeID, upperBound, "transactions",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTCTimeTrackStatsDesc, prometheus.CounterValue, float64(transactionErrors),
			nodeID, commNodeID, upperBound, "transaction_errors",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTCTimeTrackStatsDesc, prometheus.CounterValue, float64(readKeyOps),
			nodeID, commNodeID, upperBound, "read_key_ops",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTCTimeTrackStatsDesc, prometheus.CounterValue, float64(writeKeyOps),
			nodeID, commNodeID, upperBound, "write_key_ops",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTCTimeTrackStatsDesc, prometheus.CounterValue, float64(indexKeyOps),
			nodeID, commNodeID, upperBound, "index_key_ops",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTCTimeTrackStatsDesc, prometheus.CounterValue, float64(keyOpErrors),
			nodeID, commNodeID, upperBound, "key_op_errors",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoTCTimeTrackStats{}
