package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/user/go-sender/internal/queue"
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
}

// NewWorker создает новый экземпляр воркера
func NewWorker(rc *queue.RedisClient) *Worker {
	return &Worker{redisClient: rc}
}

// Start запускает процесс обработки очереди
func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	log.Printf("Worker %d started", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d stopping...", workerID)
			return
		default:
			// BLPop блокирует выполнение до появления сообщения в очереди (таймаут 5 секунд)
			result, err := w.redisClient.Client.BLPop(ctx, 5 * time.Second, "messages_queue").Result()
			if err != nil {
				if err.Error() != "redis: nil" {
					log.Printf("Worker %d error: %v", workerID, err)
				}
				continue
			}

			// result[0] - имя очереди, result[1] - данные
			payload := result[1]
			var msg Message
			if err := json.Unmarshal([]byte(payload), &msg); err != nil {
				log.Printf("Worker %d: error parsing message: %v", workerID, err)
				continue
			}

			w.processMessage(workerID, msg)
		}
	}
}

// processMessage имитирует отправку сообщения
func (w *Worker) processMessage(workerID int, msg Message) {
	log.Printf("Worker %d: Processing message ID %d to %s", workerID, msg.ID, msg.Recipient)
	
	// Имитация задержки отправки (например, запрос к SMTP или API)
	time.Sleep(100 * time.Millisecond)
	
	fmt.Printf("Message SENT: ID=%d, To=%s, Content=%s\n", msg.ID, msg.Recipient, msg.Content)
	
	// В будущем здесь можно добавить логирование результата обратно в БД
}
