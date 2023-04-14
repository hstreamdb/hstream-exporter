package collector

import (
	"github.com/hstreamdb/hstream-exporter/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	connectorSubsystem = "connector"
)

type ConnectorMetrics struct {
	Metrics []scraper.Metrics
}

func NewConnectorMetrics() *ConnectorMetrics {
	deliveredInBytes := scraper.Metrics{
		Type: scraper.ConnectorDeliveredInBytes,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, connectorSubsystem, scraper.ConnectorDeliveredInBytes.String()),
			"Connector successfully delivered in bytes.",
			[]string{"task", "server_host"}, nil,
		),
	}
	deliveredInRecords := scraper.Metrics{
		Type: scraper.ConnectorDeliveredInRecords,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, connectorSubsystem, scraper.ConnectorDeliveredInRecords.String()),
			"Connector successfully delivered in records.",
			[]string{"task", "server_host"}, nil,
		),
	}
	return &ConnectorMetrics{
		Metrics: []scraper.Metrics{deliveredInBytes, deliveredInRecords},
	}
}
