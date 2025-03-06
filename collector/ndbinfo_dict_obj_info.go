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

// Scrape `ndbinfo.dict_obj_info`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoDictObjInfoQuery = `
	SELECT
		dot1.type_name AS type,
		doi.id,
		doi.version,
		doi.state,
		COALESCE(dot2.type_name, '0') AS parent_obj_type,
		doi.parent_obj_id,
		doi.fq_name
	FROM ndbinfo.dict_obj_info doi
	JOIN ndbinfo.dict_obj_types dot1
		ON doi.type = dot1.type_id
	LEFT JOIN ndbinfo.dict_obj_types dot2
		ON doi.parent_obj_type = dot2.type_id
		AND doi.parent_obj_type > 0;
`

// Metric descriptors.
var (
	ndbinfoDictObjInfoVersionDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "dict_obj_info_version"),
		"The object version by type/id/parent_obj_type/parent_obj_id/fq_name.",
		[]string{"type", "id", "parent_obj_type", "parent_obj_id", "fq_name"}, nil,
	)
	ndbinfoDictObjInfoStateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "dict_obj_info_state"),
		"The object state by type/id/parent_obj_type/parent_obj_id/fq_name. For human presentation of states, see documentation.",
		[]string{"type", "id", "parent_obj_type", "parent_obj_id", "fq_name"}, nil,
	)
)

// ScrapeNDBInfoDictObjInfo collects from `ndbinfo.dict_obj_info`.
type ScrapeNDBInfoDictObjInfo struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoDictObjInfo) Name() string {
	return NDBInfo + ".dict_obj_info"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoDictObjInfo) Help() string {
	return "Collect metrics from ndbinfo.dict_obj_info"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoDictObjInfo) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoDictObjInfo) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoDictObjInfoRows, err := db.QueryContext(ctx, ndbinfoDictObjInfoQuery)
	if err != nil {
		return err
	}
	defer ndbinfoDictObjInfoRows.Close()

	var (
		objType, id                        string
		version, state                     uint64
		parentObjType, parentObjID, fqName string
	)

	for ndbinfoDictObjInfoRows.Next() {
		if err := ndbinfoDictObjInfoRows.Scan(
			&objType, &id,
			&version, &state,
			&parentObjType, &parentObjID, &fqName,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDictObjInfoVersionDesc, prometheus.GaugeValue, float64(version),
			objType, id,
			parentObjType, parentObjID, fqName,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDictObjInfoStateDesc, prometheus.GaugeValue, float64(state),
			objType, id,
			parentObjType, parentObjID, fqName,
		)
	}

	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoDictObjInfo{}
