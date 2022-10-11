package main

import (
	"flag"
	"fmt"
	"github.com/hstreamdb/hstream-exporter/collector"
	"github.com/hstreamdb/hstream-exporter/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
)

var (
	httpServerAddr         = flag.String("addr", "127.0.0.1:8080", "Http server addr")
	listenAddr             = flag.String("listen-addr", ":9200", "Port on which to expose metrics")
	disableExporterMetrics = flag.Bool("disable-exporter-metrics", false, "Exclude metrics about the exporter itself")
	maxScrapeRequest       = flag.Int("max-request", 40, "Maximum number of parallel scrape requests. Use 0 to disable.")
	logLevel               = flag.String("log-level", "info", "Exporter log level")
)

func newHandler(serverUrl string, includeExporterMetrics bool, maxRequests int) (http.Handler, error) {
	exporterMetricsRegistry := prometheus.NewRegistry()
	if includeExporterMetrics {
		exporterMetricsRegistry.MustRegister(
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
		)
	}

	exporter, err := collector.NewHStreamCollector(serverUrl)
	if err != nil {
		return nil, err
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)
	handler := promhttp.HandlerFor(
		prometheus.Gatherers{exporterMetricsRegistry, registry},
		promhttp.HandlerOpts{
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: maxRequests,
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

func main() {
	flag.Parse()
	util.InitLogger(*logLevel)
	handler, err := newHandler(*httpServerAddr, !(*disableExporterMetrics), *maxScrapeRequest)
	if err != nil {
		panic(fmt.Sprintf("create http handler err: %s", err.Error()))
	}
	http.Handle("/metrics", handler)
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
	if err := server.ListenAndServe(); err != nil {
		os.Exit(1)
	}
}
