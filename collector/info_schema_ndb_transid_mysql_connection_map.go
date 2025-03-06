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

// Scrape `information_schema.ndb_transid_mysql_connection_map`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbConnectionMapQuery = `
	SELECT
		mysql_connection_id,
		node_id,
		ndb_transid
	FROM information_schema.ndb_transid_mysql_connection_map;
`

// Metric descriptors.
var (
	infoSchemaInfoSchemaNDBConnectionMapDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, informationSchema, "ndb_transid_mysql_connection_map"),
		"NDB connections by mysql_connection_id/node_id/ndb_transid. Returns 1 if query is successful.",
		[]string{"mysql_connection_id", "node_id", "ndb_transid"}, nil,
	)
)

// ScrapeInfoSchemaNDBConnectionMap collects from `information_schema.ndb_transid_mysql_connection_map`.
type ScrapeInfoSchemaNDBConnectionMap struct{}

// Name of the Scraper. Should be unique.
func (ScrapeInfoSchemaNDBConnectionMap) Name() string {
	return informationSchema + ".ndb_transid_mysql_connection_map"
}

// Help describes the role of the Scraper.
func (ScrapeInfoSchemaNDBConnectionMap) Help() string {
	return "Collect metrics from information_schema.ndb_transid_mysql_connection_map"
}

// Version of MySQL from which scraper is available.
func (ScrapeInfoSchemaNDBConnectionMap) Version() float64 {
	return 5.1
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeInfoSchemaNDBConnectionMap) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	informationSchemaInfoSchemaNDBConnectionMapRows, err := db.QueryContext(ctx, ndbConnectionMapQuery)
	if err != nil {
		return err
	}
	defer informationSchemaInfoSchemaNDBConnectionMapRows.Close()

	var (
		mysqlConnectionID, nodeID, ndbTransID string
	)

	for informationSchemaInfoSchemaNDBConnectionMapRows.Next() {
		err = informationSchemaInfoSchemaNDBConnectionMapRows.Scan(
			&mysqlConnectionID, &nodeID, &ndbTransID,
		)
		if err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			infoSchemaInfoSchemaNDBConnectionMapDesc, prometheus.GaugeValue, float64(1),
			mysqlConnectionID, nodeID, ndbTransID,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeInfoSchemaNDBConnectionMap{}
