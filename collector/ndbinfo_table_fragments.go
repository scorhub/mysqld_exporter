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

// Scrape `ndbinfo.table_fragments`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_id'
// to enhance readability of returned values and to improve debugging.
const ndbinfoTableFragmentsQuery = `
	SELECT
		node_id,
		table_id,
		partition_id,
		fragment_id,
		partition_order,
		log_part_id,
		no_of_replicas,
		current_primary,
		preferred_primary,
		current_first_backup,
		current_second_backup,
		current_third_backup,
		num_alive_replicas,
		num_dead_replicas,
		num_lcp_replicas
	FROM ndbinfo.table_fragments;
`

// Metric descriptors.
var (
	ndbinfoTableFragmentsNumReplicasDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "table_fragments_num_replicas"),
		"The current/total number of replicas by node_id/table_id/partition_id/fragment_id/partition_order/log_part_id/status.",
		[]string{"node_id", "table_id", "partition_id", "fragment_id", "partition_order", "log_part_id", "status"}, nil,
	)
	ndbinfoTableFragmentsPrimaryDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "table_fragments_primary"),
		"The primary node ID by node_id/table_id/partition_id/fragment_id/partition_order/log_part_id/type.",
		[]string{"node_id", "table_id", "partition_id", "fragment_id", "partition_order", "log_part_id", "type"}, nil,
	)
	ndbinfoTableFragmentsCurrentBackupDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "table_fragments_current_backup"),
		"The current backup node ID by node_id/table_id/partition_id/fragment_id/partition_order/log_part_id/order.",
		[]string{"node_id", "table_id", "partition_id", "fragment_id", "partition_order", "log_part_id", "order"}, nil,
	)
)

// ScrapeNDBInfoTableFragments collects from `ndbinfo.table_fragments`.
type ScrapeNDBInfoTableFragments struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoTableFragments) Name() string {
	return NDBInfo + ".table_fragments"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoTableFragments) Help() string {
	return "Collect metrics from ndbinfo.table_fragments"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoTableFragments) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoTableFragments) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoTableFragmentsRows, err := db.QueryContext(ctx, ndbinfoTableFragmentsQuery)
	if err != nil {
		return err
	}
	defer ndbinfoTableFragmentsRows.Close()

	var (
		nodeID, tableID, partitionId                                string
		fragmentId, partitionOrder, logPartId                       string
		noOfReplicas, currentPrimary, preferredPrimary              uint64
		currentFirstBackup, currentSecondBackup, currentThirdBackup uint64
		numAliveReplicas, numDeadReplicas, numLCPReplicas           uint64
	)

	for ndbinfoTableFragmentsRows.Next() {
		if err := ndbinfoTableFragmentsRows.Scan(
			&nodeID, &tableID, &partitionId,
			&fragmentId, &partitionOrder, &logPartId,
			&noOfReplicas, &currentPrimary, &preferredPrimary,
			&currentFirstBackup, &currentSecondBackup, &currentThirdBackup,
			&numAliveReplicas, &numDeadReplicas, &numLCPReplicas,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableFragmentsNumReplicasDesc, prometheus.GaugeValue, float64(noOfReplicas),
			nodeID, tableID, partitionId,
			fragmentId, partitionOrder, logPartId,
			"total",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableFragmentsPrimaryDesc, prometheus.GaugeValue, float64(currentPrimary),
			nodeID, tableID, partitionId,
			fragmentId, partitionOrder, logPartId,
			"current",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableFragmentsPrimaryDesc, prometheus.GaugeValue, float64(preferredPrimary),
			nodeID, tableID, partitionId,
			fragmentId, partitionOrder, logPartId,
			"preferred",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableFragmentsCurrentBackupDesc, prometheus.GaugeValue, float64(currentFirstBackup),
			nodeID, tableID, partitionId,
			fragmentId, partitionOrder, logPartId,
			"first",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableFragmentsCurrentBackupDesc, prometheus.GaugeValue, float64(currentSecondBackup),
			nodeID, tableID, partitionId,
			fragmentId, partitionOrder, logPartId,
			"second",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableFragmentsCurrentBackupDesc, prometheus.GaugeValue, float64(currentThirdBackup),
			nodeID, tableID, partitionId,
			fragmentId, partitionOrder, logPartId,
			"third",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableFragmentsNumReplicasDesc, prometheus.GaugeValue, float64(numAliveReplicas),
			nodeID, tableID, partitionId,
			fragmentId, partitionOrder, logPartId,
			"alive",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableFragmentsNumReplicasDesc, prometheus.GaugeValue, float64(numDeadReplicas),
			nodeID, tableID, partitionId,
			fragmentId, partitionOrder, logPartId,
			"dead",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTableFragmentsNumReplicasDesc, prometheus.GaugeValue, float64(numLCPReplicas),
			nodeID, tableID, partitionId,
			fragmentId, partitionOrder, logPartId,
			"lcp",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoTableFragments{}
