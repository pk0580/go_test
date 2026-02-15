<?php

namespace App\Console\Commands;

use App\Services\GrpcClientService;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\Redis;
use Illuminate\Support\Facades\Http;
use App\Models\Message;

class BenchmarkGrpcCommand extends Command
{
    protected $signature = 'benchmark:grpc {count=100}';
    protected $description = 'Сравнение производительности: Redis Queue vs gRPC';

    public function handle(GrpcClientService $grpcService): int
    {
        $count = (int) $this->argument('count');
        $this->info("Запуск бенчмарка для {$count} сообщений...");

        // 1. Тест gRPC (Синхронно)
        $this->info("\nТестирование gRPC (Синхронно)...");
        $startGrpc = microtime(true);
        for ($i = 0; $i < $count; $i++) {
            $grpcService->sendEmail(
                "test-grpc-{$i}@example.com",
                "Benchmark",
                "Message content {$i}"
            );
        }
        $endGrpc = microtime(true);
        $grpcTime = $endGrpc - $startGrpc;
        $this->info("gRPC общее время: " . number_format($grpcTime, 4) . " сек");
        $this->info("gRPC среднее время: " . number_format(($grpcTime / $count) * 1000, 2) . " мс/запрос");

        // 2. Тест Redis (Только постановка в очередь)
        $this->info("\nТестирование Redis (Постановка в очередь)...");
        $startRedis = microtime(true);
        for ($i = 0; $i < $count; $i++) {
            // Имитируем логику из MessageController
            // Мы не создаем записи в БД, чтобы замерить только накладные расходы очереди
            $payload = json_encode([
                'id' => 999000 + $i,
                'recipient' => "test-redis-{$i}@example.com",
                'content' => "Message content {$i}",
            ]);
            Redis::rpush('messages_queue', $payload);
        }
        $endRedis = microtime(true);
        $redisTime = $endRedis - $startRedis;
        $this->info("Redis общее время: " . number_format($redisTime, 4) . " сек");
        $this->info("Redis среднее время: " . number_format(($redisTime / $count) * 1000, 2) . " мс/запрос");

        $this->info("\n--- Результаты ---");
        if ($redisTime < $grpcTime) {
            $ratio = $grpcTime / $redisTime;
            $this->warn("Redis (очередь) быстрее gRPC в " . number_format($ratio, 1) . " раз");
            $this->line("Примечание: Redis замеряет только время записи в очередь, а не время доставки.");
        } else {
            $ratio = $redisTime / $grpcTime;
            $this->info("gRPC быстрее Redis в " . number_format($ratio, 1) . " раз");
        }

        return 0;
    }
}
