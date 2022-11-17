package collector

import (
	"github.com/hstreamdb/hstream-exporter/util"
	"github.com/hstreamdb/hstreamdb-go/hstream"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	streamSubsystem = "stream"
)

type StreamCollector struct {
	metrics    []Metrics
	initUrl    string
	initClient *hstream.HStreamClient
	serverUrls []string
	clients    []*hstream.HStreamClient
}

func NewStreamCollector(serverUrl string) (Collector, error) {
	appendBytes := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_in_bytes"),
			"Successfully written bytes to the stream in seconds.",
			[]string{"stream", "server_host"}, nil,
		),
		reqArgs:    []string{"stream", "append_in_bytes"},
		mainKey:    StreamName,
		metricType: Gauge,
	}
	appendRecords := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_in_records"),
			"Successfully written records to the stream in seconds.",
			[]string{"stream", "server_host"}, nil,
		),
		reqArgs:    []string{"stream", "append_in_records"},
		mainKey:    StreamName,
		metricType: Gauge,
	}
	appendQPS := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_qps"),
			"Rate of append requests received per stream.",
			[]string{"stream", "server_host"}, nil,
		),
		reqArgs:    []string{"stream", "append_in_requests"},
		mainKey:    StreamName,
		metricType: Gauge,
	}
	appendTotal := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_total"),
			"Total number of append requests of a stream.",
			[]string{"stream", "server_host"}, nil,
		),
		reqArgs:    []string{"stream_counter", "append_total"},
		mainKey:    StreamName,
		metricType: Counter,
	}
	appendFailed := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_failed"),
			"Total number of failed append requests of a stream.",
			[]string{"stream", "server_host"}, nil,
		),
		reqArgs:    []string{"stream_counter", "append_failed"},
		mainKey:    StreamName,
		metricType: Counter,
	}
	appendRequestLatency := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, streamSubsystem, "append_latency"),
			"Append request latency.",
			[]string{"server_host"}, nil,
		),
		reqArgs:    []string{"server_histogram", "append_request_latency"},
		mainKey:    ServerHistogram,
		metricType: Summary,
	}

	client, err := hstream.NewHStreamClient(serverUrl)
	if err != nil {
		return nil, err
	}
	serverUrls, err := client.Client.GetServerInfo()
	if err != nil {
		return nil, err
	}

	clients := make([]*hstream.HStreamClient, len(serverUrls))
	for _, serverUrl := range serverUrls {
		if client, err := hstream.NewHStreamClient(serverUrl); err == nil {
			clients = append(clients, client)
		} else {
			util.Logger().Warn("connect to server " + serverUrl + " error, " + err.Error())
		}
	}

	return &StreamCollector{
		metrics:    []Metrics{appendBytes, appendRecords, appendQPS, appendTotal, appendFailed, appendRequestLatency},
		initUrl:    serverUrl,
		initClient: client,
		serverUrls: serverUrls,
		clients:    clients,
	}, nil
}

func (s *StreamCollector) CollectorName() string {
	return "StreamCollector"
}

func (s *StreamCollector) Collect(ch chan<- prometheus.Metric) error {
	return ScrapeHServerMetrics(ch, s.metrics, s.serverUrls)
}
