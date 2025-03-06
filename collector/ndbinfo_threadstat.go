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

// Scrape `ndbinfo.threadstat`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoThreadStatQuery = `
	SELECT
		node_id,
		thr_no,
		thr_nm,
		c_loop,
		c_exec,
		c_wait,
		c_l_sent_prioa,
		c_l_sent_priob,
		c_r_sent_prioa,
		c_r_sent_priob,
		os_tid,
		os_ru_utime,
		os_ru_stime,
		os_ru_minflt,
		os_ru_majflt,
		os_ru_nvcsw,
		os_ru_nivcsw
	FROM ndbinfo.threadstat;
`

// Metric descriptors.
var (
	ndbinfoThreadStatTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "threadstat_total"),
		"The amount of actions by node_id/thr_no/thr_nm/action.",
		[]string{"node_id", "thr_no", "thr_nm", "action"}, nil,
	)
	ndbinfoThreadStatOsTIDDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "threadstat_os_tid"),
		"The operating system thread ID for node_id/thr_no/thr_nm.",
		[]string{"node_id", "thr_no", "thr_nm"}, nil,
	)
	ndbinfoThreadStatTimeTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "threadstat_time_total"),
		"The action CPU time in microseconds for the thread by node_id/thr_no/thr_nm/action.",
		[]string{"node_id", "thr_no", "thr_nm", "action"}, nil,
	)
)

// ScrapeNDBInfoThreadStat collects from `ndbinfo.threadstat`.
type ScrapeNDBInfoThreadStat struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoThreadStat) Name() string {
	return NDBInfo + ".threadstat"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoThreadStat) Help() string {
	return "Collect metrics from ndbinfo.threadstat"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoThreadStat) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoThreadStat) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoThreadStatRows, err := db.QueryContext(ctx, ndbinfoThreadStatQuery)
	if err != nil {
		return err
	}
	defer ndbinfoThreadStatRows.Close()

	var (
		nodeID, thrNo, thrNm                                                       string
		cLoop, cExec, cWait, cLSentPrioA, cLSentPrioB, cRSentPrioA, cRSentPrioB    uint64
		osTID, osRuUtime, osRuStime, osRuMinFlt, osRuMajFlt, osRuNvcsw, osRuNivcsw uint64
	)

	for ndbinfoThreadStatRows.Next() {
		if err := ndbinfoThreadStatRows.Scan(
			&nodeID, &thrNo, &thrNm,
			&cLoop, &cExec, &cWait, &cLSentPrioA, &cLSentPrioB, &cRSentPrioA, &cRSentPrioB,
			&osTID, &osRuUtime, &osRuStime, &osRuMinFlt, &osRuMajFlt, &osRuNvcsw, &osRuNivcsw,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadStatTotalDesc, prometheus.CounterValue, float64(cLoop),
			nodeID, thrNo, thrNm, "c_loop",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadStatTotalDesc, prometheus.CounterValue, float64(cExec),
			nodeID, thrNo, thrNm, "c_exec",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadStatTotalDesc, prometheus.CounterValue, float64(cWait),
			nodeID, thrNo, thrNm, "c_wait",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadStatTotalDesc, prometheus.CounterValue, float64(cLSentPrioA),
			nodeID, thrNo, thrNm, "c_l_sent_prioa",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadStatTotalDesc, prometheus.CounterValue, float64(cLSentPrioB),
			nodeID, thrNo, thrNm, "c_l_sent_priob",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadStatTotalDesc, prometheus.CounterValue, float64(cRSentPrioA),
			nodeID, thrNo, thrNm, "c_r_sent_prioa",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadStatTotalDesc, prometheus.CounterValue, float64(cRSentPrioB),
			nodeID, thrNo, thrNm, "c_r_sent_priob",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadStatOsTIDDesc, prometheus.GaugeValue, float64(osTID),
			nodeID, thrNo, thrNm,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadStatTimeTotalDesc, prometheus.CounterValue, float64(osRuUtime),
			nodeID, thrNo, thrNm, "os_ru_utime",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadStatTimeTotalDesc, prometheus.CounterValue, float64(osRuStime),
			nodeID, thrNo, thrNm, "os_ru_stime",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadStatTotalDesc, prometheus.CounterValue, float64(osRuMinFlt),
			nodeID, thrNo, thrNm, "os_ru_minflt",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadStatTotalDesc, prometheus.CounterValue, float64(osRuMajFlt),
			nodeID, thrNo, thrNm, "os_ru_majflt",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadStatTotalDesc, prometheus.CounterValue, float64(osRuNvcsw),
			nodeID, thrNo, thrNm, "os_ru_nvcsw",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadStatTotalDesc, prometheus.CounterValue, float64(osRuNivcsw),
			nodeID, thrNo, thrNm, "os_ru_nivcsw",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoThreadStat{}
