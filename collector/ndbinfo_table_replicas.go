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

// Scrape `ndbinfo.table_replicas`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoTableReplicasQuery = `
	SELECT
		node_id,
		table_id,
		fragment_id,
		initial_gci,
		replica_node_id,
		is_lcp_ongoing,
		num_crashed_replicas,
		last_max_gci_started,
		last_max_gci_completed,
		last_lcp_id,
		prev_lcp_id,
		prev_max_gci_started,
		prev_max_gci_completed,
		last_create_gci,
		last_replica_gci,
		is_replica_alive
	FROM ndbinfo.table_replicas;
`

// Metric descriptors.
var (
	ndbinfoTableReplicasIsReplicaAliveDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "table_replicas_is_replica_alive"),
		"Indicates if the replica is alive (1) or not (0) by node_id/table_id/fragment_id/replica_node_id.",
		[]string{"node_id", "table_id", "fragment_id", "initial_gci", "replica_node_id"}, nil,
	)
	ndbinfoTableReplicasNumCrashedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "table_replicas_num_crashed"),
		"Number of crashed replicas by node_id/table_id/fragment_id/initial_gci/replica_node_id.",
		[]string{"node_id", "table_id", "fragment_id", "initial_gci", "replica_node_id"}, nil,
	)
	ndbinfoTableReplicasIsLCPOngoingDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "table_replicas_is_lcp_ongoing"),
		"Is 1 if LCP is ongoing on this fragment, 0 otherwise by node_id/table_id/fragment_id/initial_gci/replica_node_id.",
		[]string{"node_id", "table_id", "fragment_id", "initial_gci", "replica_node_id"}, nil,
	)
	ndbinfoTableReplicasLCPDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "table_replicas_lcp"),
		"Last / previous LCP ID by node_id/table_id/fragment_id/initial_gci/replica_node_id/type.",
		[]string{"node_id", "table_id", "fragment_id", "initial_gci", "replica_node_id", "type"}, nil,
	)
	ndbinfoTableReplicasGCIDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "table_replicas_gci"),
		"Highest / Last GCI started by node_id/table_id/fragment_id/initial_gci/replica_node_id/type.",
		[]string{"node_id", "table_id", "fragment_id", "initial_gci", "replica_node_id", "type"}, nil,
	)
)

// ScrapeNDBInfoTableReplicas collects from `ndbinfo.table_replicas`.
type ScrapeNDBInfoTableReplicas struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoTableReplicas) Name() string {
	return NDBInfo + ".table_replicas"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoTableReplicas) Help() string {
	return "Collect metrics from ndbinfo.table_replicas"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoTableReplicas) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoTableReplicas) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoTableReplicasRows, err := db.QueryContext(ctx, ndbinfoTableReplicasQuery)
	if err != nil {
		return err
	}
	defer ndbinfoTableReplicasRows.Close()

	var (
		nodeID, tableID                                     string
		fragmentId, initialGCI, replicaNodeId               string
		isLCPOngoing, numCrashedReplicas, lastMaxGCIStarted uint64
		lastMaxGCICompleted, lastLCPId, prevLCPId           uint64
		prevMaxGCIStarted, prevMaxGCICompleted              uint64
		lastCreateGCI, lastReplicaGCI, isReplicaAlive       uint64
	)

	for ndbinfoTableReplicasRows.Next() {
		if err := ndbinfoTableReplicasRows.Scan(
			&nodeID, &tableID,
			&fragmentId, &initialGCI, &replicaNodeId,
			&isLCPOngoing, &numCrashedReplicas, &lastMaxGCIStarted,
			&lastMaxGCICompleted, &lastLCPId, &prevLCPId,
			&prevMaxGCIStarted, &prevMaxGCICompleted,
			&lastCreateGCI, &lastReplicaGCI, &isReplicaAlive,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableReplicasIsReplicaAliveDesc, prometheus.GaugeValue, float64(isReplicaAlive),
			nodeID, tableID, fragmentId, initialGCI, replicaNodeId,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableReplicasNumCrashedDesc, prometheus.GaugeValue, float64(numCrashedReplicas),
			nodeID, tableID, fragmentId, initialGCI, replicaNodeId,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableReplicasIsLCPOngoingDesc, prometheus.GaugeValue, float64(isLCPOngoing),
			nodeID, tableID, fragmentId, initialGCI, replicaNodeId,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableReplicasLCPDesc, prometheus.GaugeValue, float64(lastLCPId),
			nodeID, tableID, fragmentId, initialGCI, replicaNodeId, "last_lcp_id",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableReplicasLCPDesc, prometheus.GaugeValue, float64(prevLCPId),
			nodeID, tableID, fragmentId, initialGCI, replicaNodeId, "prev_lcp_id",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableReplicasGCIDesc, prometheus.GaugeValue, float64(lastMaxGCIStarted),
			nodeID, tableID, fragmentId, initialGCI, replicaNodeId, "last_max_gci_started",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableReplicasGCIDesc, prometheus.GaugeValue, float64(lastMaxGCICompleted),
			nodeID, tableID, fragmentId, initialGCI, replicaNodeId, "last_max_gci_completed",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableReplicasGCIDesc, prometheus.GaugeValue, float64(prevMaxGCIStarted),
			nodeID, tableID, fragmentId, initialGCI, replicaNodeId, "prev_max_gci_started",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableReplicasGCIDesc, prometheus.GaugeValue, float64(prevMaxGCICompleted),
			nodeID, tableID, fragmentId, initialGCI, replicaNodeId, "prev_max_gci_completed",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableReplicasGCIDesc, prometheus.GaugeValue, float64(lastCreateGCI),
			nodeID, tableID, fragmentId, initialGCI, replicaNodeId, "last_create_gci",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableReplicasGCIDesc, prometheus.GaugeValue, float64(lastReplicaGCI),
			nodeID, tableID, fragmentId, initialGCI, replicaNodeId, "last_replica_gci",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoTableReplicas{}
