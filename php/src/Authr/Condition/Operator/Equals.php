<?php

namespace Cloudflare\Authr\Condition\Operator;

use Cloudflare\Authr\Condition\OperatorInterface;

class Equals implements OperatorInterface
{
    public function __invoke($left, $right)
    {
        return $left == $right;
    }

    public function jsonSerialize()
    {
        return '=';
    }
}
