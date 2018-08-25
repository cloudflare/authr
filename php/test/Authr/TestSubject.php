<?php

namespace Cloudflare\Test\Authr;

use Cloudflare\Authr\SubjectInterface;
use Cloudflare\Authr\RuleList;

class TestSubject implements SubjectInterface
{
    /** @var \Cloudflare\Authr\Rule[] */
    protected $rules = null;

    /**
     * @param \Cloudflare\Authr\Rule[] $rules
     */
    public function setRules(array $rules): void
    {
        $this->rules = $rules;
    }

    public function getRules(): RuleList
    {
        $rules = new RuleList();
        $rules->push(...$this->rules);
        return $rules;
    }
}
