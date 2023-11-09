package collector

import (
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hstreamdb/hstream-exporter/scraper"
	"github.com/hstreamdb/hstream-exporter/util"
	"github.com/hstreamdb/hstreamdb-go/hstream"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
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
	StreamMetrics        *StreamMetrics
	SubMetrics           *SubscriptionMetrics
	ConnMetrics          *ConnectorMetrics
	QueryMetrics         *QueryMetrics
	ViewMetrics          *ViewMetrics
	scraper              scraper.Scrape
	serverUpdateDuration time.Duration

	client *hstream.HStreamClient

	// The following fields are protected by the lock
	lock       sync.RWMutex
	TargetUrls []string
}

func (h *HStreamCollector) getServerInfo() {
	ticker := time.NewTicker(h.serverUpdateDuration)
	defer func() {
		util.Logger().Info("exit get server info loop.")
		ticker.Stop()
	}()

	util.Logger().Info("start get server info loop.", zap.String("duration", h.serverUpdateDuration.String()))

	for range ticker.C {
		urls, err := h.client.GetServerInfo()
		if err != nil {
			util.Logger().Error("get server info return error", zap.String("error", err.Error()))
			continue
		}

		h.lock.Lock()
		h.TargetUrls = urls
		util.Logger().Debug("get server info", zap.String("urls", fmt.Sprintf("%+v", urls)))
		h.lock.Unlock()
	}
}

func NewHStreamCollector(serverUrl string, caPath string, token string, duration int, registry *prometheus.Registry) (*HStreamCollector, error) {
	var (
		client *hstream.HStreamClient
		err    error
	)

	authOpts := []hstream.AuthOpts{}
	if len(token) != 0 {
		authOpts = append(authOpts, hstream.WithAuthToken(token))
	}
	if len(caPath) != 0 {
		authOpts = append(authOpts, hstream.WithCaCert(caPath))
	}

	client, err = hstream.NewHStreamClient(serverUrl, authOpts...)

	if err != nil {
		return nil, errors.WithMessage(err, "Create HStream client error")
	}

	client.SetLogLevel(zap.WarnLevel)

	urls, err := client.GetServerInfo()
	if err != nil {
		return nil, errors.WithMessage(err, "Get server info error")
	}

	util.Logger().Info("Get server urls", zap.String("urls", fmt.Sprintf("%v", urls)))
	registry.MustRegister(scrapeLatencyDesc)
	collector := &HStreamCollector{
		TargetUrls:           urls,
		StreamMetrics:        NewStreamMetrics(),
		SubMetrics:           NewSubscriptionMetrics(),
		ConnMetrics:          NewConnectorMetrics(),
		QueryMetrics:         NewQueryMetrics(),
		ViewMetrics:          NewViewMetrics(),
		scraper:              scraper.NewScraper(client),
		serverUpdateDuration: time.Duration(duration) * time.Second,
		client:               client,
	}
	go collector.getServerInfo()

	return collector, nil
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
	for _, m := range h.QueryMetrics.Metrics {
		metrics = append(metrics, m)
	}
	for _, m := range h.ViewMetrics.Metrics {
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
	metrics := h.getScrapedMetrics()
	h.lock.RLock()
	wg.Add(len(h.TargetUrls))
	for _, u := range h.TargetUrls {
		go func(url string) {
			defer wg.Done()
			h.execute(metrics, url, ch)
		}(u)
	}
	h.lock.RUnlock()
	wg.Wait()
}

func (h *HStreamCollector) execute(metrics []scraper.Metrics, target string, ch chan<- prometheus.Metric) {
	start := time.Now()
	success, faild := h.scraper.Scrape(target, metrics, ch)
	diff := time.Now().Sub(start)
	util.Logger().Debug("Scrape target done", zap.String("url", target),
		zap.Int64("milliseconds latency", diff.Milliseconds()),
		zap.Int32("success request", success),
		zap.Int32("failed request", faild))

	totalSuccessedScrap.Add(uint64(success))
	totalFailedScrap.Add(uint64(faild))

	// only record latency when successed
	if success != 0 && faild == 0 {
		scrapeLatencyDesc.WithLabelValues(target).Observe(float64(diff.Milliseconds()))
	}

	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.CounterValue, float64(totalSuccessedScrap.Load()), target)
	ch <- prometheus.MustNewConstMetric(scrapeFailedDesc, prometheus.CounterValue, float64(totalFailedScrap.Load()), target)

	if faild != 0 {
		util.Logger().Info("Scrape target failed, remove the url", zap.String("url", target))
		h.lock.Lock()
		defer h.lock.Unlock()

		idx := slices.Index(h.TargetUrls, target)
		if idx == -1 {
			util.Logger().Warn("Try to remove url from url list, but not found",
				zap.String("url", target), zap.String("url list", fmt.Sprintf("%v", h.TargetUrls)))
			return
		} else if h.TargetUrls[idx] != target {
			util.Logger().Fatal("url should equal to TargetUrl[idx]",
				zap.String("url", target), zap.Int("index", idx), zap.String("TargetUrl[idx]", h.TargetUrls[idx]))
		} else {
			h.TargetUrls = append(h.TargetUrls[:idx], h.TargetUrls[idx+1:]...)
		}
	}
}
