<?php

namespace Cloudflare\Authr\Condition\Operator\RegExp;

use Cloudflare\Authr\Condition\OperatorInterface;

class CaseInsensitive implements OperatorInterface
{
    public function __invoke($left, $right)
    {
        return preg_match("/$right/i", $left) === 1;
    }

    public function jsonSerialize()
    {
        return '~*';
    }
}
