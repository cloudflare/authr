<?php declare(strict_types=1);

namespace Cloudflare\Authr;

use ArrayIterator;
use IteratorAggregate;

/**
 * RuleList is a container for the list of rules returned by subjects. By making
 * it a separate class, we can leverage PHP7's stronger typing and skip a lot of
 * user-space type checks in the core evaluator, thus getting some speed
 * improvements.
 */
final class RuleList implements IteratorAggregate
{
    /** @var \Cloudflare\Authr\Rule[] */
    private $rules = [];

    public function push(Rule ...$rules): void
    {
        if (count($rules) > 0) {
            array_push($this->rules, ...$rules);
        }
    }

    public function getIterator()
    {
        return new ArrayIterator($this->rules);
    }
}
