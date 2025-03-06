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

// Scrape `ndbinfo.pgman_time_track_stats`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoPGManTimeTrackStatsQuery = `
	SELECT
		node_id,
		block_number,
		block_instance,
		upper_bound,
		page_reads,
		page_writes,
		log_waits,
		get_page
	FROM ndbinfo.pgman_time_track_stats;
`

// Metric descriptors.
var (
	ndbinfoPGManTimeTrackStatsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "pgman_time_track_stats_count"),
		"Based on duration of action latency by node_id/block_number/block_instance/upper_bound/action.",
		[]string{"node_id", "block_number", "block_instance", "upper_bound", "action"}, nil,
	)
)

// ScrapeNDBInfoPGManTimeTrackStats collects from `ndbinfo.pgman_time_track_stats`.
type ScrapeNDBInfoPGManTimeTrackStats struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoPGManTimeTrackStats) Name() string {
	return NDBInfo + ".pgman_time_track_stats"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoPGManTimeTrackStats) Help() string {
	return "Collect metrics from ndbinfo.pgman_time_track_stats"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoPGManTimeTrackStats) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoPGManTimeTrackStats) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoPGManTimeTrackStatsRows, err := db.QueryContext(ctx, ndbinfoPGManTimeTrackStatsQuery)
	if err != nil {
		return err
	}
	defer ndbinfoPGManTimeTrackStatsRows.Close()
	var (
		nodeID, blockNumber, blockInstance, upperBound string
		pageReads, pageWrites, logWaits, getPage       uint64
	)

	for ndbinfoPGManTimeTrackStatsRows.Next() {
		if err := ndbinfoPGManTimeTrackStatsRows.Scan(
			&nodeID, &blockNumber, &blockInstance, &upperBound,
			&pageReads, &pageWrites, &logWaits, &getPage,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoPGManTimeTrackStatsDesc, prometheus.CounterValue, float64(pageReads),
			nodeID, blockNumber, blockInstance, upperBound, "page_reads",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoPGManTimeTrackStatsDesc, prometheus.CounterValue, float64(pageWrites),
			nodeID, blockNumber, blockInstance, upperBound, "page_writes",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoPGManTimeTrackStatsDesc, prometheus.CounterValue, float64(logWaits),
			nodeID, blockNumber, blockInstance, upperBound, "log_waits",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoPGManTimeTrackStatsDesc, prometheus.CounterValue, float64(getPage),
			nodeID, blockNumber, blockInstance, upperBound, "get_page",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoPGManTimeTrackStats{}
