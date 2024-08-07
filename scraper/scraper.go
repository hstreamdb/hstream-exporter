package scraper

import (
	"encoding/json"
	"fmt"
	"github.com/hstreamdb/hstream-exporter/util"
	"github.com/hstreamdb/hstreamdb-go/hstream"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	getStatsCmd         = "server stats %s %s -i %s -p 0.5 -p 0.75 -p 0.90 -p 0.99"
	summaryStatInterval = "1min"
)

type Scrape interface {
	Scrape(target string, metrics []Metrics, ch chan<- prometheus.Metric) (success int32, failed int32)
}

var summaryMetricSet = map[StatType]struct{}{
	StreamAppendLatency:      {},
	StreamReadLatency:        {},
	CacheStoreAppendLatency:  {},
	CacheStoreReadLatency:    {},
	CheckStoreClusterLatency: {},
	CheckMetaClusterLatency:  {},
}

type Scraper struct {
	client *hstream.HStreamClient
}

func NewScraper(client *hstream.HStreamClient) Scrape {
	return &Scraper{client: client}
}

func (s *Scraper) Scrape(target string, metrics []Metrics, ch chan<- prometheus.Metric) (int32, int32) {
	batchedMetrics := make(map[hstream.StatType]*prometheus.Desc, len(metrics))
	summaryMetrics := make(map[StatType]*prometheus.Desc)
	for _, m := range metrics {
		if _, ok := summaryMetricSet[m.Type]; !ok {
			batchedMetrics[m.Type.ToHStreamStatType()] = m.Metric
		} else {
			summaryMetrics[m.Type] = m.Metric
		}
	}

	successScrapeRequest := atomic.Int32{}
	failedScrapeRequest := atomic.Int32{}
	wg := sync.WaitGroup{}
	wg.Add(2)
	// only fetch connector alive state once
	connectorAliveStatOnce := atomic.Bool{}
	connectorAliveStatOnce.Store(false)
	s.batchScrape(&wg, target, batchedMetrics, &successScrapeRequest, &failedScrapeRequest, &connectorAliveStatOnce, ch)
	s.scrapeSummary(&wg, target, summaryMetrics, &successScrapeRequest, &failedScrapeRequest, ch)
	wg.Wait()

	return successScrapeRequest.Load(), failedScrapeRequest.Load()
}

func (s *Scraper) batchScrape(wg *sync.WaitGroup, target string, metrics map[hstream.StatType]*prometheus.Desc,
	success *atomic.Int32, failed *atomic.Int32, connectorAliveStatOnce *atomic.Bool, ch chan<- prometheus.Metric) {
	go func() {
		defer wg.Done()
		mc := make([]hstream.StatType, 0, len(metrics))
		for k := range metrics {
			if k == hstream.ConnectorIsAlive && !connectorAliveStatOnce.CompareAndSwap(false, true) {
				continue
			}
			mc = append(mc, k)
		}

		statsResult, err := s.client.GetStatsRequest(target, mc)
		if err != nil {
			util.Logger().Error("send batch stats request to HStream server error",
				zap.String("target", target), zap.String("error", err.Error()))
			failed.Add(1)
			return
		}

		addr := strings.Split(target, ":")[0]
		for _, st := range statsResult {
			switch st.(type) {
			case hstream.StatValue:
				stat := st.(hstream.StatValue)
				for k, v := range stat.Value {
					switch stat.Type {
					case hstream.SubCheckListSize, hstream.ConnectorIsAlive,
						hstream.CacheStoreAppendTotal, hstream.CacheStoreAppendFailed,
						hstream.CacheStoreDeliveredTotal, hstream.CacheStoreDeliveredFailed:
						ch <- prometheus.MustNewConstMetric(metrics[stat.Type], prometheus.GaugeValue, float64(v), k, addr)
					default:
						ch <- prometheus.MustNewConstMetric(metrics[stat.Type], prometheus.CounterValue, float64(v), k, addr)
					}
					util.Logger().Debug(fmt.Sprintf("scrape counter [%s]", stat.Type),
						zap.String("host", addr),
						zap.String("metrics", fmt.Sprintf("%s: %v", k, v)),
					)
				}
			case hstream.StatError:
				stErr := st.(hstream.StatError)
				util.Logger().Error("scrape stats error",
					zap.String("stat name", stErr.Type.String()),
					zap.String("target", target),
					zap.String("message", stErr.Message))
			}
		}
		success.Add(1)
	}()
}

