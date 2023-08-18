package collector

import (
	"github.com/hstreamdb/hstream-exporter/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	subSubsystem = "subscription"
)

type SubscriptionMetrics struct {
	Metrics []scraper.Metrics
}

func NewSubscriptionMetrics() *SubscriptionMetrics {
	sendBytes := scraper.Metrics{
		Type: scraper.SubSendOutBytes,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, scraper.SubSendOutBytes.String()),
			"Bytes send by each subscription.",
			[]string{"subId", "server_host"}, nil,
		),
	}
	sendRecords := scraper.Metrics{
		Type: scraper.SubSendOutRecords,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, scraper.SubSendOutRecords.String()),
			"Records send by each subscription.",
			[]string{"subId", "server_host"}, nil,
		),
	}
	sendRecordsFailed := scraper.Metrics{
		Type: scraper.SubSendOutRecordsFailed,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, scraper.SubSendOutRecordsFailed.String()),
			"Records send failed by each subscription.",
			[]string{"subId", "server_host"}, nil,
		),
	}
	acks := scraper.Metrics{
		Type: scraper.SubReceivedAcks,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, scraper.SubReceivedAcks.String()),
			"Acknowledgements received per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
	}
	resendRecords := scraper.Metrics{
		Type: scraper.SubResendRecords,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, scraper.SubResendRecords.String()),
			"Total number of resent records per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
	}
	resendRecordsFailed := scraper.Metrics{
		Type: scraper.SubResendRecordsFailed,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, scraper.SubResendRecordsFailed.String()),
			"Total number of failed resent records per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
	}
	msgRequestRate := scraper.Metrics{
		Type: scraper.SubRequestMessages,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, scraper.SubRequestMessages.String()),
			"Requests received from clients per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
	}
	msgResponseRate := scraper.Metrics{
		Type: scraper.SubResponseMessages,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subSubsystem, scraper.SubResponseMessages.String()),
			"Response sent to clients per subscription.",
			[]string{"subId", "server_host"}, nil,
		),
	}
	//checkListSize := scraper.Metrics{
	//	Type: scraper.SubCheckListSize,
	//	Metric: prometheus.NewDesc(
	//		prometheus.BuildFQName(namespace, subSubsystem, scraper.SubCheckListSize.String()),
	//		"Checklist size per subscription.",
	//		[]string{"subId", "server_host"}, nil,
	//	),
	//}

	return &SubscriptionMetrics{
		Metrics: []scraper.Metrics{sendBytes, sendRecords, acks,
			//sendRecordsFailed, resendRecords, resendRecordsFailed, msgRequestRate, msgResponseRate, checkListSize},
			sendRecordsFailed, resendRecords, resendRecordsFailed, msgRequestRate, msgResponseRate},
	}
}
