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

// Scrape `ndbinfo.diskpagebuffer`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoDiskPageBufferQuery = `
	SELECT
		node_id,
		block_instance,
		pages_written,
		pages_written_lcp,
		pages_read,
		log_waits,
		page_requests_direct_return,
		page_requests_wait_queue,
		page_requests_wait_io
	FROM ndbinfo.diskpagebuffer;
`

// Metric descriptors.
var (
	ndbinfoDiskPageBufferPagesWrittenDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "diskpagebuffer_pages_written"),
		"The amount of pages written to disk by node_id/block_instance.",
		[]string{"node_id", "block_instance"}, nil,
	)
	ndbinfoDiskPageBufferPagesWrittenLCPDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "diskpagebuffer_pages_written_lcp"),
		"The amount of pages written by local checkpoints by node_id/block_instance.",
		[]string{"node_id", "block_instance"}, nil,
	)
	ndbinfoDiskPageBufferPagesReadDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "diskpagebuffer_pages_read"),
		"The amount of pages read from disk by node_id/block_instance.",
		[]string{"node_id", "block_instance"}, nil,
	)
	ndbinfoDiskPageBufferLogWaitsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "diskpagebuffer_log_waits"),
		"The amount of page writes waiting for log to be written to disk by node_id/block_instance.",
		[]string{"node_id", "block_instance"}, nil,
	)
	ndbinfoDiskPageBufferPageRequestDirectReturnDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "diskpagebuffer_page_requests_direct_return"),
		"The amount of requests for pages that were available in buffer by node_id/block_instance.",
		[]string{"node_id", "block_instance"}, nil,
	)
	ndbinfoDiskPageBufferPageRequestWaitQueueDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "diskpagebuffer_page_requests_wait_queue"),
		"The amount of requests that had to wait for pages to become available in buffer by node_id/block_instance.",
		[]string{"node_id", "block_instance"}, nil,
	)
	ndbinfoDiskPageBufferPageRequestWaitIODesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "diskpagebuffer_page_requests_wait_io"),
		"The amount of requests that had to be read from pages on disk (pages were unavailable in buffer) by node_id/block_instance.",
		[]string{"node_id", "block_instance"}, nil,
	)
)

// ScrapeNDBInfoDiskPageBuffer collects from `ndbinfo.diskpagebuffer`.
type ScrapeNDBInfoDiskPageBuffer struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoDiskPageBuffer) Name() string {
	return NDBInfo + ".diskpagebuffer"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoDiskPageBuffer) Help() string {
	return "Collect metrics from ndbinfo.diskpagebuffer"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoDiskPageBuffer) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoDiskPageBuffer) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoDiskPageBufferRows, err := db.QueryContext(ctx, ndbinfoDiskPageBufferQuery)
	if err != nil {
		return err
	}
	defer ndbinfoDiskPageBufferRows.Close()
	var (
		nodeID, blockInstance                                               string
		pagesWritten, pagesWrittenLCP, pagesRead, logWaits                  uint64
		pageRequestsDirectReturn, pageRequestsWaitQueue, pageRequestsWaitIO uint64
	)

	for ndbinfoDiskPageBufferRows.Next() {
		if err := ndbinfoDiskPageBufferRows.Scan(
			&nodeID, &blockInstance,
			&pagesWritten, &pagesWrittenLCP, &pagesRead, &logWaits,
			&pageRequestsDirectReturn, &pageRequestsWaitQueue, &pageRequestsWaitIO,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskPageBufferPagesWrittenDesc, prometheus.GaugeValue, float64(pagesWritten),
			nodeID, blockInstance,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskPageBufferPagesWrittenLCPDesc, prometheus.GaugeValue, float64(pagesWrittenLCP),
			nodeID, blockInstance,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskPageBufferPagesReadDesc, prometheus.GaugeValue, float64(pagesRead),
			nodeID, blockInstance,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskPageBufferLogWaitsDesc, prometheus.GaugeValue, float64(logWaits),
			nodeID, blockInstance,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskPageBufferPageRequestDirectReturnDesc, prometheus.GaugeValue, float64(pageRequestsDirectReturn),
			nodeID, blockInstance,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskPageBufferPageRequestWaitQueueDesc, prometheus.GaugeValue, float64(pageRequestsWaitQueue),
			nodeID, blockInstance,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskPageBufferPageRequestWaitIODesc, prometheus.GaugeValue, float64(pageRequestsWaitIO),
			nodeID, blockInstance,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoDiskPageBuffer{}
