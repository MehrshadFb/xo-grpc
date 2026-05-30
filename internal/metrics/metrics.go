package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	GRPCRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "xo_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "code", "type"},
	)

	GRPCRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "xo_grpc_request_duration_seconds",
			Help: "Duration of gRPC requests in seconds",
		},
		[]string{"method", "code", "type"},
	)

	ActiveWatchStreams = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "xo_active_watch_streams",
			Help: "Number of active WatchGame streams",
		},
	)
)

func Register() {
	prometheus.MustRegister(
		GRPCRequestsTotal,
		GRPCRequestDurationSeconds,
		ActiveWatchStreams,
	)
}
