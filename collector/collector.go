package collector

import (
	"github.com/hstreamdb/hstream-exporter/util"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"sync"
)

const (
	namespace       = "hstream_exporter"
	requestTemplate = "v1/cluster/stats?category=%s&metrics=%s&interval=5s"
)

var (
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "target_up"),
		"hstream_exporter: Whether a collector for hstream server succeeded.",
		[]string{"collector"},
		nil,
	)
)

// Collector is an interface which a collector need to implement
type Collector interface {
	CollectorName() string
	// Collect Get new metrics and expose them via prometheus registry.
	Collect(ch chan<- prometheus.Metric) error
}

type Metrics struct {
	metric *prometheus.Desc

	reqPath    string
	mainKey    KeyType
	metricType MetricType
}

// HStreamCollector implements the prometheus.Collector interface
type HStreamCollector struct {
	Collectors map[string]Collector
}

type newCollectorFunc func(string) (Collector, error)

// collectorRegister is a collector builder which register all needed metrics,
// then build a HStreamCollector instance.
type collectorRegister struct {
	url        string
	collectors map[string]Collector
	err        error
}

func newRegister(url string) *collectorRegister {
	return &collectorRegister{
		url:        url,
		collectors: make(map[string]Collector),
	}
}

func (r *collectorRegister) register(f newCollectorFunc) {
	if r.err != nil {
		return
	}

	collector, err := f(r.url)
	if err != nil {
		r.err = err
		return
	}
	r.collectors[collector.CollectorName()] = collector
}

func (r *collectorRegister) finish() (*HStreamCollector, error) {
	if r.err != nil {
		return nil, r.err
	}
	return &HStreamCollector{r.collectors}, nil
}

func NewHStreamCollector(serverUrl string) (*HStreamCollector, error) {
	rg := newRegister(serverUrl)
	for _, fc := range []newCollectorFunc{NewStreamCollector, NewSubCollector} {
		rg.register(fc)
	}

	return rg.finish()
}

// Describe implement prometheus.Collector interface
func (h *HStreamCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeSuccessDesc
}

// Collect implement prometheus.Collector interface
func (h *HStreamCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(h.Collectors))
	for name, c := range h.Collectors {
		go func(name string, c Collector) {
			execute(name, c, ch)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func execute(name string, c Collector, ch chan<- prometheus.Metric) {
	success := float64(1)
	if err := c.Collect(ch); err != nil {
		util.Logger().Error("collector error", zap.String("name", name), zap.String("error", err.Error()))
		success = 0
	}
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}

type KeyType int

const (
	StreamName KeyType = iota
	SubscriptionId
	ServerHistogram
	Unknown
)

func (k KeyType) String() string {
	switch k {
	case StreamName:
		return "stream_name"
	case SubscriptionId:
		return "subscription_id"
	case ServerHistogram:
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
