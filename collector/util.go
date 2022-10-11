package collector

import (
	"encoding/json"
	"fmt"
	"github.com/hstreamdb/hstream-exporter/util"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"io"
	"net/http"
	"path"
	"strconv"
	"sync"
	"sync/atomic"
)

type respTab struct {
	Headers []string          `json:"headers"`
	Value   []json.RawMessage `json:"value"`
}

// DoRequest send http request to hstream-http-server to fetch specific metrics
func DoRequest(url string) ([]map[string]string, error) {
	resp, err := http.Get("http://" + url)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("request %s error", url))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		util.Logger().Error("unexpected response status code", zap.String("url", url), zap.Int("code", resp.StatusCode), zap.String("body", string(body)))
		return nil, errors.New(url + "\n" + resp.Status + "\n" + string(body))
	}

	var tabObj respTab
	if err = json.Unmarshal(body, &tabObj); err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("parse response error for request %s", url))
	}

	var res []map[string]string
	for _, x := range tabObj.Value {
		var xMap map[string]string
		if err = json.Unmarshal(x, &xMap); err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("parse tabObj Value error, request %s", url))
		}
		res = append(res, xMap)
	}
	return res, nil
}

// ScrapeHServerMetrics gather metrics records from hstream-server and export the stats to prometheus
func ScrapeHServerMetrics(ch chan<- prometheus.Metric, metrics []Metrics, url string) error {
	wg := sync.WaitGroup{}
	wg.Add(len(metrics))
	var firstErr error
	errorp := atomic.Pointer[error]{}
	errorp.Store(&firstErr)
	for _, m := range metrics {
		go func(metric Metrics) {
			defer wg.Done()
			resourcePath := path.Join(url, metric.reqPath)
			res, err := DoRequest(resourcePath)
			if err != nil && *errorp.Load() == nil {
				errorp.Store(&err)
				return
			}
			util.Logger().Debug("get response for metrics", zap.String("metrics path", resourcePath), zap.String("res", fmt.Sprintf("%+v", res)))

			switch metric.metricType {
			case Summary:
				err = handleSummary(metric, res, ch)
			default:
				err = handleCounterAndGauge(metric, res, ch)
			}

			if err != nil && *errorp.Load() == nil {
				errorp.Store(&err)
				return
			}
		}(m)
	}
	wg.Wait()
	return firstErr
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

		summary, err := prometheus.NewConstSummary(metric.metric, 0, 0, map[float64]float64{0.5: p50, 0.95: p95, 0.99: p99}, mp["server_host"])
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
			if k != metric.mainKey.String() && k != "server_host" {
				value, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return err
				}
				switch metric.metricType {
				case Gauge:
					ch <- prometheus.MustNewConstMetric(metric.metric, prometheus.GaugeValue, value, metricMp[metric.mainKey.String()], metricMp["server_host"])
				case Counter:
					ch <- prometheus.MustNewConstMetric(metric.metric, prometheus.CounterValue, value, metricMp[metric.mainKey.String()], metricMp["server_host"])
				}
			}
		}
	}
	return nil
}
