package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	DBQueryCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "db_query_total",
			Help: "Total number of DB queries made",
		},
	)

	UserRegistrationCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "user_registration_total",
			Help: "Total number of user registrations",
		},
	)

	RequestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of response time for handler",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
)

func RegisterCustomMetrics() {
	prometheus.MustRegister(DBQueryCounter)
	prometheus.MustRegister(UserRegistrationCounter)
	prometheus.MustRegister(RequestLatency)
}
