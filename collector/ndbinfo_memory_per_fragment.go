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

// Scrape `ndbinfo.memory_per_fragment`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoMemoryPerFragmentQuery = `
	SELECT
		table_id,
		node_id,
		fragment_num,
		fixed_elem_alloc_bytes,
		fixed_elem_free_bytes,
		fixed_elem_size_bytes,
		fixed_elem_count,
		fixed_elem_free_count,
		var_elem_alloc_bytes,
		var_elem_free_bytes,
		var_elem_count,
		hash_index_alloc_bytes
	FROM ndbinfo.memory_per_fragment;
`

// Metric descriptors.
var (
	ndbinfoMemoryPerFragmentAllocBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "memory_per_fragment_alloc_bytes"),
		"The amount of bytes allocated for type by table_id/node_id/fragment_num/type.",
		[]string{"table_id", "node_id", "fragment_num", "type"}, nil,
	)
	ndbinfoMemoryPerFragmentElemFreeBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "memory_per_fragment_elem_free_bytes"),
		"The amount of free bytes remaining in pages allocated to elements by table_id/node_id/fragment_num/type.",
		[]string{"table_id", "node_id", "fragment_num", "type"}, nil,
	)
	ndbinfoMemoryPerFragmentElemCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "memory_per_fragment_elem_count"),
		"The amount of elements by table_id/node_id/fragment_num/type.",
		[]string{"table_id", "node_id", "fragment_num", "type"}, nil,
	)
	ndbinfoMemoryPerFragmentElemSizeBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "memory_per_fragment_elem_size_bytes"),
		"The length of each element in bytes by table_id/node_id/fragment_num/type.",
		[]string{"table_id", "node_id", "fragment_num", "type"}, nil,
	)
	ndbinfoMemoryPerFragmentElemFreeCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "memory_per_fragment_elem_free_count"),
		"The amount of free rows for elements by table_id/node_id/fragment_num/type.",
		[]string{"table_id", "node_id", "fragment_num", "type"}, nil,
	)
)

// ScrapeNDBInfoMemoryPerFragment collects from `ndbinfo.memory_per_fragment`.
type ScrapeNDBInfoMemoryPerFragment struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoMemoryPerFragment) Name() string {
	return NDBInfo + ".memory_per_fragment"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoMemoryPerFragment) Help() string {
	return "Collect metrics from ndbinfo.memory_per_fragment"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoMemoryPerFragment) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoMemoryPerFragment) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoMemoryPerFragmentRows, err := db.QueryContext(ctx, ndbinfoMemoryPerFragmentQuery)
	if err != nil {
		return err
	}
	defer ndbinfoMemoryPerFragmentRows.Close()
	var (
		tableID, nodeID, fragmentNum                                string
		fixedElemAllocBytes, fixedElemFreeBytes, fixedElemSizeBytes uint64
		fixedElemCount, fixedElemFreeCount                          uint64
		varElemAllocBytes, varElemFreeBytes, varElemCount           uint64
		hashIndexAllocBytes                                         uint64
	)

	for ndbinfoMemoryPerFragmentRows.Next() {
		if err := ndbinfoMemoryPerFragmentRows.Scan(
			&tableID, &nodeID, &fragmentNum,
			&fixedElemAllocBytes, &fixedElemFreeBytes, &fixedElemSizeBytes,
			&fixedElemCount, &fixedElemFreeCount,
			&varElemAllocBytes, &varElemFreeBytes, &varElemCount,
			&hashIndexAllocBytes,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoMemoryPerFragmentAllocBytesDesc, prometheus.GaugeValue, float64(fixedElemAllocBytes),
			tableID, nodeID, fragmentNum, "fixed",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoMemoryPerFragmentAllocBytesDesc, prometheus.GaugeValue, float64(varElemAllocBytes),
			tableID, nodeID, fragmentNum, "var",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoMemoryPerFragmentAllocBytesDesc, prometheus.GaugeValue, float64(hashIndexAllocBytes),
			tableID, nodeID, fragmentNum, "hash_index",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoMemoryPerFragmentElemFreeBytesDesc, prometheus.GaugeValue, float64(fixedElemFreeBytes),
			tableID, nodeID, fragmentNum, "fixed",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoMemoryPerFragmentElemFreeBytesDesc, prometheus.GaugeValue, float64(varElemFreeBytes),
			tableID, nodeID, fragmentNum, "var",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoMemoryPerFragmentElemCountDesc, prometheus.GaugeValue, float64(fixedElemCount),
			tableID, nodeID, fragmentNum, "fixed",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoMemoryPerFragmentElemCountDesc, prometheus.GaugeValue, float64(varElemCount),
			tableID, nodeID, fragmentNum, "var",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoMemoryPerFragmentElemSizeBytesDesc, prometheus.GaugeValue, float64(fixedElemSizeBytes),
			tableID, nodeID, fragmentNum, "fixed",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoMemoryPerFragmentElemFreeCountDesc, prometheus.GaugeValue, float64(fixedElemFreeCount),
			tableID, nodeID, fragmentNum, "fixed",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoMemoryPerFragment{}
