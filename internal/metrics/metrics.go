package metrics

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HTTP metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)

	// Business metrics
	itemsCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "items_created_total",
			Help: "Total number of items created",
		},
	)

	itemsRetrievedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "items_retrieved_total",
			Help: "Total number of items retrieved",
		},
	)

	itemsUpdatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "items_updated_total",
			Help: "Total number of items updated",
		},
	)

	itemsDeletedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "items_deleted_total",
			Help: "Total number of items deleted",
		},
	)

	grpcClientCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_client_calls_total",
			Help: "Total number of gRPC client calls",
		},
		[]string{"service", "method", "status_code"},
	)

	grpcClientCallDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_client_call_duration_seconds",
			Help:    "gRPC client call duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)
)

// init registers all metrics
func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestsInFlight)
	prometheus.MustRegister(itemsCreatedTotal)
	prometheus.MustRegister(itemsRetrievedTotal)
	prometheus.MustRegister(itemsUpdatedTotal)
	prometheus.MustRegister(itemsDeletedTotal)
	prometheus.MustRegister(grpcClientCallsTotal)
	prometheus.MustRegister(grpcClientCallDuration)
}

// HTTPMiddleware provides Prometheus metrics for HTTP requests
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Increment in-flight requests
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)
		durationSeconds := float64(duration) / float64(time.Second)

		// Extract endpoint from route
		endpoint := "unknown"
		if route := mux.CurrentRoute(r); route != nil {
			if pathTemplate, err := route.GetPathTemplate(); err == nil {
				endpoint = pathTemplate
			}
		}

		// Record metrics
		httpRequestsTotal.WithLabelValues(r.Method, endpoint, http.StatusText(wrapped.statusCode)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, endpoint).Observe(durationSeconds)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Business metrics functions
func RecordItemCreated() {
	itemsCreatedTotal.Inc()
}

func RecordItemRetrieved() {
	itemsRetrievedTotal.Inc()
}

func RecordItemUpdated() {
	itemsUpdatedTotal.Inc()
}

func RecordItemDeleted() {
	itemsDeletedTotal.Inc()
}

// gRPC client metrics functions
func RecordGRPCClientCall(service, method, statusCode string) {
	grpcClientCallsTotal.WithLabelValues(service, method, statusCode).Inc()
}

func RecordGRPCClientCallDuration(service, method string, duration float64) {
	grpcClientCallDuration.WithLabelValues(service, method).Observe(duration)
}

// MetricsHandler returns the Prometheus metrics handler
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
