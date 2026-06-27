package metrics

import (
	"sync"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests served.",
		},
		[]string{"method", "route", "status"},
	)

	HTTPRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "HTTP request duration in seconds.",
		},
		[]string{"method", "route", "status"},
	)

	ShortURLCacheHitsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "shorturl_cache_hits_total",
		Help: "Total number of short URL cache hits.",
	})
	ShortURLCacheMissesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "shorturl_cache_misses_total",
		Help: "Total number of short URL cache misses.",
	})
	ShortURLNegativeCacheHitsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "shorturl_negative_cache_hits_total",
		Help: "Total number of short URL negative cache hits.",
	})
	ShortURLDBFallbackTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "shorturl_db_fallback_total",
		Help: "Total number of short URL repository fallbacks.",
	})
	ShortURLCacheErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "shorturl_cache_errors_total",
		Help: "Total number of short URL cache errors.",
	})

	AnalyticsEventsProducedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "analytics_events_produced_total",
		Help: "Total number of analytics events produced.",
	})
	AnalyticsProducerErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "analytics_producer_errors_total",
		Help: "Total number of analytics producer errors.",
	})

	RateLimitedRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limited_requests_total",
			Help: "Total number of requests rejected by the distributed rate limiter.",
		},
		[]string{"type"},
	)
	RateLimitErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_errors_total",
			Help: "Total number of distributed rate limiter errors.",
		},
		[]string{"type"},
	)

	AnalyticsWorkerEventsInsertedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "analytics_worker_events_inserted_total",
		Help: "Total number of analytics events inserted by the worker.",
	})
	AnalyticsWorkerBatchesInsertedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "analytics_worker_batches_inserted_total",
		Help: "Total number of analytics batches inserted by the worker.",
	})
	AnalyticsWorkerInsertErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "analytics_worker_insert_errors_total",
		Help: "Total number of analytics worker insert errors.",
	})
	AnalyticsWorkerAckErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "analytics_worker_ack_errors_total",
		Help: "Total number of analytics worker acknowledgement errors.",
	})
	AnalyticsWorkerDecodeErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "analytics_worker_decode_errors_total",
		Help: "Total number of analytics worker decode errors.",
	})

	registerOnce sync.Once
)

// Register adds the application's collectors to Prometheus's default registry.
func Register() {
	registerOnce.Do(func() {
		prometheus.MustRegister(
			HTTPRequestsTotal,
			HTTPRequestDurationSeconds,
			ShortURLCacheHitsTotal,
			ShortURLCacheMissesTotal,
			ShortURLNegativeCacheHitsTotal,
			ShortURLDBFallbackTotal,
			ShortURLCacheErrorsTotal,
			AnalyticsEventsProducedTotal,
			AnalyticsProducerErrorsTotal,
			RateLimitedRequestsTotal,
			RateLimitErrorsTotal,
			AnalyticsWorkerEventsInsertedTotal,
			AnalyticsWorkerBatchesInsertedTotal,
			AnalyticsWorkerInsertErrorsTotal,
			AnalyticsWorkerAckErrorsTotal,
			AnalyticsWorkerDecodeErrorsTotal,
		)
	})
}
