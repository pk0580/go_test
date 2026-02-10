<?php

namespace App\Http\Controllers;

use App\Models\Message;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Redis;

class MessageController extends Controller
{
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
     * Создать сообщение и поместить его в очередь Redis.
     */
    public function store(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'recipient' => 'required|string',
            'content' => 'required|string',
        ]);

        // 1. Сохранение в БД
        $message = Message::create([
            'recipient' => $validated['recipient'],
            'content' => $validated['content'],
            'status' => 'pending',
        ]);

        // 2. Постановка в очередь Redis
        // Используем формат JSON, чтобы Go мог его легко распарсить
        $payload = json_encode([
            'id' => $message->id,
            'recipient' => $message->recipient,
            'content' => $message->content,
        ]);

        Redis::rpush('messages_queue', $payload);

        return response()->json([
            'success' => true,
            'message' => 'Сообщение поставлено в очередь',
            'data' => $message
        ], 201);
    }
}
