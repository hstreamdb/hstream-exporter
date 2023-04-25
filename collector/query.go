package collector

import (
	"github.com/hstreamdb/hstream-exporter/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	querySubsystem = "query"
)

type QueryMetrics struct {
	Metrics []scraper.Metrics
}

func NewQueryMetrics() *QueryMetrics {
	totalInputRecords := scraper.Metrics{
		Type: scraper.QueryTotalInputRecords,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, querySubsystem, scraper.QueryTotalInputRecords.String()),
			"Total number of records read from source.",
			[]string{"query_id", "server_host"}, nil,
		),
	}
	totalOutputRecords := scraper.Metrics{
		Type: scraper.QueryTotalOutputRecords,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, querySubsystem, scraper.QueryTotalOutputRecords.String()),
			"Total number of records write to sink.",
			[]string{"query_id", "server_host"}, nil,
		),
	}
	totalExecuteErrors := scraper.Metrics{
		Type: scraper.QueryTotalExecuteErrors,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, querySubsystem, scraper.QueryTotalExecuteErrors.String()),
			"Total number of query execute errors.",
			[]string{"query_id", "server_host"}, nil,
		),
	}
	return &QueryMetrics{
		Metrics: []scraper.Metrics{totalInputRecords, totalOutputRecords, totalExecuteErrors},
	}
}
