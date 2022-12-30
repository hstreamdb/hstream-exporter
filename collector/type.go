package collector

import (
	"fmt"
	"github.com/hstreamdb/hstreamdb-go/hstream"
)

const (
	getStatsCmd = "server stats %s %s -i %s"
)

type KeyType int

const (
	StreamName KeyType = iota
	SubscriptionId
	ServerSummary
	Unknown
)

func (k KeyType) String() string {
	switch k {
	case StreamName:
		return "stream_name"
	case SubscriptionId:
		return "subscription_id"
	case ServerSummary:
		return "server_host"
	case Unknown:
		return "Unknown"
	}
	return ""
}

type MetricType int

const (
	Gauge MetricType = iota
	Counter
	Summary
)

type HStreamMetrics interface {
	StatCmd(interval string) string
	GetMetricName() string
	GetMetricKey() string
	GetCounterStatsType() hstream.StatsType
}

type baseMetrics struct {
	Metric string
	Key    KeyType
}

func (b baseMetrics) GetMetricName() string {
	return b.Metric
}

func (b baseMetrics) GetMetricKey() string {
	return b.Key.String()
}

func (b baseMetrics) GetCounterStatsType() hstream.StatsType {
	panic("overload this function")
}

type StreamMetrics struct {
	baseMetrics
}

func NewStreamMetrics(metric string, key KeyType) StreamMetrics {
	return StreamMetrics{
		baseMetrics{metric, key},
	}
}

func (s StreamMetrics) StatCmd(interval string) string {
	return fmt.Sprintf(getStatsCmd, "stream", s.Metric, interval)
}

type StreamCounterMetrics struct {
	baseMetrics
}

func NewStreamCounterMetrics(metric string, key KeyType) StreamCounterMetrics {
	return StreamCounterMetrics{
		baseMetrics{metric, key},
	}
}

func (s StreamCounterMetrics) StatCmd(interval string) string {
	return fmt.Sprintf(getStatsCmd, "stream_counter", s.Metric, interval)
}

func (s StreamCounterMetrics) GetCounterStatsType() hstream.StatsType {
	var (
		statsType hstream.StreamStatsType
	)

	switch s.Metric {
	case "append_in_bytes":
		statsType = hstream.StreamAppendInBytes
	case "append_in_records":
		statsType = hstream.StreamAppendInRecords
	case "append_total":
		statsType = hstream.TotalAppend
	case "append_failed":
		statsType = hstream.FailedAppend
	}
	return statsType
}

type ServerHistogramMetrics struct {
	baseMetrics
}

func NewServerHistogramMetrics(metric string, key KeyType) ServerHistogramMetrics {
	return ServerHistogramMetrics{
		baseMetrics{metric, key},
	}
}

func (s ServerHistogramMetrics) StatCmd(interval string) string {
	return fmt.Sprintf(getStatsCmd, "server_histogram", s.Metric, interval)
}

type SubscriptionMetrics struct {
	baseMetrics
}

func NewSubscriptionMetrics(metric string, key KeyType) SubscriptionMetrics {
	return SubscriptionMetrics{
		baseMetrics{metric, key},
	}
}

func (s SubscriptionMetrics) StatCmd(interval string) string {
	return fmt.Sprintf(getStatsCmd, "subscription", s.Metric, interval)
}

type SubscriptionCounterMetrics struct {
	baseMetrics
}

func NewSubscriptionCounterMetrics(metric string, key KeyType) SubscriptionCounterMetrics {
	return SubscriptionCounterMetrics{
		baseMetrics{metric, key},
	}
}

func (s SubscriptionCounterMetrics) StatCmd(interval string) string {
	return fmt.Sprintf(getStatsCmd, "subscription_counter", s.Metric, interval)
}

func (s SubscriptionCounterMetrics) GetCounterStatsType() hstream.StatsType {
	var (
		statsType hstream.SubscriptionStatsType
	)

	switch s.Metric {
	case "send_out_bytes":
		statsType = hstream.SubDeliveryInBytes
	case "acks":
		statsType = hstream.AckReceived
	case "resend_records":
		statsType = hstream.ResendRecords
	case "send_out_records":
		statsType = hstream.SubDeliveryInRecords
	case "request_messages":
		statsType = hstream.SubMessageRequestCnt
	case "response_messages":
		statsType = hstream.SubMessageResponseCnt
	}
	return statsType
}
