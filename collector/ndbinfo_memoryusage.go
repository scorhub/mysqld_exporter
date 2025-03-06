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

// Scrape `ndbinfo.memoryusage`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoMemoryUsageQuery = `
	SELECT
		node_id,
		memory_type,
		used,
		used_pages,
		total,
		total_pages
	FROM ndbinfo.memoryusage;
`

// Metric descriptors.
var (
	ndbinfoMemoryUsageUsedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "memoryusage_used_bytes"),
		"The amount of bytes currently used for data memory or index memory by node_id/memory_type.",
		[]string{"node_id", "memory_type"}, nil,
	)
	ndbinfoMemoryUsageUsedPagesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "memoryusage_used_pages"),
		"The amount of pages currently used for data memory or index memory by node_id/memory_type.",
		[]string{"node_id", "memory_type"}, nil,
	)
	ndbinfoMemoryUsageTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "memoryusage_available_bytes"),
		"Total number of bytes of data memory or index memory available by node_id/memory_type.",
		[]string{"node_id", "memory_type"}, nil,
	)
	ndbinfoMemoryUsageTotalPagesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "memoryusage_available_pages"),
		"The amount of memory pages available for data memory or index memory by node_id/memory_type.",
		[]string{"node_id", "memory_type"}, nil,
	)
)

// ScrapeNDBInfoMemoryUsage collects from `ndbinfo.memoryusage`.
type ScrapeNDBInfoMemoryUsage struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoMemoryUsage) Name() string {
	return NDBInfo + ".memoryusage"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoMemoryUsage) Help() string {
	return "Collect metrics from ndbinfo.memoryusage"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoMemoryUsage) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoMemoryUsage) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoMemoryUsageRows, err := db.QueryContext(ctx, ndbinfoMemoryUsageQuery)
	if err != nil {
		return err
	}
	defer ndbinfoMemoryUsageRows.Close()
	var (
		nodeID, memoryType                 string
		used, usedPages, total, totalPages uint64
	)

	for ndbinfoMemoryUsageRows.Next() {
		if err := ndbinfoMemoryUsageRows.Scan(
			&nodeID, &memoryType,
			&used, &usedPages, &total, &totalPages,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoMemoryUsageUsedDesc, prometheus.GaugeValue, float64(used),
			nodeID, memoryType,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoMemoryUsageUsedPagesDesc, prometheus.GaugeValue, float64(usedPages),
			nodeID, memoryType,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoMemoryUsageTotalDesc, prometheus.GaugeValue, float64(total),
			nodeID, memoryType,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoMemoryUsageTotalPagesDesc, prometheus.GaugeValue, float64(totalPages),
			nodeID, memoryType,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoMemoryUsage{}
