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

// Scrape `performance_schema.ndb_sync_pending_objects`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const perfNDBSyncPendingObjectsQuery = `
	SELECT
		IFNULL(SCHEMA_NAME, '') AS SCHEMA_NAME,
		IFNULL(NAME, '') AS NAME,
		TYPE
	FROM performance_schema.ndb_sync_pending_objects;
`

// Metric descriptors.
var (
	performanceSchemaNDBSyncPendingObjectsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, performanceSchema, "ndb_sync_pending_objects"),
		"Returns 1 if query is successful, 0 if error encountered.",
		[]string{"schema_name", "name", "type"}, nil,
	)
)

// ScrapePerfNDBSyncPendingObjects
// Events collects from `performance_schema.ndb_sync_pending_objects`.
type ScrapePerfNDBSyncPendingObjects struct{}

// Name of the Scraper. Should be unique.
func (ScrapePerfNDBSyncPendingObjects) Name() string {
	return performanceSchema + ".ndb_sync_pending_objects"
}

// Help describes the role of the Scraper.
func (ScrapePerfNDBSyncPendingObjects) Help() string {
	return "Collect metrics from performance_schema.ndb_sync_pending_objects"
}

// Version of MySQL from which scraper is available.
func (ScrapePerfNDBSyncPendingObjects) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapePerfNDBSyncPendingObjects) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	perfSchemaNDBSyncPendingObjectsRows, err := db.QueryContext(ctx, perfNDBSyncPendingObjectsQuery)
	if err != nil {
		return err
	}
	defer perfSchemaNDBSyncPendingObjectsRows.Close()

	var (
		schemaName, name, objType string
	)
	for perfSchemaNDBSyncPendingObjectsRows.Next() {
		if err = perfSchemaNDBSyncPendingObjectsRows.Scan(
			&schemaName, &name, &objType,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			performanceSchemaNDBSyncPendingObjectsDesc, prometheus.UntypedValue, float64(1),
			schemaName, name, objType,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapePerfNDBSyncPendingObjects{}
