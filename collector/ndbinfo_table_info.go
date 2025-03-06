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

// Scrape `ndbinfo.table_info`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoTableInfoQuery = `
	SELECT
		table_id,
		logged_table,
		row_contains_gci,
		row_contains_checksum,
		read_backup,
		fully_replicated,
		storage_type,
		hashmap_id,
		partition_balance,
		create_gci
	FROM ndbinfo.table_info;
`

// Metric descriptors.
var (
	ndbinfoTableInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "table_info"),
		"Whether table is contains info (1) or not (0) by table_id/storage_type/hashmap_id/partition_balance/info.",
		[]string{"table_id", "storage_type", "hashmap_id", "partition_balance", "info"}, nil,
	)
	ndbinfoTableInfoCreateGCIDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "table_info_create_gci"),
		"GCI from which table was created by table_id/storage_type/hashmap_id/partition_balance.",
		[]string{"table_id", "storage_type", "hashmap_id", "partition_balance"}, nil,
	)
)

// ScrapeNDBInfoTableInfo collects from `ndbinfo.table_info`.
type ScrapeNDBInfoTableInfo struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoTableInfo) Name() string {
	return NDBInfo + ".table_info"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoTableInfo) Help() string {
	return "Collect metrics from ndbinfo.table_info"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoTableInfo) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoTableInfo) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoTableInfoRows, err := db.QueryContext(ctx, ndbinfoTableInfoQuery)
	if err != nil {
		return err
	}
	defer ndbinfoTableInfoRows.Close()

	var (
		tableID                                          string
		loggedTable, rowContainsGCI, rowContainsChecksum uint64
		readBackup, fullyReplicated                      uint64
		storageType, hashmapId, partitionBalance         string
		createGCI                                        uint64
	)

	for ndbinfoTableInfoRows.Next() {
		if err := ndbinfoTableInfoRows.Scan(
			&tableID,
			&loggedTable, &rowContainsGCI, &rowContainsChecksum, &readBackup, &fullyReplicated,
			&storageType, &hashmapId, &partitionBalance,
			&createGCI,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableInfoDesc, prometheus.GaugeValue, float64(loggedTable),
			tableID, storageType, hashmapId, partitionBalance, "logged_table",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableInfoDesc, prometheus.GaugeValue, float64(rowContainsGCI),
			tableID, storageType, hashmapId, partitionBalance, "row_contains_gci",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableInfoDesc, prometheus.GaugeValue, float64(rowContainsChecksum),
			tableID, storageType, hashmapId, partitionBalance, "row_contains_checksum",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableInfoDesc, prometheus.GaugeValue, float64(readBackup),
			tableID, storageType, hashmapId, partitionBalance, "read_backup",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableInfoDesc, prometheus.GaugeValue, float64(fullyReplicated),
			tableID, storageType, hashmapId, partitionBalance, "fully_replicated",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableInfoCreateGCIDesc, prometheus.GaugeValue, float64(createGCI),
			tableID, storageType, hashmapId, partitionBalance,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoTableInfo{}
