package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ActiveWSConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "playingfield",
		Subsystem: "websocket",
		Name:      "active_connections_total",
		Help:      "Current number of active WebSocket connections",
	})

	ActiveChatConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "playingfield",
		Subsystem: "websocket",
		Name:      "active_chat_connections_total",
		Help:      "Current number of WebSocket connections with a project chat context (ProjectID ≠ 0)",
	})

	WSMessagesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "playingfield",
		Subsystem: "websocket",
		Name:      "messages_total",
		Help:      "Total number of WebSocket messages sent and received",
	}, []string{"room_type", "direction"})
)

func init() {
	prometheus.MustRegister(ActiveWSConnections)
	prometheus.MustRegister(WSMessagesTotal)
	prometheus.MustRegister(ActiveChatConnections)
}

// Handler returns the standard Prometheus metrics HTTP handler
func Handler() http.Handler {
	return promhttp.Handler()
}
