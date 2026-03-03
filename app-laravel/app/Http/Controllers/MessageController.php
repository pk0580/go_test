<?php

namespace App\Http\Controllers;

use App\Models\Message;
use App\Services\GrpcClientService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Config;
use Illuminate\Support\Facades\Redis;

class MessageController extends Controller
{
    public function __construct(
        private readonly GrpcClientService $grpcService
    ) {}

    /**
     * Показать список сообщений.
     */
    public function index(): JsonResponse
    {
        $messages = Message::latest()->get();

        return response()->json([
            'success' => true,
            'data' => $messages
        ]);
    }

    /**
     * Создать сообщение и отправить его (через Redis или gRPC).
     */
    public function store(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'recipient' => 'required|string',
            'content' => 'required|string',
        ]);

        $mode = Config::get('services.sender.mode', 'redis');

        // 1. Сохранение в БД
        $message = Message::create([
            'recipient' => $validated['recipient'],
            'content' => $validated['content'],
            'status' => 'pending',
        ]);

        return $mode === 'grpc'
            ? $this->sendViaGrpc($message)
            : $this->sendViaRedis($message);
    }

    /**
     * Отправка через Redis (асинхронно).
     */
    private function sendViaRedis(Message $message): JsonResponse
    {
        $payload = json_encode([
            'id' => $message->id,
            'recipient' => $message->recipient,
            'content' => $message->content,
        ]);

        Redis::rpush('messages_queue', $payload);

        return response()->json([
            'success' => true,
            'message' => 'Сообщение поставлено в очередь (Redis)',
            'data' => $message
        ], 201);
    }

    /**
     * Отправка через gRPC (синхронно).
     */
    private function sendViaGrpc(Message $message): JsonResponse
    {
        try {
            $response = $this->grpcService->sendEmail(
                $message->recipient,
                'Уведомление',
                $message->content
            );

            if ($response->getSuccess()) {
                $message->update(['status' => 'sent']);
                return response()->json([
                    'success' => true,
                    'message' => 'Сообщение успешно отправлено через gRPC',
                    'data' => $message
                ], 201);
            }

            $message->update(['status' => 'failed']);
            return response()->json([
                'success' => false,
                'message' => 'Ошибка gRPC: ' . $response->getError(),
                'data' => $message
            ], 500);

        } catch (\Exception $e) {
            $message->update(['status' => 'failed']);
            return response()->json([
                'success' => false,
                'message' => 'Исключение gRPC: ' . $e->getMessage(),
                'data' => $message
            ], 500);
        }
    }
}
