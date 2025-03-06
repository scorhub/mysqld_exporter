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

// Scrape `performance_schema.ndb_sync_excluded_objects`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const perfNDBSyncExcludedObjectsQuery = `
	SELECT
		IFNULL(SCHEMA_NAME, '') AS SCHEMA_NAME,
		IFNULL(NAME, '') AS NAME,
		TYPE,
		REASON
	FROM performance_schema.ndb_sync_excluded_objects;
`

// Metric descriptors.
var (
	performanceSchemaNDBSyncExcludedObjectssDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, performanceSchema, "ndb_sync_excluded_objects"),
		"Returns 1 if query is successful, 0 if error encountered.",
		[]string{"schema_name", "name", "type", "reason"}, nil,
	)
)

// ScrapePerfNDBSyncExcludedObjects
// Events collects from `performance_schema.ndb_sync_excluded_objects`.
type ScrapePerfNDBSyncExcludedObjects struct{}

// Name of the Scraper. Should be unique.
func (ScrapePerfNDBSyncExcludedObjects) Name() string {
	return performanceSchema + ".ndb_sync_excluded_objects"
}

// Help describes the role of the Scraper.
func (ScrapePerfNDBSyncExcludedObjects) Help() string {
	return "Collect metrics from performance_schema.ndb_sync_excluded_objects"
}

// Version of MySQL from which scraper is available.
func (ScrapePerfNDBSyncExcludedObjects) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapePerfNDBSyncExcludedObjects) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	perfSchemaNDBSyncExcludedObjectsRows, err := db.QueryContext(ctx, perfNDBSyncExcludedObjectsQuery)
	if err != nil {
		return err
	}
	defer perfSchemaNDBSyncExcludedObjectsRows.Close()

	var (
		schemaName, name, objType, reason string
	)
	for perfSchemaNDBSyncExcludedObjectsRows.Next() {
		if err = perfSchemaNDBSyncExcludedObjectsRows.Scan(
			&schemaName, &name, &objType, &reason,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			performanceSchemaNDBSyncExcludedObjectssDesc, prometheus.UntypedValue, float64(1),
			schemaName, name, objType, reason,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapePerfNDBSyncExcludedObjects{}
