package collector

import (
	"github.com/hstreamdb/hstreamdb-go/hstream"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

const (
	namespace = "hstream_exporter"
)

var (
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "success_scrape_count"),
		"hstream_exporter: Number of times the target state was successfully scraped",
		[]string{"collector"},
		nil,
	)
	scrapeFailedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "failed_scrape_count"),
		"hstream_exporter: Number of times the target state was failed scraped",
		[]string{"collector"},
		nil,
	)
)

// Collector is an interface which a collector need to implement
type Collector interface {
	CollectorName() string
	// Collect Get new metrics and expose them via prometheus registry.
	Collect(ch chan<- prometheus.Metric) (successCnt uint32, failedCnt uint32)
}

type Metrics struct {
	metric        *prometheus.Desc
	hstreamMetric HStreamMetrics
	metricType    MetricType
}

// HStreamCollector implements the prometheus.Collector interface
type HStreamCollector struct {
	Collectors map[string]Collector
}

type newCollectorFunc func(*hstream.HStreamClient, []string) (Collector, error)

// collectorRegister is a collector builder which register all needed metrics,
// then build a HStreamCollector instance.
type collectorRegister struct {
	urls       []string
	client     *hstream.HStreamClient
	collectors map[string]Collector
	err        error
}

func newRegister(client *hstream.HStreamClient, urls []string) *collectorRegister {
	return &collectorRegister{
		urls:       urls,
		client:     client,
		collectors: make(map[string]Collector),
	}
}

func (r *collectorRegister) register(f newCollectorFunc) {
	if r.err != nil {
		return
	}

	collector, err := f(r.client, r.urls)
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
	client, err := hstream.NewHStreamClient(serverUrl)
	if err != nil {
		return nil, errors.WithMessage(err, "Create HStream client error")
	}

	urls, err := client.GetServerInfo()
	if err != nil {
		return nil, errors.WithMessage(err, "Get server info error")
	}

	rg := newRegister(client, urls)
	for _, fc := range []newCollectorFunc{NewStreamCollector, NewSubCollector} {
		rg.register(fc)
	}

	return rg.finish()
}

// Describe implement prometheus.Collector interface
func (h *HStreamCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeSuccessDesc
	ch <- scrapeFailedDesc
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
	success, faild := c.Collect(ch)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, float64(success), name)
	ch <- prometheus.MustNewConstMetric(scrapeFailedDesc, prometheus.GaugeValue, float64(faild), name)
}
