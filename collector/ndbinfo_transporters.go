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

// Scrape `ndbinfo.transporters`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoTransportersQuery = `
	SELECT
		node_id,
		remote_node_id,
		status,
		remote_address,
		bytes_sent,
		bytes_received,
		connect_count,
		overloaded,
		overload_count,
		slowdown,
		slowdown_count
	FROM ndbinfo.transporters;
`

// Metric descriptors.
var (
	ndbinfoTransportersBytesTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "transporters_bytes_total"),
		"The total amount of bytes sent/received using this connection by node_id/remote_node_id/status/remote_address/action.",
		[]string{"node_id", "remote_node_id", "status", "remote_address", "action"}, nil,
	)
	ndbinfoTransportersStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "transporters_state"),
		"Whether transporter has state (1) or not (0) by node_id/remote_node_id/status/remote_address/state.",
		[]string{"node_id", "remote_node_id", "status", "remote_address", "state"}, nil,
	)
	ndbinfoTransportersCountTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "transporters_count_total"),
		"The total amount of times action has happened on this transporter by node_id/remote_node_id/status/remote_address/action.",
		[]string{"node_id", "remote_node_id", "status", "remote_address", "action"}, nil,
	)
)

// ScrapeNDBInfoTransporters collects from `ndbinfo.transporters`.
type ScrapeNDBInfoTransporters struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoTransporters) Name() string {
	return NDBInfo + ".transporters"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoTransporters) Help() string {
	return "Collect metrics from ndbinfo.transporters"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoTransporters) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoTransporters) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoTransportersRows, err := db.QueryContext(ctx, ndbinfoTransportersQuery)
	if err != nil {
		return err
	}
	defer ndbinfoTransportersRows.Close()

	var (
		nodeID, remoteNodeID, status, remoteAddress        string
		bytesSent, bytesReceived, connectCount             uint64
		overloaded, overloadCount, slowdown, slowdownCount uint64
	)

	for ndbinfoTransportersRows.Next() {
		if err := ndbinfoTransportersRows.Scan(
			&nodeID, &remoteNodeID, &status, &remoteAddress,
			&bytesSent, &bytesReceived, &connectCount,
			&overloaded, &overloadCount, &slowdown, &slowdownCount,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransportersBytesTotalDesc, prometheus.CounterValue, float64(bytesSent),
			nodeID, remoteNodeID, status, remoteAddress, "sent",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransportersBytesTotalDesc, prometheus.CounterValue, float64(bytesReceived),
			nodeID, remoteNodeID, status, remoteAddress, "received",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransportersCountTotalDesc, prometheus.CounterValue, float64(connectCount),
			nodeID, remoteNodeID, status, remoteAddress, "connect",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransportersStatusDesc, prometheus.GaugeValue, float64(overloaded),
			nodeID, remoteNodeID, status, remoteAddress, "overload",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransportersCountTotalDesc, prometheus.CounterValue, float64(overloadCount),
			nodeID, remoteNodeID, status, remoteAddress, "overload",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransportersStatusDesc, prometheus.GaugeValue, float64(slowdown),
			nodeID, remoteNodeID, status, remoteAddress, "slowdown",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransportersCountTotalDesc, prometheus.CounterValue, float64(slowdownCount),
			nodeID, remoteNodeID, status, remoteAddress, "slowdown",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoTransporters{}
