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

// Scrape `ndbinfo.threadblocks`.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// Include full query, even in case of 'SELECT * FROM database.table_name'
// to enhance readability of returned values and to improve debugging.
const ndbinfoThreadBlocksQuery = `
	SELECT
		node_id,
		thr_no,
		block_name,
		block_instance
	FROM ndbinfo.threadblocks;
`

// Metric descriptors.
var (
	ndbinfoThreadBlocksDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, NDBInfo, "threadblocks"),
		"Returns 1 if query is successful, 0 if error encountered.",
		[]string{"node_id", "thr_no", "block_name", "block_instance"}, nil,
	)
)

// ScrapeNDBInfoThreadBlocks collects from `ndbinfo.threadblocks`.
type ScrapeNDBInfoThreadBlocks struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNDBInfoThreadBlocks) Name() string {
	return NDBInfo + ".threadblocks"
}

// Help describes the role of the Scraper.
func (ScrapeNDBInfoThreadBlocks) Help() string {
	return "Collect metrics from ndbinfo.threadblocks"
}

// Version of MySQL from which scraper is available.
func (ScrapeNDBInfoThreadBlocks) Version() float64 {
	return 8.0
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeNDBInfoThreadBlocks) Scrape(ctx context.Context, instance *instance, ch chan<- prometheus.Metric, _ *slog.Logger) error {
	db := instance.getDB()
	ndbinfoThreadBlocksRows, err := db.QueryContext(ctx, ndbinfoThreadBlocksQuery)
	if err != nil {
		return err
	}
	defer ndbinfoThreadBlocksRows.Close()

	var (
		nodeID, thrNo, blockName, blockInstance string
	)

	for ndbinfoThreadBlocksRows.Next() {
		if err := ndbinfoThreadBlocksRows.Scan(
			&nodeID, &thrNo, &blockName, &blockInstance,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoThreadBlocksDesc, prometheus.UntypedValue, float64(1),
			nodeID, thrNo, blockName, blockInstance,
		)
	}

	return nil
}

// check interface
var _ Scraper = ScrapeNDBInfoThreadBlocks{}
