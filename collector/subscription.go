package collector

import (
	"github.com/hstreamdb/hstream-exporter/util"
	"github.com/hstreamdb/hstreamdb-go/hstream"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	subSubsystem = "subscription"
)

type SubCollector struct {
	metrics    []Metrics
	initUrl    string
	initClient *hstream.HStreamClient
	serverUrls []string
	clients    []*hstream.HStreamClient
}

func NewSubCollector(serverUrl string) (Collector, error) {
	sendBytes := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "send_bytes"),
			"Bytes send by each subscription in seconds.",
			[]string{"subId", "server_host"}, nil,
		),
		reqArgs:    []string{"subscription", "send_out_bytes"},
		mainKey:    SubscriptionId,
		metricType: Gauge,
	}
	msgRequestRate := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "request_messages"),
			"Rate of requests received from clients per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		reqArgs:    []string{"subscription", "request_messages"},
		mainKey:    SubscriptionId,
		metricType: Gauge,
	}
	msgResponseRate := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "response_messages"),
			"Rate of response sent to clients per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		reqArgs:    []string{"subscription", "response_messages"},
		mainKey:    SubscriptionId,
		metricType: Gauge,
	}
	acks := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "acks"),
			"Rate of acknowledgements received per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		reqArgs:    []string{"subscription", "acks"},
		mainKey:    SubscriptionId,
		metricType: Gauge,
	}
	resendRate := Metrics{
		metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, "resend_records"),
			"Total number of resent records per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
		reqArgs:    []string{"subscription_counter", "resend_records"},
		mainKey:    SubscriptionId,
		metricType: Gauge,
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

	return &SubCollector{
		metrics:    []Metrics{sendBytes, msgRequestRate, msgResponseRate, acks, resendRate},
		initUrl:    serverUrl,
		initClient: client,
		serverUrls: serverUrls,
		clients:    clients,
	}, nil
}

func (s *SubCollector) CollectorName() string {
	return "SubCollector"
}

func (s *SubCollector) Collect(ch chan<- prometheus.Metric) error {
	return ScrapeHServerMetrics(ch, s.metrics, s.serverUrls)
}
