<?php declare(strict_types=1);

namespace Cloudflare\Authr\Condition\Operator\RegExp;

use Cloudflare\Authr\Condition\OperatorInterface;

class InverseCaseInsensitive implements OperatorInterface
{
    public function __invoke($left, $right): bool
    {
        return preg_match("/$right/i", $left) === 0;
    }

    public function jsonSerialize()
    {
        return '!~*';
    }
}