func (s *Scraper) scrapeSummary(wg *sync.WaitGroup, target string, metrics map[StatType]*prometheus.Desc,
	success *atomic.Int32, failed *atomic.Int32, ch chan<- prometheus.Metric) {
	defer wg.Done()

	wg1 := sync.WaitGroup{}
	wg1.Add(len(metrics))

	for m := range metrics {
		go func(addr string, metric StatType) {
			defer wg1.Done()
			cmd := getSummaryStatsCmd(metric)
			resp, err := s.client.AdminRequestToServer(addr, cmd)
			if err != nil {
				failed.Add(1)
				util.Logger().Error("send admin request to HStream server error",
					zap.String("cmd", cmd),
					zap.String("url", addr), zap.String("error", err.Error()))
				return
			}

			table, err := parseResponse(resp)
			if err != nil {
				failed.Add(1)
				util.Logger().Error("decode admin request error", zap.String("cmd", cmd),
					zap.String("url", addr), zap.String("error", err.Error()))
				return
			}
			table["server_host"] = addr
			if err = handleSummary(metrics[metric], metric, table, ch); err != nil {
				failed.Add(1)
				util.Logger().Error("handle summary stats error", zap.String("stat", metric.String()),
					zap.String("target", addr), zap.Error(err))
				return
			}
			success.Add(1)
		}(target, m)
	}
	wg1.Wait()
}

func getSummaryStatsCmd(stat StatType) string {
	switch stat {
	case StreamAppendLatency:
		return fmt.Sprintf(getStatsCmd, "server_histogram", "append_latency", summaryStatInterval)
	case StreamReadLatency:
		return fmt.Sprintf(getStatsCmd, "server_histogram", "read_latency", summaryStatInterval)
	case CacheStoreAppendLatency:
		return fmt.Sprintf(getStatsCmd, "server_histogram", "append_cache_store_latency", summaryStatInterval)
	case CacheStoreReadLatency:
		return fmt.Sprintf(getStatsCmd, "server_histogram", "read_cache_store_latency", summaryStatInterval)
	case CheckStoreClusterLatency:
		return fmt.Sprintf(getStatsCmd, "server_histogram", "check_store_cluster_healthy_latency", summaryStatInterval)
	case CheckMetaClusterLatency:
		return fmt.Sprintf(getStatsCmd, "server_histogram", "check_meta_cluster_healthy_latency", summaryStatInterval)
	}
	util.Logger().Error("unsupported summary stat", zap.String("stat", stat.String()))
	return ""
}

type respTable struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

// serverStatsInfo indicates the stats scraped from the server
// e.g.
// | server_host | stream_name | appends_1min | <- headers
// |   server1   |     s1      |    1829      | <- rows[0]
type serverStatsInfo = map[string]string

func parseResponse(resp string) (serverStatsInfo, error) {
	var jsonObj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(resp), &jsonObj); err != nil {
		return nil, err
	}

	var table respTable
	if content, ok := jsonObj["content"]; ok {
		if err := json.Unmarshal(content, &table); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("no content fields in admin response")
	}

	header := table.Headers
	mp := make(serverStatsInfo)
	for _, rows := range table.Rows {
		for i := 0; i < len(rows); i++ {
			mp[header[i]] = rows[i]
		}
	}

	return mp, nil
}

func handleSummary(metric *prometheus.Desc, metricType StatType, mp map[string]string, ch chan<- prometheus.Metric) error {
	var err error
	parse := func(input string) float64 {
		if err != nil {
			return 0
		}
		value, e := strconv.ParseFloat(input, 64)
		if e != nil {
			err = e
			return 0
		}
		return value
	}

	p50 := parse(mp["p50"])
	p90 := parse(mp["p90"])
	p99 := parse(mp["p99"])
	if err != nil {
		return err
	}

	util.Logger().Debug(fmt.Sprintf("scrape summary [%s]", metricType),
		zap.String("host", mp["server_host"]),
		zap.String("metrics", fmt.Sprintf("%+v", mp)),
	)

	ch <- prometheus.MustNewConstSummary(metric, 0, 0,
		map[float64]float64{0.5: p50, 0.90: p90, 0.99: p99}, mp["server_host"])

	return nil
}
