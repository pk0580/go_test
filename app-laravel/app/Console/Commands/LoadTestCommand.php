<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;

class LoadTestCommand extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'app:load-test {count=100}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'Генерация тестовых сообщений в очередь для нагрузочного тестирования';

    /**
     * Execute the console command.
     */
    public function handle(): void
    {
        $count = (int) $this->argument('count');
        $this->info("Генерация {$count} сообщений...");

        $bar = $this->output->createProgressBar($count);
        $bar->start();

        for ($i = 0; $i < $count; $i++) {
            $message = \App\Models\Message::create([
                'recipient' => "user{$i}@example.com",
                'content' => "Тестовое сообщение #{$i} для нагрузочного тестирования",
                'status' => 'pending',
            ]);

            $payload = json_encode([
                'id' => $message->id,
                'recipient' => $message->recipient,
                'content' => $message->content,
            ]);

            \Illuminate\Support\Facades\Redis::rpush('messages_queue', $payload);
            $bar->advance();
        }

        $bar->finish();
        $this->newLine();
        $this->info('Готово!');
    }
}
