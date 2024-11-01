package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var MethodCount = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "method_count",
		Help: "Count of failed/successful RPC call",
	},
	[]string{"method", "status", "podIP"},
)

var MethodDuration = promauto.NewSummaryVec(
	prometheus.SummaryOpts{
		Name: "method_duration",
		Help: "for latency of each call, in 99, 95, 50 quantiles",
		Objectives: map[float64]float64{0.50: 0.05, 0.95: 0.005, 0.99: 0.001},
	},
	[]string{"method", "podIP"},
)

var ActiveSubscribers = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "active_subscribers",
		Help: "Total active subscriptions",
	},
)

func Init()  {
	// prometheus.MustRegister(MethodCount)
	// prometheus.MustRegister(MethodDuration)
	// prometheus.MustRegister(ActiveSubscribers)
	//brew services start grafana
}