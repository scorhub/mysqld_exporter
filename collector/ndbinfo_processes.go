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

// Scrape `ndbinfo.processes`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoProcessesQuery = `
	SELECT
		node_id,
		node_type,
		node_version,
		process_id,
		IFNULL(angel_process_id, 0) AS angel_process_id,
		process_name,
		service_URI
	FROM ndbinfo.processes;
`

// Metric descriptors.
var (
	ndbinfoProcessesAngelProcessIDDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "processes_angel_process_id"),
		"Process ID of this node's angel process by node_id/node_type/node_version/process_id/process_name/service_URI. Returns 0 if no angel process is set.",
		[]string{"node_id", "node_type", "node_version", "process_id", "process_name", "service_URI"}, nil,
	)
)

// ScrapeNDBInfoProcesses collects from `ndbinfo.processes`.
type ScrapeNDBInfoProcesses struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoProcesses) Name() string {
	return NDBInfo + ".processes"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoProcesses) Help() string {
	return "Collect metrics from ndbinfo.processes"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoProcesses) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoProcesses) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoProcessesRows, err := db.QueryContext(ctx, ndbinfoProcessesQuery)
	if err != nil {
		return err
	}
	defer ndbinfoProcessesRows.Close()

	var (
		nodeID, nodeType, nodeVersion, processID string
		angelProcessID                           uint64
		processName, serviceURI                  string
	)

	for ndbinfoProcessesRows.Next() {
		if err := ndbinfoProcessesRows.Scan(
			&nodeID, &nodeType, &nodeVersion, &processID,
			&angelProcessID,
			&processName, &serviceURI,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoProcessesAngelProcessIDDesc, prometheus.GaugeValue, float64(angelProcessID),
			nodeID, nodeType, nodeVersion, processID, processName, serviceURI,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoProcesses{}
