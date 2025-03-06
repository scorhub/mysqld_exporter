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

// Scrape `ndbinfo.table_distribution_status`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoTableDistributionStatusQuery = `
	SELECT
		node_id,
		table_id,
		tab_copy_status,
		tab_update_status,
		tab_lcp_status,
		tab_status,
		tab_storage,
		tab_partitions,
		tab_fragments,
		current_scan_count,
		scan_count_wait,
		is_reorg_ongoing
	FROM ndbinfo.table_distribution_status;
`

// Metric descriptors.
var (
	ndbinfoTableDistributionStatusCurrentScanCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "table_distribution_status_current_scan_count"),
		"The current number of active scans by node_id/table_id/tab_copy_status/tab_update_status/tab_lcp_status/tab_status/tab_storage/tab_partitions/tab_fragments",
		[]string{"node_id", "table_id", "tab_copy_status", "tab_update_status", "tab_lcp_status", "tab_status", "tab_storage", "tab_partitions", "tab_fragments"}, nil,
	)
	ndbinfoTableDistributionStatusScanCountWaitDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "table_distribution_status_scan_count_wait"),
		"The current number of scans waiting to be performed before ALTER TABLE can complete by node_id/table_id/tab_copy_status/tab_update_status/tab_lcp_status/tab_status/tab_storage/tab_partitions/tab_fragments",
		[]string{"node_id", "table_id", "tab_copy_status", "tab_update_status", "tab_lcp_status", "tab_status", "tab_storage", "tab_partitions", "tab_fragments"}, nil,
	)
	ndbinfoTableDistributionStatusIsReorgOngoingDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "table_distribution_status_is_reorg_ongoing"),
		"Whether the table is currently being reorganized (1 if true) by node_id/table_id/tab_copy_status/tab_update_status/tab_lcp_status/tab_status/tab_storage/tab_partitions/tab_fragments",
		[]string{"node_id", "table_id", "tab_copy_status", "tab_update_status", "tab_lcp_status", "tab_status", "tab_storage", "tab_partitions", "tab_fragments"}, nil,
	)
)

// ScrapeNDBInfoTableDistributionStatus collects from `ndbinfo.table_distribution_status`.
type ScrapeNDBInfoTableDistributionStatus struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoTableDistributionStatus) Name() string {
	return NDBInfo + ".table_distribution_status"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoTableDistributionStatus) Help() string {
	return "Collect metrics from ndbinfo.table_distribution_status"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoTableDistributionStatus) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoTableDistributionStatus) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoTableDistributionStatusRows, err := db.QueryContext(ctx, ndbinfoTableDistributionStatusQuery)
	if err != nil {
		return err
	}
	defer ndbinfoTableDistributionStatusRows.Close()

	var (
		nodeID, tableID, tabCopyStatus                  string
		tabUpdateStatus, tabLCPStatus, tabStatus        string
		tabStorage, tabPartitions, tabFragments         string
		currentScanCount, scanCountWait, isReorgOngoing uint64
	)

	for ndbinfoTableDistributionStatusRows.Next() {
		if err := ndbinfoTableDistributionStatusRows.Scan(
			&nodeID, &tableID, &tabCopyStatus,
			&tabUpdateStatus, &tabLCPStatus, &tabStatus,
			&tabStorage, &tabPartitions, &tabFragments,
			&currentScanCount, &scanCountWait, &isReorgOngoing,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableDistributionStatusCurrentScanCountDesc, prometheus.CounterValue, float64(currentScanCount),
			nodeID, tableID, tabCopyStatus,
			tabUpdateStatus, tabLCPStatus, tabStatus,
			tabStorage, tabPartitions, tabFragments,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableDistributionStatusScanCountWaitDesc, prometheus.CounterValue, float64(currentScanCount),
			nodeID, tableID, tabCopyStatus,
			tabUpdateStatus, tabLCPStatus, tabStatus,
			tabStorage, tabPartitions, tabFragments,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableDistributionStatusIsReorgOngoingDesc, prometheus.CounterValue, float64(scanCountWait),
			nodeID, tableID, tabCopyStatus,
			tabUpdateStatus, tabLCPStatus, tabStatus,
			tabStorage, tabPartitions, tabFragments,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoTableDistributionStatus{}
