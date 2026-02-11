package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	MessagesProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "go_sender_messages_processed_total",
		Help: "Total number of messages processed by status",
	}, []string{"status"})

	MessagesProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "go_sender_message_processing_duration_seconds",
		Help:    "Time taken to process a message",
		Buckets: prometheus.DefBuckets,
	})

	ActiveWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "go_sender_active_workers",
		Help: "Number of active worker routines",
	})
)
