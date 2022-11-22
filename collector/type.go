package collector

import "fmt"

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
