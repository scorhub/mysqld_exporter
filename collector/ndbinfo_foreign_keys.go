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

// Scrape `ndbinfo.foreign_keys`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoForeignKeysQuery = `
	SELECT
		object_id,
		name,
		parent_table,
		parent_columns,
		child_table,
		child_columns,
		parent_index,
		child_index,
		on_update_action,
		on_delete_action
	FROM ndbinfo.foreign_keys;
`

// Metric descriptors.
var (
	ndbinfoForeignKeysColumnsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "foreign_keys_columns"),
		"The amount of columns by object_id/name/parent_table/child_table/parent_index/child_index/on_update_action/on_delete_action/column_type.",
		[]string{"object_id", "name", "parent_table", "child_table", "parent_index", "child_index", "on_update_action", "on_delete_action", "column_type"}, nil,
	)
)

// ScrapeNDBInfoForeignKeys collects from `ndbinfo.foreign_keys`.
type ScrapeNDBInfoForeignKeys struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoForeignKeys) Name() string {
	return NDBInfo + ".foreign_keys"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoForeignKeys) Help() string {
	return "Collect metrics from ndbinfo.foreign_keys"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoForeignKeys) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoForeignKeys) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoForeignKeysRows, err := db.QueryContext(ctx, ndbinfoForeignKeysQuery)
	if err != nil {
		return err
	}
	defer ndbinfoForeignKeysRows.Close()

	var (
		objectID, name, parentTable                             string
		parentColumns                                           uint64
		childTable                                              string
		childColumns                                            uint64
		parentIndex, childIndex, onUpdateAction, onDeleteAction string
	)

	for ndbinfoForeignKeysRows.Next() {
		if err := ndbinfoForeignKeysRows.Scan(
			&objectID, &name, &parentTable,
			&parentColumns,
			&childTable,
			&childColumns,
			&parentIndex, &childIndex, &onUpdateAction, &onDeleteAction,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoForeignKeysColumnsDesc, prometheus.GaugeValue, float64(parentColumns),
			objectID, name, parentTable, childTable,
			parentIndex, childIndex, onUpdateAction, onDeleteAction,
			"parent",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoForeignKeysColumnsDesc, prometheus.GaugeValue, float64(childColumns),
			objectID, name, parentTable, childTable,
			parentIndex, childIndex, onUpdateAction, onDeleteAction,
			"child",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoForeignKeys{}
