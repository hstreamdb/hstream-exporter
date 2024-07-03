package collector

import (
	"github.com/hstreamdb/hstream-exporter/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	healthyCheckerSubsystem = "healthyChecker"
)

type HealthyCheckerMetrics struct {
	Metrics []scraper.Metrics
}

func NewHealthyCheckerMetrics() *HealthyCheckerMetrics {
	checkStoreClusterLatency := scraper.Metrics{
		Type: scraper.CheckStoreClusterLatency,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, healthyCheckerSubsystem, scraper.CheckStoreClusterLatency.String()),
			"Check store cluster healthy latency.",
			[]string{"server_host"}, nil,
		),
	}
	checkMetaClusterLatency := scraper.Metrics{
		Type: scraper.CheckMetaClusterLatency,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, healthyCheckerSubsystem, scraper.CheckMetaClusterLatency.String()),
			"Check meta cluster healthy latency.",
			[]string{"server_host"}, nil,
		),
	}
	return &HealthyCheckerMetrics{
		Metrics: []scraper.Metrics{checkStoreClusterLatency, checkMetaClusterLatency},
	}
}
