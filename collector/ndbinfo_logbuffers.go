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

// Scrape `ndbinfo.logbuffers`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoLogBuffersQuery = `
	SELECT
		node_id,
		log_type,
		log_id,
		log_part,
		total,
		used
	FROM ndbinfo.logbuffers;
`

// Metric descriptors.
var (
	ndbinfoLogBuffersTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "logbuffers_bytes_available"),
		"The amount of total space available for this log in bytes by node_id/log_type/log_id/log_part.",
		[]string{"node_id", "log_type", "log_id", "log_part"}, nil,
	)
	ndbinfoLogBuffersUsedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "logbuffers_bytes_used"),
		"The amount of space used by this log in bytes by node_id/log_type/log_id/log_part.",
		[]string{"node_id", "log_type", "log_id", "log_part"}, nil,
	)
)

// ScrapeNDBInfoLogBuffers collects from `ndbinfo.logbuffers`.
type ScrapeNDBInfoLogBuffers struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoLogBuffers) Name() string {
	return NDBInfo + ".logbuffers"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoLogBuffers) Help() string {
	return "Collect metrics from ndbinfo.logbuffers"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoLogBuffers) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoLogBuffers) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoLogBuffersRows, err := db.QueryContext(ctx, ndbinfoLogBuffersQuery)
	if err != nil {
		return err
	}
	defer ndbinfoLogBuffersRows.Close()
	var (
		nodeID, logType, logID, logPart string
		total, used                     uint64
	)

	for ndbinfoLogBuffersRows.Next() {
		if err := ndbinfoLogBuffersRows.Scan(
			&nodeID, &logType, &logID, &logPart,
			&total, &used,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLogBuffersTotalDesc, prometheus.GaugeValue, float64(total),
			nodeID, logType, logID, logPart,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLogBuffersUsedDesc, prometheus.GaugeValue, float64(used),
			nodeID, logType, logID, logPart,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoLogBuffers{}
