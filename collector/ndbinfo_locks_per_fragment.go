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

// Scrape `ndbinfo.locks_per_fragment`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoLocksPerFragmentQuery = `
	SELECT
		table_id,
		node_id,
		fragment_num,
		ex_req,
		ex_imm_ok,
		ex_wait_ok,
		ex_wait_fail,
		sh_req,
		sh_imm_ok,
		sh_wait_ok,
		sh_wait_fail,
		wait_ok_millis,
		wait_fail_millis
	FROM ndbinfo.locks_per_fragment;
`

// Metric descriptors.
var (
	ndbinfoLocksPerFragmentExDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "locks_per_fragment_ex_total"),
		"The amount of exclusive lock requests by table_id/node_id/fragment_num/action.",
		[]string{"table_id", "node_id", "fragment_num", "action"}, nil,
	)
	ndbinfoLocksPerFragmentShDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "locks_per_fragment_sh_total"),
		"The amount of shared lock requests by table_id/node_id/fragment_num/action.",
		[]string{"table_id", "node_id", "fragment_num", "action"}, nil,
	)
	ndbinfoLocksPerFragmentWaitDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "locks_per_fragment_wait_millis_total"),
		"The amount of time spent waiting for lock requests that were granted/failed, in milliseconds by table_id/node_id/fragment_num/action.",
		[]string{"table_id", "node_id", "fragment_num", "action"}, nil,
	)
)

// ScrapeNDBInfoLocksPerFragment collects from `ndbinfo.locks_per_fragment`.
type ScrapeNDBInfoLocksPerFragment struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoLocksPerFragment) Name() string {
	return NDBInfo + ".locks_per_fragment"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoLocksPerFragment) Help() string {
	return "Collect metrics from ndbinfo.locks_per_fragment"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoLocksPerFragment) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoLocksPerFragment) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoLocksPerFragmentRows, err := db.QueryContext(ctx, ndbinfoLocksPerFragmentQuery)
	if err != nil {
		return err
	}
	defer ndbinfoLocksPerFragmentRows.Close()
	var (
		tableID, nodeID, fragmentNum         string
		exReq, exImmOk, exWaitOk, exWaitFail uint64
		shReq, shImmOk, shWaitOk, shWaitFail uint64
		waitOkMillis, waitFailMillis         uint64
	)

	for ndbinfoLocksPerFragmentRows.Next() {
		if err := ndbinfoLocksPerFragmentRows.Scan(
			&tableID, &nodeID, &fragmentNum,
			&exReq, &exImmOk, &exWaitOk, &exWaitFail,
			&shReq, &shImmOk, &shWaitOk, &shWaitFail,
			&waitOkMillis, &waitFailMillis,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLocksPerFragmentExDesc, prometheus.CounterValue, float64(exReq),
			tableID, nodeID, fragmentNum, "req",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLocksPerFragmentExDesc, prometheus.CounterValue, float64(exImmOk),
			tableID, nodeID, fragmentNum, "imm_ok",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLocksPerFragmentExDesc, prometheus.CounterValue, float64(exWaitOk),
			tableID, nodeID, fragmentNum, "wait_ok",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLocksPerFragmentExDesc, prometheus.CounterValue, float64(exWaitFail),
			tableID, nodeID, fragmentNum, "wait_fail",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLocksPerFragmentShDesc, prometheus.CounterValue, float64(shReq),
			tableID, nodeID, fragmentNum, "req",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLocksPerFragmentShDesc, prometheus.CounterValue, float64(shImmOk),
			tableID, nodeID, fragmentNum, "imm_ok",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLocksPerFragmentShDesc, prometheus.CounterValue, float64(shWaitOk),
			tableID, nodeID, fragmentNum, "wait_ok",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLocksPerFragmentShDesc, prometheus.CounterValue, float64(shWaitFail),
			tableID, nodeID, fragmentNum, "wait_fail",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLocksPerFragmentWaitDesc, prometheus.CounterValue, float64(waitOkMillis),
			tableID, nodeID, fragmentNum, "ok",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLocksPerFragmentWaitDesc, prometheus.CounterValue, float64(waitFailMillis),
			tableID, nodeID, fragmentNum, "fail",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoLocksPerFragment{}
