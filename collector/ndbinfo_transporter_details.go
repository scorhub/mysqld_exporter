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

// Scrape `ndbinfo.transporter_details`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoTransporterDetailsQuery = `
	SELECT
		node_id,
		trp_id,
		remote_node_id,
		status,
		remote_address,
		bytes_sent,
		bytes_received,
		connect_count,
		overloaded,
		overload_count,
		slowdown,
		slowdown_count,
		encrypted,
		sendbuffer_used_bytes,
		sendbuffer_max_used_bytes,
		sendbuffer_alloc_bytes,
		sendbuffer_max_alloc_bytes,
		type
	FROM ndbinfo.transporter_details;
`

// Metric descriptors.
var (
	ndbinfoTransporterDetailsBytesTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "transporter_details_bytes_total"),
		"The total amount of bytes sent/received using this connection by node_id/trp_id/remote_node_id/status/remote_address/type/action.",
		[]string{"node_id", "trp_id", "remote_node_id", "status", "remote_address", "type", "action"}, nil,
	)
	ndbinfoTransporterDetailsCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "transporter_details_count_total"),
		"The total amount of times action has happened on this transporter by node_id/trp_id/remote_node_id/status/remote_address/type/action.",
		[]string{"node_id", "trp_id", "remote_node_id", "status", "remote_address", "type", "action"}, nil,
	)
	ndbinfoTransporterDetailsStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "transporter_details_state"),
		"Whether transporter has state (1) or not (0) by node_id/trp_id/remote_node_id/status/remote_address/type/state.",
		[]string{"node_id", "trp_id", "remote_node_id", "status", "remote_address", "type", "state"}, nil,
	)
	ndbinfoTransporterDetailsSendbufferBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "transporter_details_sendbuffer_bytes"),
		"The amount of bytes for the action by this transporter by node_id/trp_id/remote_node_id/status/remote_address/type/action.",
		[]string{"node_id", "trp_id", "remote_node_id", "status", "remote_address", "type", "action"}, nil,
	)
)

// ScrapeNDBInfoTransporterDetails collects from `ndbinfo.transporter_details`.
type ScrapeNDBInfoTransporterDetails struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoTransporterDetails) Name() string {
	return NDBInfo + ".transporter_details"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoTransporterDetails) Help() string {
	return "Collect metrics from ndbinfo.transporter_details"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoTransporterDetails) Version() float64 {
	return 8.4
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoTransporterDetails) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoTransporterDetailsRows, err := db.QueryContext(ctx, ndbinfoTransporterDetailsQuery)
	if err != nil {
		return err
	}
	defer ndbinfoTransporterDetailsRows.Close()

	var (
		nodeID, trpID, remoteNodeID, status, remoteAddress string
		bytesSent, bytesReceived, connectCount             uint64
		overloaded, overloadCount, slowdown, slowdownCount uint64
		encrypted                                          uint64
		sendBufferUsedBytes, sendBufferMaxUsedBytes        uint64
		sendBufferAllocBytes, sendBufferMaxAllocBytes      uint64
		connectionType                                     string
	)

	for ndbinfoTransporterDetailsRows.Next() {
		if err := ndbinfoTransporterDetailsRows.Scan(
			&nodeID, &remoteNodeID, &trpID, &status, &remoteAddress,
			&bytesSent, &bytesReceived, &connectCount,
			&overloaded, &overloadCount, &slowdown, &slowdownCount,
			&encrypted,
			&sendBufferUsedBytes, &sendBufferMaxUsedBytes,
			&sendBufferAllocBytes, &sendBufferMaxAllocBytes,
			&connectionType,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransporterDetailsBytesTotalDesc, prometheus.CounterValue, float64(bytesSent),
			nodeID, trpID, remoteNodeID, status, remoteAddress, connectionType, "sent",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransporterDetailsBytesTotalDesc, prometheus.CounterValue, float64(bytesReceived),
			nodeID, trpID, remoteNodeID, status, remoteAddress, connectionType, "received",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransporterDetailsCountDesc, prometheus.CounterValue, float64(connectCount),
			nodeID, trpID, remoteNodeID, status, remoteAddress, connectionType, "connect",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransporterDetailsStatusDesc, prometheus.GaugeValue, float64(overloaded),
			nodeID, trpID, remoteNodeID, status, remoteAddress, connectionType, "overload",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransporterDetailsCountDesc, prometheus.CounterValue, float64(overloadCount),
			nodeID, trpID, remoteNodeID, status, remoteAddress, connectionType, "overload",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransporterDetailsStatusDesc, prometheus.GaugeValue, float64(slowdown),
			nodeID, trpID, remoteNodeID, status, remoteAddress, connectionType, "slowdown",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransporterDetailsCountDesc, prometheus.CounterValue, float64(slowdownCount),
			nodeID, trpID, remoteNodeID, status, remoteAddress, connectionType, "slowdown",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransporterDetailsStatusDesc, prometheus.GaugeValue, float64(encrypted),
			nodeID, trpID, remoteNodeID, status, remoteAddress, connectionType, "encrypted",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransporterDetailsSendbufferBytesDesc, prometheus.GaugeValue, float64(sendBufferUsedBytes),
			nodeID, trpID, remoteNodeID, status, remoteAddress, connectionType, "used",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransporterDetailsSendbufferBytesDesc, prometheus.GaugeValue, float64(sendBufferMaxUsedBytes),
			nodeID, trpID, remoteNodeID, status, remoteAddress, connectionType, "max_used",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransporterDetailsSendbufferBytesDesc, prometheus.GaugeValue, float64(sendBufferAllocBytes),
			nodeID, trpID, remoteNodeID, status, remoteAddress, connectionType, "alloc",
		)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTransporterDetailsSendbufferBytesDesc, prometheus.GaugeValue, float64(sendBufferMaxAllocBytes),
			nodeID, trpID, remoteNodeID, status, remoteAddress, connectionType, "max_alloc",
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoTransporterDetails{}
