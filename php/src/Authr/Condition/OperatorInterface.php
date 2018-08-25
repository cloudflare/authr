<?php declare(strict_types=1);

namespace Cloudflare\Authr\Condition;

use JsonSerializable;

interface OperatorInterface extends JsonSerializable
{
    public function __invoke($left, $right): bool;
}
