<?php

namespace App\Http\Controllers;

use App\Models\Message;
use Illuminate\Http\Response;

class MetricsController extends Controller
{
    /**
     * Экспорт базовых метрик для Prometheus.
     */
    public function index(): Response
    {
        $totalMessages = Message::count();
        $pendingMessages = Message::where('status', 'pending')->count();
        $sentMessages = Message::where('status', 'sent')->count();
        $failedMessages = Message::where('status', 'failed')->count();

        $metrics = [
            '# HELP laravel_messages_total Total messages in database',
            '# TYPE laravel_messages_total counter',
            "laravel_messages_total $totalMessages",

            '# HELP laravel_messages_status_count Messages count by status',
            '# TYPE laravel_messages_status_count gauge',
            "laravel_messages_status_count{status=\"pending\"} $pendingMessages",
            "laravel_messages_status_count{status=\"sent\"} $sentMessages",
            "laravel_messages_status_count{status=\"failed\"} $failedMessages",
        ];

        return response(implode("\n", $metrics) . "\n")
            ->header('Content-Type', 'text/plain; version=0.0.4');
    }
}
