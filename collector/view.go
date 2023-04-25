package collector

import (
	"github.com/hstreamdb/hstream-exporter/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	viewSubsystem = "view"
)

type ViewMetrics struct {
	Metrics []scraper.Metrics
}

func NewViewMetrics() *ViewMetrics {
	totalExecuteQueries := scraper.Metrics{
		Type: scraper.ViewTotalExecuteQueries,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, viewSubsystem, scraper.ViewTotalExecuteQueries.String()),
			"Total execute queries in view.",
			[]string{"view_id", "server_host"}, nil,
		),
	}
	return &ViewMetrics{
		Metrics: []scraper.Metrics{totalExecuteQueries},
	}
}
