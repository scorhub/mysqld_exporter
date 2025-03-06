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

// Scrape `ndbinfo.cpudata`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoCPUDataQuery = `
	SELECT
		node_id,
		cpu_no,
		cpu_online,
		cpu_userspace_time,
		cpu_idle_time,
		cpu_system_time,
		cpu_interrupt_time,
		cpu_exec_vm_time
	FROM ndbinfo.cpudata;
`

// Metric descriptors.
var (
	ndbinfoCPUDataOnlineDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "cpudata_cpu_online"),
		"The status of CPU by node_id/cpu_no. 1 if the CPU is currently online, otherwise 0.",
		[]string{"node_id", "cpu_no"}, nil,
	)
	ndbinfoCPUTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "cpudata_cpu_time"),
		"The amount of time the CPUs spent in each mode during last second by node_id/cpu_no/mode.",
		[]string{"node_id", "cpu_no", "mode"}, nil,
	)
)

// ScrapeNDBInfoCPUData collects from `ndbinfo.cpudata`.
type ScrapeNDBInfoCPUData struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoCPUData) Name() string {
	return NDBInfo + ".cpudata"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoCPUData) Help() string {
	return "Collect metrics from ndbinfo.cpudata"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoCPUData) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoCPUData) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoCPUDataRows, err := db.QueryContext(ctx, ndbinfoCPUDataQuery)
	if err != nil {
		return err
	}
	defer ndbinfoCPUDataRows.Close()

	var (
		nodeID, cpuNo                                  string
		cpuOnline, cpuUserspaceTime, cpuIdleTime       uint64
		cpuSystemTime, cpuInterruptTime, cpuExecVMTime uint64
	)

	for ndbinfoCPUDataRows.Next() {
		if err := ndbinfoCPUDataRows.Scan(
			&nodeID, &cpuNo,
			&cpuOnline, &cpuUserspaceTime, &cpuIdleTime,
			&cpuSystemTime, &cpuInterruptTime, &cpuExecVMTime,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUDataOnlineDesc, prometheus.GaugeValue, float64(cpuOnline),
			nodeID, cpuNo,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUTimeDesc, prometheus.GaugeValue, float64(cpuUserspaceTime),
			nodeID, cpuNo, "userspace",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUTimeDesc, prometheus.GaugeValue, float64(cpuIdleTime),
			nodeID, cpuNo, "idle",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUTimeDesc, prometheus.GaugeValue, float64(cpuSystemTime),
			nodeID, cpuNo, "system",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUTimeDesc, prometheus.GaugeValue, float64(cpuInterruptTime),
			nodeID, cpuNo, "interrupt",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUTimeDesc, prometheus.GaugeValue, float64(cpuExecVMTime),
			nodeID, cpuNo, "exec_vm",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoCPUData{}
