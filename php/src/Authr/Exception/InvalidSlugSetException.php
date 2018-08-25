<?php declare(strict_types=1);

namespace Cloudflare\Authr\Exception;

use Cloudflare\Authr\Exception as BaseException;
use RuntimeException;

class InvalidSlugSetException extends RuntimeException implements BaseException
{
}
