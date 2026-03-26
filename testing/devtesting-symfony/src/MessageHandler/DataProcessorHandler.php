<?php

namespace App\MessageHandler;

use App\Message\DataProcessorMessage;
use OpenTelemetry\API\Trace\Span;
use OpenTelemetry\API\Trace\StatusCode;
use Symfony\Component\Messenger\Attribute\AsMessageHandler;
use Traceway\OpenTelemetryBundle\TracingInterface;

#[AsMessageHandler]
class DataProcessorHandler
{
    public function __construct(
        private readonly TracingInterface $tracing,
    ) {}

    public function __invoke(DataProcessorMessage $message): void
    {
        $this->tracing->trace('loading data', function () {
            usleep(random_int(100, 2000) * 1000);
        });

        $rootSpan = Span::getCurrent();

        for ($i = 0; $i < $message->batchSize; $i++) {
            $rootSpan->addEvent("data loaded successfully $i");
        }

        $rootSpan->setStatus(StatusCode::STATUS_ERROR, 'what an error');
        $rootSpan->recordException(new \RuntimeException('what an error'));
    }
}
