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

// Scrape `ndbinfo.server_operations`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoServerOperationsCurrentQuery = `
	SELECT
		node_id,
		block_instance,
		operation_type,
		IFNULL(state, '') AS state,
		tableid,
		fragmentid,
		client_node_id,
		tc_node_id,
		tc_block_no,
		tc_block_instance,
		COUNT(*) AS current_ops
	FROM ndbinfo.server_operations
	GROUP BY node_id, block_instance, operation_type, state, tableid, fragmentid, client_node_id, tc_node_id, tc_block_no, tc_block_instance;
`

// Metric descriptors.
var (
	ndbinfoServerOperationsCurrentDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "server_operations_current"),
		"The amount of current server operations by node_id/block_instance/operation_type/state/tableid/fragmentid/client_node_id/tc_node_id/tc_block_no/tc_block_instance.",
		[]string{"node_id", "block_instance", "operation_type", "state", "tableid", "fragmentid", "client_node_id", "tc_node_id", "tc_block_no", "tc_block_instance"}, nil,
	)
)

// ScrapeNDBInfoServerOperations collects from `ndbinfo.server_operations`.
type ScrapeNDBInfoServerOperations struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoServerOperations) Name() string {
	return NDBInfo + ".server_operations"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoServerOperations) Help() string {
	return "Collect metrics from ndbinfo.server_operations"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoServerOperations) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoServerOperations) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoServerOperationsRows, err := db.QueryContext(ctx, ndbinfoServerOperationsCurrentQuery)
	if err != nil {
		return err
	}
	defer ndbinfoServerOperationsRows.Close()

	var (
		nodeID, blockInstance, operationType, state, tableID           string
		fragmentID, clientNodeID, tcNodeID, tcBlockNo, tcBlockInstance string
		currentOps                                                     uint64
	)

	for ndbinfoServerOperationsRows.Next() {
		if err := ndbinfoServerOperationsRows.Scan(
			&nodeID, &blockInstance, &operationType, &state, &tableID,
			&fragmentID, &clientNodeID, &tcNodeID, &tcBlockNo, &tcBlockInstance,
			&currentOps,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoServerOperationsCurrentDesc, prometheus.GaugeValue, float64(currentOps),
			nodeID, blockInstance, operationType, state, tableID, fragmentID, clientNodeID, tcNodeID, tcBlockNo, tcBlockInstance,
		)
	}

	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoServerOperations{}
