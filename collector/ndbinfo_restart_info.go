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

// Scrape `ndbinfo.restart_info`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoRestartInfoQuery = `
	SELECT
		node_id,
		node_restart_status_int,
		secs_to_complete_node_failure,
		secs_to_allocate_node_id,
		secs_to_include_in_heartbeat_protocol,
		secs_until_wait_for_ndbcntr_master,
		secs_wait_for_ndbcntr_master,
		secs_to_get_start_permitted,
		secs_to_wait_for_lcp_for_copy_meta_data,
		secs_to_copy_meta_data,
		secs_to_include_node,
		secs_starting_node_to_request_local_recovery,
		secs_for_local_recovery,
		secs_restore_fragments,
		secs_undo_disk_data,
		secs_exec_redo_log,
		secs_index_rebuild,
		secs_to_synchronize_starting_node,
		secs_wait_lcp_for_restart,
		secs_wait_subscription_handover,
		total_restart_secs
	FROM ndbinfo.restart_info;
`

// Metric descriptors.
var (
	ndbinfoRestartInfoRestartStatusIntDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "restart_info_node_restart_status_int"),
		"The node restart status code by node_id. See manual for human readable value.",
		[]string{"node_id"}, nil,
	)
	ndbinfoRestartInfoSecsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "restart_info_secs"),
		"Time in seconds to complete action by node_id/action.",
		[]string{"node_id", "action"}, nil,
	)
)

// ScrapeNDBInfoRestartInfo collects from `ndbinfo.restart_info`.
type ScrapeNDBInfoRestartInfo struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoRestartInfo) Name() string {
	return NDBInfo + ".restart_info"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoRestartInfo) Help() string {
	return "Collect metrics from ndbinfo.restart_info"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoRestartInfo) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoRestartInfo) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoRestartInfoRows, err := db.QueryContext(ctx, ndbinfoRestartInfoQuery)
	if err != nil {
		return err
	}
	defer ndbinfoRestartInfoRows.Close()
	var (
		nodeID                                                       string
		nodeRestartStatusInt, secsToCompleteNodeFailure              uint64
		secsToAllocateNodeId, secsToIncludeInHeartbeatProtocol       uint64
		secsUntilWaitForNDBCntrMaster, secsWaitForNDBCntrMaster      uint64
		secsToGetStartPermitted, secsToWaitForLCPForCopyMetaData     uint64
		secsToCopyMetaData, secsToIncludeNode                        uint64
		secsStartingNodeToRequestLocalRecovery, secsForLocalRecovery uint64
		secsRestoreFragments, secsUndoDiskData, secsExecRedoLog      uint64
		secsIndexRebuild, secsToSynchronizeStartingNode              uint64
		secsWaitLCPForRestart, secsWaitSubscriptionHandover          uint64
		totalRestartSecs                                             uint64
	)

	for ndbinfoRestartInfoRows.Next() {
		if err := ndbinfoRestartInfoRows.Scan(
			&nodeID,
			&nodeRestartStatusInt, &secsToCompleteNodeFailure,
			&secsToAllocateNodeId, &secsToIncludeInHeartbeatProtocol,
			&secsUntilWaitForNDBCntrMaster, &secsWaitForNDBCntrMaster,
			&secsToGetStartPermitted, &secsToWaitForLCPForCopyMetaData,
			&secsToCopyMetaData, &secsToIncludeNode,
			&secsStartingNodeToRequestLocalRecovery, &secsForLocalRecovery,
			&secsRestoreFragments, &secsUndoDiskData, &secsExecRedoLog,
			&secsIndexRebuild, &secsToSynchronizeStartingNode,
			&secsWaitLCPForRestart, &secsWaitSubscriptionHandover,
			&totalRestartSecs,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoRestartStatusIntDesc, prometheus.GaugeValue, float64(nodeRestartStatusInt),
			nodeID,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsToCompleteNodeFailure),
			nodeID, "complete_node_failure",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsToAllocateNodeId),
			nodeID, "allocate_node_id",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsToIncludeInHeartbeatProtocol),
			nodeID, "include_in_heartbeat_protocol",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsUntilWaitForNDBCntrMaster),
			nodeID, "until_wait_for_ndbcntr_master",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsWaitForNDBCntrMaster),
			nodeID, "wait_for_ndbcntr_master",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsToGetStartPermitted),
			nodeID, "get_start_permitted",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsToWaitForLCPForCopyMetaData),
			nodeID, "wait_for_lcp_for_copy_meta_data",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsToCopyMetaData),
			nodeID, "copy_meta_data",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsToIncludeNode),
			nodeID, "include_node",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsStartingNodeToRequestLocalRecovery),
			nodeID, "starting_node_to_request_local_recovery",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsForLocalRecovery),
			nodeID, "local_recovery",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsRestoreFragments),
			nodeID, "restore_fragments",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsUndoDiskData),
			nodeID, "undo_disk_data",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsExecRedoLog),
			nodeID, "exec_redo_log",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsIndexRebuild),
			nodeID, "index_rebuild",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsToSynchronizeStartingNode),
			nodeID, "synchronize_starting_node",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsWaitLCPForRestart),
			nodeID, "wait_lcp_for_restart",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(secsWaitSubscriptionHandover),
			nodeID, "wait_subscription_handover",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoRestartInfoSecsDesc, prometheus.GaugeValue, float64(totalRestartSecs),
			nodeID, "total",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoRestartInfo{}
