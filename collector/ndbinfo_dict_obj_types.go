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

// Scrape `ndbinfo.dict_obj_types`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoDictObjTypesQuery = `
	SELECT
		type_id,
		type_name
	FROM ndbinfo.dict_obj_types;
`

// Metric descriptors.
var (
	ndbinfoDictObjTypesLevelDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "dict_obj_types"),
		"Returns 1 if query is successful, 0 if error encountered.",
		[]string{"type_id", "type_name"}, nil,
	)
)

// ScrapeNDBInfoDictObjTypes collects from `ndbinfo.dict_obj_types`.
type ScrapeNDBInfoDictObjTypes struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoDictObjTypes) Name() string {
	return NDBInfo + ".dict_obj_types"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoDictObjTypes) Help() string {
	return "Collect metrics from ndbinfo.dict_obj_types"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoDictObjTypes) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoDictObjTypes) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoDictObjTypesRows, err := db.QueryContext(ctx, ndbinfoDictObjTypesQuery)
	if err != nil {
		return err
	}
	defer ndbinfoDictObjTypesRows.Close()
	var (
		typeID, typeName string
	)

	for ndbinfoDictObjTypesRows.Next() {
		if err := ndbinfoDictObjTypesRows.Scan(
			&typeID, &typeName,
		); err != nil {
			ch <- prometheus.MustNewConstMetric(
				ndbinfoDictObjTypesLevelDesc, prometheus.GaugeValue, float64(0),
				typeID, typeName,
			)
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDictObjTypesLevelDesc, prometheus.GaugeValue, float64(1),
			typeID, typeName,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoDictObjTypes{}
