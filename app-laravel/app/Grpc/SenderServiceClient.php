<?php

namespace App\Grpc;

use Grpc\BaseStub;
use App\Grpc\Sender\EmailRequest;
use App\Grpc\Sender\EmailResponse;
use App\Grpc\Sender\StatusRequest;
use App\Grpc\Sender\StatusResponse;

/**
 * Клиент для SenderService
 */
class SenderServiceClient extends BaseStub
{
    /**
     * @param string $hostname хост
     * @param array $opts специфические настройки
     * @param \Grpc\Channel $channel (опционально) переиспользование объекта канала
     */
    public function __construct(string $hostname, array $opts, $channel = null)
    {
        parent::__construct($hostname, $opts, $channel);
    }

    /**
     * Синхронная отправка Email
     * @param EmailRequest $argument входной аргумент
     * @param array $metadata метаданные
     * @param array $options параметры вызова
     * @return \Grpc\UnaryCall
     */
    public function SendEmail(
        EmailRequest $argument,
        array $metadata = [],
        array $options = []
    ): \Grpc\UnaryCall {
        return $this->_simpleRequest(
            '/sender.SenderService/SendEmail',
            $argument,
            [EmailResponse::class, 'decode'],
            $metadata,
            $options
        );
    }

    /**
     * Получение текущего статуса воркера
     * @param StatusRequest $argument входной аргумент
     * @param array $metadata метаданные
     * @param array $options параметры вызова
     * @return \Grpc\UnaryCall
     */
    public function GetWorkerStatus(
        StatusRequest $argument,
        array $metadata = [],
        array $options = []
    ): \Grpc\UnaryCall {
        return $this->_simpleRequest(
            '/sender.SenderService/GetWorkerStatus',
            $argument,
            [StatusResponse::class, 'decode'],
            $metadata,
            $options
        );
    }
}
