package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/user/go-sender/internal/db"
	"github.com/user/go-sender/internal/grpcserver"
	"github.com/user/go-sender/internal/queue"
	"github.com/user/go-sender/internal/sender"
	"github.com/user/go-sender/internal/worker"
)

func main() {
	// Настройка zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(os.Stdout)

	log.Info().Msg("Starting Go Message Sender Service...")

	// Запуск HTTP сервера для метрик Prometheus
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Info().Msg("Metrics server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal().Err(err).Msg("Failed to start metrics server")
		}
	}()

	// Инициализация Redis
	redisClient, err := queue.NewRedisClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	log.Info().Msg("Connected to Redis successfully")

	// Инициализация MySQL
	mysqlDB, err := db.NewDBConnection()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MySQL")
	}
	defer mysqlDB.Close()
	log.Info().Msg("Connected to MySQL successfully")

	// Инициализация отправителя
	emailSender := sender.NewSMTPSender()

	// Получение количества воркеров из переменной окружения
	numWorkers := 5
	if nwStr := os.Getenv("WORKER_COUNT"); nwStr != "" {
		if nw, err := strconv.Atoi(nwStr); err == nil && nw > 0 {
			numWorkers = nw
		}
	}

	// Контекст для корректного завершения (Graceful Shutdown)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Инициализация воркера
	msgWorker := worker.NewWorker(redisClient, mysqlDB, emailSender)

	// Запуск диспетчера и воркеров
	wg.Add(1)
	go msgWorker.Start(ctx, &wg, numWorkers)

	// Запуск gRPC сервера
	grpcSrv := grpcserver.NewServer(emailSender)
	wg.Add(1)
	go grpcSrv.Start(ctx, &wg, 50051)

	// Ожидание сигнала для остановки
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Int("workers", numWorkers).Msg("Service is running. Press CTRL+C to stop.")
	<-stop

	log.Info().Msg("Shutting down service...")
	cancel()   // Сигнализируем воркерам о необходимости остановиться
	wg.Wait() // Ждем завершения всех воркеров

	log.Info().Msg("Service stopped.")
}
