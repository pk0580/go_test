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
        $status = $grpcService->getWorkerStatus();

        if ($status['success']) {
            $this->info("Статус: {$status['status']}");
            $this->info("Обработано сообщений: {$status['messages_processed']}");
        } else {
            $this->error("Ошибка: {$status['error']}");
        }

        $email = $this->argument('email');
        $this->info("Отправка тестового письма на {$email}...");

        $result = $grpcService->sendEmail(
            $email,
            'Тест gRPC Laravel 12',
            'Это сообщение отправлено напрямую через gRPC из Laravel в Go.'
        );

        if ($result['success']) {
            $this->info("Письмо успешно отправлено! ID: {$result['message_id']}");
        } else {
            $this->error("Ошибка при отправке: {$result['error']}");
        }

        return 0;
    }
}
