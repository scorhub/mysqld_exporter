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

// Scrape `ndbinfo.operations_per_fragment`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoOperationsPerFragmentQuery = `
	SELECT
		table_id,
		node_id,
		fragment_num,
		tot_key_reads,
		tot_key_inserts,
		tot_key_updates,
		tot_key_writes,
		tot_key_deletes,
		tot_key_refs,
		tot_key_attrinfo_bytes,
		tot_key_keyinfo_bytes,
		tot_key_prog_bytes,
		tot_key_inst_exec,
		tot_key_bytes_returned,
		tot_frag_scans,
		tot_scan_rows_examined,
		tot_scan_rows_returned,
		tot_scan_bytes_returned,
		tot_scan_prog_bytes,
		tot_scan_bound_bytes,
		tot_scan_inst_exec,
		tot_qd_frag_scans,
		conc_frag_scans,
		conc_qd_frag_scans,
		tot_commits
	FROM ndbinfo.operations_per_fragment;
`

// Metric descriptors.
var (
	ndbinfoOperationsPerFragmentKeyTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "operations_per_fragment_key_total"),
		"The total number of key actions by table_id/node_id/fragment_num/action.",
		[]string{"table_id", "node_id", "fragment_num", "action"}, nil,
	)
	ndbinfoOperationsPerFragmentKeyBytesTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "operations_per_fragment_key_bytes_total"),
		"The total size of all action by table_id/node_id/fragment_num/action.",
		[]string{"table_id", "node_id", "fragment_num", "action"}, nil,
	)
	ndbinfoOperationsPerFragmentScansTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "operations_per_fragment_scans_total"),
		"The total number of scans performed action by table_id/node_id/fragment_num/action.",
		[]string{"table_id", "node_id", "fragment_num", "action"}, nil,
	)
	ndbinfoOperationsPerFragmentScanBytesTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "operations_per_fragment_scan_bytes_total"),
		"The total size of data used in scan actions by table_id/node_id/fragment_num/action.",
		[]string{"table_id", "node_id", "fragment_num", "action"}, nil,
	)
	ndbinfoOperationsPerFragmentCurrentScansDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "operations_per_fragment_current_scans"),
		"The number of non-queued/queued scans currently active by table_id/node_id/fragment_num/action.",
		[]string{"table_id", "node_id", "fragment_num", "action"}, nil,
	)
	ndbinfoOperationsPerFragmentCommitsTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "operations_per_fragment_commits_total"),
		"The total number of row changes committed by table_id/node_id/fragment_num.",
		[]string{"table_id", "node_id", "fragment_num"}, nil,
	)
)

// ScrapeNDBInfoOperationsPerFragment collects from `ndbinfo.operations_per_fragment`.
type ScrapeNDBInfoOperationsPerFragment struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoOperationsPerFragment) Name() string {
	return NDBInfo + ".operations_per_fragment"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoOperationsPerFragment) Help() string {
	return "Collect metrics from ndbinfo.operations_per_fragment"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoOperationsPerFragment) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoOperationsPerFragment) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoOperationsPerFragmentRows, err := db.QueryContext(ctx, ndbinfoOperationsPerFragmentQuery)
	if err != nil {
		return err
	}
	defer ndbinfoOperationsPerFragmentRows.Close()
	var (
		tableID, nodeID, fragmentNum                                                       string
		totKeyReads, totKeyInserts, totKeyUpdates, totKeyWrites, totKeyDeletes, totKeyRefs uint64
		totKeyAttrInfoBytes, totKeyKeyInfoBytes, totKeyProgBytes                           uint64
		totKeyInstExec, totKeyBytesReturned, totFragScans, totScanRowsExamined             uint64
		totScanRowsReturned, totScanBytesReturned, totScanProgBytes, totScanBoundBytes     uint64
		totScanInstExec, totQdFragScans, concFragScans, concQdFragScans, totCommits        uint64
	)

	for ndbinfoOperationsPerFragmentRows.Next() {
		if err := ndbinfoOperationsPerFragmentRows.Scan(
			&tableID, &nodeID, &fragmentNum,
			&totKeyReads, &totKeyInserts, &totKeyUpdates, &totKeyWrites, &totKeyDeletes, &totKeyRefs,
			&totKeyAttrInfoBytes, &totKeyKeyInfoBytes, &totKeyProgBytes,
			&totKeyInstExec, &totKeyBytesReturned, &totFragScans, &totScanRowsExamined,
			&totScanRowsReturned, &totScanBytesReturned, &totScanProgBytes, &totScanBoundBytes,
			&totScanInstExec, &totQdFragScans, &concFragScans, &concQdFragScans, &totCommits,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentKeyTotalDesc, prometheus.CounterValue, float64(totKeyReads),
			tableID, nodeID, fragmentNum, "read",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentKeyTotalDesc, prometheus.CounterValue, float64(totKeyInserts),
			tableID, nodeID, fragmentNum, "insert",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentKeyTotalDesc, prometheus.CounterValue, float64(totKeyUpdates),
			tableID, nodeID, fragmentNum, "update",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentKeyTotalDesc, prometheus.CounterValue, float64(totKeyWrites),
			tableID, nodeID, fragmentNum, "write",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentKeyTotalDesc, prometheus.CounterValue, float64(totKeyDeletes),
			tableID, nodeID, fragmentNum, "delete",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentKeyTotalDesc, prometheus.CounterValue, float64(totKeyRefs),
			tableID, nodeID, fragmentNum, "refs",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentKeyTotalDesc, prometheus.CounterValue, float64(totKeyInstExec),
			tableID, nodeID, fragmentNum, "inst_exec",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentKeyBytesTotalDesc, prometheus.CounterValue, float64(totKeyBytesReturned),
			tableID, nodeID, fragmentNum, "returned",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentKeyBytesTotalDesc, prometheus.CounterValue, float64(totKeyAttrInfoBytes),
			tableID, nodeID, fragmentNum, "attrinfo",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentKeyBytesTotalDesc, prometheus.CounterValue, float64(totKeyKeyInfoBytes),
			tableID, nodeID, fragmentNum, "keyinfo",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentKeyBytesTotalDesc, prometheus.CounterValue, float64(totKeyProgBytes),
			tableID, nodeID, fragmentNum, "prog",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentScansTotalDesc, prometheus.CounterValue, float64(totFragScans),
			tableID, nodeID, fragmentNum, "total",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentScansTotalDesc, prometheus.CounterValue, float64(totScanRowsExamined),
			tableID, nodeID, fragmentNum, "rows_examined",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentScansTotalDesc, prometheus.CounterValue, float64(totScanRowsReturned),
			tableID, nodeID, fragmentNum, "rows_returned",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentScansTotalDesc, prometheus.CounterValue, float64(totScanInstExec),
			tableID, nodeID, fragmentNum, "inst_exec",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentScansTotalDesc, prometheus.CounterValue, float64(totQdFragScans),
			tableID, nodeID, fragmentNum, "queued",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentCurrentScansDesc, prometheus.GaugeValue, float64(concFragScans),
			tableID, nodeID, fragmentNum, "conc",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentCurrentScansDesc, prometheus.GaugeValue, float64(concQdFragScans),
			tableID, nodeID, fragmentNum, "conc_qd",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentScanBytesTotalDesc, prometheus.CounterValue, float64(totScanBytesReturned),
			tableID, nodeID, fragmentNum, "total",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentScanBytesTotalDesc, prometheus.CounterValue, float64(totScanProgBytes),
			tableID, nodeID, fragmentNum, "prog",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentScanBytesTotalDesc, prometheus.CounterValue, float64(totScanBoundBytes),
			tableID, nodeID, fragmentNum, "bound",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoOperationsPerFragmentCommitsTotalDesc, prometheus.CounterValue, float64(totCommits),
			tableID, nodeID, fragmentNum,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoOperationsPerFragment{}
