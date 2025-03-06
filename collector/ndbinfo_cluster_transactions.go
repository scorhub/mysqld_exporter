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

// Scrape `ndbinfo.cluster_transactions`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoClusterTransactionsQuery = `
	SELECT
		node_id,
		state,
		SUM(count_operations) AS total_count_operations,
		SUM(outstanding_operations) AS total_outstanding_operations,
		MAX(inactive_seconds) AS max_inactive_seconds,
		client_node_id,
		COUNT(*) AS total_transactions
	FROM ndbinfo.cluster_transactions
	GROUP BY node_id, state, client_node_id;
`

// Metric descriptors.
var (
	ndbinfoClusterTransactionsTotalCountOperationsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "cluster_transactions_total_count_operations"),
		"The amount of stateful primary key operations in transaction (includes reads with locks, as well as DML operations) by node_id/state/client_node_id.",
		[]string{"node_id", "state", "client_node_id"}, nil,
	)
	ndbinfoClusterTransactionsTotalOutstandingOperationsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "cluster_transactions_total_outstanding_operations"),
		"The amount of operations still being executed in local data management blocks by node_id/state/client_node_id.",
		[]string{"node_id", "state", "client_node_id"}, nil,
	)
	ndbinfoClusterTransactionsMaxInactivityDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "cluster_transactions_max_inactive_seconds"),
		"The maximum inactivity seconds by node_id/state/client_node_id.",
		[]string{"node_id", "state", "client_node_id"}, nil,
	)
	ndbinfoClusterTransactionsTotalTransactionDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "cluster_transactions_total_transactions"),
		"The total amount of transactions by node_id/state/client_node_id.",
		[]string{"node_id", "state", "client_node_id"}, nil,
	)
)

// ScrapeNDBInfoClusterTransactions collects from `ndbinfo.cluster_transactions`.
type ScrapeNDBInfoClusterTransactions struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoClusterTransactions) Name() string {
	return NDBInfo + ".cluster_transactions"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoClusterTransactions) Help() string {
	return "Collect metrics from ndbinfo.cluster_transactions"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoClusterTransactions) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoClusterTransactions) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoClusterTransactionsRows, err := db.QueryContext(ctx, ndbinfoClusterTransactionsQuery)
	if err != nil {
		return err
	}
	defer ndbinfoClusterTransactionsRows.Close()

	var (
		nodeID, state                                                        string
		totalCountOperations, totalOutstandingOperations, maxInactiveSeconds uint64
		clientNodeID                                                         string
		totalTransactions                                                    uint64
	)

	for ndbinfoClusterTransactionsRows.Next() {
		if err := ndbinfoClusterTransactionsRows.Scan(
			&nodeID, &state,
			&totalCountOperations, &totalOutstandingOperations, &maxInactiveSeconds,
			&clientNodeID,
			&totalTransactions,
		); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			ndbinfoClusterTransactionsTotalCountOperationsDesc, prometheus.GaugeValue, float64(totalCountOperations),
			nodeID, state, clientNodeID,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoClusterTransactionsTotalOutstandingOperationsDesc, prometheus.GaugeValue, float64(totalOutstandingOperations),
			nodeID, state, clientNodeID,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoClusterTransactionsMaxInactivityDesc, prometheus.GaugeValue, float64(maxInactiveSeconds),
			nodeID, state, clientNodeID,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoClusterTransactionsTotalTransactionDesc, prometheus.GaugeValue, float64(totalTransactions),
			nodeID, state, clientNodeID,
		)
	}

	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoClusterTransactions{}
