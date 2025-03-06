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

// Scrape `ndbinfo.files`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoFilesQuery = `
	SELECT
		id,
		type,
		name,
		parent,
		parent_name,
		free_extents,
		total_extents,
		(extent_size * 1024 * 1024) AS extent_size_bytes, -- Convert MB to bytes.
		initial_size,
		maximum_size,
		autoextend_size
	FROM ndbinfo.files;
`

// Metric descriptors.
var (
	ndbinfoFilesExtentsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "files_extents"),
		"The amount of free extents by id/type/name/parent/parent_name/amount_type.",
		[]string{"id", "type", "name", "parent", "parent_name", "amount_type"}, nil,
	)
	ndbinfoFilesSizeBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "files_size_bytes"),
		"The extent size in bytes by id/type/name/parent/parent_name/size_type.",
		[]string{"id", "type", "name", "parent", "parent_name", "size_type"}, nil,
	)
)

// ScrapeNDBInfoFiles collects from `ndbinfo.files`.
type ScrapeNDBInfoFiles struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoFiles) Name() string {
	return NDBInfo + ".files"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoFiles) Help() string {
	return "Collect metrics from ndbinfo.files"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoFiles) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoFiles) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoFilesRows, err := db.QueryContext(ctx, ndbinfoFilesQuery)
	if err != nil {
		return err
	}
	defer ndbinfoFilesRows.Close()

	var (
		id, fileType, name, parent, parentName   string
		freeExtents, totalExtents, extentSize    uint64
		initialSize, maximumSize, autoExtendSize uint64
	)
	for ndbinfoFilesRows.Next() {
		if err := ndbinfoFilesRows.Scan(
			&id, &fileType, &name, &parent, &parentName,
			&freeExtents, &totalExtents, &extentSize,
			&initialSize, &maximumSize, &autoExtendSize,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoFilesExtentsDesc, prometheus.GaugeValue, float64(freeExtents),
			id, fileType, name, parent, parentName, "free",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoFilesExtentsDesc, prometheus.GaugeValue, float64(totalExtents),
			id, fileType, name, parent, parentName, "total",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoFilesSizeBytesDesc, prometheus.GaugeValue, float64(extentSize),
			id, fileType, name, parent, parentName, "extent",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoFilesSizeBytesDesc, prometheus.GaugeValue, float64(initialSize),
			id, fileType, name, parent, parentName, "initial",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoFilesSizeBytesDesc, prometheus.GaugeValue, float64(maximumSize),
			id, fileType, name, parent, parentName, "maximum",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoFilesSizeBytesDesc, prometheus.GaugeValue, float64(autoExtendSize),
			id, fileType, name, parent, parentName, "autoextend",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoFiles{}
