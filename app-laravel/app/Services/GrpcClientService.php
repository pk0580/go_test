<?php

namespace App\Services;

use App\Grpc\SenderServiceClient;
use App\Grpc\Sender\EmailRequest;
use App\Grpc\Sender\StatusRequest;
use Illuminate\Support\Facades\Log;

class GrpcClientService
{
    private SenderServiceClient $client;

    public function __construct()
    {
        $address = env('GRPC_GO_SERVICE_ADDR', 'go-sender:50051');

        // В Laravel 12 можно использовать новые возможности типизации и context,
        // но базовое подключение gRPC остается стандартным.
        $this->client = new SenderServiceClient($address, [
            'credentials' => \Grpc\ChannelCredentials::createInsecure(),
        ]);
    }

    /**
     * Отправляет email через gRPC сервис на Go
     */
    public function sendEmail(string $to, string $subject, string $body): array
    {
        $request = new EmailRequest();
        $request->setTo($to);
        $request->setSubject($subject);
        $request->setBody($body);

        [$response, $status] = $this->client->SendEmail($request)->wait();

        if ($status->code !== \Grpc\STATUS_OK) {
            Log::error('gRPC Error: ' . $status->details, ['code' => $status->code]);
            return [
                'success' => false,
                'error' => "gRPC error: {$status->details} (code: {$status->code})",
            ];
        }

        return [
            'success' => $response->getSuccess(),
            'message_id' => $response->getMessageId(),
            'error' => $response->getError(),
        ];
    }

    /**
     * Получает статус воркера из Go сервиса
     */
    public function getWorkerStatus(): array
    {
        $request = new StatusRequest();

        [$response, $status] = $this->client->GetWorkerStatus($request)->wait();

        if ($status->code !== \Grpc\STATUS_OK) {
            return [
                'success' => false,
                'error' => "gRPC error: {$status->details}",
            ];
        }

        return [
            'success' => true,
            'status' => $response->getStatus(),
            'messages_processed' => $response->getMessagesProcessed(),
        ];
    }
}
