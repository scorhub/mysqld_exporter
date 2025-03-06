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

// Scrape `ndbinfo.server_transactions`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoServerTransactionsQuery = `
	SELECT
		node_id,
		state,
		SUM(count_operations) AS total_count_operations,
		SUM(outstanding_operations) AS total_outstanding_operations,
		MAX(inactive_seconds) AS max_inactive_seconds,
		client_node_id,
		COUNT(*) AS total_transactions
	FROM ndbinfo.server_transactions
	GROUP BY node_id, state, client_node_id;
`

// Metric descriptors.
var (
	ndbinfoServerTransactionsTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "server_transactions_total"),
		"The total amount of operations by node_id/state/client_node_id/action.",
		[]string{"node_id", "state", "client_node_id", "action"}, nil,
	)
	ndbinfoServerTransactionsMaxInactivityDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "server_transactions_max_inactive_seconds"),
		"The maximum inactivity seconds by node_id/state/client_node_id.",
		[]string{"node_id", "state", "client_node_id"}, nil,
	)
)

// ScrapeNDBInfoServerTransactions collects from `ndbinfo.server_transactions`.
type ScrapeNDBInfoServerTransactions struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoServerTransactions) Name() string {
	return NDBInfo + ".server_transactions"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoServerTransactions) Help() string {
	return "Collect metrics from ndbinfo.server_transactions"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoServerTransactions) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoServerTransactions) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoServerTransactionsRows, err := db.QueryContext(ctx, ndbinfoServerTransactionsQuery)
	if err != nil {
		return err
	}
	defer ndbinfoServerTransactionsRows.Close()

	var (
		nodeID, state                                                        string
		totalCountOperations, totalOutstandingOperations, maxInactiveSeconds uint64
		clientNodeID                                                         string
		totalTransactions                                                    uint64
	)

	for ndbinfoServerTransactionsRows.Next() {
		if err := ndbinfoServerTransactionsRows.Scan(
			&nodeID, &state,
			&totalCountOperations, &totalOutstandingOperations, &maxInactiveSeconds,
			&clientNodeID,
			&totalTransactions,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoServerTransactionsTotalDesc, prometheus.GaugeValue, float64(totalCountOperations),
			nodeID, state, clientNodeID, "count",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoServerTransactionsTotalDesc, prometheus.GaugeValue, float64(totalOutstandingOperations),
			nodeID, state, clientNodeID, "outstanding",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoServerTransactionsTotalDesc, prometheus.GaugeValue, float64(totalTransactions),
			nodeID, state, clientNodeID, "transactions",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoServerTransactionsMaxInactivityDesc, prometheus.GaugeValue, float64(maxInactiveSeconds),
			nodeID, state, clientNodeID,
		)
	}

	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoServerTransactions{}
