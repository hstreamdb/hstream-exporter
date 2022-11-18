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
			"Bytes send by each subscription in seconds.",
			[]string{"subId", "server_host"}, nil,
		),
		hstreamMetric: NewSubscriptionMetrics("send_out_bytes", SubscriptionId),
		metricType:    Gauge,
	}
	msgRequestRate := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "request_messages"),
			"Rate of requests received from clients per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		hstreamMetric: NewSubscriptionMetrics("request_messages", SubscriptionId),
		metricType:    Gauge,
	}
	msgResponseRate := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "response_messages"),
			"Rate of response sent to clients per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		hstreamMetric: NewSubscriptionMetrics("response_messages", SubscriptionId),
		metricType:    Gauge,
	}
	acks := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "acks"),
			"Rate of acknowledgements received per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		hstreamMetric: NewSubscriptionMetrics("acks", SubscriptionId),
		metricType:    Gauge,
	}
	resendRate := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "resend_records"),
			"Total number of resent records per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		hstreamMetric: NewSubscriptionCounterMetrics("resend_records", SubscriptionId),
		metricType:    Gauge,
	}

	return &SubCollector{
		client:     client,
		metrics:    []Metrics{sendBytes, msgRequestRate, msgResponseRate, acks, resendRate},
		serverUrls: serverUrls,
	}, nil
}

func (s *SubCollector) CollectorName() string {
	return "SubCollector"
}

func (s *SubCollector) Collect(ch chan<- prometheus.Metric) (uint32, uint32) {
	return ScrapeHServerMetrics(ch, s.client, s.metrics, s.serverUrls)
}
