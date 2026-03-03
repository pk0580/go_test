# Руководство по созданию Go-сервиса с gRPC

Данное руководство описывает процесс создания микросервиса на языке Go, который общается с основным приложением (например, на Laravel) через протокол gRPC. Мы разберем это на примере сервиса рассылки сообщений, реализованного в этом проекте.

---

## Что такое gRPC и почему мы его используем?

**gRPC** (Google Remote Procedure Call) — это современный высокопроизводительный фреймворк для вызова удаленных процедур. 
*   **Почему Go?** Go идеально подходит для микросервисов благодаря своей производительности, легковесным потокам (горутинам) и простоте развертывания.
*   **Почему gRPC?** В отличие от традиционного REST (JSON через HTTP/1.1), gRPC использует:
    *   **Protocol Buffers (Protobuf)** — бинарный формат передачи данных, который компактнее и быстрее JSON.
    *   **HTTP/2** — позволяет передавать данные в обе стороны одновременно (стриминг) и мультиплексировать запросы в одном соединении.
    *   **Строгую типизацию** — контракт взаимодействия (API) описывается один раз, и из него генерируется код для любого языка (Go, PHP, Python и т.д.).

---

## Шаг 0: Подготовка окружения

Для разработки на Go в данном проекте вам понадобятся:
1.  **Go** (версия 1.20+).
2.  **Protocol Buffers Compiler (`protoc`)**.
3.  **Плагины для Go**:
    ```bash
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    ```

---

## Шаг 1: Описание контракта (`.proto` файл)

Все начинается с описания структуры данных и сервиса в файле `.proto`. Это "закон", которому будут следовать и сервер (Go), и клиент (Laravel).

Создайте файл `proto/sender.proto`:

```protobuf
syntax = "proto3"; // Версия протокола

package sender; // Имя пакета

// Опции для генерации кода под конкретные языки
option go_package = "github.com/user/go-sender/internal/grpcserver/pb";
option php_namespace = "App\\Grpc\\Sender";

// Описание сервиса и его методов
service SenderService {
  // Метод SendEmail принимает EmailRequest и возвращает EmailResponse
  rpc SendEmail (EmailRequest) returns (EmailResponse);
}

// Структура входящего сообщения
message EmailRequest {
  string to = 1;      // Поле 1: кому
  string subject = 2; // Поле 2: тема
  string body = 3;    // Поле 3: текст
}

// Структура ответа
message EmailResponse {
  bool success = 1;     // Поле 1: успешно ли
  string message_id = 2; // Поле 2: ID сообщения
  string error = 3;      // Поле 3: текст ошибки (если есть)
}
```

---

## Шаг 2: Генерация кода

После того как контракт описан, нужно сгенерировать файлы, с которыми мы будем работать в Go.

Из корня проекта выполните команду:
```bash
protoc --go_out=./service-go --go-grpc_out=./service-go proto/sender.proto
```
Это создаст папку `internal/grpcserver/pb` с файлами `sender.pb.go` и `sender_grpc.pb.go`. **Их не нужно редактировать вручную!**

---

## Шаг 3: Инициализация Go-модуля

Перейдите в папку сервиса и создайте файл зависимостей `go.mod`:

```bash
cd service-go
go mod init github.com/user/go-sender
go mod tidy
```

---

## Шаг 4: Реализация бизнес-логики (Sender)

Создадим интерфейс и реализацию для отправки почты. Это отделяет логику отправки от логики gRPC.

Файл `internal/sender/sender.go`:

```go
package sender

import "fmt"

// Интерфейс для отправителя
type EmailSender interface {
	Send(to string, body string) error
}

// Реализация (например, через SMTP)
type SMTPSender struct{}

func NewSMTPSender() *SMTPSender {
	return &SMTPSender{}
}

func (s *SMTPSender) Send(to string, body string) error {
	// Здесь логика реальной отправки через SMTP
	fmt.Printf("Отправка письма на %s: %s\n", to, body)
	return nil
}
```

---

## Шаг 5: Создание gRPC сервера

Теперь реализуем сам сервер, который будет принимать вызовы.

Файл `internal/grpcserver/server.go`:

```go
package grpcserver

import (
	"context"
	"fmt"
	"net"
	"github.com/user/go-sender/internal/grpcserver/pb"
	"github.com/user/go-sender/internal/sender"
	"google.golang.org/grpc"
)

// Наша структура сервера, в которую мы встраиваем сгенерированный код
type Server struct {
	pb.UnimplementedSenderServiceServer
	emailSender sender.EmailSender
}

// Конструктор
func NewServer(s sender.EmailSender) *Server {
	return &Server{emailSender: s}
}

// Реализация метода SendEmail из proto-файла
func (s *Server) SendEmail(ctx context.Context, req *pb.EmailRequest) (*pb.EmailResponse, error) {
	// 1. Вызываем бизнес-логику
	err := s.emailSender.Send(req.GetTo(), req.GetBody())
	
	// 2. Обрабатываем ошибку
	if err != nil {
		return &pb.EmailResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 3. Возвращаем успешный ответ
	return &pb.EmailResponse{
		Success: true,
		MessageId: "generated-id-123",
	}, nil
}

// Метод для запуска сервера
func (s *Server) Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSenderServiceServer(grpcServer, s)

	fmt.Printf("gRPC сервер запущен на порту %d\n", port)
	return grpcServer.Serve(lis)
}
```

---

## Шаг 6: Точка входа (`main.go`)

Файл `cmd/main.go` собирает все части воедино и запускает приложение.

```go
package main

import (
	"log"
	"github.com/user/go-sender/internal/grpcserver"
	"github.com/user/go-sender/internal/sender"
)

func main() {
	// 1. Инициализируем зависимости
	emailSender := sender.NewSMTPSender()

	// 2. Создаем сервер
	srv := grpcserver.NewServer(emailSender)

	// 3. Запускаем
	if err := srv.Start(50051); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
```

---

## Шаг 7: Запуск в Docker

Для работы в микросервисной архитектуре добавьте сервис в `docker-compose.yml`:

```yaml
  service-go:
    build:
      context: ./service-go
    ports:
      - "50051:50051"
    environment:
      - SMTP_HOST=mailpit
      - SMTP_PORT=1025
```

И создайте `service-go/Dockerfile`:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

---

## Как это работает вместе? (Для новичков)

1.  **Laravel** хочет отправить письмо. Он создает объект `EmailRequest`, упаковывает его в бинарный формат и отправляет по адресу `service-go:50051`.
2.  **Go-сервис** слушает этот порт. Когда приходит запрос, gRPC-библиотека автоматически распаковывает его и вызывает функцию `SendEmail` в нашем коде.
3.  **Go-код** выполняет отправку (быстро и параллельно) и возвращает `EmailResponse`.
4.  **Laravel** получает ответ почти мгновенно и может продолжить свою работу.

Это гораздо эффективнее, чем если бы Laravel сам подключался к SMTP для каждого письма в высоконагруженной системе.
