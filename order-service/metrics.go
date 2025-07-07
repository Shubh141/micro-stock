package main

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	OrdersPlacedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_placed_total",
			Help: "Total number of orders placed by item",
		},
		[]string{"item"},
	)

	HTTPStatusCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_responses_total",
			Help: "Count of HTTP status codes returned",
		},
		[]string{"path", "method", "code"},
	)

	HTTPErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Count of HTTP errors",
		},
		[]string{"path", "method", "code"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)
)

// InitMetrics registers all Prometheus metrics
func InitMetrics() {
	prometheus.MustRegister(OrdersPlacedCounter, HTTPStatusCounter, HTTPErrorCounter, RequestDuration)
}

// PrometheusHandler exposes the /metrics endpoint
func PrometheusHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// TrackDurationMiddleware observes request durations
func TrackDurationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		RequestDuration.WithLabelValues(c.FullPath(), c.Request.Method).Observe(duration)
	}
}

// TrackStatusMiddleware tracks all HTTP statuses (and errors separately)
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
