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

// Scrape `ndbinfo.cpustat`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoCPUStatQuery = `
	SELECT
		node_id,
		thr_no,
		OS_user,
		OS_system,
		OS_idle,
		thread_exec,
		thread_sleeping,
		thread_spinning,
		thread_send,
		thread_buffer_full,
		elapsed_time
	FROM ndbinfo.cpustat;
`

// Metric descriptors.
var (
	ndbinfoCPUStatOSUserDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "cpustat"),
		"The percentage of last second spend in each mode by node_id/thr_no/mode.",
		[]string{"node_id", "thr_no", "mode"}, nil,
	)
	ndbinfoCPUStatElapsedTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "cpustat_elapsed_time"),
		"The total time over which the CPU statistics were collected by node_id/thr_no.",
		[]string{"node_id", "thr_no"}, nil,
	)
)

// ScrapeNDBInfoCPUStat collects from `ndbinfo.cpustat`.
type ScrapeNDBInfoCPUStat struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoCPUStat) Name() string {
	return NDBInfo + ".cpustat"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoCPUStat) Help() string {
	return "Collect CPU usage metrics during the last second from ndbinfo.cpustat"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoCPUStat) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoCPUStat) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoCPUStatRows, err := db.QueryContext(ctx, ndbinfoCPUStatQuery)
	if err != nil {
		return err
	}
	defer ndbinfoCPUStatRows.Close()

	var (
		nodeID, thrNo                                             string
		OSUser, OSSystem, OSIdle, threadExec, threadSleeping      uint64
		threadSpinning, threadSend, threadBufferFull, elapsedTime uint64
	)

	for ndbinfoCPUStatRows.Next() {
		if err := ndbinfoCPUStatRows.Scan(
			&nodeID, &thrNo,
			&OSUser, &OSSystem, &OSIdle, &threadExec, &threadSleeping,
			&threadSpinning, &threadSend, &threadBufferFull, &elapsedTime,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUStatOSUserDesc, prometheus.GaugeValue, float64(OSUser),
			nodeID, thrNo, "os_user",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUStatOSUserDesc, prometheus.GaugeValue, float64(OSSystem),
			nodeID, thrNo, "os_system",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUStatOSUserDesc, prometheus.GaugeValue, float64(OSIdle),
			nodeID, thrNo, "os_idle",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUStatOSUserDesc, prometheus.GaugeValue, float64(threadExec),
			nodeID, thrNo, "thread_exec",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUStatOSUserDesc, prometheus.GaugeValue, float64(threadSleeping),
			nodeID, thrNo, "thread_sleeping",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUStatOSUserDesc, prometheus.GaugeValue, float64(threadSpinning),
			nodeID, thrNo, "thread_spinning",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUStatOSUserDesc, prometheus.GaugeValue, float64(threadSend),
			nodeID, thrNo, "thread_send",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUStatOSUserDesc, prometheus.GaugeValue, float64(threadBufferFull),
			nodeID, thrNo, "thread_buffer_full",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUStatElapsedTimeDesc, prometheus.GaugeValue, float64(elapsedTime),
			nodeID, thrNo,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoCPUStat{}
