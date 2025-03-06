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

// Scrape `ndbinfo.logspaces`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoLogSpacesQuery = `
	SELECT
		node_id,
		log_type,
		log_id,
		log_part,
		total,
		used
	FROM ndbinfo.logspaces;
`

// Metric descriptors.
var (
	ndbinfoLogSpacesTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "logspaces_bytes_available"),
		"The amount of total space available for this log in bytes by node_id/log_type/log_id/log_part.",
		[]string{"node_id", "log_type", "log_id", "log_part"}, nil,
	)
	ndbinfoLogSpacesUsedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "logspaces_bytes_used"),
		"The amount of space used by this log in bytes by node_id/log_type/log_id/log_part.",
		[]string{"node_id", "log_type", "log_id", "log_part"}, nil,
	)
)

// ScrapeNDBInfoLogSpaces collects from `ndbinfo.logspaces`.
type ScrapeNDBInfoLogSpaces struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoLogSpaces) Name() string {
	return NDBInfo + ".logspaces"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoLogSpaces) Help() string {
	return "Collect metrics from ndbinfo.logspaces"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoLogSpaces) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoLogSpaces) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoLogSpacesRows, err := db.QueryContext(ctx, ndbinfoLogSpacesQuery)
	if err != nil {
		return err
	}
	defer ndbinfoLogSpacesRows.Close()
	var (
		nodeID, logType, logID, logPart string
		total, used                     uint64
	)

	for ndbinfoLogSpacesRows.Next() {
		if err := ndbinfoLogSpacesRows.Scan(
			&nodeID, &logType, &logID, &logPart,
			&total, &used,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLogSpacesTotalDesc, prometheus.GaugeValue, float64(total),
			nodeID, logType, logID, logPart,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLogSpacesUsedDesc, prometheus.GaugeValue, float64(used),
			nodeID, logType, logID, logPart,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoLogSpaces{}
