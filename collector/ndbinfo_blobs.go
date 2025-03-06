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

// Scrape `ndbinfo.blobs`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoBlobsQuery = `
	SELECT
		table_id,
		database_name,
		table_name,
		column_id,
		inline_size,
		part_size,
		stripe_size
	FROM ndbinfo.blobs;
`

// Metric descriptors.
var (
	ndbinfoBlobsSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "blobs_size"),
		"The size of blobs column by table_id/database_name/table_name/column_id/type.",
		[]string{"table_id", "database_name", "table_name", "column_id", "type"}, nil,
	)
)

// ScrapeNDBInfoBlobs collects from `ndbinfo.blobs`.
type ScrapeNDBInfoBlobs struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoBlobs) Name() string {
	return NDBInfo + ".blobs"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoBlobs) Help() string {
	return "Collect metrics from ndbinfo.blobs"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoBlobs) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoBlobs) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoBlobsRows, err := db.QueryContext(ctx, ndbinfoBlobsQuery)
	if err != nil {
		return err
	}
	defer ndbinfoBlobsRows.Close()

	var (
		tableID, databaseName, tableName, columnID string
		inlineSize, partSize, stripeSize           uint64
	)
	for ndbinfoBlobsRows.Next() {
		if err := ndbinfoBlobsRows.Scan(
			&tableID, &databaseName, &tableName, &columnID,
			&inlineSize, &partSize, &stripeSize,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoBlobsSizeDesc, prometheus.GaugeValue, float64(inlineSize),
			tableID, databaseName, tableName, columnID, "inline",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoBlobsSizeDesc, prometheus.GaugeValue, float64(partSize),
			tableID, databaseName, tableName, columnID, "part",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoBlobsSizeDesc, prometheus.GaugeValue, float64(stripeSize),
			tableID, databaseName, tableName, columnID, "stripe",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoBlobs{}
