<?php

namespace App\Services;

use App\Grpc\Sender\EmailRequest;
use App\Grpc\Sender\EmailResponse;
use App\Grpc\Sender\StatusRequest;
use App\Grpc\Sender\StatusResponse;
use App\Grpc\SenderServiceClient;
use Grpc\ChannelCredentials;
use Illuminate\Support\Facades\Config;
use Illuminate\Support\Facades\Log;
use RuntimeException;
use const Grpc\STATUS_OK;

/**
 * Сервис для взаимодействия с Go-сервисом через gRPC.
 * Реализован с использованием возможностей PHP 8.4 и Laravel 12.
 */
class GrpcClientService
{
    private SenderServiceClient $client;

    public function __construct()
    {
        $address = Config::get('services.grpc.go_sender.address');

        // Создаем клиент gRPC с использованием небезопасных учетных данных для локальной разработки.
        // Проверяем наличие класса, так как расширение может быть не загружено
        if (!class_exists(ChannelCredentials::class)) {
            Log::warning('gRPC extension is not loaded. GrpcClientService might not work.');
        }

        $this->client = new SenderServiceClient($address, [
            'credentials' => class_exists(ChannelCredentials::class)
                ? ChannelCredentials::createInsecure()
                : null,
        ]);
    }

    /**
     * Синхронная отправка email сообщения через Go-сервис.
     *
     * @param string $to Получатель
     * @param string $subject Тема письма
     * @param string $body Текст письма
     * @return EmailResponse
     * @throws RuntimeException
     */
    public function sendEmail(string $to, string $subject, string $body): EmailResponse
    {
        $request = new EmailRequest();
        $request->setTo($to);
        $request->setSubject($subject);
        $request->setBody($body);

        /** @var array{0: ?EmailResponse, 1: \stdClass} $call */
        $call = $this->client->SendEmail($request)->wait();
        [$response, $status] = $call;

        if ($status->code !== STATUS_OK) {
            Log::error('gRPC SendEmail failed', [
                'code' => $status->code,
                'details' => $status->details,
            ]);
            throw new RuntimeException("gRPC error: {$status->details}", $status->code);
        }

        return $response;
    }

    /**
     * Получение текущего статуса воркера из Go-сервиса.
     *
     * @return StatusResponse
     * @throws RuntimeException
     */
    public function getWorkerStatus(): StatusResponse
    {
        $request = new StatusRequest();

        /** @var array{0: ?StatusResponse, 1: \stdClass} $call */
        $call = $this->client->GetWorkerStatus($request)->wait();
        [$response, $status] = $call;

        if ($status->code !== STATUS_OK) {
            Log::error('gRPC GetWorkerStatus failed', [
                'code' => $status->code,
                'details' => $status->details,
            ]);
            throw new RuntimeException("gRPC error: {$status->details}", $status->code);
        }

        return $response;
    }
}
