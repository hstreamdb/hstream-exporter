package collector

import (
	"github.com/hstreamdb/hstream-exporter/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	cacheStoreSubsystem = "cacheStore"
)

type CacheStoreMetrics struct {
	Metrics []scraper.Metrics
}

func NewCacheStoreMetrics() *CacheStoreMetrics {
	appendInBytes := scraper.Metrics{
		Type: scraper.CacheStoreAppendInBytes,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, cacheStoreSubsystem, scraper.CacheStoreAppendInBytes.String()),
			"Successfully written bytes to the cache store.",
			[]string{"column_family", "server_host"}, nil,
		),
	}
	appendInRecords := scraper.Metrics{
		Type: scraper.CacheStoreAppendInReccords,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, cacheStoreSubsystem, scraper.CacheStoreAppendInReccords.String()),
			"Successfully written records to the cache store.",
			[]string{"column_family", "server_host"}, nil,
		),
	}
	appendTotal := scraper.Metrics{
		Type: scraper.CacheStoreAppendTotal,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, cacheStoreSubsystem, scraper.CacheStoreAppendTotal.String()),
			"Number of success append requests of a cache store.",
			[]string{"column_family", "server_host"}, nil,
		),
	}
	appendFailed := scraper.Metrics{
		Type: scraper.CacheStoreAppendFailed,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, cacheStoreSubsystem, scraper.CacheStoreAppendFailed.String()),
			"Number of failed append requests of a cache store.",
			[]string{"column_family", "server_host"}, nil,
		),
	}
	appendRequestLatency := scraper.Metrics{
		Type: scraper.CacheStoreAppendLatency,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, cacheStoreSubsystem, scraper.CacheStoreAppendLatency.String()),
			"Append cache store latency.",
			[]string{"server_host"}, nil,
		),
	}
	readInBytes := scraper.Metrics{
		Type: scraper.CacheStoreReadInBytes,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, cacheStoreSubsystem, scraper.CacheStoreReadInBytes.String()),
			"Successfully read bytes from the cache store.",
			[]string{"column_family", "server_host"}, nil,
		),
	}
	readInBatches := scraper.Metrics{
		Type: scraper.CacheStoreReadInRecords,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, cacheStoreSubsystem, scraper.CacheStoreReadInRecords.String()),
			"Successfully read records from the cache store.",
			[]string{"column_family", "server_host"}, nil,
		),
	}
	readCacheStoreLatency := scraper.Metrics{
		Type: scraper.CacheStoreReadLatency,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, cacheStoreSubsystem, scraper.CacheStoreReadLatency.String()),
			"Read cache store latency.",
			[]string{"server_host"}, nil,
		),
	}
	deliveredInRecords := scraper.Metrics{
		Type: scraper.CacheStoreDeliveredInRecords,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, cacheStoreSubsystem, scraper.CacheStoreDeliveredInRecords.String()),
			"Successfully delivered records from the cache store.",
			[]string{"column_family", "server_host"}, nil,
		),
	}
	deliveredTotal := scraper.Metrics{
		Type: scraper.CacheStoreDeliveredTotal,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, cacheStoreSubsystem, scraper.CacheStoreDeliveredTotal.String()),
			"Total delivered records from the cache store.",
			[]string{"column_family", "server_host"}, nil,
		),
	}
	deliveredFiled := scraper.Metrics{
		Type: scraper.CacheStoreDeliveredFailed,
		Metric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, cacheStoreSubsystem, scraper.CacheStoreDeliveredFailed.String()),
			"Failed delivered records from the cache store.",
			[]string{"column_family", "server_host"}, nil,
		),
	}
	return &CacheStoreMetrics{
		Metrics: []scraper.Metrics{appendInBytes, appendInRecords, appendTotal, appendFailed, appendRequestLatency,
			readInBytes, readInBatches, readCacheStoreLatency, deliveredInRecords, deliveredTotal, deliveredFiled},
	}
}
