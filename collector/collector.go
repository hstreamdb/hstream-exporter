package collector

import (
	"fmt"
	"github.com/hstreamdb/hstream-exporter/scraper"
	"github.com/hstreamdb/hstream-exporter/util"
	"github.com/hstreamdb/hstreamdb-go/hstream"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
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
		[]string{"server_host"},
		nil,
	)
	scrapeLatencyDesc = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hstream_exporter_scrape_latency",
			Help:    "Histogram for per scrape latency.",
			Buckets: prometheus.LinearBuckets(0, 10, 10),
		},
		[]string{"server_host"},
	)

	totalSuccessedScrap = atomic.Uint64{}
	totalFailedScrap    = atomic.Uint64{}
)

// HStreamCollector implements the prometheus.Collector interface
type HStreamCollector struct {
	TargetUrls    []string
	StreamMetrics *StreamMetrics
	SubMetrics    *SubscriptionMetrics
	ConnMetrics   *ConnectorMetrics
	scraper       scraper.Scrape
}

func NewHStreamCollector(serverUrl string, registry *prometheus.Registry) (*HStreamCollector, error) {
	client, err := hstream.NewHStreamClient(serverUrl)
	if err != nil {
		return nil, errors.WithMessage(err, "Create HStream client error")
	}

	urls, err := client.GetServerInfo()
	if err != nil {
		return nil, errors.WithMessage(err, "Get server info error")
	}
	util.Logger().Info("Get server urls", zap.String("urls", fmt.Sprintf("%v", urls)))
	registry.MustRegister(scrapeLatencyDesc)
	return &HStreamCollector{
		TargetUrls:    urls,
		StreamMetrics: NewStreamMetrics(),
		SubMetrics:    NewSubscriptionMetrics(),
		ConnMetrics:   NewConnectorMetrics(),
		scraper:       scraper.NewScraper(client),
	}, nil
}

func (h *HStreamCollector) getScrapedMetrics() []scraper.Metrics {
	metrics := []scraper.Metrics{}
	for _, m := range h.StreamMetrics.Metrics {
		metrics = append(metrics, m)
	}
	for _, m := range h.SubMetrics.Metrics {
		metrics = append(metrics, m)
	}
	for _, m := range h.ConnMetrics.Metrics {
		metrics = append(metrics, m)
	}
	return metrics
}

// Describe implement prometheus.Collector interface
func (h *HStreamCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeSuccessDesc
	ch <- scrapeFailedDesc
}

// Collect implement prometheus.Collector interface
func (h *HStreamCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(h.TargetUrls))
	metrics := h.getScrapedMetrics()
	for _, u := range h.TargetUrls {
		go func(url string) {
			defer wg.Done()
			execute(h.scraper, metrics, url, ch)
		}(u)
	}
	wg.Wait()
}

func execute(scraper scraper.Scrape, metrics []scraper.Metrics, target string, ch chan<- prometheus.Metric) {
	start := time.Now()
	success, faild := scraper.Scrape(target, metrics, ch)
	diff := time.Now().Sub(start)
	util.Logger().Debug("Scrape target done", zap.String("url", target),
		zap.Int64("milliseconds latency", diff.Milliseconds()),
		zap.Int32("success request", success),
		zap.Int32("failed request", faild))

	scrapeLatencyDesc.WithLabelValues(target).Observe(float64(diff.Milliseconds()))
	totalSuccessedScrap.Add(uint64(success))
	totalFailedScrap.Add(uint64(faild))

	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.CounterValue, float64(totalSuccessedScrap.Load()), target)
	ch <- prometheus.MustNewConstMetric(scrapeFailedDesc, prometheus.CounterValue, float64(totalFailedScrap.Load()), target)
}
