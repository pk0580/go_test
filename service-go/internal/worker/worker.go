package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/user/go-sender/internal/queue"
	"github.com/user/go-sender/internal/sender"
)

// Message структура сообщения из очереди
type Message struct {
	ID        int64  `json:"id"`
	Recipient string `json:"recipient"`
	Content   string `json:"content"`
}

// Worker структура воркера
type Worker struct {
	redisClient *queue.RedisClient
	db          *sql.DB
	sender      sender.EmailSender
	taskChan    chan Message
	processed   uint64
}

// NewWorker создает новый экземпляр воркера
func NewWorker(rc *queue.RedisClient, db *sql.DB, s sender.EmailSender) *Worker {
	return &Worker{
		redisClient: rc,
		db:          db,
		sender:      s,
		taskChan:    make(chan Message, 100),
	}
}

// Start запускает процесс обработки очереди
func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup, numWorkers int) {
	defer wg.Done()
	log.Printf("Starting %d worker routines...", numWorkers)

	// Запуск пула воркеров, которые слушают канал задач
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go w.workerRoutine(ctx, wg, i)
	}

	// Запуск мониторинга производительности
	wg.Add(1)
	go w.monitor(ctx, wg)

	// Основной цикл: чтение из Redis и отправка в канал задач
	for {
		select {
		case <-ctx.Done():
			log.Println("Dispatcher stopping...")
			close(w.taskChan)
			return
		default:
			// BLPop блокирует выполнение до появления сообщения в очереди
			result, err := w.redisClient.Client.BLPop(ctx, 5 * time.Second, "laravel-database-messages_queue").Result()
			if err != nil {
				if err.Error() != "redis: nil" && ctx.Err() == nil {
					log.Printf("Dispatcher error: %v", err)
				}
				continue
			}

            // result[0] - имя очереди, result[1] - данные
			var msg Message
			if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
				log.Printf("Dispatcher: error parsing message: %v", err)
				continue
			}

			// Отправляем задачу в канал
			select {
                case w.taskChan <- msg:
                    // Задача успешно отправлена в канал
                case <-ctx.Done():
                    return
			}
		}
	}
}

// workerRoutine отдельная горутина-воркер
func (w *Worker) workerRoutine(ctx context.Context, wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	log.Printf("Worker routine %d started", workerID)

	for msg := range w.taskChan {
		w.processMessage(workerID, msg)
	}
	log.Printf("Worker routine %d stopped", workerID)
}

// monitor логирует метрики производительности
func (w *Worker) monitor(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastProcessed uint64

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			currentProcessed := atomic.LoadUint64(&w.processed)
			diff := currentProcessed - lastProcessed
			log.Printf("METRIC: Processed %d messages in last 5s (Total: %d, Avg: %.2f msg/s)",
				diff, currentProcessed, float64(diff)/5.0)
			lastProcessed = currentProcessed
		}
	}
}

// processMessage обрабатывает сообщение
func (w *Worker) processMessage(workerID int, msg Message) {
	log.Printf("Worker %d: Processing message ID %d to %s", workerID, msg.ID, msg.Recipient)

	// Отправка через отправителя (SMTP или Mock)
	err := w.sender.Send(msg.Recipient, msg.Content)
	
	status := "sent"
	if err != nil {
		log.Printf("Worker %d: Failed to send message ID %d: %v", workerID, msg.ID, err)
		status = "failed"
	}

	// Обновление статуса в БД MySQL
	_, dbErr := w.db.Exec("UPDATE messages SET status = ?, updated_at = ? WHERE id = ?", status, time.Now(), msg.ID)
	if dbErr != nil {
		log.Printf("Worker %d: Failed to update status in DB for ID %d: %v", workerID, msg.ID, dbErr)
	} else {
		log.Printf("Worker %d: Updated status to '%s' for message ID %d", workerID, status, msg.ID)
	}

	if status == "sent" {
		atomic.AddUint64(&w.processed, 1)
		fmt.Printf("Message PROCESSED: ID=%d, To=%s\n", msg.ID, msg.Recipient)
	}
}
