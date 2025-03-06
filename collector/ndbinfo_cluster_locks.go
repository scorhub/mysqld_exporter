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

// Scrape `ndbinfo.cluster_locks`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoClusterLocksQuery = `
	SELECT
		node_id,
		tableid,
		fragmentid,
		rowid,
		mode,
		state,
		detail,
		op,
		duration_millis,
		lock_num,
		IFNULL(waiting_for, '') AS waiting_for
	FROM ndbinfo.cluster_locks;
`

// Metric descriptors.
var (
	ndbinfoClusterLocksDurationMillisDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "cluster_locks_duration_millis"),
		"The amount of time spent waiting or holding lock in milliseconds by node_id/tableid/fragmentid/rowid/mode/state/detail/op/lock_num/waiting_for.",
		[]string{"node_id", "tableid", "fragmentid", "rowid", "mode", "state", "detail", "op", "lock_num", "waiting_for"}, nil,
	)
)

// ScrapeNDBInfoClusterLocks collects from `ndbinfo.cluster_locks`.
type ScrapeNDBInfoClusterLocks struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoClusterLocks) Name() string {
	return NDBInfo + ".cluster_locks"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoClusterLocks) Help() string {
	return "Collect metrics from ndbinfo.cluster_locks"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoClusterLocks) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoClusterLocks) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoClusterLocksRows, err := db.QueryContext(ctx, ndbinfoClusterLocksQuery)
	if err != nil {
		return err
	}
	defer ndbinfoClusterLocksRows.Close()

	var (
		nodeID, tableID, fragmentID, rowID string
		mode, state, detail, op            string
		durationMillis                     uint64
		lockNum, waitingFor                string
	)

	for ndbinfoClusterLocksRows.Next() {
		if err := ndbinfoClusterLocksRows.Scan(
			&nodeID, &tableID, &fragmentID, &rowID,
			&mode, &state, &detail, &op,
			&durationMillis,
			&lockNum, &waitingFor,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoClusterLocksDurationMillisDesc, prometheus.GaugeValue, float64(durationMillis),
			nodeID, tableID, fragmentID, rowID,
			mode, state, detail, op,
			lockNum, waitingFor,
		)
	}

	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoClusterLocks{}
