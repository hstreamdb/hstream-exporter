package collector

import (
	"github.com/hstreamdb/hstream-exporter/scraper"
	"github.com/hstreamdb/hstreamdb-go/hstream"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
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
)

// HStreamCollector implements the prometheus.Collector interface
type HStreamCollector struct {
	TargetUrls    []string
	StreamMetrics *StreamMetrics
	SubMetrics    *SubscriptionMetrics
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
	registry.MustRegister(scrapeLatencyDesc)
	return &HStreamCollector{
		TargetUrls:    urls,
		StreamMetrics: NewStreamMetrics(),
		SubMetrics:    NewSubscriptionMetrics(),
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

	scrapeLatencyDesc.WithLabelValues(target).Observe(float64(diff.Milliseconds()))

	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, float64(success), target)
	ch <- prometheus.MustNewConstMetric(scrapeFailedDesc, prometheus.GaugeValue, float64(faild), target)
}
