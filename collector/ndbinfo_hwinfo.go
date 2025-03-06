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

// Scrape `ndbinfo.hwinfo`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoHWInfoQuery = `
	SELECT
		node_id,
		cpu_cnt_max,
		cpu_cnt,
		num_cpu_cores,
		num_cpu_sockets,
		HW_memory_size,
		model_name
	FROM ndbinfo.hwinfo;
`

// Metric descriptors.
var (
	ndbinfoHWInfoCPUCntMaxDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "hwinfo_cpu_cnt_max"),
		"The amount of processors on host by node_id/model_name.",
		[]string{"node_id", "model_name"}, nil,
	)
	ndbinfoHWInfoCPUCntDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "hwinfo_cpu_cnt"),
		"The amount of processors allocated by node_id/model_name.",
		[]string{"node_id", "model_name"}, nil,
	)
	ndbinfoHWInfoNumCPUCoresDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "hwinfo_num_cpu_cores"),
		"The amount of CPU cores on this host by node_id/model_name.",
		[]string{"node_id", "model_name"}, nil,
	)
	ndbinfoHWInfoNumCPUSocketsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "hwinfo_num_cpu_sockets"),
		"The amount of sockets on this host by node_id/model_name.",
		[]string{"node_id", "model_name"}, nil,
	)
	ndbinfoHWInfoHWMemorySizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "hwinfo_hw_memory_size"),
		"The amount of memory available on this host by node_id/model_name.",
		[]string{"node_id", "model_name"}, nil,
	)
)

// ScrapeNDBInfoHWInfo collects from `ndbinfo.hwinfo`.
type ScrapeNDBInfoHWInfo struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoHWInfo) Name() string {
	return NDBInfo + ".hwinfo"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoHWInfo) Help() string {
	return "Collect metrics from ndbinfo.hwinfo"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoHWInfo) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoHWInfo) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoHWInfoRows, err := db.QueryContext(ctx, ndbinfoHWInfoQuery)
	if err != nil {
		return err
	}
	defer ndbinfoHWInfoRows.Close()
	var (
		nodeID                                                      string
		cpuCntMax, cpuCnt, numCPUCores, numCPUSockets, HWMemorySize uint64
		modelName                                                   string
	)

	for ndbinfoHWInfoRows.Next() {
		if err := ndbinfoHWInfoRows.Scan(
			&nodeID,
			&cpuCntMax, &cpuCnt, &numCPUCores, &numCPUSockets, &HWMemorySize,
			&modelName,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoHWInfoCPUCntMaxDesc, prometheus.GaugeValue, float64(cpuCntMax),
			nodeID, modelName,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoHWInfoCPUCntDesc, prometheus.GaugeValue, float64(cpuCnt),
			nodeID, modelName,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoHWInfoNumCPUCoresDesc, prometheus.GaugeValue, float64(numCPUCores),
			nodeID, modelName,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoHWInfoNumCPUSocketsDesc, prometheus.GaugeValue, float64(numCPUSockets),
			nodeID, modelName,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoHWInfoHWMemorySizeDesc, prometheus.GaugeValue, float64(HWMemorySize),
			nodeID, modelName,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoHWInfo{}
