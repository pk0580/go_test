<?php

namespace App\Console\Commands;

use App\Services\GrpcClientService;
use Illuminate\Console\Command;

class TestGrpcCommand extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'test:grpc {email=test@example.com}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'Проверка работы gRPC клиента';

    /**
     * Execute the console command.
     */
    public function handle(GrpcClientService $grpcService): int
    {
        $this->info('Проверка статуса воркера через gRPC...');

        try {
            $response = $grpcService->getWorkerStatus();
            $this->info("Статус: {$response->getStatus()}");
            $this->info("Обработано сообщений: {$response->getMessagesProcessed()}");
            $this->info("Активных воркеров: {$response->getActiveWorkers()}");
        } catch (\Exception $e) {
            $this->error("Ошибка при получении статуса: " . $e->getMessage());
        }

        $email = $this->argument('email');
        $this->info("Отправка тестового письма на {$email}...");

        try {
            $response = $grpcService->sendEmail(
                $email,
                'Тест gRPC Laravel 12',
                'Это сообщение отправлено напрямую через gRPC из Laravel в Go.'
            );

            if ($response->getSuccess()) {
                $this->info("Письмо успешно отправлено! ID: {$response->getMessageId()}");
            } else {
                $this->error("Сервер вернул ошибку: {$response->getError()}");
            }
        } catch (\Exception $e) {
            $this->error("Ошибка при отправке: " . $e->getMessage());
        }

        return 0;
    }
}
