package collector

import (
	"github.com/hstreamdb/hstream-exporter/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	streamSubsystem = "stream"
)

type StreamMetrics struct {
	Metrics []scraper.Metrics
}

func NewStreamMetrics() *StreamMetrics {
	appendInBytes := scraper.Metrics{
		Type: scraper.StreamAppendInBytes,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, scraper.StreamAppendInBytes.String()),
			"Successfully written bytes to the stream.",
			[]string{"stream", "server_host"}, nil,
		),
	}
	appendInRecords := scraper.Metrics{
		Type: scraper.StreamAppendInReccords,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, scraper.StreamAppendInReccords.String()),
			"Successfully written records to the stream.",
			[]string{"stream", "server_host"}, nil,
		),
	}
	appendTotal := scraper.Metrics{
		Type: scraper.StreamAppendTotal,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, scraper.StreamAppendTotal.String()),
			"Number of append requests of a stream.",
			[]string{"stream", "server_host"}, nil,
		),
	}
	appendFailed := scraper.Metrics{
		Type: scraper.StreamAppendFailed,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, scraper.StreamAppendFailed.String()),
			"Number of failed append requests of a stream.",
			[]string{"stream", "server_host"}, nil,
		),
	}
	appendRequestLatency := scraper.Metrics{
		Type: scraper.StreamAppendLatency,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, scraper.StreamAppendLatency.String()),
			"Append request latency.",
			[]string{"server_host"}, nil,
		),
	}
	return &StreamMetrics{
		Metrics: []scraper.Metrics{appendInBytes, appendInRecords, appendTotal, appendFailed, appendRequestLatency},
	}
}
