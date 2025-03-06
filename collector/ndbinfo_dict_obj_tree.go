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

// Scrape `ndbinfo.dict_obj_tree`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoDictObjTreeQuery = `
	SELECT
		type,
		id,
		name,
		parent_type,
		parent_id,
		parent_name,
		root_type,
		root_id,
		root_name,
		level,
		path,
		indented_name
	FROM ndbinfo.dict_obj_tree;
`

// Metric descriptors.
var (
	ndbinfoDictObjTreeLevelDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "dict_obj_tree_level"),
		"Level of the object in the hierarchy by type/id/name/parent_type/parent_id/parent_name/root_type/root_id/root_name/path/indented_name.",
		[]string{"type", "id", "name", "parent_type", "parent_id", "parent_name", "root_type", "root_id", "root_name", "path", "indented_name"}, nil,
	)
)

// ScrapeNDBInfoDictObjTree collects from `ndbinfo.dict_obj_tree`.
type ScrapeNDBInfoDictObjTree struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoDictObjTree) Name() string {
	return NDBInfo + ".dict_obj_tree"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoDictObjTree) Help() string {
	return "Collect metrics from ndbinfo.dict_obj_tree"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoDictObjTree) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoDictObjTree) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoDictObjTreeRows, err := db.QueryContext(ctx, ndbinfoDictObjTreeQuery)
	if err != nil {
		return err
	}
	defer ndbinfoDictObjTreeRows.Close()
	var (
		dictType, id, name               string
		parentType, parentID, parentName string
		rootType, rootID, rootName       string
		level                            uint64
		path, indentedName               string
	)

	for ndbinfoDictObjTreeRows.Next() {
		if err := ndbinfoDictObjTreeRows.Scan(
			&dictType, &id, &name,
			&parentType, &parentID, &parentName,
			&rootType, &rootID, &rootName,
			&level,
			&path, &indentedName,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDictObjTreeLevelDesc, prometheus.GaugeValue, float64(level),
			dictType, id, name,
			parentType, parentID, parentName,
			rootType, rootID, rootName,
			path, indentedName,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoDictObjTree{}
