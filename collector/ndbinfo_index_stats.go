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

// Scrape `ndbinfo.index_stats`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoIndexStatsQuery = `
	SELECT
		index_id,
		index_version,
		sample_version,
		load_time,
		sample_count
	FROM ndbinfo.index_stats;
`

// Metric descriptors.
var (
	ndbinfoIndexStatsVersionDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "index_stats_version"),
		"The current version of index by index_id/type.",
		[]string{"index_id", "type"}, nil,
	)
	ndbinfoIndexStatsLoadTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "index_stats_load_time"),
		"The UNIX timestamp of last loaded statistics by index_id.",
		[]string{"index_id"}, nil,
	)
	ndbinfoIndexStatsSampleCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "index_stats_sample_count"),
		"The amount of samples collected for the index statistics by index_id.",
		[]string{"index_id"}, nil,
	)
)

// ScrapeNDBInfoIndexStats collects from `ndbinfo.index_stats`.
type ScrapeNDBInfoIndexStats struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoIndexStats) Name() string {
	return NDBInfo + ".index_stats"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoIndexStats) Help() string {
	return "Collect metrics from ndbinfo.index_stats"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoIndexStats) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoIndexStats) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoIndexStatsRows, err := db.QueryContext(ctx, ndbinfoIndexStatsQuery)
	if err != nil {
		return err
	}
	defer ndbinfoIndexStatsRows.Close()

	var (
		indexId                                            string
		indexVersion, sampleVersion, loadTime, sampleCount uint64
	)
	for ndbinfoIndexStatsRows.Next() {
		if err := ndbinfoIndexStatsRows.Scan(
			&indexId,
			&indexVersion, &sampleVersion, &loadTime, &sampleCount,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoIndexStatsVersionDesc, prometheus.GaugeValue, float64(indexVersion),
			indexId, "index",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoIndexStatsVersionDesc, prometheus.GaugeValue, float64(sampleVersion),
			indexId, "sample",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoIndexStatsLoadTimeDesc, prometheus.GaugeValue, float64(loadTime),
			indexId,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoIndexStatsSampleCountDesc, prometheus.GaugeValue, float64(sampleCount),
			indexId,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoIndexStats{}
