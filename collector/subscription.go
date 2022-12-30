package collector

import (
	"github.com/hstreamdb/hstreamdb-go/hstream"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	subSubsystem = "subscription"
)

type SubCollector struct {
	client     *hstream.HStreamClient
	metrics    []Metrics
	serverUrls []string
}

func NewSubCollector(client *hstream.HStreamClient, serverUrls []string) (Collector, error) {
	sendBytes := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "send_bytes"),
			"Bytes send by each subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		hstreamMetric: NewSubscriptionCounterMetrics("send_out_bytes", SubscriptionId),
		metricType:    Counter,
	}
	sendRecords := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "send_records"),
			"Records send by each subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		hstreamMetric: NewSubscriptionCounterMetrics("send_out_records", SubscriptionId),
		metricType:    Counter,
	}
	msgRequestRate := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "request_messages"),
			"Requests received from clients per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		hstreamMetric: NewSubscriptionCounterMetrics("request_messages", SubscriptionId),
		metricType:    Counter,
	}
	msgResponseRate := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "response_messages"),
			"Response sent to clients per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		hstreamMetric: NewSubscriptionCounterMetrics("response_messages", SubscriptionId),
		metricType:    Counter,
	}
	acks := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "acks"),
			"Acknowledgements received per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		hstreamMetric: NewSubscriptionCounterMetrics("acks", SubscriptionId),
		metricType:    Counter,
	}
	resendRate := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "resend_records"),
			"Total number of resent records per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		hstreamMetric: NewSubscriptionCounterMetrics("resend_records", SubscriptionId),
		metricType:    Counter,
	}

	return &SubCollector{
		client:     client,
		metrics:    []Metrics{sendBytes, sendRecords, msgRequestRate, msgResponseRate, acks, resendRate},
		serverUrls: serverUrls,
	}, nil
}

func (s *SubCollector) CollectorName() string {
	return "SubCollector"
}

func (s *SubCollector) Collect(ch chan<- prometheus.Metric) (uint32, uint32) {
	return ScrapeHServerMetrics(ch, s.client, s.metrics, s.serverUrls)
}
