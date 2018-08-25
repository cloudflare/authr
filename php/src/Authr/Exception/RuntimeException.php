<?php declare(strict_types=1);

namespace Cloudflare\Authr\Exception;

use Cloudflare\Authr\Exception as BaseException;
use RuntimeException as PHPRuntimeException;

class RuntimeException extends PHPRuntimeException implements BaseException
{
}
