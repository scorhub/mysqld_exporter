// Copyright The Prometheus Authors
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

package collector

import (
	"context"
	"database/sql/driver"
	"errors"
	"maps"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/promslog"
)

var ndbinfoFilesQueryRegexp = "^" + regexp.QuoteMeta(strings.Join(strings.Fields(ndbinfoFilesQuery), " ")) + "$"

var ndbinfoFilesColumns = []string{
	"id", "type", "name", "parent", "parent_name",
	"free_extents", "total_extents", "extent_size_bytes",
	"initial_size", "maximum_size", "autoextend_size",
}

func scrapeNDBInfoFilesMetrics(ctx context.Context, inst *instance) ([]prometheus.Metric, error) {
	ch := make(chan prometheus.Metric)
	result := make(chan error, 1)

	go func() {
		result <- (ScrapeNDBInfoFiles{}).Scrape(ctx, inst, ch, promslog.NewNopLogger())
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	return metrics, <-result
}

func TestScrapeNDBInfoFiles(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// Log file groups and tablespaces report NULL in parent_name and every extent and
	// size column, so they must not stop the files listed after them from being scraped.
	rows := sqlmock.NewRows(ndbinfoFilesColumns).
		AddRow(1, "Log file group", "lg_1", 0, nil, nil, nil, nil, nil, nil, nil).
		AddRow(2, "Undo file", "undo_1.log", 1, "lg_1", 0, 0, 0, 134217728, 0, 0).
		AddRow(3, "Tablespace", "ts_1", 0, nil, nil, nil, nil, nil, nil, nil).
		AddRow(4, "Data file", "data_1.dat", 3, "ts_1", 30, 32, 1048576, 33554432, 0, 0).
		// Each column is checked on its own: a NULL in one does not drop the others.
		AddRow(5, "Data file", "data_2.dat", 3, "ts_1", 16, 16, 1048576, 16777216, nil, nil)
	mock.ExpectQuery(ndbinfoFilesQueryRegexp).WillReturnRows(rows).RowsWillBeClosed()

	metrics, err := scrapeNDBInfoFilesMetrics(t.Context(), &instance{db: db})
	if err != nil {
		t.Fatal(err)
	}

	undo := labelMap{"id": "2", "type": "Undo file", "name": "undo_1.log", "parent": "1", "parent_name": "lg_1"}
	data1 := labelMap{"id": "4", "type": "Data file", "name": "data_1.dat", "parent": "3", "parent_name": "ts_1"}
	data2 := labelMap{"id": "5", "type": "Data file", "name": "data_2.dat", "parent": "3", "parent_name": "ts_1"}
	with := func(labels labelMap, name, value string) labelMap {
		labels = maps.Clone(labels)
		labels[name] = value
		return labels
	}
	want := []struct {
		desc   *prometheus.Desc
		labels labelMap
		value  float64
	}{
		{ndbinfoFilesExtentsDesc, with(undo, "amount_type", "free"), 0},
		{ndbinfoFilesExtentsDesc, with(undo, "amount_type", "total"), 0},
		{ndbinfoFilesSizeBytesDesc, with(undo, "size_type", "extent"), 0},
		{ndbinfoFilesSizeBytesDesc, with(undo, "size_type", "initial"), 134217728},
		{ndbinfoFilesSizeBytesDesc, with(undo, "size_type", "maximum"), 0},
		{ndbinfoFilesSizeBytesDesc, with(undo, "size_type", "autoextend"), 0},
		{ndbinfoFilesExtentsDesc, with(data1, "amount_type", "free"), 30},
		{ndbinfoFilesExtentsDesc, with(data1, "amount_type", "total"), 32},
		{ndbinfoFilesSizeBytesDesc, with(data1, "size_type", "extent"), 1048576},
		{ndbinfoFilesSizeBytesDesc, with(data1, "size_type", "initial"), 33554432},
		{ndbinfoFilesSizeBytesDesc, with(data1, "size_type", "maximum"), 0},
		{ndbinfoFilesSizeBytesDesc, with(data1, "size_type", "autoextend"), 0},
		{ndbinfoFilesExtentsDesc, with(data2, "amount_type", "free"), 16},
		{ndbinfoFilesExtentsDesc, with(data2, "amount_type", "total"), 16},
		{ndbinfoFilesSizeBytesDesc, with(data2, "size_type", "extent"), 1048576},
		{ndbinfoFilesSizeBytesDesc, with(data2, "size_type", "initial"), 16777216},
	}
	if len(metrics) != len(want) {
		t.Fatalf("got %d metrics, want %d", len(metrics), len(want))
	}
	for i, w := range want {
		if got := metrics[i].Desc(); got != w.desc {
			t.Errorf("metric %d descriptor = %s, want %s", i, got, w.desc)
		}
		wantResult := MetricResult{labels: w.labels, value: w.value, metricType: dto.MetricType_GAUGE}
		if diff := cmp.Diff(wantResult, readMetric(metrics[i]), cmp.AllowUnexported(MetricResult{})); diff != "" {
			t.Errorf("metric %d mismatch (-want +got):\n%s", i, diff)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestScrapeNDBInfoFilesErrors(t *testing.T) {
	t.Parallel()

	queryErr := errors.New("query failed")
	iterationErr := errors.New("iteration failed")
	closeErr := errors.New("close failed")

	dataFile := []driver.Value{4, "Data file", "data_1.dat", 3, "ts_1", 30, 32, 1048576, 33554432, 0, 0}
	invalidFile := []driver.Value{5, "Data file", "data_2.dat", 3, "ts_1", "invalid", 32, 1048576, 33554432, 0, 0}

	tests := []struct {
		name      string
		rows      *sqlmock.Rows
		queryErr  error
		wantErr   error
		wantScan  bool
		wantCount int
	}{
		{
			name:     "query",
			queryErr: queryErr,
			wantErr:  queryErr,
		},
		{
			name:     "invalid free extents",
			rows:     sqlmock.NewRows(ndbinfoFilesColumns).AddRow(invalidFile...),
			wantScan: true,
		},
		{
			name: "iteration after a valid row",
			rows: sqlmock.NewRows(ndbinfoFilesColumns).
				AddRow(dataFile...).AddRow(dataFile...).RowError(1, iterationErr),
			wantErr:   iterationErr,
			wantCount: 6,
		},
		{
			name:    "close at end of rows",
			rows:    sqlmock.NewRows(ndbinfoFilesColumns).CloseError(closeErr),
			wantErr: closeErr,
		},
		{
			name:     "scan and close",
			rows:     sqlmock.NewRows(ndbinfoFilesColumns).AddRow(invalidFile...).CloseError(closeErr),
			wantErr:  closeErr,
			wantScan: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })

			query := mock.ExpectQuery(ndbinfoFilesQueryRegexp)
			if tt.queryErr != nil {
				query.WillReturnError(tt.queryErr)
			} else {
				query.WillReturnRows(tt.rows).RowsWillBeClosed()
			}

			metrics, err := scrapeNDBInfoFilesMetrics(t.Context(), &instance{db: db})
			if err == nil {
				t.Fatal("expected scrape error")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantScan && !strings.Contains(err.Error(), "Scan error") {
				t.Errorf("error = %v, want scan error", err)
			}
			if len(metrics) != tt.wantCount {
				t.Errorf("got %d metrics, want %d", len(metrics), tt.wantCount)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestScrapeNDBInfoFilesCanceled(t *testing.T) {
	t.Parallel()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	metrics, err := scrapeNDBInfoFilesMetrics(ctx, &instance{db: db})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
	if len(metrics) != 0 {
		t.Errorf("got %d metrics with a canceled context", len(metrics))
	}
}
