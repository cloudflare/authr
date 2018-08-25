<?php declare(strict_types=1);

namespace Cloudflare\Authr\Condition\Operator;

use Cloudflare\Authr\Condition\OperatorInterface;

class ArrayIntersect implements OperatorInterface
{
    public function __invoke($left, $right): bool
    {
        if (!is_array($left) || !is_array($right)) {
            return false;
        }
        $isect = array_intersect($left, $right);
        return count($isect) > 0;
    }

    public function jsonSerialize()
    {
        return '&';
    }
}
