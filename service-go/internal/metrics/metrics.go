package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	MessagesProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "go_sender_messages_processed_total",
		Help: "Общее количество обработанных сообщений по статусам",
	}, []string{"status"})

	MessagesProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "go_sender_message_processing_duration_seconds",
		Help:    "Время, затраченное на обработку сообщения",
		Buckets: prometheus.DefBuckets,
	})

	ActiveWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "go_sender_active_workers",
		Help: "Количество активных горутин-воркеров",
	})
)
