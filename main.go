package main

import (
	"flag"
	"fmt"
	"github.com/hstreamdb/hstream-exporter/collector"
	"github.com/hstreamdb/hstream-exporter/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"
)

var (
	hServerAddr            = flag.String("addr", "hstream://127.0.0.1:6570", "HStream server addr")
	listenAddr             = flag.String("listen-addr", ":9200", "Port on which to expose metrics")
	disableExporterMetrics = flag.Bool("disable-exporter-metrics", false, "Exclude metrics about the exporter itself")
	maxScrapeRequest       = flag.Int("max-request", 0, "Maximum number of parallel scrape requests. Use 0 to disable.")
	// TODO: the prometheus scrap timeout must greater than hstream rpc request timeout(default 5s), add a validation
	timeout               = flag.Int("timeout", 10, "Time out in seconds for each prometheus scrap request.")
	logLevel              = flag.String("log-level", "info", "Exporter log level")
	getServerInfoDuration = flag.Int("get-server-info-duration", 30, "Get server info in second duration.")
)

func newHandler(serverUrl string, includeExporterMetrics bool, maxRequests int, timeout int) (http.Handler, error) {
	exporterMetricsRegistry := prometheus.NewRegistry()
	if includeExporterMetrics {
		exporterMetricsRegistry.MustRegister(
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
		)
	}

	registry := prometheus.NewRegistry()
	exporter, err := collector.NewHStreamCollector(serverUrl, *getServerInfoDuration, registry)
	if err != nil {
		return nil, err
	}
	util.Logger().Info("create connection with hstream server", zap.String("url", serverUrl))
	registry.MustRegister(exporter)
	handler := promhttp.HandlerFor(
		prometheus.Gatherers{exporterMetricsRegistry, registry},
		promhttp.HandlerOpts{
			ErrorLog:            util.NewPromErrLogger(),
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: maxRequests,
			Timeout:             time.Duration(timeout) * time.Second,
			Registry:            exporterMetricsRegistry,
		},
	)

	if includeExporterMetrics {
		// Note that we have to use h.exporterMetricsRegistry here to
		// use the same promhttp metrics for all expositions.
		handler = promhttp.InstrumentMetricHandler(
			exporterMetricsRegistry, handler,
		)
	}
	return handler, nil
}

// updateLogLevel handle update log level request, e.g.: curl -X POST localhost:9200/log_level?debug
func updateLogLevel(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "update log level only accept a post request", http.StatusBadRequest)
		return
	}

	level := r.URL.RawQuery

	util.Logger().Info("receive update log-level request", zap.String("level", level))
	if err := util.UpdateLogLevel(level); err != nil {
		util.Logger().Error("update logger error", zap.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	util.Logger().Info("update log level success", zap.String("level", level))
	return
}

func main() {
	flag.Parse()
	if err := util.InitLogger(*logLevel); err != nil {
		util.Logger().Error("init logger error", zap.Error(err))
		os.Exit(1)
	}

	handler, err := newHandler(*hServerAddr, !(*disableExporterMetrics), *maxScrapeRequest, *timeout)
	if err != nil {
		panic(fmt.Sprintf("create http handler err: %s", err.Error()))
	}
	http.Handle("/metrics", handler)
	http.HandleFunc("/log_level", updateLogLevel)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>HStream Exporter</title></head>
			<body>
			<h1>HStream Exporter</h1>
			<p><a href="` + "/metrics" + `">Metrics</a></p>
			</body>
			</html>`))
	})

	server := &http.Server{Addr: *listenAddr}
	util.Logger().Info("HStream Exporter start", zap.String("address", *listenAddr))
	if err = server.ListenAndServe(); err != nil {
		os.Exit(1)
	}
}
