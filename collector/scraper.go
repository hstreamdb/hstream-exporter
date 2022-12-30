package collector

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
	getStatsInterval = "5s"
)

type respTable struct {
	url     string
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

// ScrapeHServerMetrics gather metrics records from hstream-server and export the stats to prometheus
// TODO: use goroutine pool to reuse goroutine ???
func ScrapeHServerMetrics(ch chan<- prometheus.Metric, client *hstream.HStreamClient,
	metrics []Metrics, serverUrls []string) (successScrape uint32, failedScrape uint32) {
	wg := sync.WaitGroup{}
	wg.Add(len(metrics))

	for _, m := range metrics {
		go func(metric Metrics) {
			defer wg.Done()
			switch metric.hstreamMetric.(type) {
			case StreamCounterMetrics, SubscriptionCounterMetrics:
				if metric.metricType == Counter {
					if !scrapeV2(metric, client, serverUrls, ch) {
						atomic.AddUint32(&failedScrape, 1)
						return
					}

				} else {
					if !scrapeV1(metric, client, serverUrls, ch) {
						atomic.AddUint32(&failedScrape, 1)
						return
					}
				}
			default:
				if !scrapeV1(metric, client, serverUrls, ch) {
					atomic.AddUint32(&failedScrape, 1)
					return
				}
			}
			atomic.AddUint32(&successScrape, 1)
		}(m)
	}
	wg.Wait()
	return
}

func scrapeV2(metric Metrics, client *hstream.HStreamClient, serverUrls []string, ch chan<- prometheus.Metric) bool {
	wg := sync.WaitGroup{}
	wg.Add(len(serverUrls))
	atLeastOneSucc := atomic.Bool{}

	for _, url := range serverUrls {
		go func(addr string) {
			defer wg.Done()
			statsType := metric.hstreamMetric.GetCounterStatsType()
			var (
				resp map[string]int64
				err  error
			)
			switch statsType.(type) {
			case hstream.StreamStatsType:
				streamStats := statsType.(hstream.StreamStatsType)
				resp, err = client.GetStreamStatsRequest(addr, streamStats)
				if err != nil {
					util.Logger().Error("send stream stats request to HStream server error",
						zap.String("stat type", streamStats.String()),
						zap.String("url", addr), zap.Error(err))
					return
				}
			case hstream.SubscriptionStatsType:
				subStats := statsType.(hstream.SubscriptionStatsType)
				resp, err = client.GetSubscriptionStatsRequest(addr, subStats)
				if err != nil {
					util.Logger().Error("send subscription stats request to HStream server error",
						zap.String("stat type", subStats.String()),
						zap.String("url", addr), zap.Error(err))
					return
				}
			}

			for k, v := range resp {
				ch <- prometheus.MustNewConstMetric(metric.metric, prometheus.CounterValue, float64(v), k, addr)
			}
			atLeastOneSucc.CompareAndSwap(false, true)
		}(url)
	}
	wg.Wait()
	return atLeastOneSucc.Load()
}

func scrapeV1(metric Metrics, client *hstream.HStreamClient, serverUrls []string, ch chan<- prometheus.Metric) bool {
	res, err := scrape(client, serverUrls, metric)
	if err != nil {
		util.Logger().Error("scrape metrics error", zap.String("metric", metric.hstreamMetric.GetMetricName()))
		return false
	}

	util.Logger().Debug("get response for metrics",
		zap.String("server urls", strings.Join(serverUrls, " ")),
		zap.String("metric", metric.hstreamMetric.GetMetricName()),
		zap.String("res", fmt.Sprintf("%+v", res)))

	switch metric.metricType {
	case Summary:
		if err = handleSummary(metric, res, ch); err != nil {
			util.Logger().Error("create prometheus metric error", zap.String("metric", metric.hstreamMetric.GetMetricName()),
				zap.Error(err))
			return false
		}
	default:
		if err = handleCounterAndGauge(metric, res, ch); err != nil {
			util.Logger().Error("create prometheus metric error", zap.String("metric", metric.hstreamMetric.GetMetricName()),
				zap.Error(err))
			return false
		}
	}
	return true
}

// serverStatsInfo indicates the stats scraped from the server
// e.g.
// | server_host | stream_name | appends_1min |
// |   server1   |     s1      |    1829      |
type serverStatsInfo = map[string]string

// scrape send admin request to all hserver, merge all returned records.
func scrape(client *hstream.HStreamClient, serverUrls []string, metric Metrics) ([]serverStatsInfo, error) {
	wg := sync.WaitGroup{}
	wg.Add(len(serverUrls))
	mutex := sync.Mutex{}

	responseTables := make([]*respTable, 0, len(serverUrls))
	for _, url := range serverUrls {
		go func(addr string) {
			defer wg.Done()
			cmd := metric.hstreamMetric.StatCmd(getStatsInterval)
			resp, err := client.AdminRequestToServer(addr, cmd)
			if err != nil {
				util.Logger().Error("send admin request to HStream server error",
					zap.String("cmd", cmd),
					zap.String("url", addr), zap.Error(err))
				return
			}

			table, err := parseResponse(resp)
			if err != nil {
				util.Logger().Error("decode admin request error", zap.String("cmd", cmd),
					zap.String("url", addr), zap.Error(err))
				return
			}
			table.url = strings.Split(addr, ":")[0]
			mutex.Lock()
			responseTables = append(responseTables, table)
			mutex.Unlock()
		}(url)
	}
	wg.Wait()

	return mergeResponseTable(responseTables)
}

// mergeResponseTable merge stats scraped from different server together
func mergeResponseTable(records []*respTable) ([]serverStatsInfo, error) {
	if len(records) == 0 {
		return nil, errors.New("get empty response")
	}
	header := records[0].Headers
	res := make([]serverStatsInfo, 0, len(records))
	for _, table := range records {
		for _, rows := range table.Rows {
			mp := make(serverStatsInfo)
			mp["server_host"] = table.url
			for i := 0; i < len(rows); i++ {
				mp[header[i]] = rows[i]
			}
			res = append(res, mp)
		}
	}
	return res, nil
}

// parseResponse convert admin response to respTable
func parseResponse(resp string) (*respTable, error) {
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

	return &table, nil
}

func handleSummary(metric Metrics, res []map[string]string, ch chan<- prometheus.Metric) error {
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

	for _, mp := range res {
		p50 := parse(mp["p50"])
		p95 := parse(mp["p95"])
		p99 := parse(mp["p99"])
		if err != nil {
			return err
		}

		summary, err := prometheus.NewConstSummary(metric.metric, 0, 0,
			map[float64]float64{0.5: p50, 0.95: p95, 0.99: p99}, mp["server_host"])
		if err != nil {
			return errors.WithMessage(err, "create summary error")
		}
		ch <- summary
	}
	return nil
}

func handleCounterAndGauge(metric Metrics, res []map[string]string, ch chan<- prometheus.Metric) error {
	for _, metricMp := range res {
		for k, v := range metricMp {
			if k != metric.hstreamMetric.GetMetricKey() && k != "server_host" {
				value, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return err
				}
				ch <- prometheus.MustNewConstMetric(metric.metric, prometheus.GaugeValue, value,
					metricMp[metric.hstreamMetric.GetMetricKey()], metricMp["server_host"])
			}
		}
	}
	return nil
}
