<?php

namespace Cloudflare\Authr\Condition\Operator;

use Cloudflare\Authr\Condition\OperatorInterface;

class ArrayDifference implements OperatorInterface
{
    public function __invoke($left, $right)
    {
        if (!is_array($left) || !is_array($right)) {
            return false;
        }
        $isect = array_intersect($left, $right);
        return count($isect) === 0;
    }

    public function jsonSerialize()
    {
        return '-';
    }
}
