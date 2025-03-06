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

// Scrape `ndbinfo.counters`.

package collector

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoCountersQuery = `
	SELECT
		node_id,
		counter_name,
		val
	FROM ndbinfo.counters;
`

// Metric descriptors.
var (
	// Map known counter names to types. Unknown types will be mapped as untyped.
	ndbinfoCountersTypes = map[string]struct {
		vtype prometheus.ValueType
		desc  *prometheus.Desc
	}{
		"ATTRINFO": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_attrinfo_total"),
				"The number of times an interpreted program is sent to the data node by node_id.",
				[]string{"node_id"}, nil)},
		"TRANSACTIONS": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_transactions_total"),
				"The total number of transactions initiated by node_id.",
				[]string{"node_id"}, nil)},
		"COMMITS": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_commits_total"),
				"The number of transactions that have been committed by node_id.",
				[]string{"node_id"}, nil)},
		"READS": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_reads_total"),
				"The amount of all read operations by node_id.",
				[]string{"node_id"}, nil)},
		"SIMPLE_READS": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_simple_reads_total"),
				"The amount of reads where the is both the beginning and ending operation for a given transaction by node_id.",
				[]string{"node_id"}, nil)},
		"WRITES": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_writes_total"),
				"The amount of the writes by node_id.",
				[]string{"node_id"}, nil)},
		"ABORTS": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_aborts_total"),
				"The amount of transactions that have been aborted by node_id.",
				[]string{"node_id"}, nil)},
		"TABLE_SCANS": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_table_scans_total"),
				"The amount of table scan operations performed by node_id.",
				[]string{"node_id"}, nil)},
		"RANGE_SCANS": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_range_scans_total"),
				"The amount of range scan operations performed by node_id.",
				[]string{"node_id"}, nil)},
		"OPERATIONS": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_operations_total"),
				"The amount of operations processed by node_id.",
				[]string{"node_id"}, nil)},
		"READS_RECEIVED": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_reads_received_total"),
				"The amount of read requests received by node_id.",
				[]string{"node_id"}, nil)},
		"LOCAL_READS_SENT": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_local_reads_sent_total"),
				"The amount of local read requests sent by node_id.",
				[]string{"node_id"}, nil)},
		"REMOTE_READS_SENT": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_remote_reads_sent_total"),
				"The amount of remote read requests sent by node_id.",
				[]string{"node_id"}, nil)},
		"READS_NOT_FOUND": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_reads_not_found_total"),
				"The amount of read requests that did not find a matching record by node_id.",
				[]string{"node_id"}, nil)},
		"TABLE_SCANS_RECEIVED": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_table_scans_received_total"),
				"The amount of table scan requests received by node_id.",
				[]string{"node_id"}, nil)},
		"LOCAL_TABLE_SCANS_SENT": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_local_table_scans_sent_total"),
				"The amount of local table scan requests sent by node_id.",
				[]string{"node_id"}, nil)},
		"RANGE_SCANS_RECEIVED": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_range_scans_received_total"),
				"The amount of range scan requests received by node_id.",
				[]string{"node_id"}, nil)},
		"LOCAL_RANGE_SCANS_SENT": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_local_range_scans_sent_total"),
				"The amount of local range scan requests sent by node_id.",
				[]string{"node_id"}, nil)},
		"REMOTE_RANGE_SCANS_SENT": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_remote_range_scans_sent_total"),
				"The amount of remote range scan requests sent by node_id.",
				[]string{"node_id"}, nil)},
		"SCAN_BATCHES_RETURNED": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_scan_batches_returned_total"),
				"The amount of scan batches returned by node_id.",
				[]string{"node_id"}, nil)},
		"SCAN_ROWS_RETURNED": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_scan_rows_returned_total"),
				"The amount of rows returned by scan operations by node_id.",
				[]string{"node_id"}, nil)},
		"PRUNED_RANGE_SCANS_RECEIVED": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_pruned_range_scans_received_total"),
				"The amount of pruned range scan requests received by node_id.",
				[]string{"node_id"}, nil)},
		"CONST_PRUNED_RANGE_SCANS_RECEIVED": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_const_pruned_range_scans_received_total"),
				"The amount of constant pruned range scan requests received by node_id.",
				[]string{"node_id"}, nil)},
		"LOCAL_READS": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_local_reads_total"),
				"The amount of only those reads of the primary fragment replica on the same node as the transaction coordinator by node_id.",
				[]string{"node_id"}, nil)},
		"LOCAL_WRITES": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_local_writes_total"),
				"The amount of primary-key write operations using a transaction coordinator in a node that also holds the primary fragment replica of the record by node_id.",
				[]string{"node_id"}, nil)},
		"LQHKEY_OVERLOAD": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_lqhkey_overload_total"),
				"The amount of primary key requests rejected at the LQH block instance due to transporter overload by node_id.",
				[]string{"node_id"}, nil)},
		"LQHKEY_OVERLOAD_TC": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_lqhkey_overload_tc_total"),
				"The amount of instances where the TC node transporter was overloaded by node_id.",
				[]string{"node_id"}, nil)},
		"LQHKEY_OVERLOAD_READER": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_lqhkey_overload_reader_total"),
				"The amount of instances where the API reader (reads only) node was overloaded by node_id.",
				[]string{"node_id"}, nil)},
		"LQHKEY_OVERLOAD_NODE_PEER": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_lqhkey_overload_node_peer_total"),
				"The amount of instances where the next backup data node (writes only) was overloaded by node_id.",
				[]string{"node_id"}, nil)},
		"LQHKEY_OVERLOAD_SUBSCRIBER": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_lqhkey_overload_subscriber_total"),
				"The amount of instances where an event subscriber (writes only) was overloaded by node_id.",
				[]string{"node_id"}, nil)},
		"LQHSCAN_SLOWDOWNS": {prometheus.CounterValue,
			prometheus.NewDesc(prometheus.BuildFQName(namespace, NDBInfo, "counters_lqhkey_slowdowns_total"),
				"The amount of instances where a fragment scan batch size was reduced due to scanning API transporter overload by node_id.",
				[]string{"node_id"}, nil)},
	}
)

// ScrapeNDBInfoCounters collects from `ndbinfo.counters`.
type ScrapeNDBInfoCounters struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoCounters) Name() string {
	return NDBInfo + ".counters"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoCounters) Help() string {
	return "Collect metrics from ndbinfo.counters"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoCounters) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoCounters) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoCountersRows, err := db.QueryContext(ctx, ndbinfoCountersQuery)
	if err != nil {
		return err
	}
	defer ndbinfoCountersRows.Close()

	var (
		nodeID, counterName string
		val                 uint64
	)

	for ndbinfoCountersRows.Next() {
		if err := ndbinfoCountersRows.Scan(
			&nodeID, &counterName,
			&val,
		); err != nil {
			return err
		}

		if metricType, ok := ndbinfoCountersTypes[counterName]; ok {
			ch <- prometheus.MustNewConstMetric(
				metricType.desc, metricType.vtype, float64(val),
				nodeID,
			)
			continue
		}

		// Unknown metric. Report as untyped.
		desc := prometheus.NewDesc(prometheus.BuildFQName(namespace, informationSchema, fmt.Sprintf("counter_%s", strings.ToLower(counterName))),
			fmt.Sprintf("Unsupported metric from column %s", counterName),
			[]string{"node_id"}, nil)
		ch <- prometheus.MustNewConstMetric(
			desc, prometheus.UntypedValue, float64(val),
			nodeID,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoCounters{}
