<?php declare(strict_types=1);

namespace Cloudflare\Authr\Condition\Operator;

use Cloudflare\Authr\Condition\OperatorInterface;

class In implements OperatorInterface
{
    public function __invoke($leftNeedle, $rightHaystack): bool
    {
        if (!is_array($rightHaystack)) {
            return false;
        }

        return in_array($leftNeedle, $rightHaystack);
    }

    public function jsonSerialize()
    {
        return '$in';
    }
}
