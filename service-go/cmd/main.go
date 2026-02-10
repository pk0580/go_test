package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/user/go-sender/internal/queue"
	"github.com/user/go-sender/internal/sender"
	"github.com/user/go-sender/internal/worker"
)

func main() {
	fmt.Println("Starting Go Message Sender Service...")

	// Инициализация Redis
	redisClient, err := queue.NewRedisClient()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")

	// Инициализация отправителя
	emailSender := sender.NewSMTPSender()

	// Контекст для корректного завершения (Graceful Shutdown)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Инициализация воркера
	msgWorker := worker.NewWorker(redisClient, emailSender)

	// Запуск диспетчера и воркеров
	numWorkers := 5
	wg.Add(1)
	go msgWorker.Start(ctx, &wg, numWorkers)

	// Ожидание сигнала для остановки
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Service is running with %d workers. Press CTRL+C to stop.", numWorkers)
	<-stop

	log.Println("Shutting down service...")
	cancel()   // Сигнализируем воркерам о необходимости остановиться
	wg.Wait() // Ждем завершения всех воркеров

	log.Println("Service stopped.")
}
