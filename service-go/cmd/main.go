package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Starting Go Message Sender Service...")

	// Здесь будет инициализация очереди и воркеров
	// go worker.Start()

	// Ожидание сигнала для Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Service is running. Press CTRL+C to stop.")
	<-stop

	log.Println("Shutting down service...")
}
