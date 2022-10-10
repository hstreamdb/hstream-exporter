package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	subSubsystem = "subscription"
)

type SubCollector struct {
	metrics []Metrics
	url     string
}

func NewSubCollector(serverUrl string) (Collector, error) {
	sendBytes := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "send_bytes"),
			"Bytes send by each subscription in seconds.",
			[]string{"subId", "server_host"}, nil,
		),
		reqPath:    fmt.Sprintf(requestTemplate, "subscription", "send_out_bytes"),
		mainKey:    SubscriptionId,
		metricType: Gauge,
	}
	msgRequestRate := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "request_messages"),
			"Rate of requests received from clients per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		reqPath:    fmt.Sprintf(requestTemplate, "subscription", "request_messages"),
		mainKey:    SubscriptionId,
		metricType: Gauge,
	}
	msgResponseRate := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "response_messages"),
			"Rate of response sent to clients per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		reqPath:    fmt.Sprintf(requestTemplate, "subscription", "response_messages"),
		mainKey:    SubscriptionId,
		metricType: Gauge,
	}
	acks := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "acks"),
			"Rate of acknowledgements received per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		reqPath:    fmt.Sprintf(requestTemplate, "subscription", "acks"),
		mainKey:    SubscriptionId,
		metricType: Gauge,
	}
	resendRate := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "resend_records"),
			"Total number of resent records per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		reqPath:    fmt.Sprintf(requestTemplate, "subscription_counter", "resend_records"),
		mainKey:    SubscriptionId,
		metricType: Gauge,
	}
	return &SubCollector{
		metrics: []Metrics{sendBytes, msgRequestRate, msgResponseRate, acks, resendRate},
		url:     serverUrl,
	}, nil
}

func (s *SubCollector) CollectorName() string {
	return "SubCollector"
}

func (s *SubCollector) Collect(ch chan<- prometheus.Metric) error {
	return ScrapeHServerMetrics(ch, s.metrics, s.url)
}
