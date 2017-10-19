<?php

namespace Cloudflare\Authr\Condition\Operator\RegExp;

use Cloudflare\Authr\Condition\OperatorInterface;

class InverseCaseSensitive implements OperatorInterface
{
    public function __invoke($left, $right)
    {
        return preg_match("/$right/", $left) === 0;
    }

    public function jsonSerialize()
    {
        return '!~';
    }
}
