# Мониторинг системы

В проекте реализована система мониторинга на базе Prometheus и Grafana, а также структурное логирование.

## 1. Логирование (Logs)

Все сервисы выводят логи в формате JSON в стандартный поток вывода (stdout/stderr).

### Как смотреть логи:
*   **Go-сервис (Sender):**
    ```bash
    docker-compose logs -f go-sender
    ```
*   **Laravel (Web):**
    ```bash
    docker-compose logs -f app
    ```
*   **Все сервисы:**
    ```bash
    docker-compose logs -f
    ```

Логи Go-сервиса используют `zerolog` и содержат такие поля, как `level`, `worker_id`, `message_id`, `status` и др.
Логи Laravel настроены на канал `stderr` с использованием `JsonFormatter`.

## 2. Метрики (Metrics)

Метрики экспортируются в формате Prometheus.

### Эндпоинты:
*   **Go Service:** `http://localhost:8080/metrics` (внутри сети docker: `go-sender:8080/metrics`)
*   **Laravel Service:** `http://localhost:8080/api/metrics` (внутри сети docker: `web:80/api/metrics`)

### Основные метрики Go:
*   `go_sender_messages_processed_total`: Счетчик обработанных сообщений (с разделением по статусам `sent`/`failed`).
*   `go_sender_message_processing_duration_seconds`: Гистограмма времени обработки одного сообщения.
*   `go_sender_active_workers`: Количество активных воркеров в данный момент.

### Основные метрики Laravel:
*   `laravel_messages_total`: Общее количество сообщений в базе.
*   `laravel_messages_status_count`: Количество сообщений в разрезе статусов (`pending`, `sent`, `failed`).

## 3. Визуализация

### Prometheus
Доступен по адресу: [http://localhost:9090](http://localhost:9090)
Здесь можно выполнять прямые PromQL запросы и проверять состояние таргетов (Status -> Targets).

### Grafana
Доступна по адресу: [http://localhost:3000](http://localhost:3000)
*   **Логин/Пароль:** `admin` / `admin` (по умолчанию)
*   **Источник данных (Data Source):** Prometheus уже преднастроен автоматически через provisioning.
*   **Дашборды:** Вы можете создать свой дашборд, используя метрики, описанные выше.

### Mailpit
Используется для перехвата и просмотра исходящих email-сообщений в процессе разработки.
*   **Web UI:** [http://localhost:8025](http://localhost:8025)
*   **SMTP сервер:** `localhost:1025` (внутри docker-сети: `mailpit:1025`)

## 4. Проверка работоспособности

Чтобы увидеть изменения на графиках:
1. Отправьте тестовое сообщение через API Laravel:
   ```bash
   curl -X POST http://localhost:8080/api/messages \
     -H "Content-Type: application/json" \
     -d '{"recipient": "test@example.com", "content": "Hello Monitoring!"}'
   ```
2. Перейдите в Grafana или Prometheus и проверьте метрику `go_sender_messages_processed_total`.
