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

// Scrape `ndbinfo.hash_maps`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoHashMapsQuery = `
	SELECT 
		id,
		version,
		state
	FROM ndbinfo.hash_maps;
`

// Metric descriptors.
var (
	ndbinfoHashMapsVersionDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "hash_maps_version"),
		"The version of the hash map by id.",
		[]string{"id"}, nil,
	)
	ndbinfoHashMapsStateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "hash_maps_state"),
		"The state of the hash map by id. For human presentation of states, see documentation.",
		[]string{"id"}, nil,
	)
)

// ScrapeNDBInfoHashMaps collects from `ndbinfo.hash_maps`.
type ScrapeNDBInfoHashMaps struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoHashMaps) Name() string {
	return NDBInfo + ".hash_maps"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoHashMaps) Help() string {
	return "Collect metrics from ndbinfo.hash_maps"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoHashMaps) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoHashMaps) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoHashMapsRows, err := db.QueryContext(ctx, ndbinfoHashMapsQuery)
	if err != nil {
		return err
	}
	defer ndbinfoHashMapsRows.Close()

	var (
		id             string
		version, state uint64
	)
	for ndbinfoHashMapsRows.Next() {
		if err := ndbinfoHashMapsRows.Scan(
			&id,
			&version, &state,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoHashMapsVersionDesc, prometheus.GaugeValue, float64(version),
			id,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoHashMapsStateDesc, prometheus.GaugeValue, float64(state),
			id,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoHashMaps{}
