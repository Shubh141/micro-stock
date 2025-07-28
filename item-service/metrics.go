package main

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Counter for total items created
	ItemCreatedCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "items_created_total",
			Help: "Total number of items created via SetStock",
		},
	)

	// Counter for total items deleted
	ItemDeletedCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "items_deleted_total",
			Help: "Total number of items deleted via DeleteStock",
		},
	)

	// Gauge for current stock level per item
	ItemStockGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "item_stock_level",
			Help: "Current stock level per item",
		},
		[]string{"item"},
	)

	// Histogram for request duration by path + method
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	// Counter for HTTP status codes
	HTTPStatusCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_responses_total",
			Help: "Count of HTTP responses by status code",
		},
		[]string{"path", "method", "code"},
	)

	// Counter for error status codes
	HTTPErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Count of HTTP errors (status >= 400)",
		},
		[]string{"path", "method", "code"},
	)
)

// Registers all Prometheus metrics
func InitMetrics() {
	prometheus.MustRegister(
		ItemCreatedCounter,
		ItemDeletedCounter,
		ItemStockGauge,
		RequestDuration,
		HTTPStatusCounter,
		HTTPErrorCounter,
	)
}

// Exposes the /metrics endpoint
func PrometheusHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// Observes request durations
func TrackDurationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		RequestDuration.WithLabelValues(c.FullPath(), c.Request.Method).Observe(duration)
	}
}

// Tracks all HTTP statuses (and errors separately)
func TrackStatusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		status := c.Writer.Status()
		statusStr := strconv.Itoa(status)
		path := c.FullPath()
		method := c.Request.Method

		HTTPStatusCounter.WithLabelValues(path, method, statusStr).Inc()
		if status >= 400 {
			HTTPErrorCounter.WithLabelValues(path, method, statusStr).Inc()
		}
	}
}
