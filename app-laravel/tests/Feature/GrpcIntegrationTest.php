<?php

namespace Tests\Feature;

use App\Services\GrpcClientService;
use Tests\TestCase;
use Exception;

class GrpcIntegrationTest extends TestCase
{
    /**
     * Проверяет интеграцию с Go-сервисом через gRPC.
     * Ожидается, что сервис go-sender запущен и доступен.
     */
    public function test_grpc_get_worker_status_returns_valid_response(): void
    {
        try {
            /** @var GrpcClientService $service */
            $service = $this->app->make(GrpcClientService::class);
            $response = $service->getWorkerStatus();

            $this->assertEquals('running', $response->getStatus());
            $this->assertIsInt($response->getMessagesProcessed());
        } catch (Exception $e) {
            $this->fail("gRPC call failed: " . $e->getMessage());
        }
    }

    /**
     * Проверяет отправку email через gRPC.
     */
    public function test_grpc_send_email_returns_success(): void
    {
        try {
            /** @var GrpcClientService $service */
            $service = $this->app->make(GrpcClientService::class);
            $response = $service->sendEmail(
                'test@example.com',
                'Integration Test',
                'This is a test message from Laravel'
            );

            $this->assertTrue($response->getSuccess());
            $this->assertStringContainsString('gRPC-', $response->getMessageId());
        } catch (Exception $e) {
            $this->fail("gRPC call failed: " . $e->getMessage());
        }
    }
}
