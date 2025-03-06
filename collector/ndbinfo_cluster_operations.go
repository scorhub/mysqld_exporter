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

// Scrape `ndbinfo.cluster_operations`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoClusterOperationsCurrentQuery = `
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
	FROM ndbinfo.cluster_operations
	GROUP BY node_id, block_instance, operation_type, state, tableid, fragmentid, client_node_id, tc_node_id, tc_block_no, tc_block_instance;
`

// Metric descriptors.
var (
	ndbinfoClusterOperationsCurrentDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "cluster_operations_current"),
		"The amount of current cluster operations by node_id/block_instance/operation_type/state/tableid/fragmentid/client_node_id/tc_node_id/tc_block_no/tc_block_instance.",
		[]string{"node_id", "block_instance", "operation_type", "state", "tableid", "fragmentid", "client_node_id", "tc_node_id", "tc_block_no", "tc_block_instance"}, nil,
	)
)

// ScrapeNDBInfoClusterOperations collects from `ndbinfo.cluster_operations`.
type ScrapeNDBInfoClusterOperations struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoClusterOperations) Name() string {
	return NDBInfo + ".cluster_operations"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoClusterOperations) Help() string {
	return "Collect metrics from ndbinfo.cluster_operations"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoClusterOperations) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoClusterOperations) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoClusterOperationsRows, err := db.QueryContext(ctx, ndbinfoClusterOperationsCurrentQuery)
	if err != nil {
		return err
	}
	defer ndbinfoClusterOperationsRows.Close()

	var (
		nodeID, blockInstance, operationType, state, tableID           string
		fragmentID, clientNodeID, tcNodeID, tcBlockNo, tcBlockInstance string
		currentOps                                                     uint64
	)

	for ndbinfoClusterOperationsRows.Next() {
		if err := ndbinfoClusterOperationsRows.Scan(
			&nodeID, &blockInstance, &operationType, &state, &tableID,
			&fragmentID, &clientNodeID, &tcNodeID, &tcBlockNo, &tcBlockInstance,
			&currentOps,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoClusterOperationsCurrentDesc, prometheus.GaugeValue, float64(currentOps),
			nodeID, blockInstance, operationType, state, tableID,
			fragmentID, clientNodeID, tcNodeID, tcBlockNo, tcBlockInstance,
		)
	}

	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoClusterOperations{}
