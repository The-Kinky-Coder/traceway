<?php

namespace App\MessageHandler;

use App\Message\PaymentProcessorMessage;
use OpenTelemetry\API\Trace\Span;
use OpenTelemetry\API\Trace\SpanKind;
use OpenTelemetry\API\Trace\StatusCode;
use Symfony\Component\Messenger\Attribute\AsMessageHandler;
use Traceway\OpenTelemetryBundle\TracingInterface;

#[AsMessageHandler]
class PaymentProcessorHandler
{
    public function __construct(
        private readonly TracingInterface $tracing,
    ) {}

    public function __invoke(PaymentProcessorMessage $message): void
    {
        $this->tracing->trace('validate.payment', function () use ($message) {
            usleep(random_int(20, 100) * 1000);
        }, ['payment.amount' => $message->amount, 'payment.currency' => $message->currency]);

        try {
            $this->tracing->trace('charge.gateway', function () {
                usleep(random_int(100, 500) * 1000);
                throw new \RuntimeException('gateway timeout: payment provider unreachable');
            }, kind: SpanKind::KIND_CLIENT);
        } catch (\RuntimeException $e) {
            $rootSpan = Span::getCurrent();
            $rootSpan->setStatus(StatusCode::STATUS_ERROR, $e->getMessage());
            $rootSpan->recordException($e);
        }
    }
}
