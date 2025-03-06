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

// Scrape `ndbinfo.resources`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoResourcesQuery = `
	SELECT
		node_id,
		resource_name,
		reserved,
		used,
		max,
		spare
	FROM ndbinfo.resources;
`

// Metric descriptors.
var (
	ndbinfoResourcesReservedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "resources"),
		"The amount, as a number of 32KB pages by node_id/resource_name/type.",
		[]string{"node_id", "resource_name", "type"}, nil,
	)
)

// ScrapeNDBInfoResources collects from `ndbinfo.resources`.
type ScrapeNDBInfoResources struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoResources) Name() string {
	return NDBInfo + ".resources"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoResources) Help() string {
	return "Collect metrics from ndbinfo.resources"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoResources) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoResources) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoResourcesRows, err := db.QueryContext(ctx, ndbinfoResourcesQuery)
	if err != nil {
		return err
	}
	defer ndbinfoResourcesRows.Close()
	var (
		nodeID, resourceName string
		// Use pagesMax instead of max to avoid collision with builtin function.
		reserved, used, pagesMax, spare uint64
	)

	for ndbinfoResourcesRows.Next() {
		if err := ndbinfoResourcesRows.Scan(
			&nodeID, &resourceName,
			&reserved, &used, &pagesMax, &spare,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoResourcesReservedDesc, prometheus.GaugeValue, float64(reserved),
			nodeID, resourceName, "reserved",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoResourcesReservedDesc, prometheus.GaugeValue, float64(used),
			nodeID, resourceName, "used",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoResourcesReservedDesc, prometheus.GaugeValue, float64(pagesMax),
			nodeID, resourceName, "max",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoResourcesReservedDesc, prometheus.GaugeValue, float64(spare),
			nodeID, resourceName, "spare",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoResources{}
