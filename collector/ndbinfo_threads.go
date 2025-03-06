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

// Scrape `ndbinfo.threads`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoThreadsQuery = `
	SELECT
		node_id,
		thr_no,
		thread_name,
		thread_description
	FROM ndbinfo.threads;
`

// Metric descriptors.
var (
	ndbinfoThreadsThreadNoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "threads"),
		"Returns 1 if query is successful, 0 if error encountered. Grouped by node_id/thr_no/thread_name/thread_description.",
		[]string{"node_id", "thr_no", "thread_name", "thread_description"}, nil,
	)
)

// ScrapeNDBInfoThreads collects from `ndbinfo.threads`.
type ScrapeNDBInfoThreads struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoThreads) Name() string {
	return NDBInfo + ".threads"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoThreads) Help() string {
	return "Collect metrics from ndbinfo.threads"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoThreads) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoThreads) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoThreadsRows, err := db.QueryContext(ctx, ndbinfoThreadsQuery)
	if err != nil {
		return err
	}
	defer ndbinfoThreadsRows.Close()

	var (
		nodeID, thrNo, threadName, threadDescription string
	)

	for ndbinfoThreadsRows.Next() {
		if err := ndbinfoThreadsRows.Scan(
			&nodeID, &thrNo, &threadName, &threadDescription,
		); err != nil {
			ch <- prometheus.MustNewConstMetric(
				ndbinfoThreadsThreadNoDesc, prometheus.GaugeValue, float64(0),
				nodeID, thrNo, threadName, threadDescription,
			)
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadsThreadNoDesc, prometheus.GaugeValue, float64(1),
			nodeID, thrNo, threadName, threadDescription,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoThreads{}
