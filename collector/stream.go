package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	streamSubsystem = "stream"
)

type StreamCollector struct {
	metrics []Metrics
	url     string
}

func NewStreamCollector(serverUrl string) (Collector, error) {
	appendBytes := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_in_bytes"),
			"Successfully written bytes to the stream in seconds.",
			[]string{"stream", "server_host"}, nil,
		),
		reqPath:    fmt.Sprintf(requestTemplate, "stream", "append_in_bytes"),
		mainKey:    StreamName,
		metricType: Gauge,
	}
	appendRecords := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_in_records"),
			"Successfully written records to the stream in seconds.",
			[]string{"stream", "server_host"}, nil,
		),
		reqPath:    fmt.Sprintf(requestTemplate, "stream", "append_in_records"),
		mainKey:    StreamName,
		metricType: Gauge,
	}
	appendQPS := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_qps"),
			"Rate of append requests received per stream.",
			[]string{"stream", "server_host"}, nil,
		),
		reqPath:    fmt.Sprintf(requestTemplate, "stream", "append_in_requests"),
		mainKey:    StreamName,
		metricType: Gauge,
	}
	appendTotal := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_total"),
			"Total number of append requests of a stream.",
			[]string{"stream", "server_host"}, nil,
		),
		reqPath:    fmt.Sprintf(requestTemplate, "stream_counter", "append_total"),
		mainKey:    StreamName,
		metricType: Counter,
	}
	appendFailed := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_failed"),
			"Total number of failed append requests of a stream.",
			[]string{"stream", "server_host"}, nil,
		),
		reqPath:    fmt.Sprintf(requestTemplate, "stream_counter", "append_failed"),
		mainKey:    StreamName,
		metricType: Counter,
	}
	return &StreamCollector{
		metrics: []Metrics{appendBytes, appendRecords, appendQPS, appendTotal, appendFailed},
		url:     serverUrl,
	}, nil
}

func (s *StreamCollector) CollectorName() string {
	return "StreamCollector"
}

func (s *StreamCollector) Collect(ch chan<- prometheus.Metric) error {
	return ScrapeHServerMetrics(ch, s.metrics, s.url)
}
