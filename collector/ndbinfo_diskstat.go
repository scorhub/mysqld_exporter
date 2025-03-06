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

// Scrape `ndbinfo.diskstat`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoDiskStatQuery = `
	SELECT
		node_id,
		block_instance,
		pages_made_dirty,
		reads_issued,
		reads_completed,
		writes_issued,
		writes_completed,
		log_writes_issued,
		log_writes_completed,
		get_page_calls_issued,
		get_page_reqs_issued,
		get_page_reqs_completed
	FROM ndbinfo.diskstat;
`

// Metric descriptors.
var (
	ndbinfoDiskStatPagesMadeDirtyDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "diskstat_pages_made_dirty"),
		"The amount of pages made dirty during the past second by node_id/block_instance.",
		[]string{"node_id", "block_instance"}, nil,
	)
	ndbinfoDiskStatReadsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "diskstat_reads"),
		"The amount of reads issued/completed during the past second by node_id/block_instance/action.",
		[]string{"node_id", "block_instance", "action"}, nil,
	)
	ndbinfoDiskStatWritesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "diskstat_writes"),
		"The amount of writes issued/completed during the past second by node_id/block_instance/action.",
		[]string{"node_id", "block_instance", "action"}, nil,
	)
	ndbinfoDiskStatLogWritesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "diskstat_log_writes"),
		"The amount of log writes issued/completed during the past second by node_id/block_instance/action.",
		[]string{"node_id", "block_instance", "action"}, nil,
	)
	ndbinfoDiskStatGetPageCallsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "diskstat_get_page_calls"),
		"The amount of get_page() call actions during the past second by node_id/block_instance/action.",
		[]string{"node_id", "block_instance", "action"}, nil,
	)
	ndbinfoDiskStatGetPageReqsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "diskstat_get_page_reqs"),
		"The amount of get_page() requests issued/completed during the past second by node_id/block_instance/action.",
		[]string{"node_id", "block_instance", "action"}, nil,
	)
)

// ScrapeNDBInfoDiskStat collects from `ndbinfo.diskstat`.
type ScrapeNDBInfoDiskStat struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoDiskStat) Name() string {
	return NDBInfo + ".diskstat"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoDiskStat) Help() string {
	return "Collect metrics from ndbinfo.diskstat"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoDiskStat) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoDiskStat) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoDiskStatRows, err := db.QueryContext(ctx, ndbinfoDiskStatQuery)
	if err != nil {
		return err
	}
	defer ndbinfoDiskStatRows.Close()
	var (
		nodeID, blockInstance                          string
		pagesMadeDirty, readsIssued, readsCompleted    uint64
		writesIssued, writesCompleted, logWritesIssued uint64
		logWritesCompleted, getPageCallsIssued         uint64
		getPageReqsIssued, getPageReqsCompleted        uint64
	)

	for ndbinfoDiskStatRows.Next() {
		if err := ndbinfoDiskStatRows.Scan(
			&nodeID, &blockInstance,
			&pagesMadeDirty, &readsIssued, &readsCompleted,
			&writesIssued, &writesCompleted, &logWritesIssued,
			&logWritesCompleted, &getPageCallsIssued,
			&getPageReqsIssued, &getPageReqsCompleted,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskStatPagesMadeDirtyDesc, prometheus.GaugeValue, float64(pagesMadeDirty),
			nodeID, blockInstance,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskStatReadsDesc, prometheus.GaugeValue, float64(readsIssued),
			nodeID, blockInstance, "issued",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskStatReadsDesc, prometheus.GaugeValue, float64(readsCompleted),
			nodeID, blockInstance, "completed",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskStatWritesDesc, prometheus.GaugeValue, float64(writesIssued),
			nodeID, blockInstance, "issued",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskStatWritesDesc, prometheus.GaugeValue, float64(writesCompleted),
			nodeID, blockInstance, "completed",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskStatLogWritesDesc, prometheus.GaugeValue, float64(logWritesIssued),
			nodeID, blockInstance, "issued",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskStatLogWritesDesc, prometheus.GaugeValue, float64(logWritesCompleted),
			nodeID, blockInstance, "completed",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskStatGetPageCallsDesc, prometheus.GaugeValue, float64(getPageCallsIssued),
			nodeID, blockInstance, "issued",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskStatGetPageReqsDesc, prometheus.GaugeValue, float64(getPageReqsIssued),
			nodeID, blockInstance, "issued",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskStatGetPageReqsDesc, prometheus.GaugeValue, float64(getPageReqsCompleted),
			nodeID, blockInstance, "completed",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoDiskStat{}
