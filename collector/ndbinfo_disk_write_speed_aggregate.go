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

// Scrape `ndbinfo.disk_write_speed_aggregate`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoDiskWriteSpeedAggregateQuery = `
	SELECT
		node_id,
		thr_no,
		backup_lcp_speed_last_sec,
		redo_speed_last_sec,
		backup_lcp_speed_last_10sec,
		redo_speed_last_10sec,
		std_dev_backup_lcp_speed_last_10sec,
		std_dev_redo_speed_last_10sec,
		backup_lcp_speed_last_60sec,
		redo_speed_last_60sec,
		std_dev_backup_lcp_speed_last_60sec,
		std_dev_redo_speed_last_60sec,
		slowdowns_due_to_io_lag,
		slowdowns_due_to_high_cpu,
		disk_write_speed_set_to_min,
		current_target_disk_write_speed  
	FROM ndbinfo.disk_write_speed_aggregate;
`

// Metric descriptors.
var (
	ndbinfoDiskWriteSpeedAggregateBackupLCPSpeedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "disk_write_speed_aggregate_backup_lcp_speed"),
		"The amount of bytes written to disk by backup and LCP processes in the last time interval by node_id/thr_no/time.",
		[]string{"node_id", "thr_no", "time"}, nil,
	)
	ndbinfoDiskWriteSpeedAggregateREDOSpeedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "disk_write_speed_aggregate_redo_speed"),
		"The amount of bytes written to REDO log in the last time interval by node_id/thr_no/time.",
		[]string{"node_id", "thr_no", "time"}, nil,
	)
	ndbinfoDiskWriteSpeedAggregateStdDevBackupLCPSpeedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "disk_write_speed_aggregate_std_dev_backup_lcp_speed"),
		"Standard deviation in number of bytes written to disk by backup and LCP processes per second, averaged over the last x seconds by node_id/thr_no/time.",
		[]string{"node_id", "thr_no", "time"}, nil,
	)
	ndbinfoDiskWriteSpeedAggregateStdDevREDOSpeedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "disk_write_speed_aggregate_std_dev_redo_speed"),
		"Standard deviation in number of bytes written to REDO log per second, averaged over the last x seconds by node_id/thr_no/time.",
		[]string{"node_id", "thr_no", "time"}, nil,
	)
	ndbinfoDiskWriteSpeedAggregateSlowdownsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "disk_write_speed_aggregate_slowdowns"),
		"The amount of seconds since last node start that disk writes were slowed due to reason by node_id/thr_no/reason.",
		[]string{"node_id", "thr_no", "reason"}, nil,
	)
	ndbinfoDiskWriteSpeedAggregateDiskWriteSpeedSetToMinDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "disk_write_speed_aggregate_disk_write_speed_set_to_min"),
		"The amount of seconds since last node start that disk write speed was set to minimum by node_id/thr_no.",
		[]string{"node_id", "thr_no"}, nil,
	)
	ndbinfoDiskWriteSpeedAggregateCurrentTargetDistWriteSpeedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "disk_write_speed_aggregate_current_target_disk_write_speed"),
		"The actual speed of disk writes (aggregated) by node_id/thr_no.",
		[]string{"node_id", "thr_no"}, nil,
	)
)

// ScrapeNDBInfoDiskWriteSpeedAggregate collects from `ndbinfo.disk_write_speed_aggregate`.
type ScrapeNDBInfoDiskWriteSpeedAggregate struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoDiskWriteSpeedAggregate) Name() string {
	return NDBInfo + ".disk_write_speed_aggregate"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoDiskWriteSpeedAggregate) Help() string {
	return "Collect metrics from ndbinfo.disk_write_speed_aggregate"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoDiskWriteSpeedAggregate) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoDiskWriteSpeedAggregate) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoDiskWriteSpeedAggregateRows, err := db.QueryContext(ctx, ndbinfoDiskWriteSpeedAggregateQuery)
	if err != nil {
		return err
	}
	defer ndbinfoDiskWriteSpeedAggregateRows.Close()
	var (
		nodeID, thrNo                                           string
		backupLCPSpeedLastSec, redoSpeedLastSec                 uint64
		backupLCPSpeedLast10sec, redoSpeedLast10sec             uint64
		stdDevBackupLCPSpeedLast10sec, stdDevREDOSpeedLast10sec uint64
		backupLCPSpeedLast60sec, redoSpeedLast60sec             uint64
		stdDevBackupLCPSpeedLast60sec, stdDevREDOSpeedLast60sec uint64
		slowdownsDueToIoLag, slowdownsDueToHighCPU              uint64
		diskWriteSpeedSetToMin, currentTargetDiskWriteSpeed     uint64
	)

	for ndbinfoDiskWriteSpeedAggregateRows.Next() {
		if err := ndbinfoDiskWriteSpeedAggregateRows.Scan(
			&nodeID, &thrNo,
			&backupLCPSpeedLastSec, &redoSpeedLastSec,
			&backupLCPSpeedLast10sec, &redoSpeedLast10sec,
			&stdDevBackupLCPSpeedLast10sec, &stdDevREDOSpeedLast10sec,
			&backupLCPSpeedLast60sec, &redoSpeedLast60sec,
			&stdDevBackupLCPSpeedLast60sec, &stdDevREDOSpeedLast60sec,
			&slowdownsDueToIoLag, &slowdownsDueToHighCPU,
			&diskWriteSpeedSetToMin, &currentTargetDiskWriteSpeed,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateBackupLCPSpeedDesc, prometheus.GaugeValue, float64(backupLCPSpeedLastSec),
			nodeID, thrNo, "sec",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateBackupLCPSpeedDesc, prometheus.GaugeValue, float64(backupLCPSpeedLast10sec),
			nodeID, thrNo, "10sec",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateBackupLCPSpeedDesc, prometheus.GaugeValue, float64(backupLCPSpeedLast60sec),
			nodeID, thrNo, "60sec",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateREDOSpeedDesc, prometheus.GaugeValue, float64(redoSpeedLastSec),
			nodeID, thrNo, "sec",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateREDOSpeedDesc, prometheus.GaugeValue, float64(redoSpeedLast10sec),
			nodeID, thrNo, "10sec",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateREDOSpeedDesc, prometheus.GaugeValue, float64(redoSpeedLast60sec),
			nodeID, thrNo, "60sec",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateStdDevBackupLCPSpeedDesc, prometheus.GaugeValue, float64(stdDevBackupLCPSpeedLast10sec),
			nodeID, thrNo, "10sec",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateStdDevBackupLCPSpeedDesc, prometheus.GaugeValue, float64(stdDevBackupLCPSpeedLast60sec),
			nodeID, thrNo, "60sec",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateStdDevREDOSpeedDesc, prometheus.GaugeValue, float64(stdDevREDOSpeedLast10sec),
			nodeID, thrNo, "10sec",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateStdDevREDOSpeedDesc, prometheus.GaugeValue, float64(stdDevREDOSpeedLast60sec),
			nodeID, thrNo, "60sec",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateSlowdownsDesc, prometheus.GaugeValue, float64(slowdownsDueToIoLag),
			nodeID, thrNo, "io_lag",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateSlowdownsDesc, prometheus.GaugeValue, float64(slowdownsDueToHighCPU),
			nodeID, thrNo, "high_cpu",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateDiskWriteSpeedSetToMinDesc, prometheus.GaugeValue, float64(diskWriteSpeedSetToMin),
			nodeID, thrNo,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateCurrentTargetDistWriteSpeedDesc, prometheus.GaugeValue, float64(currentTargetDiskWriteSpeed),
			nodeID, thrNo,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoDiskWriteSpeedAggregate{}
