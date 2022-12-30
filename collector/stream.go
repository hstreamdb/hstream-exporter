package collector

import (
	"github.com/hstreamdb/hstreamdb-go/hstream"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	streamSubsystem = "stream"
)

type StreamCollector struct {
	client     *hstream.HStreamClient
	metrics    []Metrics
	serverUrls []string
}

func NewStreamCollector(client *hstream.HStreamClient, serverUrls []string) (Collector, error) {
	appendBytes := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_in_bytes"),
			"Successfully written bytes to the stream.",
			[]string{"stream", "server_host"}, nil,
		),
		hstreamMetric: NewStreamCounterMetrics("append_in_bytes", StreamName),
		metricType:    Counter,
	}
	appendRecords := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_in_records"),
			"Successfully written records to the stream.",
			[]string{"stream", "server_host"}, nil,
		),
		hstreamMetric: NewStreamCounterMetrics("append_in_records", StreamName),
		metricType:    Counter,
	}
	//appendQPS := Metrics{
	//	metric: prometheus.NewDesc(
	//		prometheus.BuildFQName(namespace, streamSubsystem, "append_qps"),
	//		"Rate of append requests received per stream.",
	//		[]string{"stream", "server_host"}, nil,
	//	),
	//	hstreamMetric: NewStreamCounterMetrics("append_in_requests", StreamName),
	//	metricType:    Counter,
	//}
	appendTotal := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_total"),
			"Number of append requests of a stream.",
			[]string{"stream", "server_host"}, nil,
		),
		hstreamMetric: NewStreamCounterMetrics("append_total", StreamName),
		metricType:    Counter,
	}
	appendFailed := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_failed"),
			"Number of failed append requests of a stream.",
			[]string{"stream", "server_host"}, nil,
		),
		hstreamMetric: NewStreamCounterMetrics("append_failed", StreamName),
		metricType:    Counter,
	}
	appendRequestLatency := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_latency"),
			"Append request latency.",
			[]string{"server_host"}, nil,
		),
		hstreamMetric: NewServerHistogramMetrics("append_request_latency", ServerSummary),
		metricType:    Summary,
	}

	return &StreamCollector{
		client:     client,
		metrics:    []Metrics{appendBytes, appendRecords, appendTotal, appendFailed, appendRequestLatency},
		serverUrls: serverUrls,
	}, nil
}

func (s *StreamCollector) CollectorName() string {
	return "StreamCollector"
}

func (s *StreamCollector) Collect(ch chan<- prometheus.Metric) (uint32, uint32) {
	return ScrapeHServerMetrics(ch, s.client, s.metrics, s.serverUrls)
}
