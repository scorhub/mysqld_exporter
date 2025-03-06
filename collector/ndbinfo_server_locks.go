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

// Scrape `ndbinfo.server_locks`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoServerLocksQuery = `
	SELECT
		mysql_connection_id,
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
	FROM ndbinfo.server_locks;
`

// Metric descriptors.
var (
	ndbinfoServerLocksDurationMillisDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "server_locks_duration_millis"),
		"The amount of time spent waiting or holding lock in milliseconds by mysql_connection_id/node_id/tableid/fragmentid/rowid/mode/state/detail/op/lock_num/waiting_for.",
		[]string{"mysql_connection_id", "node_id", "tableid", "fragmentid", "rowid", "mode", "state", "detail", "op", "lock_num", "waiting_for"}, nil,
	)
)

// ScrapeNDBInfoServerLocks collects from `ndbinfo.server_locks`.
type ScrapeNDBInfoServerLocks struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoServerLocks) Name() string {
	return NDBInfo + ".server_locks"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoServerLocks) Help() string {
	return "Collect metrics from ndbinfo.server_locks"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoServerLocks) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoServerLocks) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoServerLocksRows, err := db.QueryContext(ctx, ndbinfoServerLocksQuery)
	if err != nil {
		return err
	}
	defer ndbinfoServerLocksRows.Close()

	var (
		mysqlConnectionID, nodeID, tableID         string
		fragmentID, rowID, mode, state, detail, op string
		durationMillis                             uint64
		lockNum, waitingFor                        string
	)

	for ndbinfoServerLocksRows.Next() {
		if err := ndbinfoServerLocksRows.Scan(
			&mysqlConnectionID, &nodeID, &tableID,
			&fragmentID, &rowID, &mode, &state, &detail, &op,
			&durationMillis,
			&lockNum, &waitingFor,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoServerLocksDurationMillisDesc, prometheus.GaugeValue, float64(durationMillis),
			mysqlConnectionID, nodeID, tableID,
			fragmentID, rowID, mode, state, detail, op,
			lockNum, waitingFor,
		)
	}

	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoServerLocks{}
