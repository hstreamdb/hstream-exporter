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
	getStatsCmd = "server stats %s %s -i %s"
)

type respTab struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

type resultTab struct {
	Headers []string            `json:"headers"`
	Value   []map[string]string `json:"value"`
}

func doRequestForServer(serverUrl, category, metrics, interval string) (*respTab, error) {
	client, err := hstream.NewHStreamClient(serverUrl)
	if err != nil {
		return nil, err
	}
	getStatsCmd := fmt.Sprintf(getStatsCmd, category, metrics, interval)
	resp, err := client.AdminRequest(getStatsCmd)
	if err != nil {
		util.Logger().Error("send admin request to HStream server error: ",
			zap.String("cmd", getStatsCmd),
			zap.String("url", serverUrl), zap.String("body", err.Error()))
		return nil, errors.New(serverUrl + "\n" + err.Error())
	}
	var jsonObj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(resp), &jsonObj); err != nil {
		return nil, err
	}

	var tab respTab
	if content, ok := jsonObj["content"]; ok {
		if err := json.Unmarshal(content, &tab); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New(fmt.Sprintf("no content fields in admin response: %+v", jsonObj))
	}

	return &tab, nil
}

func getFromCluster(serverUrls []string, category, metrics, interval string) map[string]*respTab {
	respTabs := make(map[string]*respTab, 0)
	for _, serverUrl := range serverUrls {
		if resp, err := doRequestForServer(serverUrl, category, metrics, interval); err != nil {
			continue
		} else {
			respTabs[serverUrl] = resp
		}
	}
	return respTabs
}

func DoRequest(serverUrls []string, category, metrics, interval string) ([]map[string]string, error) {
	switch category {
	case "server_histogram":
		resp, err := aggLatencyHist(serverUrls, category, metrics, interval)
		if err != nil {
			return nil, err
		}
		return resp.Value, nil
	default:
		resp, err := aggAppendSum(serverUrls, category, metrics, interval)
		if err != nil {
			return nil, err
		}
		return resp.Value, nil
	}
}

func aggLatencyHist(serverUrls []string, category, metrics, interval string) (*resultTab, error) {
	records := getFromCluster(serverUrls, category, metrics, interval)
	if len(records) == 0 {
		return nil, errors.New("no response from cluster, got stats of length 0")
	}

	var headers []string
	setHeader := false
	values := make([]map[string]string, 0, len(records))
	for addr, table := range records {
		if !setHeader {
			headers = make([]string, len(table.Headers)+1)
			headers[0] = "server_host"
			copy(headers[1:], table.Headers)
			setHeader = true
			util.Logger().Debug(fmt.Sprintf("headers = %s", headers))
		}

		mp := make(map[string]string, len(table.Headers)+1)
		for _, rows := range table.Rows {
			for idx := 0; idx < len(rows); idx++ {
				mp[headers[idx+1]] = rows[idx]
			}
			host := strings.Split(addr, ":")[0]
			mp[headers[0]] = host
		}
		values = append(values, mp)
		util.Logger().Debug(fmt.Sprintf("%+v", mp))
	}

	return &resultTab{
		Headers: headers,
		Value:   values,
	}, nil
}

func aggAppendSum(serverUrls []string, category, metrics, interval string) (*resultTab, error) {
	records := getFromCluster(serverUrls, category, metrics, interval)

	if len(records) == 0 {
		return nil, errors.New("no response from cluster, got stats of length 0")
	}

	dataVec := make([]*respTab, 0, len(records))
	servers := make([]string, 0, len(records))
	for addr, v := range records {
		dataVec = append(dataVec, v)
		servers = append(servers, addr)
	}

	return sum(dataVec, servers)
}

func sum(records []*respTab, servers []string) (*resultTab, error) {
	headers := records[0].Headers
	// statistics: {resource: {metrics1: value1, metrics2: value2}}
	statistics := map[string]map[string]int64{}
	dataSize := len(records[0].Headers) - 1
	resourceOriginMp := map[string]string{}
	for i, table := range records {
		serverUrl := strings.Split(servers[i], ":")[0]
		// e.g. table.Rows[0]
		// |stream_name|appends_1min|appends_5min|appends_10min|  <- headers
		// |    s1     |      0     |    1829    |    1829     |  <- row0
		// |    s2     |      0     |    7270    |    7270     |  <- row1
		for _, rows := range table.Rows {
			target := rows[0]
			if _, ok := statistics[target]; !ok {
				statistics[target] = make(map[string]int64, dataSize)
				resourceOriginMp[target] = serverUrl
			}

			for idx := 1; idx <= dataSize; idx++ {
				value, err := strconv.ParseInt(rows[idx], 10, 64)
				if err != nil {
					return nil, err
				}
				statistics[target][headers[idx]] += value
			}
		}
	}

	// construct result
	rows := []map[string]string{}
	for resource, metricsMp := range statistics {
		res := map[string]string{}
		res[headers[0]] = resource
		for metrics, v := range metricsMp {
			res[metrics] = strconv.FormatInt(v, 10)
		}
		res["server_host"] = resourceOriginMp[resource]
		rows = append(rows, res)
	}

	headers = append([]string{"server_host"}, headers...)
	allStats := resultTab{
		Headers: headers,
		Value:   rows,
	}
	util.Logger().Debug(fmt.Sprintf("res = %+v", statistics))
	return &allStats, nil
}

// ScrapeHServerMetrics gather metrics records from hstream-server and export the stats to prometheus
func ScrapeHServerMetrics(ch chan<- prometheus.Metric, metrics []Metrics, serverUrls []string) error {
	wg := sync.WaitGroup{}
	wg.Add(len(metrics))
	var firstErr error
	errorPtr := atomic.Pointer[error]{}
	errorPtr.Store(&firstErr)
	for _, m := range metrics {
		m := m
		m.reqArgs = append(m.reqArgs, "5s")
		go func(metric Metrics) {
			defer wg.Done()
			metricsArgs := strings.Join(metric.reqArgs, " ")
			res, err := DoRequest(serverUrls, m.reqArgs[0], m.reqArgs[1], m.reqArgs[2])
			if err != nil && *errorPtr.Load() == nil {
				errorPtr.Store(&err)
				return
			}
			util.Logger().Debug("get response for metrics", zap.String("server urls", strings.Join(serverUrls, " ")), zap.String("metrics args", metricsArgs), zap.String("res", fmt.Sprintf("%+v", res)))

			switch metric.metricType {
			case Summary:
				err = handleSummary(metric, res, ch)
			default:
				err = handleCounterAndGauge(metric, res, ch)
			}

			if err != nil && *errorPtr.Load() == nil {
				errorPtr.Store(&err)
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
