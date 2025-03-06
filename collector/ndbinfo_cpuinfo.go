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

// Scrape `ndbinfo.cpuinfo`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoCPUInfoQuery = `
	SELECT
		node_id,
		cpu_no,
		cpu_online,
		core_id,
		socket_id
	FROM ndbinfo.cpuinfo;
`

// Metric descriptors.
var (
	ndbinfoCPUInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "cpuinfo_cpu_online"),
		"The status of CPU by node_id/cpu_no/core_id/socket_id. 1 if the CPU is online, otherwise 0 .",
		[]string{"node_id", "cpu_no", "core_id", "socket_id"}, nil,
	)
)

// ScrapeNDBInfoCPUInfo collects from `ndbinfo.cpuinfo`.
type ScrapeNDBInfoCPUInfo struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoCPUInfo) Name() string {
	return NDBInfo + ".cpuinfo"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoCPUInfo) Help() string {
	return "Collect metrics from ndbinfo.cpuinfo"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoCPUInfo) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoCPUInfo) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoCPUInfoRows, err := db.QueryContext(ctx, ndbinfoCPUInfoQuery)
	if err != nil {
		return err
	}
	defer ndbinfoCPUInfoRows.Close()

	var (
		nodeID, cpuNo    string
		cpuOnline        uint64
		coreId, socketID string
	)

	for ndbinfoCPUInfoRows.Next() {
		if err := ndbinfoCPUInfoRows.Scan(
			&nodeID, &cpuNo,
			&cpuOnline,
			&coreId, &socketID,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoCPUInfoDesc, prometheus.GaugeValue, float64(cpuOnline),
			nodeID, cpuNo, coreId, socketID,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoCPUInfo{}
