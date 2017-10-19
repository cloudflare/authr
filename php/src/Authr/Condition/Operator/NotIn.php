<?php

namespace Cloudflare\Authr\Condition\Operator;

use Cloudflare\Authr\Condition\OperatorInterface;

class NotIn implements OperatorInterface
{
    public function __invoke($leftNeedle, $rightHaystack)
    {
        if (!is_array($rightHaystack)) {
            return false;
        }

        return !in_array($leftNeedle, $rightHaystack);
    }

    public function jsonSerialize()
    {
        return '$nin';
    }
}
