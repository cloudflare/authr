<?php declare(strict_types=1);

namespace Cloudflare\Authr\Condition\Operator;

use Cloudflare\Authr\Condition\OperatorInterface;

class Equals implements OperatorInterface
{
    public function __invoke($left, $right): bool
    {
        return $left == $right;
    }

    public function jsonSerialize()
    {
        return '=';
    }
}
