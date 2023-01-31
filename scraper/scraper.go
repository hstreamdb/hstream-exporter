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
	getStatsCmd         = "server stats %s %s -i %s"
	summaryStatInterval = "5s"
)

type Scrape interface {
	Scrape(target string, metrics []Metrics, ch chan<- prometheus.Metric) (success int32, failed int32)
}

var summaryMetricSet = map[StatType]struct{}{
	StreamAppendLatency: {},
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
			batchedMetrics[m.Type.ToHstreamStatType()] = m.Metric
		} else {
			summaryMetrics[m.Type] = m.Metric
		}
	}

	successScrapeRequest := atomic.Int32{}
	failedScrapeRequest := atomic.Int32{}
	wg := sync.WaitGroup{}
	wg.Add(2)
	s.batchScrape(&wg, target, batchedMetrics, &successScrapeRequest, &failedScrapeRequest, ch)
	s.scrapeSummary(&wg, target, summaryMetrics, &successScrapeRequest, &failedScrapeRequest, ch)
	wg.Wait()

	return successScrapeRequest.Load(), failedScrapeRequest.Load()
}

func (s *Scraper) batchScrape(wg *sync.WaitGroup, target string, metrics map[hstream.StatType]*prometheus.Desc,
	success *atomic.Int32, failed *atomic.Int32, ch chan<- prometheus.Metric) {
	go func() {
		defer wg.Done()
		mc := make([]hstream.StatType, 0, len(metrics))
		for k := range metrics {
			mc = append(mc, k)
		}

		statsResult, err := s.client.GetStatsRequest(target, mc)
		if err != nil {
			util.Logger().Error("send batch stats request to HStream server error",
				zap.String("target", target), zap.Error(err))
			failed.Add(1)
			return
		}

		for _, st := range statsResult {
			switch st.(type) {
			case hstream.StatValue:
				stat := st.(hstream.StatValue)
				for k, v := range stat.Value {
					ch <- prometheus.MustNewConstMetric(metrics[stat.Type], prometheus.CounterValue, float64(v), k, target)
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
					zap.String("url", addr), zap.Error(err))
				return
			}

			table, err := parseResponse(resp)
			if err != nil {
				failed.Add(1)
				util.Logger().Error("decode admin request error", zap.String("cmd", cmd),
					zap.String("url", addr), zap.Error(err))
				return
			}
			table["server_host"] = strings.Split(addr, ":")[0]
			if err = handleSummary(metrics[metric], table, ch); err != nil {
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
		return fmt.Sprintf(getStatsCmd, "server_histogram", "append_request_latency", summaryStatInterval)
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

func handleSummary(metric *prometheus.Desc, mp map[string]string, ch chan<- prometheus.Metric) error {
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
	p95 := parse(mp["p95"])
	p99 := parse(mp["p99"])
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstSummary(metric, 0, 0,
		map[float64]float64{0.5: p50, 0.95: p95, 0.99: p99}, mp["server_host"])

	return nil
}
