<?php declare(strict_types=1);

namespace Cloudflare\Authr\Condition\Operator;

use Cloudflare\Authr\Condition\OperatorInterface;

class Like implements OperatorInterface
{
    public function __invoke($left, $right): bool
    {
        $pleft = '^';
        $pright = '$';
        if ($right[0] === '*') {
            $pleft = '';
        }
        if ($right[strlen($right) - 1] === '*') {
            $pright = '';
        }

        return (bool) preg_match(str_replace('*', '', "/$pleft$right$pright/i"), $left);
    }

    public function jsonSerialize()
    {
        return '~';
    }
}
