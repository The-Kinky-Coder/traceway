<?php

use App\Kernel;
use Symfony\Component\Dotenv\Dotenv;

require_once dirname(__DIR__) . '/vendor/autoload.php';

(new Dotenv())->loadEnv(dirname(__DIR__) . '/.env');
\OpenTelemetry\SDK\SdkAutoloader::autoload();

$kernel = new Kernel('dev', true);
$request = \Symfony\Component\HttpFoundation\Request::createFromGlobals();
$response = $kernel->handle($request);
$response->send();
$kernel->terminate($request, $response);
