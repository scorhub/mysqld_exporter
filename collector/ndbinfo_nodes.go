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

// Scrape `ndbinfo.nodes`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoNodesQuery = `
	SELECT
		node_id,
		uptime,
		status,
		start_phase,
		config_generation
	FROM ndbinfo.nodes;
`

// Metric descriptors.
var (
	ndbinfoNodesUptimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "nodes_uptime_total"),
		"The time since the node was last started, in seconds by node_id/status/start_phase.",
		[]string{"node_id", "status", "start_phase"}, nil,
	)
	ndbinfoNodesConfigGenerationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "nodes_config_generation"),
		"The version of the cluster configuration file in use by node_id/status/start_phase.",
		[]string{"node_id", "status", "start_phase"}, nil,
	)
)

// ScrapeNDBInfoNodes collects from `ndbinfo.nodes`.
type ScrapeNDBInfoNodes struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoNodes) Name() string {
	return NDBInfo + ".nodes"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoNodes) Help() string {
	return "Collect metrics from ndbinfo.nodes"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoNodes) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoNodes) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoNodesRows, err := db.QueryContext(ctx, ndbinfoNodesQuery)
	if err != nil {
		return err
	}
	defer ndbinfoNodesRows.Close()
	var (
		nodeID             string
		uptime             uint64
		status, startPhase string
		configGeneration   uint64
	)

	for ndbinfoNodesRows.Next() {
		if err := ndbinfoNodesRows.Scan(
			&nodeID,
			&uptime,
			&status, &startPhase,
			&configGeneration,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoNodesUptimeDesc, prometheus.GaugeValue, float64(uptime),
			nodeID, status, startPhase,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoNodesConfigGenerationDesc, prometheus.GaugeValue, float64(configGeneration),
			nodeID, status, startPhase,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoNodes{}
